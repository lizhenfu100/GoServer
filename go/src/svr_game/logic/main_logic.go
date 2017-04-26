package logic

import (
	"time"

	"svr_game/logic/player"
)

func MainLoop() {
	// init func list
	player.InitSvrMailLst()

	for {
		player.G_auto_write_db.RunSevice()

		time.Sleep(30 * time.Millisecond)
	}
}
