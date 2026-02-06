// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"io"
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/files"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetFileContentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetFileContentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := files.NewGetFileContentLogic(r.Context(), svcCtx)
		reader, contentType, err := l.GetFileContent(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer reader.Close()

		// 设置响应头
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		
		// 流式传输文件内容
		w.WriteHeader(http.StatusOK)
		io.Copy(w, reader)
	}
}
