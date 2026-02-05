// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PreUploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreUploadFileLogic {
	return &PreUploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreUploadFileLogic) PreUploadFile(req *types.PreUploadReq) (resp *types.PreUploadResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.ProjectId <= 0 || req.Name == "" || req.FileCategory == "" {
		return nil, model.InputParamInvalid
	}

	// Check project membership
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).Where("project_id = ? AND user_id = ?", req.ProjectId, userId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	// ensure file exists or create
	var file *model.Files
	// try find by unique (project_id,name)
	err = l.svcCtx.DB.WithContext(l.ctx).Where("project_id = ? AND name = ?", req.ProjectId, req.Name).First(&file).Error
	if err != nil {
		// create new file
		newFile := &model.Files{
			ProjectId:    uint64(req.ProjectId),
			Name:         req.Name,
			FileCategory: req.FileCategory,
		}
		_, err = l.svcCtx.FilesModel.Insert(l.ctx, newFile)
		if err != nil {
			return nil, err
		}
		file = newFile
	}
	// get max version
	var maxVerNumber int64
	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.FileVersions{}).Where("file_id = ?", file.Id).Select("COALESCE(MAX(version_number),0)").Scan(&maxVerNumber).Error
	if err != nil {
		return nil, err
	}
	nextVer := maxVerNumber + 1

	// storage path
	objectPath := fmt.Sprintf("projects/%d/%s/%d", req.ProjectId, req.Name, nextVer)
	// create version row (without actual upload)
	newVer := &model.FileVersions{
		FileId:        file.Id,
		VersionNumber: uint64(nextVer),
		SizeBytes:     uint64(req.SizeBytes),
		Hash:          req.Hash,
		StoragePath:   objectPath,
		MimeType:      req.MimeType,
		CreatedBy:     uint64(userId),
	}
	_, err = l.svcCtx.FileVersionsModel.Insert(l.ctx, newVer)
	if err != nil {
		return nil, err
	}
	// pre-signed url
	url, err := l.svcCtx.OSSBucket.SignURL(objectPath, "PUT", int64(l.svcCtx.Config.OSS.ExpireSeconds))
	if err != nil {
		return nil, err
	}
	resp = &types.PreUploadResp{
		UploadUrl:     url,
		FileId:        int64(file.Id),
		VersionId:     int64(newVer.Id),
		VersionNumber: int64(newVer.VersionNumber),
	}

	return resp, nil
}
