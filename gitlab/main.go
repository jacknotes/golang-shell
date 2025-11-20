package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"flag"

	gitlab "github.com/xanzy/go-gitlab"
)

func main() {
	// === 1. 定义命令行参数 ===
	var (
		gitlabURL     = flag.String("gitlab-url", getEnv("GITLAB_URL", "http://gitlab.hs.com"), "GitLab base URL")
		token         = flag.String("token", getEnv("GITLAB_TOKEN", ""), "GitLab private token")
		projectName   = flag.String("project", getEnv("PROJECT_NAME", ""), "Project name (or path)")
		webhookURL    = flag.String("webhook-url", getEnv("WEBHOOK_URL", "https://argocd.k8s.hs.com/api/webhook"), "Webhook URL")
		webhookToken  = flag.String("webhook-token", getEnv("WEBHOOK_TOKEN", "uv6uHEyPI6Xbvh7I4b5tDfdNs1bBBtOL"), "Webhook secret token")
		namespaceName = flag.String("namespace", getEnv("NAMESPACE", "k8s-deploy"), "Namespace (group) name")
		deleteProject = flag.Bool("delete-project", false, "If true, delete the project instead of creating it")
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("Error: GITLAB_TOKEN is required")
	}
	if *projectName == "" {
		log.Fatal("Error: PROJECT_NAME is required")
	}

	// 替换 . 为 -（仅用于路径）
	sanitizedProjectName := strings.ReplaceAll(*projectName, ".", "-")

	client := gitlab.NewClient(nil, *token)
	client.SetBaseURL(*gitlabURL + "/api/v3")

	// === 获取 namespace ID ===
	fmt.Printf("🔍 Looking up namespace ID for group: %s\n", *namespaceName)
	group, _, err := client.Groups.GetGroup(*namespaceName)
	if err != nil {
		log.Fatalf("Failed to find group '%s': %v", *namespaceName, err)
	}
	namespaceID := group.ID
	fmt.Printf("✅ Found group '%s' with ID: %d\n", group.Name, namespaceID)

	// 构造完整项目路径（用于查找）
	fullPath := fmt.Sprintf("%s/%s", *namespaceName, sanitizedProjectName)

	if *deleteProject {
		// ======================
		// 🔥 删除项目逻辑
		// ======================
		fmt.Printf("🗑️  Attempting to delete project: %s\n", fullPath)

		// GitLab API v3 的 DeleteProject 接受 project ID 或 "namespace/project" 字符串
		_, err := client.Projects.DeleteProject(fullPath)
		if err != nil {
			log.Fatalf("❌ Failed to delete project '%s': %v", fullPath, err)
		}
		fmt.Printf("✅ Project deleted successfully: %s\n", fullPath)
		return // 删除后直接退出
	}

	// ======================
	// ➕ 创建项目逻辑（原有）
	// ======================
	description := "Auto-created project for " + sanitizedProjectName
	visibility := gitlab.PrivateVisibility

	createOpts := &gitlab.CreateProjectOptions{
		Name:        &sanitizedProjectName,
		Path:        &sanitizedProjectName,
		Description: &description,
		NamespaceID: &namespaceID,
		Visibility:  &visibility,
	}

	fmt.Printf("🚀 Creating project: %s in group %s (ID: %d)\n", sanitizedProjectName, *namespaceName, namespaceID)
	project, _, err := client.Projects.CreateProject(createOpts)
	if err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}
	fmt.Printf("✅ Project created: %s/%s (ID: %d)\n", *gitlabURL, project.PathWithNamespace, project.ID)

	// 添加 Webhook
	hookOpts := &gitlab.AddProjectHookOptions{
		URL:                   webhookURL,
		PushEvents:            &[]bool{true}[0],
		MergeRequestsEvents:   &[]bool{false}[0],
		TagPushEvents:         &[]bool{false}[0],
		EnableSSLVerification: &[]bool{false}[0],
		Token:                 webhookToken,
	}

	fmt.Println("🔗 Adding webhook...")
	hook, _, err := client.Projects.AddProjectHook(project.ID, hookOpts)
	if err != nil {
		log.Fatalf("Failed to add webhook: %v", err)
	}

	if hook != nil {
		fmt.Printf("✅ Webhook added: %s\n", (*hook).URL)
	} else {
		fmt.Println("✅ Webhook added (URL not returned by API)")
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
