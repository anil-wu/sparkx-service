package previews

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/previews"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PreviewBuildPreuploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PreviewBuildPreuploadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := previews.NewPreviewBuildLogic(r.Context(), svcCtx)
		uploadUrl, contentType, storageKey, err := l.PreuploadBuildFile(
			req.BuildVersionId,
			req.Name,
			req.FileFormat,
			req.SizeBytes,
			req.Hash,
			req.ContentType,
		)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, &types.PreviewBuildPreuploadResp{
			UploadUrl:   uploadUrl,
			ContentType: contentType,
			StorageKey:  storageKey,
		})
	}
}
