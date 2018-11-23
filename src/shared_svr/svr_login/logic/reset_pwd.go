package logic

import (
	"common"
	"common/email"
	"common/sign"
	"fmt"
	"generate_out/err"
	"generate_out/rpc/enum"
	"net/http"
	"net/url"
	"netConfig"
	"strconv"
	"time"
)

func Http_ask_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	name := q.Get("name")
	passwd := q.Get("passwd")

	//FIXME:zhoumf: 限制同ip调用频率

	//! 创建回复
	ack := "失败，无此信息"
	defer func() {
		w.Write([]byte(ack))
	}()

	svrId := netConfig.HashCenterID(name)
	netConfig.CallRpcCenter(svrId, enum.Rpc_center_ask_reset_password, func(buf *common.NetPack) {
		buf.WriteString(name)
	}, func(recvBuf *common.NetPack) {
		if errCode := recvBuf.ReadUInt16(); errCode == err.Success {
			accountId := recvBuf.ReadUInt32()
			emailAddr := recvBuf.ReadString()
			//1、创建url
			httpAddr := fmt.Sprintf("http://%s:%d/reset_password", netConfig.G_Local_Meta.OutIP, netConfig.G_Local_Meta.Port())
			u, _ := url.Parse(httpAddr)
			q := u.Query()
			//2、写入参数
			q.Set("svrid", strconv.Itoa(svrId))
			q.Set("id", strconv.Itoa(int(accountId)))
			q.Set("pwd", passwd)
			flag := strconv.FormatInt(time.Now().Unix(), 10)
			q.Set("flag", flag)
			q.Set("sign", sign.CalcSign(passwd+flag))
			//3、生成完整url
			u.RawQuery = q.Encode()
			sUrl := u.String()

			email.SendMail("密码重置", emailAddr, sUrl)
			ack = "Ok: send verification to your email. Please check it."
		} else if errCode == err.Account_none {
			ack = "Error: Account_none"
		} else if errCode == err.Account_without_bind_info {
			ack = "Error: Account_without_bind_info"
		} else {
			ack = "Error: unknown"
		}
	})
}
func Http_reset_password(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	svrId, _ := strconv.Atoi(q.Get("svrid"))
	aid, _ := strconv.Atoi(q.Get("id"))
	passwd := q.Get("pwd")
	flag := q.Get("flag")
	timeFlag, _ := strconv.ParseInt(flag, 10, 0)

	//! 创建回复
	ack := "失败，无此信息"
	defer func() {
		w.Write([]byte(ack))
	}()

	if sign.CalcSign(passwd+flag) != q.Get("sign") {
		ack = "Error: sign failed"
	} else {
		if time.Now().Unix()-timeFlag > 3600 {
			ack = "Error: url expire"
		} else {
			netConfig.CallRpcCenter(svrId, enum.Rpc_center_reset_password, func(buf *common.NetPack) {
				buf.WriteUInt32(uint32(aid))
				buf.WriteString(passwd)
			}, func(recvBuf *common.NetPack) {
				errCode := recvBuf.ReadUInt16()
				switch errCode {
				case err.Success:
					ack = "Reset password ok"
				case err.Account_none:
					ack = "Error: Account_none"
				case err.Passwd_format_err:
					ack = "Error: Passwd_format_err"
				default:
					ack = "Error: unknown"
				}
			})
		}
	}
}
