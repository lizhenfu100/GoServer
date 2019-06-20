package gm

import (
	"common"
	"encoding/json"
	"net/http"
	"strconv"
	"svr_game/player"
)

func Http_show_player(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id, _ := strconv.Atoi(q.Get("aid"))

	if ptr := player.FindWithDB(uint32(id)); ptr != nil {
		str, _ := json.MarshalIndent(ptr, "", "     ")
		w.Write(str)
	} else {
		w.Write(common.S2B("none player"))
	}
}
