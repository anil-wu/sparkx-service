// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/anil-wu/spark-x/internal/config"
	"github.com/anil-wu/spark-x/internal/handler"
	agents "github.com/anil-wu/spark-x/internal/handler/agents"
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

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/agents/configs",
				Handler: agents.ListAgentConfigsHandler(ctx),
			},
		},
		rest.WithJwt(ctx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/v1"),
	)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
