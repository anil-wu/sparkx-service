// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFileLogic {
	return &DeleteFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFileLogic) DeleteFile(req *types.DeleteFileReq) (resp *types.BaseResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 {
		return nil, model.InputParamInvalid
	}

	// 获取文件信息
	file, err := l.svcCtx.FilesModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}

	// Get project_id from project_files
	var projectFile model.ProjectFiles
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("file_id = ?", file.Id).First(&projectFile).Error; err != nil {
		return nil, err
	}

	// 检查项目成员权限（需要 admin 或 owner 角色）
	var member model.ProjectMembers
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("project_id = ? AND user_id = ?", projectFile.ProjectId, userId).
		First(&member).Error; err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("project not found or permission denied")
		}
		return nil, err
	}

	// 只有 owner 和 admin 可以删除文件
	if member.Role != "owner" && member.Role != "admin" {
		return nil, errors.New("permission denied: only owner or admin can delete files")
	}

	// 获取文件的所有版本
	var versions []model.FileVersions
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("file_id = ?", file.Id).
		Find(&versions).Error; err != nil {
		l.Errorf("[DeleteFile] Failed to find versions: %v", err)
		return nil, err
	}

	// 从 OSS 删除所有版本的文件
	for _, version := range versions {
		if version.StorageKey != "" {
			if err := l.svcCtx.OSSBucket.DeleteObject(version.StorageKey); err != nil {
				l.Errorf("[DeleteFile] Failed to delete OSS object %s: %v", version.StorageKey, err)
				// 继续删除其他版本，记录错误但不中断
			}
		}
	}
	l.Infof("[DeleteFile] Deleted %d OSS objects for fileId=%d", len(versions), req.Id)

	// 删除数据库中的版本记录
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("file_id = ?", file.Id).
		Delete(&model.FileVersions{}).Error; err != nil {
		l.Errorf("[DeleteFile] Failed to delete versions: %v", err)
		return nil, err
	}

	// 软删除文件记录（使用 GORM 的 Delete 方法会自动处理软删除）
	_, err = l.svcCtx.FilesModel.Delete(l.ctx, uint64(req.Id))
	if err != nil {
		l.Errorf("[DeleteFile] Failed to delete file: %v", err)
		return nil, err
	}

	l.Infof("[DeleteFile] Successfully deleted fileId=%d, name=%s", req.Id, file.Name)

	resp = &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}
	return resp, nil
}
