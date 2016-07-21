package activity

import (
	"fmt"
	"time"
)

//! 活动配置表
type TActivityCsv struct {
	ID           int
	BeginTime    int
	EndTime      int
	TimeType     int //! 活动时间 1->月 2->周 3->开服 4->指定日期
	AwardTime    int //! 活动奖励时间
	ActivityType int //! 活动套用模板
	AwardType    int //! 活动奖励套用模板
}

var G_ActivityCsv []TActivityCsv

func GetActivityCsvInfo(id int) *TActivityCsv {
	if id <= 0 || id >= len(G_ActivityCsv) {
		return nil
	}
	return &G_ActivityCsv[id]
}

func GetActivityNextBeginTime(activityID int) (beginTime int64, endTime int64) {
	csv := GetActivityCsvInfo(activityID)
	now := time.Now()
	if csv.TimeType == 1 { //! 按照月计算
		beginDate := time.Date(now.Year(), now.Month(), csv.BeginTime, 0, 0, 0, 0, now.Location())
		endDate := time.Date(now.Year(), now.Month(), csv.EndTime, 23, 59, 59, 59, now.Location())
		beginDate.AddDate(0, 1, 0)
		endDate.AddDate(0, 1, 0)
		beginTime = beginDate.Unix()
		endTime = endDate.Unix()
	} else if csv.TimeType == 2 { //! 按照日计算
		weekDay := int(now.Weekday())
		if weekDay == 0 {
			weekDay = 7
		}

		if csv.EndTime == 7 && csv.BeginTime == 1 { // 永久活动
			beginTime = 0
			endTime = 0
		} else {
			beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			beginDate = beginDate.AddDate(0, 0, csv.BeginTime-weekDay)
			beginDate.AddDate(0, 0, 7)
			endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, now.Location())
			endDate = endDate.AddDate(0, 0, csv.EndTime-weekDay)
			endDate.AddDate(0, 0, 7)
			beginTime = beginDate.Unix()
			endTime = endDate.Unix()
		}
	} else if csv.TimeType == 3 {
		return -1, -1 //! 开服活动无下次开启时间
	} else if csv.TimeType == 4 { //! 按照 月*100+日格式写 比如 310 = 3月10日
		day := csv.BeginTime % 100
		month := (csv.BeginTime - day) / 100
		if day < 1 || day > 31 || month < 1 || month > 12 {
			fmt.Printf("Invalid Activity BeginTime: %d", csv.BeginTime)
			return -1, -1
		}

		beginData := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location())
		beginData.AddDate(1, 0, 0)
		beginTime = beginData.Unix()

		day = csv.EndTime % 100
		month = (csv.EndTime - day) / 100
		if day < 1 || day > 31 || month < 1 || month > 12 {
			fmt.Printf("Invalid Activity BeginTime: %d", csv.BeginTime)
			return -1, -1
		}

		endData := time.Date(now.Year(), time.Month(month), day, 23, 59, 59, 59, now.Location())
		endData.AddDate(1, 0, 0)
		endTime = endData.Unix()
	}

	return beginTime, endTime
}
func GetActivityEndTime(activityID int) (beginTime int64, endTime int64) {
	csv := GetActivityCsvInfo(activityID)

	if csv.EndTime == 0 {
		return 0, 0
	}

	now := time.Now()
	if csv.TimeType == 1 { //! 按照月计算
		beginDate := time.Date(now.Year(), now.Month(), csv.BeginTime, 0, 0, 0, 0, now.Location())
		endDate := time.Date(now.Year(), now.Month(), csv.EndTime, 23, 59, 59, 59, now.Location())
		beginTime = beginDate.Unix()
		endTime = endDate.Unix()
	} else if csv.TimeType == 2 { //! 按照日计算
		weekDay := int(now.Weekday())
		if weekDay == 0 { //! 特殊处理周末
			weekDay = 7
		}

		if csv.EndTime == 7 && csv.BeginTime == 1 {
			//! 永久活动
			beginTime = 0
			endTime = 0
		} else {
			beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			beginDate = beginDate.AddDate(0, 0, csv.BeginTime-weekDay)
			beginTime = beginDate.Unix()

			endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, now.Location())
			endDate = endDate.AddDate(0, 0, csv.EndTime-weekDay)
			endTime = endDate.Unix()
		}
	} else if csv.TimeType == 3 { //! 按照开服时间计算
		openDay := GetOpenServerDay()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		beginDate = beginDate.AddDate(0, 0, -1*openDay)
		beginTime = beginDate.Unix()

		endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, now.Location())
		endDate = endDate.AddDate(0, 0, csv.BeginTime-openDay)
		endTime = endDate.Unix()
	} else if csv.TimeType == 4 { //! 按照 月*100+日格式写 比如 310 = 3月10日
		day := csv.BeginTime % 100
		month := (csv.BeginTime - day) / 100
		if day < 1 || day > 31 || month < 1 || month > 12 {
			fmt.Printf("Invalid Activity BeginTime: %d", csv.BeginTime)
			return -1, -1
		}

		beginData := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location())
		beginTime = beginData.Unix()

		day = csv.EndTime % 100
		month = (csv.EndTime - day) / 100
		if day < 1 || day > 31 || month < 1 || month > 12 {
			fmt.Printf("Invalid Activity BeginTime: %d", csv.BeginTime)
			return -1, -1
		}

		endData := time.Date(now.Year(), time.Month(month), day, 23, 59, 59, 59, now.Location())
		endTime = endData.Unix()

		//! 若今年活动时间已过,则时间变更为明年
		if endTime < now.Unix() {
			beginData.AddDate(1, 0, 0)
			beginTime = beginData.Unix()

			endData.AddDate(1, 0, 0)
			endTime = endData.Unix()
		}
	}

	return beginTime, endTime
}
func GetOpenServerDay() int {
	return 1
}
