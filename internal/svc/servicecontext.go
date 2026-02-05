// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
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

	UsersModel          model.UsersModel
	UserIdentitiesModel model.UserIdentitiesModel
	ProjectsModel       model.ProjectsModel
	ProjectMembersModel model.ProjectMembersModel
	FilesModel          model.FilesModel
	FileVersionsModel   model.FileVersionsModel

	OSSClient *oss.Client
	OSSBucket *oss.Bucket
}

func NewServiceContext(c config.Config) *ServiceContext {
	var db *gorm.DB
	var err error
	if c.MySQL.DSN != "" {
		db, err = gorm.Open(mysql.Open(c.MySQL.DSN), &gorm.Config{})
		if err != nil {
			panic(err)
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
	var filesModel model.FilesModel
	var fileVersionsModel model.FileVersionsModel
	if db != nil && conn != nil {
		usersModel = model.NewUsersModel(db, conn)
		userIdentitiesModel = model.NewUserIdentitiesModel(db, conn)
		projectsModel = model.NewProjectsModel(db, conn)
		projectMembersModel = model.NewProjectMembersModel(db, conn)
		filesModel = model.NewFilesModel(db, conn)
		fileVersionsModel = model.NewFileVersionsModel(db, conn)
	}

	// init oss
	var ossClient *oss.Client
	var bucket *oss.Bucket
	if c.OSS.Endpoint != "" && c.OSS.Bucket != "" {
		ossClient, err = oss.New(c.OSS.Endpoint, c.OSS.AccessKeyId, c.OSS.AccessKeySecret)
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
		Config:              c,
		DB:                  db,
		Conn:                conn,
		UsersModel:          usersModel,
		UserIdentitiesModel: userIdentitiesModel,
		ProjectsModel:       projectsModel,
		ProjectMembersModel: projectMembersModel,
		FilesModel:          filesModel,
		FileVersionsModel:   fileVersionsModel,
		OSSClient:           ossClient,
		OSSBucket:           bucket,
	}
}
