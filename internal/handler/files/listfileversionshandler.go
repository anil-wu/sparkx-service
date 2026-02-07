// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/files"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListFileVersionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListFileVersionsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := files.NewListFileVersionsLogic(r.Context(), svcCtx)
		resp, err := l.ListFileVersions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
