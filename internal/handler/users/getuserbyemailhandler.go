// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package users

import (
	"net/http"
	"strings"

	"github.com/anil-wu/spark-x/internal/logic/users"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetUserByEmailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimPrefix(r.URL.Path, "/api/v1/users/email/")
		req := types.GetUserByEmailReq{Email: email}
		l := users.NewGetUserByEmailLogic(r.Context(), svcCtx)
		resp, err := l.GetUserByEmail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
