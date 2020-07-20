package web

import (
	"encoding/json"
	"fmt"
	"nets/http"
)

// 参数：一组账号名 逗号分隔
func (self *TemplateData) CallGame(accounts string, cb func(gameAddr string, aids []uint32)) {
	u := fmt.Sprintf("%s/game_info?game=%s&v=%s",
		g_common.CenterList[0], self.GameName, accounts)
	if buf := http.Client.Get(u); buf != nil {
		var ret [][3]int // aid, loginId, gameId
		json.Unmarshal(buf, &ret)
		addrs := map[string][]uint32{}
		for _, v := range ret {
			if addr := self.addrGame(v[1], v[2]); addr != "" {
				addrs[addr] = append(addrs[addr], uint32(v[0]))
			}
		}
		for k, v := range addrs {
			cb(k, v)
		}
	}
}
