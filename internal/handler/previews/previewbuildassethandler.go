package previews

import (
	"net/http"
	"strings"

	"github.com/anil-wu/spark-x/internal/logic/previews"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PreviewBuildAssetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PreviewBuildAssetReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		requestedPath := strings.TrimPrefix(strings.TrimSpace(req.AssetPath), "/")
		l := previews.NewPreviewBuildLogic(r.Context(), svcCtx)
		redirectUrl, err := l.GetAssetRedirectURL(req.BuildVersionId, requestedPath)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Location", redirectUrl)
		w.WriteHeader(http.StatusFound)
	}
}
