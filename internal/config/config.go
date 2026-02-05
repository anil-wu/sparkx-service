// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MySQL struct {
		DSN string
	}
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Google struct {
		ClientID string
	}
	OSS struct {
		Endpoint        string
		AccessKeyId     string
		AccessKeySecret string
		Bucket          string
		ExpireSeconds   int64
	}
}
