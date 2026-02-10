package builds

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateReleaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateReleaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateReleaseLogic {
	return &CreateReleaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateReleaseLogic) CreateRelease(req *types.CreateReleaseReq) (resp *types.CreateReleaseResp, err error) {
	userIdNumber, ok := l.ctx.Value("userId").(json.Number)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	userId, _ := userIdNumber.Int64()

	if req == nil ||
		req.ProjectId <= 0 ||
		req.BuildVersionId <= 0 ||
		req.ReleaseManifestFileId <= 0 ||
		req.ReleaseManifestFileVersionId <= 0 ||
		strings.TrimSpace(req.Name) == "" ||
		strings.TrimSpace(req.Channel) == "" ||
		strings.TrimSpace(req.Platform) == "" {
		return nil, model.InputParamInvalid
	}
	if l.svcCtx.DB == nil {
		return nil, errors.New("db not configured")
	}

	channel := strings.TrimSpace(req.Channel)
	if !isOneOf(channel, "dev", "qa", "beta", "prod") {
		return nil, model.InputParamInvalid
	}

	platform := strings.TrimSpace(req.Platform)
	if !isOneOf(platform, "web", "android", "ios", "desktop") {
		return nil, model.InputParamInvalid
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if !isOneOf(status, "active", "rolled_back", "archived") {
		return nil, model.InputParamInvalid
	}

	var publishedAt sql.NullTime
	if strings.TrimSpace(req.PublishedAt) != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(req.PublishedAt), time.Local)
		if err != nil {
			return nil, model.InputParamInvalid
		}
		publishedAt = sql.NullTime{Time: t, Valid: true}
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

	var bv model.BuildVersions
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("`id` = ?", req.BuildVersionId).First(&bv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	if bv.ProjectId != uint64(req.ProjectId) {
		return nil, model.InputParamInvalid
	}

	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	r := &model.Releases{
		ProjectId:                   uint64(req.ProjectId),
		BuildVersionId:              uint64(req.BuildVersionId),
		ReleaseManifestFileId:       uint64(req.ReleaseManifestFileId),
		ReleaseManifestFileVersionId: uint64(req.ReleaseManifestFileVersionId),
		Name:                        strings.TrimSpace(req.Name),
		Channel:                     channel,
		Platform:                    platform,
		Status:                      status,
		VersionTag:                  sql.NullString{String: strings.TrimSpace(req.VersionTag), Valid: strings.TrimSpace(req.VersionTag) != ""},
		ReleaseNotes:                sql.NullString{String: req.ReleaseNotes, Valid: strings.TrimSpace(req.ReleaseNotes) != ""},
		PublishedAt:                 publishedAt,
		CreatedBy:                   uint64(userId),
	}
	if req.Id > 0 {
		r.Id = uint64(req.Id)
	}
	if err := tx.Create(r).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	tx = nil

	publishedAtStr := ""
	if r.PublishedAt.Valid {
		publishedAtStr = r.PublishedAt.Time.Format("2006-01-02 15:04:05")
	}

	return &types.CreateReleaseResp{
		ReleaseId:                   int64(r.Id),
		ProjectId:                   req.ProjectId,
		BuildVersionId:              req.BuildVersionId,
		ReleaseManifestFileId:       req.ReleaseManifestFileId,
		ReleaseManifestFileVersionId: req.ReleaseManifestFileVersionId,
		Name:                        r.Name,
		Channel:                     r.Channel,
		Platform:                    r.Platform,
		Status:                      r.Status,
		VersionTag:                  r.VersionTag.String,
		ReleaseNotes:                r.ReleaseNotes.String,
		CreatedBy:                   int64(r.CreatedBy),
		CreatedAt:                   r.CreatedAt.Format("2006-01-02 15:04:05"),
		PublishedAt:                 publishedAtStr,
	}, nil
}

func isOneOf(v string, allowed ...string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}
