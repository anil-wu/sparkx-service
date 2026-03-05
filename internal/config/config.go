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
	AdminAuth struct {
		AccessSecret string
		AccessExpire int64
	}
	Google struct {
		ClientID string
	}
	Storage struct {
		Provider      string
		ExpireSeconds int64
	}
	OSS struct {
		Endpoint        string
		AccessKeyId     string
		AccessKeySecret string
		Bucket          string
		ExpireSeconds   int64
	}
	S3 struct {
		Endpoint        string
		AccessKeyId     string
		AccessKeySecret string
		Bucket          string
		Region          string
		UseSSL          bool
		ExpireSeconds   int64
	}
}
