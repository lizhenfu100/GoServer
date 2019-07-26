package email

var G_InvalidCsv = map[string]*invalidAddr{}

type invalidAddr struct {
	Addr string
}

func InCsvInvalid(addr string) bool {
	_, ok := G_InvalidCsv[addr]
	return ok
}
