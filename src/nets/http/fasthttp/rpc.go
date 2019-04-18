package fasthttp

import (
	http "github.com/valyala/fasthttp"
	mhttp "nets/http"
)

func _HandleRpc(ctx *http.RequestCtx) {
	ip := ""
	if mhttp.G_Intercept != nil {
		ip = ctx.RemoteIP().String()
	}
	mhttp.HandleRpc(ctx.Request.Body(), ctx, ip)
}

// ------------------------------------------------------------
//! player rpc
func RegHandlePlayerRpc(cb func(*http.RequestCtx)) {
	HandleFunc("/player_rpc", cb)
}
