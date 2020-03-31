package web

var (
	g_common = TCommon{
		CenterList: []string{
			"http://3.17.67.102:7000",
			"http://3.16.163.125:7000",
			"http://18.223.109.103:7000",
			"http://18.216.113.27:7000",
		},
		SdkAddrs: []string{
			"http://120.78.152.152:7002", //China
		},
	}
	g_map = map[string]TemplateData{
		"HappyDiner": {
			Logins: []TLogin{{Name: "ChinaSouth"}, {Name: "ChinaNorth"}},
			pf_id: []string{
				"Android,Official,GooglePlay,hykb,kkp,yyh,yyb", //好游快爆,酷酷跑,应用汇,应用宝
				"IOS", "4399", "9games",
				"bilibili", "coolpad",
				"huawei", "jinli", "lenovo", "meizu",
				"nubiya", "oppo", "vivo", "xiaomi",
			},
		},
		"SoulKnight": {
			Logins: []TLogin{{Name: "ChinaSouth"}, {Name: "ChinaNorth"}},
			pf_id: []string{
				"TapTap", "IOS",
				"4399_360", "4399_4399",
				"4399_4399hezi", "4399_4399jm",
				"4399_anzhi_dj", "4399_baidu_dj",
				"4399_biligame", "4399_dn_dj",
				"4399_douyu", "4399_duowan_mc",
				"4399_leap", "4399_letv",
				"4399_meiyou_dj", "4399_migu", "4399_muzhiwan",
				"4399_pps",
				"4399_samsungapp", "4399_snssdk", "4399_sogou",
				"4399_uc_dj",
				"4399_wdj_dj", "4399_wo", "4399_xiaomi",
				"4399_yxgame", "4399_yyb_dj", "4399_yyh_dj",
				"9games",
				"BlackBox", "CatClaw", "CoolPad", "GamePop", "Gourd",
				"HuaWei", "JinLi", "KuaiKan", "Lenovo",
				"LianYun001", "LianYun233", "LianYunCPS",
				"MeiZu", "OPPO", "VIVO", "Xiao7",
				"YYBShare", "YouPinWei", "Yyb",
			},
		},
		"zhmr": {
			Logins: []TLogin{
				{Name: "华南", Addrs: []string{"http://52.82.37.128:7030"}},
			},
			pf_id: []string{
				"TapTap", "IOS",
			},
		},
		//"DungeonOfWeirdos": {
		//	Logins: []TLogin{
		//		{Name: "华南", Addrs: []string{"http://52.82.109.217:7030"}},
		//	},
		//	Pf_id: []string{
		//		"TapTap", "IOS", "Android",
		//	},
		//},
	}
)
