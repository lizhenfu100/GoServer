package web

import (
	"common/std"
	"shared_svr/svr_gm/conf"
)

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
			Logins: []TLogin{{Name: "China"}},
			pf_id:  conf.G_Platform,
		},
		"DungeonOfWeirdos": {
			Logins: []TLogin{{Name: "China", Addrs: []string{"http://52.82.37.128:7030"}}},
			pf_id:  conf.G_Platform,
		},
		"Soul5": {
			pf_id: []std.StrPair{{"Wechat", ""}},
		},
	}
)
