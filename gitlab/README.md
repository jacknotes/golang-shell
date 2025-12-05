# 说明

此脚本用于gitlab v8.9.11，用来创建项目和删除项目。


## 编译
```bash
GOOS=linux GOARCH=amd64 go build -o ./dist/gitlab-create-project main.go
```



## 创建项目
```bash
GITLAB_TOKEN="abcfdf" PROJECT_NAME="dotnet-testv1-service-hs-com" ./dist/gitlab-create-project.exe
```

## 删除项目
```bash
./dist/gitlab-create-project.exe -delete-project -project frontend-testv2-hs-com -token abcfdf
```
