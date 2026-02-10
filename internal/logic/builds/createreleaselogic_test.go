package builds

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/anil-wu/spark-x/internal/model"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
)

func TestCreateRelease_InvalidParam(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userId", json.Number("1"))
	l := NewCreateReleaseLogic(ctx, &svc.ServiceContext{})
	_, err := l.CreateRelease(&types.CreateReleaseReq{})
	if err != model.InputParamInvalid {
		t.Fatalf("expected %v, got %v", model.InputParamInvalid, err)
	}
}

func TestCreateRelease_DbNotConfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), "userId", json.Number("1"))
	l := NewCreateReleaseLogic(ctx, &svc.ServiceContext{})
	_, err := l.CreateRelease(&types.CreateReleaseReq{
		ProjectId:                   1,
		BuildVersionId:              1,
		ReleaseManifestFileId:       1,
		ReleaseManifestFileVersionId: 1,
		Name:                        "r1",
		Channel:                     "dev",
		Platform:                    "web",
	})
	if err == nil || !strings.Contains(err.Error(), "db not configured") {
		t.Fatalf("expected db not configured error, got %v", err)
	}
}
