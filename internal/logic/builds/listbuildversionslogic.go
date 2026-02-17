// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package builds

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBuildVersionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBuildVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBuildVersionsLogic {
	return &ListBuildVersionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBuildVersionsLogic) ListBuildVersions(req *types.ListBuildVersionsReq) (resp *types.BuildVersionListResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.ProjectId <= 0 {
		return nil, model.InputParamInvalid
	}
	if l.svcCtx.DB == nil {
		return nil, errors.New("db not configured")
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
		Where("project_id = ? AND user_id = ?", req.ProjectId, userId).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.BuildVersions{}).
		Where("project_id = ?", req.ProjectId).
		Count(&total).Error; err != nil {
		return nil, err
	}

	var buildVersions []model.BuildVersions
	offset := (page - 1) * pageSize
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Where("project_id = ?", req.ProjectId).
		Order("id DESC").
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&buildVersions).Error; err != nil {
		return nil, err
	}

	list := make([]types.BuildVersionItem, 0, len(buildVersions))
	for _, bv := range buildVersions {
		item := types.BuildVersionItem{
			BuildVersionId:            int64(bv.Id),
			ProjectId:                 int64(bv.ProjectId),
			SoftwareManifestId:        int64(bv.SoftwareManifestId),
			Description:               bv.Description.String,
			BuildVersionFileId:        int64(bv.BuildVersionFileId),
			BuildVersionFileVersionId: int64(bv.BuildVersionFileVersionId),
			CreatedBy:                 int64(bv.CreatedBy),
			CreatedAt:                 bv.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		list = append(list, item)
	}

	return &types.BuildVersionListResp{
		List: list,
		Page: types.PageResp{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}
