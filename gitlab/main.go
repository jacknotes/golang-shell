package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"flag"

	"github.com/PuerkitoBio/goquery"
	"github.com/xanzy/go-gitlab"
)

func main() {
	// === 1. 定义命令行参数 ===
	var (
		gitlabURL         = flag.String("gitlab-url", getEnv("GITLAB_URL", "http://gitlab.hs.com"), "GitLab base URL")
		namespaceName     = flag.String("namespace", getEnv("NAMESPACE", "k8s-deploy"), "Namespace (group) name")
		token             = flag.String("token", getEnv("GITLAB_TOKEN", ""), "GitLab private token")
		projectName       = flag.String("project", getEnv("PROJECT_NAME", ""), "Project name")
		username          = flag.String("username", getEnv("GITLAB_USERNAME", ""), "gitlab login username")
		password          = flag.String("password", getEnv("GITLAB_PASSWORD", ""), "gitlab login password")
		webhookURL        = flag.String("webhook-url", getEnv("WEBHOOK_URL", ""), "Webhook URL")
		webhookToken      = flag.String("webhook-token", getEnv("WEBHOOK_TOKEN", ""), "Webhook secret token")
		addProject        = flag.Bool("add-project", false, "If true, add the project instead of creating it")
		delProject        = flag.Bool("del-project", false, "If true, delete the project instead of creating it")
		addProjectWebhook = flag.Bool("add-project-webhook", false, "If true, add the project webhook of creating it")
	)

	// 解析命令行参数
	flag.Parse()

	if !*addProject && !*delProject && !*addProjectWebhook {
		fmt.Fprintf(os.Stderr, "Error: one of --add-project, --delete-project, or --add-project-webhook is required\n\n")
		flag.Usage() // 自动打印所有 flags 的说明
		os.Exit(1)
	}

	if *addProject || *delProject {
		if *token == "" {
			log.Fatal("Error: token is required")
		}
		if *projectName == "" {
			log.Fatal("Error: project is required")
		}
	}

	if *addProjectWebhook {
		if *token == "" {
			log.Fatal("Error: token is required")
		}
		if *projectName == "" {
			log.Fatal("Error: project is required")
		}
		if *username == "" {
			log.Fatal("Error: username is required")
		}
		if *password == "" {
			log.Fatal("Error: password is required")
		}
		if *webhookURL == "" {
			log.Fatal("Error: webhook-url is required")
		}
		if *webhookToken == "" {
			log.Fatal("Error: webhook-token is required")
		}
	}

	// 替换 . 为 -（仅用于路径），并且转换为小写
	sanitizedProjectName := strings.ToLower(strings.ReplaceAll(*projectName, ".", "-"))
	// 构造完整项目路径（用于查找）
	fullPath := fmt.Sprintf("%s/%s", *namespaceName, sanitizedProjectName)

	// 创建 gitlab sdk HTTP 客户端
	client := gitlab.NewClient(nil, *token)
	client.SetBaseURL(*gitlabURL + "/api/v3")

	// === 获取 namespace ID ===
	fmt.Printf("namespace ID for group: %s\n", *namespaceName)
	group, _, err := client.Groups.GetGroup(*namespaceName)
	if err != nil {
		log.Fatalf("Failed to find group '%s': %v\n", *namespaceName, err)
	}
	namespaceID := group.ID
	fmt.Printf("Found group '%s' with ID: %d\n", group.Name, namespaceID)

	// 删除项目
	if *delProject {
		fmt.Printf("Attempting to delete project: %s\n", fullPath)
		// GitLab API v3 的 DeleteProject 接受 project ID 或 "namespace/project" 字符串
		_, err := client.Projects.DeleteProject(fullPath)
		if err != nil {
			log.Fatalf("Failed to delete project '%s': %v\n", fullPath, err)
		}
		fmt.Printf("Project deleted successfully: %s\n", fullPath)
		return // 删除后直接退出
	}

	// 创建项目
	if *addProject {
		description := "Auto-created project for " + sanitizedProjectName
		visibility := gitlab.PrivateVisibility
		createOpts := &gitlab.CreateProjectOptions{
			Name:        &sanitizedProjectName,
			Path:        &sanitizedProjectName,
			Description: &description,
			NamespaceID: &namespaceID,
			Visibility:  &visibility,
		}
		fmt.Printf("Creating project: %s in group %s (ID: %d)\n", sanitizedProjectName, *namespaceName, namespaceID)
		project, _, err := client.Projects.CreateProject(createOpts)
		if err != nil {
			log.Fatalf("Failed to create project: %v\n", err)
		}
		fmt.Printf("Project created: %s/%s (ID: %d)\n", *gitlabURL, project.PathWithNamespace, project.ID)

	}
	// 添加 Webhook
	// 为什么不用gitlab sdk添加project webhook的原因: gitlab 8.9.0 API版本不支持在创建项目中添加webhook时传入webhook的token
	if *addProjectWebhook {
		projectURL := fmt.Sprintf("%s/%s/hooks", strings.TrimSuffix(*gitlabURL, "/"), fullPath)
		addWebhook(*gitlabURL, *username, *password, projectURL, *webhookToken, *webhookURL)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func addWebhook(gitlabURL, username, password, projectURL, webhookToken, webhookURL string) {
	// fmt.Printf("gitlabURL: %s\nusername: %s\npassword: %s\nprojectURL: %s\nwebhookToken: %s\nwebhookURL: %s\n", gitlabURL, username, password, projectURL, webhookToken, webhookURL)
	// 创建 HTTP 客户端
	// 启用 CookieJar，自动管理 Cookie
	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Jar: jar}

	// 获取 CSRF token
	loginURL := fmt.Sprintf("%s/users/sign_in", strings.TrimSuffix(gitlabURL, "/"))
	csrfToken, err := getCSRFToken(loginURL, cli)
	if err != nil {
		log.Fatalf("get csrfToken fail: %v\n", err)
	}

	// 模拟登录
	err = login(loginURL, username, password, csrfToken, cli)
	if err != nil {
		log.Fatalf("web login fail: %v\n", err)
	} else {
		fmt.Printf("web login success\n")
	}

	// 添加 Webhook
	err = webhook(projectURL, webhookURL, webhookToken, cli)
	if err != nil {
		log.Fatalf("add Webhook fail: %v\n", err)
	} else {
		fmt.Printf("add Webhook success\n")
	}
}

func getCSRFToken(loginURL string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status error %d: %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("parse HTML error: %w", err)
	}
	// 查找 input[name='authenticity_token']
	csrf := ""
	doc.Find("input[name='authenticity_token']").Each(func(i int, s *goquery.Selection) {
		if val, exists := s.Attr("value"); exists {
			csrf = val
		}
	})

	if csrf == "" {
		// 调试：打印部分 HTML 确认结构
		log.Printf("DEBUG: HTML snippet:\n%.500s", string(body))
		return "", fmt.Errorf("CSRF token (authenticity_token) not found in input field")
	}

	return csrf, nil
}

