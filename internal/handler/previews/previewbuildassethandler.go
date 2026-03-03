package previews

import (
	"io"
	"net/http"
	"strconv"
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
		reader, contentType, contentLength, err := l.OpenAssetObject(req.BuildVersionId, requestedPath)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer func() { _ = reader.Close() }()

		w.Header().Set("Cache-Control", "no-store")
		if strings.TrimSpace(contentType) != "" {
			w.Header().Set("Content-Type", contentType)
		}
		if contentLength > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
		}
		_, _ = io.Copy(w, reader)
	}
}
