package common

type TRpcCsv struct {
	Name     string
	ID       int
	IsClient int // Client实现的rpc
	Comment  string
}

var G_RpcCsv map[string]*TRpcCsv = nil

func RpcToOpcode(rpc string) uint16 {
	if csv, ok := G_RpcCsv[rpc]; ok {
		return uint16(csv.ID)
	}
	print("!!!  " + rpc + " isn't in rpc.csv  !!!\n")
	return 0
}
