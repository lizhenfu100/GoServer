package msg

type CodeInfo_req struct { //client上报的第三方授权码
	Code string `json:"code"`
}
type CodeInfo_ack struct {
	Retcode_ack
	Third_account string `json:"third_account"`
}