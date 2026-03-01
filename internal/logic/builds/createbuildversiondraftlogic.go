package builds

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateBuildVersionDraftLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBuildVersionDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBuildVersionDraftLogic {
	return &CreateBuildVersionDraftLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func normalizeEntryPath(raw string) string {
	p := strings.TrimSpace(raw)
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimLeft(p, "/")
	if p == "" {
		return "index.html"
	}
	return p
}

func (l *CreateBuildVersionDraftLogic) CreateBuildVersionDraft(req *types.CreateBuildVersionDraftReq) (resp *types.CreateBuildVersionDraftResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.ProjectId <= 0 || req.SoftwareManifestId <= 0 {
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

	var sm model.SoftwareManifests
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.SoftwareManifestId).First(&sm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	if sm.ProjectId != uint64(req.ProjectId) {
		return nil, model.InputParamInvalid
	}

	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	versionNumber := uint32(req.VersionNumber)
	if versionNumber == 0 {
		var maxVersionNumber sql.NullInt64
		if err := tx.Model(&model.BuildVersions{}).
			Where("project_id = ? AND software_manifest_id = ?", req.ProjectId, req.SoftwareManifestId).
			Select("MAX(version_number)").
			Scan(&maxVersionNumber).Error; err != nil {
			return nil, err
		}
		if maxVersionNumber.Valid {
			versionNumber = uint32(maxVersionNumber.Int64) + 1
		} else {
			versionNumber = 1
		}
	}

	entryPath := normalizeEntryPath(req.EntryPath)
	bv := &model.BuildVersions{
		ProjectId:                 uint64(req.ProjectId),
		SoftwareManifestId:        uint64(req.SoftwareManifestId),
		VersionNumber:             versionNumber,
		Description:               sql.NullString{String: req.Description, Valid: strings.TrimSpace(req.Description) != ""},
		BuildVersionFileId:        0,
		BuildVersionFileVersionId: 0,
		PreviewStoragePrefix:      "",
		EntryPath:                 entryPath,
		CreatedBy:                 uint64(userId),
	}

	if err := tx.Create(bv).Error; err != nil {
		return nil, err
	}

	previewStoragePrefix := fmt.Sprintf("previews/%d/%d/%d/", req.ProjectId, sm.SoftwareId, bv.Id)
	if err := tx.Model(&model.BuildVersions{}).
		Where("id = ?", bv.Id).
		Updates(map[string]any{
			"preview_storage_prefix": previewStoragePrefix,
			"entry_path":             entryPath,
		}).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	tx = nil

	return &types.CreateBuildVersionDraftResp{
		BuildVersionId:       int64(bv.Id),
		ProjectId:            req.ProjectId,
		SoftwareManifestId:   req.SoftwareManifestId,
		VersionNumber:        int64(versionNumber),
		Description:          bv.Description.String,
		PreviewStoragePrefix: previewStoragePrefix,
		EntryPath:            entryPath,
		CreatedBy:            int64(userId),
		CreatedAt:            bv.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
