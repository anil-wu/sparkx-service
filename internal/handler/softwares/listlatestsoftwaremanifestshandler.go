package softwares

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/softwares"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListLatestSoftwareManifestsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListLatestSoftwareManifestsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := softwares.NewListLatestSoftwareManifestsLogic(r.Context(), svcCtx)
		resp, err := l.ListLatestSoftwareManifests(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
