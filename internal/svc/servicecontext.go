// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/anil-wu/spark-x/internal/config"
	"github.com/anil-wu/spark-x/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Conn   sqlx.SqlConn

	UsersModel             model.UsersModel
	UserIdentitiesModel    model.UserIdentitiesModel
	ProjectsModel          model.ProjectsModel
	ProjectMembersModel    model.ProjectMembersModel
	ProjectFilesModel      model.ProjectFilesModel
	FilesModel             model.FilesModel
	FileVersionsModel      model.FileVersionsModel
	AdminsModel            model.AdminsModel
	SoftwareTemplatesModel model.SoftwareTemplatesModel

	OSSClient *oss.Client
	OSSBucket *oss.Bucket
}

func normalizeOSSEndpoint(rawEndpoint, bucket string) string {
	raw := strings.TrimSpace(rawEndpoint)
	if raw == "" {
		return raw
	}

	b := strings.TrimSpace(bucket)

	var scheme string
	var host string

	u, err := url.Parse(raw)
	if err == nil && u.Host != "" {
		scheme = u.Scheme
		host = u.Host
	} else {
		u2, err2 := url.Parse("https://" + raw)
		if err2 == nil && u2.Host != "" {
			scheme = "https"
			host = u2.Host
		} else {
			host = raw
		}
	}

	host = strings.TrimSuffix(host, "/")

	if b != "" {
		prefix := b + "."
		for strings.HasPrefix(host, prefix) {
			host = strings.TrimPrefix(host, prefix)
		}
	}

	if scheme != "" {
		return scheme + "://" + host
	}
	return host
}

func NewServiceContext(c config.Config) *ServiceContext {
	var db *gorm.DB
	var err error
	if c.MySQL.DSN != "" {
		db, err = gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.SetMaxOpenConns(20)
			sqlDB.SetMaxIdleConns(5)
			sqlDB.SetConnMaxLifetime(5 * time.Minute)
			sqlDB.SetConnMaxIdleTime(1 * time.Minute)

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if pingErr := sqlDB.PingContext(ctx); pingErr != nil {
				logx.Errorf("mysql ping failed: %v", pingErr)
			}
		}
	}
	var conn sqlx.SqlConn
	if c.MySQL.DSN != "" {
		conn = sqlx.NewMysql(c.MySQL.DSN)
	}

	// init models
	var usersModel model.UsersModel
	var userIdentitiesModel model.UserIdentitiesModel
	var projectsModel model.ProjectsModel
	var projectMembersModel model.ProjectMembersModel
	var projectFilesModel model.ProjectFilesModel
	var filesModel model.FilesModel
	var fileVersionsModel model.FileVersionsModel
	var adminsModel model.AdminsModel
	var softwareTemplatesModel model.SoftwareTemplatesModel
	if db != nil && conn != nil {
		usersModel = model.NewUsersModel(db, conn)
		userIdentitiesModel = model.NewUserIdentitiesModel(db, conn)
		projectsModel = model.NewProjectsModel(db, conn)
		projectMembersModel = model.NewProjectMembersModel(db, conn)
		projectFilesModel = model.NewProjectFilesModel(db, conn)
		filesModel = model.NewFilesModel(db, conn)
		fileVersionsModel = model.NewFileVersionsModel(db, conn)
		adminsModel = model.NewAdminsModel(db, conn)
		softwareTemplatesModel = model.NewSoftwareTemplatesModel(db, conn)
	}

	// init oss
	var ossClient *oss.Client
	var bucket *oss.Bucket
	if c.OSS.Endpoint != "" && c.OSS.Bucket != "" {
		endpoint := normalizeOSSEndpoint(c.OSS.Endpoint, c.OSS.Bucket)
		ossClient, err = oss.New(endpoint, c.OSS.AccessKeyId, c.OSS.AccessKeySecret)
		if err != nil {
			logx.Errorf("oss init failed: %v", err)
		} else {
			bucket, err = ossClient.Bucket(c.OSS.Bucket)
			if err != nil {
				logx.Errorf("oss bucket init failed: %v", err)
			}
		}
	}

	return &ServiceContext{
		Config:                 c,
		DB:                     db,
		Conn:                   conn,
		UsersModel:             usersModel,
		UserIdentitiesModel:    userIdentitiesModel,
		ProjectsModel:          projectsModel,
		ProjectMembersModel:    projectMembersModel,
		ProjectFilesModel:      projectFilesModel,
		FilesModel:             filesModel,
		FileVersionsModel:      fileVersionsModel,
		AdminsModel:            adminsModel,
		SoftwareTemplatesModel: softwareTemplatesModel,
		OSSClient:              ossClient,
		OSSBucket:              bucket,
	}
}
