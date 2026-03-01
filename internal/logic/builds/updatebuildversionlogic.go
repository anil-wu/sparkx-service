package builds

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateBuildVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateBuildVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBuildVersionLogic {
	return &UpdateBuildVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func normalizePreviewPrefix(raw string) string {
	p := strings.TrimSpace(raw)
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimLeft(p, "/")
	if p != "" && !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func (l *UpdateBuildVersionLogic) UpdateBuildVersion(req *types.UpdateBuildVersionReq) (resp *types.UpdateBuildVersionResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.Id <= 0 {
		return nil, model.InputParamInvalid
	}
	if l.svcCtx.DB == nil {
		return nil, errors.New("db not configured")
	}

	var bv model.BuildVersions
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", req.Id).First(&bv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
		Where("project_id = ? AND user_id = ?", bv.ProjectId, userId).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("project not found or permission denied")
	}

	updates := map[string]any{}
	if strings.TrimSpace(req.PreviewStoragePrefix) != "" {
		updates["preview_storage_prefix"] = normalizePreviewPrefix(req.PreviewStoragePrefix)
	}
	if strings.TrimSpace(req.EntryPath) != "" {
		updates["entry_path"] = normalizeEntryPath(req.EntryPath)
	}

	if req.BuildVersionFileId > 0 || req.BuildVersionFileVersionId > 0 {
		if req.BuildVersionFileId <= 0 || req.BuildVersionFileVersionId <= 0 {
			return nil, errors.New("buildVersionFileId and buildVersionFileVersionId must be both provided")
		}
		updates["build_version_file_id"] = uint64(req.BuildVersionFileId)
		updates["build_version_file_version_id"] = uint64(req.BuildVersionFileVersionId)
	}

	if len(updates) == 0 {
		return &types.UpdateBuildVersionResp{
			BuildVersionId:            int64(bv.Id),
			PreviewStoragePrefix:      bv.PreviewStoragePrefix,
			EntryPath:                 bv.EntryPath,
			BuildVersionFileId:        int64(bv.BuildVersionFileId),
			BuildVersionFileVersionId: int64(bv.BuildVersionFileVersionId),
		}, nil
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.BuildVersions{}).
		Where("id = ?", bv.Id).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", bv.Id).First(&bv).Error; err != nil {
		return nil, err
	}

	return &types.UpdateBuildVersionResp{
		BuildVersionId:            int64(bv.Id),
		PreviewStoragePrefix:      bv.PreviewStoragePrefix,
		EntryPath:                 bv.EntryPath,
		BuildVersionFileId:        int64(bv.BuildVersionFileId),
		BuildVersionFileVersionId: int64(bv.BuildVersionFileVersionId),
	}, nil
}
