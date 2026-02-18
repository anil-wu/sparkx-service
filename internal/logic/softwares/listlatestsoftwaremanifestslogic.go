package softwares

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLatestSoftwareManifestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLatestSoftwareManifestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLatestSoftwareManifestsLogic {
	return &ListLatestSoftwareManifestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLatestSoftwareManifestsLogic) ListLatestSoftwareManifests(req *types.ListLatestSoftwareManifestsReq) (resp *types.LatestSoftwareManifestListResp, err error) {
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

	if req == nil || req.ProjectId <= 0 {
		return nil, model.InputParamInvalid
	}
	softwareIds, err := parseSoftwareIds(req.SoftwareIds)
	if err != nil || len(softwareIds) == 0 {
		return nil, model.InputParamInvalid
	}
	if l.svcCtx.DB == nil {
		return nil, errors.New("db not configured")
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

	subQuery := l.svcCtx.DB.WithContext(l.ctx).
		Table("software_manifests").
		Select("MAX(id) AS id").
		Where("project_id = ? AND software_id IN ?", req.ProjectId, softwareIds).
		Group("software_id")

	var manifests []model.SoftwareManifests
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("software_manifests").
		Joins("JOIN (?) AS latest ON latest.id = software_manifests.id", subQuery).
		Find(&manifests).Error; err != nil {
		return nil, err
	}

	bySoftwareId := make(map[int64]model.SoftwareManifests, len(manifests))
	for _, m := range manifests {
		bySoftwareId[int64(m.SoftwareId)] = m
	}

	items := make([]types.LatestSoftwareManifestItem, 0, len(softwareIds))
	for _, softwareId := range softwareIds {
		m, exists := bySoftwareId[softwareId]
		if !exists {
			items = append(items, types.LatestSoftwareManifestItem{
				SoftwareId: softwareId,
				HasRecord:  false,
			})
			continue
		}
		items = append(items, types.LatestSoftwareManifestItem{
			SoftwareId:            softwareId,
			HasRecord:             true,
			ManifestId:            int64(m.Id),
			ManifestFileId:        int64(m.ManifestFileId),
			ManifestFileVersionId: int64(m.ManifestFileVersionId),
			VersionNumber:         int64(m.VersionNumber),
			VersionDescription:    m.VersionDescription.String,
			CreatedBy:             int64(m.CreatedBy),
			CreatedAt:             m.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.LatestSoftwareManifestListResp{List: items}, nil
}

func parseSoftwareIds(v string) ([]int64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	parts := strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	out := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil || id <= 0 {
			return nil, errors.New("invalid software id")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}
