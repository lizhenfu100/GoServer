package player

import (
	"common"
	"dbmgo"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

const KPlayerArgs = "PlayerArgs"

func Rpc_game_player_agrs_set(req, ack *common.NetPack, conn common.Conn) {
	uid := req.ReadString()
	key := req.ReadString()
	val := req.LeftBuf()
	dbmgo.UpsertIdSync(KPlayerArgs, uid, bson.M{"$set": bson.M{key: val}})
}
func Rpc_game_player_agrs_get(req, ack *common.NetPack, conn common.Conn) {
	uid := req.ReadString()
	key := req.ReadString()
	m := bson.M{"_id": 0, key: 1}
	dbmgo.DB().C(KPlayerArgs).Find(bson.D{{"_id", uid}}).Select(m).One(&m)
	for _, k := range strings.Split(key, ".") {
		v := m[k]
		if p, ok := v.(bson.M); ok {
			m = p
		} else {
			ack.WriteBuf(v.([]byte))
			break
		}
	}
}
