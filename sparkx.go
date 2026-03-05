// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/anil-wu/spark-x/internal/config"
	"github.com/anil-wu/spark-x/internal/handler"
	"github.com/anil-wu/spark-x/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/sparkx-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		c.MySQL.DSN = dsn
	}

	if endpoint := os.Getenv("OSS_ENDPOINT"); endpoint != "" {
		c.OSS.Endpoint = endpoint
	}
	if accessKeyId := os.Getenv("OSS_ACCESS_KEY_ID"); accessKeyId != "" {
		c.OSS.AccessKeyId = accessKeyId
	}
	if accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET"); accessKeySecret != "" {
		c.OSS.AccessKeySecret = accessKeySecret
	}
	if bucket := os.Getenv("OSS_BUCKET"); bucket != "" {
		c.OSS.Bucket = bucket
	}
	if expireSeconds := os.Getenv("OSS_EXPIRE_SECONDS"); expireSeconds != "" {
		if v, err := strconv.ParseInt(expireSeconds, 10, 64); err == nil {
			c.OSS.ExpireSeconds = v
		}
	}

	if provider := strings.TrimSpace(os.Getenv("STORAGE_PROVIDER")); provider != "" {
		c.Storage.Provider = provider
	}
	if expireSeconds := strings.TrimSpace(os.Getenv("STORAGE_EXPIRE_SECONDS")); expireSeconds != "" {
		if v, err := strconv.ParseInt(expireSeconds, 10, 64); err == nil {
			c.Storage.ExpireSeconds = v
		}
	}

	if endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT")); endpoint != "" {
		c.S3.Endpoint = endpoint
	}
	if accessKeyId := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID")); accessKeyId != "" {
		c.S3.AccessKeyId = accessKeyId
	}
	if accessKeySecret := os.Getenv("S3_ACCESS_KEY_SECRET"); strings.TrimSpace(accessKeySecret) != "" {
		c.S3.AccessKeySecret = accessKeySecret
	}
	if bucket := strings.TrimSpace(os.Getenv("S3_BUCKET")); bucket != "" {
		c.S3.Bucket = bucket
	}
	if region := strings.TrimSpace(os.Getenv("S3_REGION")); region != "" {
		c.S3.Region = region
	}
	if rawUseSSL := strings.TrimSpace(os.Getenv("S3_USE_SSL")); rawUseSSL != "" {
		if v, err := strconv.ParseBool(rawUseSSL); err == nil {
			c.S3.UseSSL = v
		}
	}
	if expireSeconds := strings.TrimSpace(os.Getenv("S3_EXPIRE_SECONDS")); expireSeconds != "" {
		if v, err := strconv.ParseInt(expireSeconds, 10, 64); err == nil {
			c.S3.ExpireSeconds = v
		}
	}

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
