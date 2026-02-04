# SparkX Service

sparkx-service

## 常用命令

### goctl api 命令
作用:根据 sparkx.api 生成 api 代码

> goctl api go -api sparkx.api -dir . -style gozero

如果有包导入问题则使用 `go mod tidy` 更新依赖

[参考文档](https://go-zero.dev/docs/tasks/dsl/api)

### 数据库模型生成命令
作用:根据 sql/schema.sql 生成数据库模型代码

> goctl model mysql ddl --src sql/schema.sql --dir "./models/"

[参考文档](https://go-zero.dev/docs/tasks/cli/mysql)

## 启动命令
作用:启动 sparkx-service 服务

> go run main.go -f etc/sparkx.yaml

## 配置文件
作用:配置 sparkx-service 服务

> sparkx-api.yaml

### MySQL 配置
作用:配置 MySQL 数据库连接

> DSN: "user:password@tcp(127.0.0.1:3306)/sparkx?charset=utf8mb4&parseTime=true&loc=Local"

### OSS 配置
作用:配置 OSS 对象存储连接

> Endpoint: "https://oss-your-region.aliyuncs.com"
> AccessKeyId: ""
> AccessKeySecret: ""
> Bucket: "your-bucket-name"
> ExpireSeconds: 1800