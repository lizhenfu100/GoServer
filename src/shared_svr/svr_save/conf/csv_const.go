package conf

var Const struct {
	// 云存档
	MacBindMax         byte
	MacFreeBindMax     byte //可随意绑定的设备数，无时间间隔限制
	MacChangePeriod    int  //切换设备的周期，两周：3600*24*14
	MacUnbindPeriod    int  //解绑设备的周期，一周：3600*24*7
	RaiseBindCntPeriod int  //提升绑定次数上限，90天
}
