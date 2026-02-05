// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package projects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProjectLogic {
	return &DeleteProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteProjectLogic) DeleteProject(req *types.DeleteProjectReq) (resp *types.BaseResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 {
		return nil, errors.New("id required")
	}

	// Check ownership
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.Projects{}).Where("id = ? AND owner_id = ?", req.Id, userId).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	subFiles := tx.Table("files").Select("id").Where("project_id = ?", req.Id)
	if err = tx.Table("file_versions").Where("file_id IN (?)", subFiles).Delete(&model.FileVersions{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = tx.Table("files").Where("project_id = ?", req.Id).Delete(&model.Files{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = tx.Table("project_members").Where("project_id = ?", req.Id).Delete(&model.ProjectMembers{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = tx.Table("projects").Where("id = ?", req.Id).Delete(&model.Projects{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err = tx.Commit().Error; err != nil {
		return nil, err
	}
	resp = &types.BaseResp{Code: 0, Msg: "ok"}

	return resp, nil
}
