// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package agents

import (
	"net/http"

	"github.com/anil-wu/spark-x/internal/logic/agents"
	"github.com/anil-wu/spark-x/internal/svc"
	"github.com/anil-wu/spark-x/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAvailableAgentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAgentReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := agents.NewGetAvailableAgentLogic(r.Context(), svcCtx)
		resp, err := l.GetAvailableAgent(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
