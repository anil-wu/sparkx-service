// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package builds

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/builds"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListBuildVersionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListBuildVersionsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := builds.NewListBuildVersionsLogic(r.Context(), svcCtx)
		resp, err := l.ListBuildVersions(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
