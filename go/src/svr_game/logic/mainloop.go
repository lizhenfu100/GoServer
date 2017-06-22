package logic

import (
	"time"

	"svr_game/logic/player"
)

func MainLoop() {
	// init func list
	player.InitSvrMailLst()

	timeNow, timeOld, time_elapse := time.Now().UnixNano()/int64(time.Millisecond), int64(0), 0
	for {
		timeOld = timeNow
		timeNow = time.Now().UnixNano() / int64(time.Millisecond)
		time_elapse = int(timeNow - timeOld)

		player.G_Auto_Write_DB.RunSevice(time_elapse)

		//TODO:zhoumf: 用类似ServiceList优化AFK检查
		player.CheckAFK()

		time.Sleep(1000 * time.Millisecond)
	}
}
