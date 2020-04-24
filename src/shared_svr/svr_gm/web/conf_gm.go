package web

import "common/std"

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
	G_Platform = []std.StrPair{
		{"TapTap", ""}, {"IOS", ""}, {"IOS_CN", ""}, {"GP", ""},
		{"SY_Coolpad", "尚游_酷派"},
		{"SY_Lenovo", "尚游_联想"}, {"SY_Meizu", "尚游_魅族"},
		{"SY_Huawei", "尚游_华为"}, {"SY_Jinli", "尚游_金立"},
		{"SY_Oppo", "尚游_Oppo"}, {"SY_Vivo", "尚游_Vivo"},
		{"360", ""}, {"4399", ""}, {"4399hezi", "4399盒子"},
		{"Anzhi", "安智"}, {"Bazaar", ""}, {"Baidu", "百度"},
		{"Biligame", "b站"}, {"Douyu", "斗鱼"},
		{"Gourd", "葫芦侠"}, {"Gamepop", "好游快爆"},
		{"Huawei_Overseas", "华为海外"}, {"Leap", "微信leap"},
		{"Lianyun001_kkp", "酷酷跑"},
		{"LianYun233", "233游戏"},
		{"LianYunCPS", "咪咕CPS"},
		{"Meiyou", "魅游单机"}, {"Migu", "咪咕"}, {"Muzhiwan", "拇指玩"},
		{"Nubiya", "努比亚"},
		{"PPS", "PPS平台"}, {"Samsung", "三星"}, {"Sogou", "搜狗"},
		{"Toutiao", "今日头条"}, {"UC", "九游"},
		{"WDJ", "豌豆荚"}, {"Wo", "联通沃商店"},
		{"Xiao7", "小七游戏"}, {"XiaoMi", "小米"},
		{"Yixin", "易信"}, {"YYB", "应用宝"}, {"YYH", "应用汇"},
	}
	g_map = map[string]TemplateData{
		"HappyDiner": {
			Logins: []TLogin{{Name: "ChinaSouth"}, {Name: "ChinaNorth"}},
			pf_id: []std.StrPair{
				// 存档互通，礼包码不互通的~囧~共用首个渠道名
				{"Android,Official,GooglePlay,hykb,kkp,yyh,yyb", "官网,谷歌,好游快爆,酷酷跑,应用汇,应用宝"},
				{"IOS", ""}, {"4399", ""},
				{"coolpad", "尚游_酷派"}, {"huawei", "尚游_华为"}, {"jinli", "尚游_金立"},
				{"lenovo", "尚游_联想"}, {"meizu", "尚游_魅族"},
				{"oppo", "尚游_Oppo"}, {"vivo", "尚游_Vivo"},
				{"9games", "九游"}, {"bilibili", "B站"},
				{"nubiya", "努比亚"}, {"xiaomi", "小米"},
			},
		},
		"SoulKnight": {
			Logins: []TLogin{{Name: "ChinaSouth"}, {Name: "ChinaNorth"}},
			pf_id: []std.StrPair{
				{"TapTap", ""}, {"IOS", ""},
				{"4399_360", ""}, {"4399_4399", ""},
				{"4399_4399hezi", "4399盒子"}, {"4399_4399jm", "4399静默包"},
				{"4399_anzhi_dj", "安智"}, {"4399_baidu_dj", "百度"},
				{"4399_biligame", "B站"}, {"4399_dn_dj", "当乐单机"},
				{"4399_douyu", "斗鱼"}, {"4399_duowan_mc", ""},
				{"4399_leap", "微信leap"}, {"4399_letv", "乐视"},
				{"4399_meiyou_dj", "魅游单机"}, {"4399_migu", "咪咕"},
				{"4399_muzhiwan", "拇指玩"}, {"4399_pps", "PPS平台"},
				{"4399_samsungapp", "三星"}, {"4399_snssdk", "今日头条"},
				{"4399_sogou", "搜狗"}, {"4399_uc_dj", "九游"},
				{"4399_wdj_dj", "豌豆荚"}, {"4399_wo", "联通沃商店"}, {"4399_xiaomi", "小米"},
				{"4399_yxgame", "易信"}, {"4399_yyb_dj", "应用宝"}, {"4399_yyh_dj", "应用汇"},
				{"BlackBox", ""}, {"CatClaw", ""}, {"CoolPad", "尚游_酷派"},
				{"GamePop", "好游快爆"}, {"Gourd", "葫芦侠"},
				{"HuaWei", "尚游_华为"}, {"JinLi", "尚游_金立"}, {"KuaiKan", ""},
				{"LianYun001", "酷酷跑"}, {"LianYun233", "233游戏"}, {"LianYunCPS", "咪咕CPS"},
				{"Lenovo", "尚游_联想"},
				{"MeiZu", "尚游_魅族"}, {"OPPO", "尚游_Oppo"}, {"VIVO", "尚游_Vivo"},
				{"Xiao7", ""}, {"YYBShare", ""}, {"YouPinWei", ""},
			},
		},
		"zhmr": {
			Logins: []TLogin{
				{Name: "华南", Addrs: []string{"http://52.82.37.128:7030"}},
			},
			pf_id: G_Platform,
		},
		//"DungeonOfWeirdos": {
		//	Logins: []TLogin{
		//		{Name: "华南", Addrs: []string{"http://52.82.109.217:7030"}},
		//	},
		//	Pf_id: G_Platform,
		//},
	}
)
