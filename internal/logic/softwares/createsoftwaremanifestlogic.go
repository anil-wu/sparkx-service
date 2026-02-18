// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package softwares

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateSoftwareManifestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSoftwareManifestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSoftwareManifestLogic {
	return &CreateSoftwareManifestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSoftwareManifestLogic) CreateSoftwareManifest(req *types.CreateSoftwareManifestReq) (resp *types.CreateSoftwareManifestResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	isAdmin := false
	if !ok {
		adminIdNumber, ok2 := l.ctx.Value("adminId").(json.Number)
		if !ok2 {
			return nil, errors.New("unauthorized")
		}
		userIdNumber = adminIdNumber
		isAdmin = true
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.ProjectId <= 0 || req.SoftwareId <= 0 || req.ManifestFileId <= 0 || req.ManifestFileVersionId <= 0 {
		return nil, model.InputParamInvalid
	}

	if l.svcCtx.DB == nil {
		return nil, errors.New("db not configured")
	}

	var sw model.Softwares
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.SoftwareId).First(&sw).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	if sw.ProjectId != uint64(req.ProjectId) {
		return nil, model.InputParamInvalid
	}

	if !isAdmin {
		var count int64
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.ProjectMembers{}).
			Where("project_id = ? AND user_id = ?", req.ProjectId, userId).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, errors.New("project not found or permission denied")
		}
	}

	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// 获取当前软件工程下最大的 version_number
	var maxVersionNumber uint32
	if err := tx.Model(&model.SoftwareManifests{}).
		Where("software_id = ?", req.SoftwareId).
		Select("COALESCE(MAX(version_number), 0)").
		Scan(&maxVersionNumber).Error; err != nil {
		return nil, err
	}

	sm := &model.SoftwareManifests{
		ProjectId:             uint64(req.ProjectId),
		SoftwareId:            uint64(req.SoftwareId),
		ManifestFileId:        uint64(req.ManifestFileId),
		ManifestFileVersionId: uint64(req.ManifestFileVersionId),
		VersionNumber:         maxVersionNumber + 1,
		VersionDescription:    sql.NullString{String: req.VersionDescription, Valid: strings.TrimSpace(req.VersionDescription) != ""},
		CreatedBy:             uint64(userId),
	}
	if req.Id > 0 {
		sm.Id = uint64(req.Id)
	}
	if err := tx.Create(sm).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	tx = nil

	return &types.CreateSoftwareManifestResp{
		ManifestId:            int64(sm.Id),
		ProjectId:             req.ProjectId,
		SoftwareId:            req.SoftwareId,
		ManifestFileId:        req.ManifestFileId,
		ManifestFileVersionId: req.ManifestFileVersionId,
		VersionNumber:         int64(sm.VersionNumber),
		VersionDescription:    sm.VersionDescription.String,
		CreatedBy:             int64(sm.CreatedBy),
		CreatedAt:             sm.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
