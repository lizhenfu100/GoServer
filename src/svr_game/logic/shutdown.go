/***********************************************************************
* @ 安全的关服流程
* @ brief
	1、关闭网络服务
	2、所有在线玩家触发Logout
	3、等待 db_process 操作完毕
	4、最后自杀

* @ 凌晨6点更新
	、尽可能玩家少量在线：关服会踢所有玩家下线，瞬间大量写库

* @ author zhoumf
* @ date 2018-12-19
***********************************************************************/
package logic

import (
	"dbmgo"
	"fmt"
	"nets/http/http"
	"nets/tcp"
	"os"
	"svr_game/player"
)

func Shutdown() {
	fmt.Println("Shutdown Begin")
	tcp.CloseServer()
	http.CloseServer()

	player.G_players.Range(func(_, v interface{}) bool {
		v.(*player.TPlayer).Logout()
		return true
	})

	//关服任务，阻塞，等待任务都完成才能关服
	dbmgo.WaitStop()

	fmt.Println("Shutdown End")
	os.Exit(0)
}
