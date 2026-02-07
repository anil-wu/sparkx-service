// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package admin

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/admin"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewAdminProfileLogic(r.Context(), svcCtx)
		resp, err := l.AdminProfile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
