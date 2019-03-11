package test

import (
	"common"
	"fmt"
	"generate_out/rpc/enum"
	"http"
	"testing"
)
// go test -v ./src/svr_client/test/login_test.go
// go test -v ./src/svr_client/test/move_player_db_test.go
func Test_move_player_db(t *testing.T) { //须先登录游戏服
	accountId, newLoginId := uint32(6), 2
	gameAddr := http.Addr("127.0.0.1", 7040)
	playerRpc := http.NewPlayerRpc(gameAddr, accountId)
	playerRpc.CallRpc(enum.Rpc_game_move_player_db, func(buf *common.NetPack) {
		buf.WriteInt(newLoginId)
		buf.WriteString("IOS")
		buf.WriteString("80FD3EC9-D41A-4F41-B181-4EA58B0B4C33")
	}, func(recvBuf *common.NetPack) {
		e := recvBuf.ReadUInt16()
		fmt.Println("move_player_db: ", e)
	})
}
