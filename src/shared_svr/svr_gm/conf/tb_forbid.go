package conf

import "common/std"

const (
	Order_Mac = "mac_order" //订单sdk，封设备
	Save_Mac  = "mac_save"  //云存档，封设备
)

var (
	G_Platform = []std.StrPair{
		{"TapTap", ""}, {"IOS", ""}, {"IOSCN", ""}, {"GP", ""},
		{"CoolPad", "尚游_酷派"},
		{"Lenovo", "尚游_联想"}, {"MeiZu", "尚游_魅族"},
		{"HuaWei", "尚游_华为"}, {"JinLi", "尚游_金立"},
		{"OPPO", "尚游_Oppo"}, {"VIVO", "尚游_Vivo"},
		{"360", ""}, {"4399", ""}, {"4399hezi", "4399盒子"},
		{"Anzhi", "安智"}, {"Baidu", "百度"}, {"Bazaar", ""},
		{"Biligame", "b站"}, {"DnLe", "当乐单机"}, {"Douyu", "斗鱼"},
		{"GamePop", "好游快爆"}, {"Gourd", "葫芦侠"},
		{"HuaWeiOutseas", "华为海外"}, {"Leap", "微信leap"}, {"Letv", "乐视"},
		{"LianYun001", "酷酷跑"},
		{"LianYun233", "233游戏"},
		{"LianYunCPS", "咪咕CPS"},
		{"Meiyou", "魅游单机"}, {"Migu", "咪咕"}, {"Muzhiwan", "拇指玩"},
		{"Nubiya", "努比亚"},
		{"PPS", "PPS平台"}, {"Samsung", "三星"}, {"Sogou", "搜狗"},
		{"Toutiao", "今日头条"}, {"UC", "九游"},
		{"WDJ", "豌豆荚"}, {"Wo", "联通沃商店"},
		{"Xiao7", "小七游戏"}, {"Xiaomi", "小米"},
		{"Yixin", "易信"}, {"YYB", "应用宝"}, {"YYH", "应用汇"},
	}
)

func CheckPf(gameName, pf_id string) bool {
	switch gameName {
	case "SoulKnight", "HappyDiner":
		return true
	default:
		for _, v := range G_Platform {
			if v.K == pf_id {
				return true
			}
		}
		return false
	}
}
