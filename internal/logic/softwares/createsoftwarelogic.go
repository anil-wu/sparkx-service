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
)

type CreateSoftwareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSoftwareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSoftwareLogic {
	return &CreateSoftwareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSoftwareLogic) CreateSoftware(req *types.CreateSoftwareReq) (resp *types.CreateSoftwareResp, err error) {
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

	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, model.InputParamInvalid
	}
	if req.ProjectId < 0 || (!isAdmin && req.ProjectId <= 0) {
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

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "archived" {
		return nil, model.InputParamInvalid
	}

	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	sw := &model.Softwares{
		ProjectId:       uint64(req.ProjectId),
		Name:            strings.TrimSpace(req.Name),
		Description:     sql.NullString{String: req.Description, Valid: strings.TrimSpace(req.Description) != ""},
		TemplateId:      uint64(req.TemplateId),
		TechnologyStack: strings.TrimSpace(req.TechnologyStack),
		Status:          status,
		CreatedBy:       uint64(userId),
	}
	if err := tx.Create(sw).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	tx = nil

	return &types.CreateSoftwareResp{
		SoftwareId:      int64(sw.Id),
		ProjectId:       int64(sw.ProjectId),
		Name:            sw.Name,
		Description:     sw.Description.String,
		TemplateId:      int64(sw.TemplateId),
		TechnologyStack: sw.TechnologyStack,
		Status:          sw.Status,
		CreatedBy:       int64(sw.CreatedBy),
		CreatedAt:       sw.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
