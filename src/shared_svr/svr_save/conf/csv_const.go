package conf

//go:generate D:\server\bin\gen_conf.exe *csv conf
type csv struct {
	// 云存档
	MacBindMax      byte
	MacFreeBindMax  byte //可随意绑定的设备数，无时间间隔限制
	RaiseBindCntDay byte //提升绑定次数上限，90天
	MacChangePeriod int  //切换设备的周期，两周：3600*24*14
	MacUnbindPeriod int  //解绑设备的周期，一周：3600*24*7

	IpNew2Old map[string]string //新旧节点ip
}
