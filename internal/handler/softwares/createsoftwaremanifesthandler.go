// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package softwares

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/softwares"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateSoftwareManifestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateSoftwareManifestReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := softwares.NewCreateSoftwareManifestLogic(r.Context(), svcCtx)
		resp, err := l.CreateSoftwareManifest(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
