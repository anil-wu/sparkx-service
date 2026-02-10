package builds

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

type CreateBuildVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateBuildVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBuildVersionLogic {
	return &CreateBuildVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateBuildVersionLogic) CreateBuildVersion(req *types.CreateBuildVersionReq) (resp *types.CreateBuildVersionResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil || req.ProjectId <= 0 || req.SoftwareManifestId <= 0 || req.BuildVersionFileId <= 0 || req.BuildVersionFileVersionId <= 0 {
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

	bv := &model.BuildVersions{
		ProjectId:                 uint64(req.ProjectId),
		SoftwareManifestId:        uint64(req.SoftwareManifestId),
		Description:               sql.NullString{String: req.Description, Valid: strings.TrimSpace(req.Description) != ""},
		BuildVersionFileId:        uint64(req.BuildVersionFileId),
		BuildVersionFileVersionId: uint64(req.BuildVersionFileVersionId),
		CreatedBy:                 uint64(userId),
	}
	if req.Id > 0 {
		bv.Id = uint64(req.Id)
	}
	if err := tx.Create(bv).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	tx = nil

	return &types.CreateBuildVersionResp{
		BuildVersionId:            int64(bv.Id),
		ProjectId:                 req.ProjectId,
		SoftwareManifestId:        req.SoftwareManifestId,
		Description:               bv.Description.String,
		BuildVersionFileId:        req.BuildVersionFileId,
		BuildVersionFileVersionId: req.BuildVersionFileVersionId,
		CreatedBy:                 int64(bv.CreatedBy),
		CreatedAt:                 bv.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

