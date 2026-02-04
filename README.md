# SparkX Service

SparkX 后端服务，基于 [go-zero](https://go-zero.dev) 框架构建，提供用户管理、项目协作、文件版本控制等核心功能。

## 功能特性

- **用户系统**: 注册、登录、个人信息管理
- **项目管理**: 创建、更新、归档项目，支持项目列表分页查询
- **团队协作**: 项目成员邀请与权限管理 (Owner, Admin, Developer, Viewer)
- **资产管理**: 文件预上传、版本控制、多媒体类型支持

## 技术栈

- **Golang**: 1.24+
- **Framework**: go-zero v1.9.4
- **Database**: MySQL 8.0+
- **ORM**: GORM v1.31+ (与 go-zero sqlx 混合使用)
- **Object Storage**: Aliyun OSS

## 目录结构

```text
service/
├── etc/                # 配置文件
├── internal/
│   ├── config/         # 配置定义
│   ├── handler/        # HTTP 路由处理 (goctl 生成)
│   ├── logic/          # 业务逻辑实现
│   ├── model/          # 数据库模型 (GORM & sqlx)
│   ├── svc/            # 服务依赖上下文
│   └── types/          # API 请求/响应类型定义
├── sql/                # 数据库建表脚本
├── sparkx.api          # API 接口定义 DSL
├── sparkx.go           # 服务入口文件
└── go.mod              # 依赖管理
```

## 快速开始

### 1. 环境准备

确保本地已安装以下环境：
- Go 1.24 或更高版本
- MySQL 8.0+
- (可选) Redis (取决于业务逻辑是否启用缓存)

### 2. 数据库初始化

创建数据库 `sparkx` 并导入表结构：

```bash
# 登录 MySQL
mysql -u root -p

# 创建数据库
CREATE DATABASE sparkx DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 导入表结构
use sparkx;
source sql/schema.sql;
```

### 3. 配置文件

复制或直接修改 `etc/sparkx-api.yaml`，配置数据库和 OSS 信息：

```yaml
Name: sparkx-api
Host: 0.0.0.0
Port: 8890

MySQL:
  DSN: "user:password@tcp(127.0.0.1:3306)/sparkx?charset=utf8mb4&parseTime=true&loc=Local"

OSS:
  Endpoint: "https://oss-cn-xxx.aliyuncs.com"
  AccessKeyId: "your-access-key"
  AccessKeySecret: "your-secret"
  Bucket: "your-bucket-name"
  ExpireSeconds: 1800
```

### 4. 运行服务

```bash
# 安装依赖
go mod tidy

# 启动服务
go run sparkx.go -f etc/sparkx-api.yaml
```

服务启动后将监听 `8890` 端口。

## Docker 部署

### 1. 构建镜像

```bash
docker build -t sparkx-service:v1 .
```

### 2. 运行容器

```bash
docker run --rm -it -p 8890:8890 sparkx-service:v1
```

## 环境变量

支持通过环境变量覆盖部分配置：

- `MYSQL_DSN`: MySQL 连接字符串，格式参考 `etc/sparkx-api.yaml`。

## 开发指南

### 代码生成

本项目主要使用 `goctl` 工具维护代码结构。

**安装 goctl**:
```bash
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

**更新 API 代码**:
当修改了 `sparkx.api` 文件后，执行以下命令更新 handler 和 types：
```bash
goctl api go -api sparkx.api -dir . -style gozero
```

**更新 Model 代码**:
当修改了 `sql/schema.sql` 后，可使用以下命令生成基础 model 代码（注意不要覆盖自定义逻辑）：
```bash
goctl model mysql ddl --src sql/schema.sql --dir "./internal/model"
```
*注意：本项目 Model 层集成了 GORM，生成代码后可能需要手动调整 `NewUsersModel` 等构造函数以注入 `*gorm.DB`。*

### 常见问题

1. **包导入错误**:
   执行 `go mod tidy` 清理和下载依赖。

2. **GORM 与 sqlx**:
   项目在 `internal/svc/servicecontext.go` 中初始化了 GORM 连接，并传递给 Model 层。在 Logic 层中，根据需要选择使用 `l.svcCtx.UserModel` 提供的方法。
