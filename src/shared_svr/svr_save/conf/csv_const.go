package conf

var Const struct {
	// 云存档
	MacBindMax      byte
	MacFreeBindMax  byte //可随意绑定的设备数，无时间间隔限制
	MacChangePeriod int  //切换设备的等待时间，一周：3600*24*7
}
