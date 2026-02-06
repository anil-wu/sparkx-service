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

type RollbackVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRollbackVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackVersionLogic {
	return &RollbackVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RollbackVersionLogic) RollbackVersion(req *types.RollbackVersionReq) (resp *types.BaseResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 || req.VersionNumber <= 0 {
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

	// 检查项目成员权限（需要 developer 及以上角色）
	var member model.ProjectMembers
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("project_id = ? AND user_id = ?", file.ProjectId, userId).
		First(&member).Error; err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("project not found or permission denied")
		}
		return nil, err
	}

	// 检查权限：viewer 不能回滚版本
	if member.Role == "viewer" {
		return nil, errors.New("permission denied: viewer cannot rollback versions")
	}

	// 查找目标版本
	targetVersion, err := l.svcCtx.FileVersionsModel.FindOneByFileIdVersionNumber(
		l.ctx, uint64(req.Id), uint64(req.VersionNumber))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("target version not found")
		}
		return nil, err
	}

	// 检查目标版本是否已经是当前版本
	if file.CurrentVersionId == targetVersion.Id {
		return nil, errors.New("target version is already the current version")
	}

	// 更新文件的当前版本
	file.CurrentVersionId = targetVersion.Id
	_, err = l.svcCtx.FilesModel.Update(l.ctx, int64(file.Id), file)
	if err != nil {
		l.Errorf("[RollbackVersion] Failed to update file: %v", err)
		return nil, err
	}

	l.Infof("[RollbackVersion] Successfully rolled back fileId=%d to versionNumber=%d, versionId=%d",
		req.Id, req.VersionNumber, targetVersion.Id)

	resp = &types.BaseResp{
		Code: 0,
		Msg:  "success",
	}
	return resp, nil
}
