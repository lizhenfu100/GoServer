// Generated by GoServer/src/generate
// Don't edit !
package rpc
import (
	"nets"
	"generate_out/rpc/enum"
	"shared_svr/zookeeper/logic"
	
)
func init() {
	
		nets.RegTcpRpc(map[uint16]nets.TcpRpc{
			enum.Rpc_net_error: logic.Rpc_net_error,
			enum.Rpc_zoo_register: logic.Rpc_zoo_register,
			
		})
	
	
	
	
}
