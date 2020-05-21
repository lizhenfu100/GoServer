package logic

import (
	"common"
	"dbmgo"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"netConfig/meta"
	"text/template"
)

const KDBHash = "hash" //GM页面编辑列表，不在列表的客户端包体，不予服务

func Rpc_gm_client_hash(req, ack *common.NetPack, _ common.Conn) {
	v := req.ReadString()
	if ok, e := dbmgo.Find(KDBHash, "_id", v, &bson.M{}); e == nil {
		ack.WriteBool(ok)
	}
	ack.WriteBool(true)
}

// ------------------------------------------------------------
func Http_client_hash(w http.ResponseWriter, r *http.Request) {
	var codes []string
	var vs []bson.M
	dbmgo.FindAll(KDBHash, nil, &vs)
	for _, v := range vs {
		codes = append(codes, v["_id"].(string))
	}
	addr := fmt.Sprintf("http://%s:%d/client_hash_add", meta.G_Local.OutIP, meta.G_Local.HttpPort)
	tmp := hashAck{codes, addr}
	t, _ := template.New("").Parse(kTemplate)
	t.Execute(w, &tmp)
}
func Http_client_hash_add(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	dbmgo.Insert(KDBHash, bson.M{"_id": q.Get("hash")})
}

type hashAck struct {
	Code []string
	Addr string
}

const kTemplate = `<html>
<body bgcolor="white">
	<form action="{{$.Addr}}" method="get"><br>
		<input type="text" name="hash">
		<input type="submit" value="添加">
	</form>
    <table border="1" cellpadding="3" cellspacing="0">
{{range $_, $v := .Code}}
        <tr>
            <td>{{$v}}</td>
        </tr>
{{end}}
    </table>
</body>
</html>`