func login(loginURL, username, password, csrfToken string, client *http.Client) error {
	// 构造表单数据（不是 JSON！）
	formData := url.Values{}
	formData.Set("user[login]", username)
	formData.Set("user[password]", password)
	formData.Set("authenticity_token", csrfToken)
	formData.Set("utf8", "✓") // 注意：GitLab 用的是 ✓，不是 "?"
	formData.Set("user[remember_me]", "0")

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	// 设置正确的 Content-Type
	req.Header.Set("Referer", loginURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	// 读取 body（用于错误调试）
	body, _ := io.ReadAll(resp.Body)

	// GitLab 登录成功后通常会 302 重定向到首页
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func webhook(projectURL, projectHookUrl, projectHookToken string, client *http.Client) error {
	// 因为gitlab采用Rails架构，Rails 默认每次响应 HTML 时生成新 token（带随机 mask），防止重放攻击，所以每次需要提交表单前，重新 GET 页面获取最新 token
	csrfToken, err := getCSRFToken(projectURL, client)
	if err != nil {
		log.Fatalf("get csrfToken fail: %v", err)
	}

	// 构造表单数据（不是 JSON！）
	formData := url.Values{}
	formData.Set("utf8", "✓") // 注意：GitLab 用的是 ✓，不是 "?"
	formData.Set("authenticity_token", csrfToken)
	formData.Set("hook[url]", projectHookUrl)
	formData.Set("hook[token]", projectHookToken)
	formData.Set("hook[push_events]", "1")
	formData.Set("hook[tag_push_events]", "0")
	formData.Set("hook[note_events]", "0")
	formData.Set("hook[issues_events]", "0")
	formData.Set("hook[merge_requests_events]", "0")
	formData.Set("hook[build_events]", "0")
	formData.Set("hook[wiki_page_events]", "0")
	formData.Set("hook[enable_ssl_verification]", "0")

	req, err := http.NewRequest("POST", projectURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	// 设置正确的 Content-Type
	req.Header.Set("Referer", projectURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	// 读取 body（用于错误调试）
	body, _ := io.ReadAll(resp.Body)

	// GitLab 登录成功后通常会 302 重定向到首页
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
