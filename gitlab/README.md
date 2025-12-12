# 使用说明
此脚本用于gitlab v8.9.11，用来创建项目、添加项目webhook、删除项目。


## 编译项目
```bash
GOOS=linux GOARCH=amd64 go build -o ./dist/gitlab-create-project main.go
GOOS=windows GOARCH=amd64 go build -o ./dist/gitlab-create-project.exe main.go
```

## 创建项目
```bash
./gitlab-create-project -project 'java-testv100.service.hs.com' -token 'a45nuRz1196YmHR3kn123' -add-project
```

## 添加项目webhook
```bash
./gitlab-create-project -project 'java-testv100.service.hs.com' -token 'a45nuRz1196YmHR3kn123' -webhook-url 'https://argocd.k8s.hs.com/api/webhook' -webhook-token 'uv6uHEyPI6Xbv12345tDfdNs1bBBtOL' -username 0001 -password password123 -add-project-webhook
```

## 创建项目并添加项目webhook
```bash
./gitlab-create-project -project 'java-testv100.service.hs.com' -token 'a45nuRz1196YmHR3kn123' -webhook-url 'https://argocd.k8s.hs.com/api/webhook' -webhook-token 'uv6uHEyPI6Xbv12345tDfdNs1bBBtOL' -username 0001 -password password123 -add-project -add-project-webhook
```

## 删除项目
```bash
./gitlab-create-project -project 'java-testv100.service.hs.com' -token 'a45nuRz1196YmHR3kn123' -del-project
```

