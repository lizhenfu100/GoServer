package logic

import (
	"common"
	"common/file"
	"encoding/json"
	"fmt"
	"os"
	"shared_svr/svr_save/gm"
	"time"
)

// 敏感数据（如游戏进度）异动，记录历史存档
type TSensitive struct {
	GameSession int //进度，不同游戏含义不一
}

func (self *TSaveData) CheckSensitiveVal(newExtra string) {
	pNew, pOld := &TSensitive{}, &TSensitive{}
	json.Unmarshal(common.S2B(newExtra), pNew)
	json.Unmarshal(common.S2B(self.Extra), pOld)

	if pNew.GameSession < pOld.GameSession || gm.G_Backup.IsValid(self.Key) {
		dir := fmt.Sprintf("player/%s/", self.Key)
		name := time.Now().Format("20060102_150405") + ".save"
		if fi, e := file.CreateFile(dir, name, os.O_TRUNC|os.O_WRONLY); e == nil {
			fi.Write(self.Data)
			fi.Close()
		}
		file.DelExpired(dir, "", 30) //删除30天前的记录
	}
}
