// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package files

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anil-wu/spark-x/internal/logic/files"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListProjectFilesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
		idStr = strings.TrimSuffix(idStr, "/files")
		projectId, _ := strconv.ParseInt(idStr, 10, 64)
		req := types.ListProjectFilesReq{
			ProjectId: projectId,
			Page:      1,
			PageSize:  20,
		}
		l := files.NewListProjectFilesLogic(r.Context(), svcCtx)
		resp, err := l.ListProjectFiles(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
