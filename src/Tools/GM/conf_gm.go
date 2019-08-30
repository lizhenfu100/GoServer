package main

var (
	g_common = &TCommon{
		CenterList: []string{
			"http://3.17.67.102:7000",
			"http://3.16.163.125:7000",
			"http://18.221.148.84:7000",
			"http://18.223.109.103:7000",
			"http://18.216.113.27:7000",
		},
		SdkAddrs: []string{"",
			"http://120.78.152.152:7002", //1 北美
			"http://120.78.152.152:7002", //2 亚洲
			"http://120.78.152.152:7002", //3 欧洲
			"http://120.78.152.152:7002", //4 南美
			"http://120.78.152.152:7002", //5 中国华北
			"http://120.78.152.152:7002", //6 中国华南
		},
	}
	g_map = map[string]TemplateData{
		"HappyDiner": {
			GameName: "HappyDiner", //游戏名，以及对应的大区
			Logins: []TLogin{{}, //0位空，大区编号从1起始
				{Name: "北美", Addrs: []string{"http://3.17.67.102:7030"}},    //1 北美
				{Name: "亚洲", Addrs: []string{"http://13.229.215.168:7030"}}, //2 亚洲
				{Name: "欧洲", Addrs: []string{"http://18.185.80.202:7030"}},  //3 欧洲
				{Name: "南美", Addrs: []string{"http://54.94.211.178:7030"}},  //4 南美
				{Name: "华北", Addrs: []string{"http://39.96.196.250:7030"}},  //5 中国华北
				{Name: "华南", Addrs: []string{"http://47.106.35.74:7030"}},   //6 中国华南
			},
		},
		"SoulKnight": {
			GameName: "SoulKnight",
			Logins: []TLogin{{}, //0位留空 1起始
				{}, //1 北美
				{}, //2 亚洲
				{}, //3 欧洲
				{}, //4 南美
				{Name: "华北", Addrs: []string{"http://39.97.111.110:7030"}}, //5 中国华北
				{Name: "华南", Addrs: []string{"http://39.108.87.225:7030"}}, //6 中国华南
			},
		},
	}
)
