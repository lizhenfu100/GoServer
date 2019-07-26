/***********************************************************************
* @ fasthttp
* @ brief
    、fasthttp替代原生http，测试qps提升量
        · 外网机器
            · http       qps:2500 cpu:21%
            · fasthttp   qps:2100 cpu:7%
        · 本地机器
            · http       qps:3400
            · fasthttp   qps:4300

	、qps反而少了400……可能是内部协程池调度缘故，查看fasthttp参数配置

* @ author zhoumf
* @ date 2019-3-28
***********************************************************************/
package fasthttp

import (
	"common"
	"fmt"
	http "github.com/valyala/fasthttp"
	mhttp "nets/http"
)

var (
	_svr http.Server
	_map = map[string]http.RequestHandler{}
)

func NewHttpServer(port uint16, module string, svrId int) error {
	mhttp.InitSvr(module, svrId)

	_svr.Handler = func(ctx *http.RequestCtx) {
		if v, ok := _map[common.B2S(ctx.Path())]; ok {
			v(ctx)
		} else {
			ctx.NotFound()
		}
	}
	HandleFunc("/client_rpc", _HandleRpc)
	HandleFunc("/reg_to_svr", func(ctx *http.RequestCtx) {
		mhttp.Reg_to_svr(ctx, ctx.Request.Body())
	})
	return _svr.ListenAndServe(fmt.Sprintf(":%d", port))
}
func CloseServer() { _svr.Shutdown() }

func HandleFunc(path string, f http.RequestHandler) { _map[path] = f }

// ------------------------------------------------------------
// -- rpc
func _HandleRpc(ctx *http.RequestCtx) {
	//TODO: 检查是否需要recover()
	ip := ""
	if mhttp.G_Intercept != nil {
		ip = ctx.RemoteIP().String()
	}
	mhttp.HandleRpc(ctx.Request.Body(), ctx, ip)
}
func RegHandlePlayerRpc(cb func(*http.RequestCtx)) {
	HandleFunc("/player_rpc", cb)
}
