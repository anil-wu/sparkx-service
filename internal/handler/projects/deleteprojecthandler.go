// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package projects

import (
	"net/http"
	"strings"
	"strconv"

	"github.com/anil-wu/spark-x/internal/logic/projects"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteProjectHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// path: /api/v1/projects/:id
		idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		req := types.DeleteProjectReq{Id: id}
		l := projects.NewDeleteProjectLogic(r.Context(), svcCtx)
		resp, err := l.DeleteProject(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
