package taskhandler

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"time"
)

var WeekDayMap = map[string]int64{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    7,
}

const LayOutDate = "2006-01-02"

// Get the day of the week for the two dates entered
func GetWeekDay(startime, endtim string) (int64, int64) {
	startday, _ := time.Parse(LayOutDate, startime)
	endday, _ := time.Parse(LayOutDate, endtim)
	staweek_int := startday.Weekday().String()
	endweek_int := endday.Weekday().String()
	return WeekDayMap[staweek_int], WeekDayMap[endweek_int]
}

// String to timestamp
func StringToTimeStamp(strTime string) int64 {
	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation(LayOutDate, strTime, loc)
	if err != nil {
		logs.Error("StringToTimeStamp is abnormal：", err)
		return 0
	}
	unix_time := theTime.Unix()
	return unix_time
}

// Timestamp to string
func TimeStampToString(timeStp int64) string {
	datetime := time.Unix(timeStp, 0).Format(LayOutDate)
	return datetime
}

// Time converted into a list of days of the week
func ChangeToWeek(startime, endtim string) []map[string]string {
	staweek_int, endweek_int := GetWeekDay(startime, endtim)
	start_stamp := StringToTimeStamp(startime)
	end_stamp := StringToTimeStamp(endtim)
	fmt.Println("start_stamp==", start_stamp, "end_stamp==", end_stamp)
	var week_list = make([]map[string]string, 0)
	if (end_stamp-start_stamp)/604800 <= 1 && endweek_int-staweek_int >= 0 {
		if end_stamp-start_stamp < 604800 && endweek_int-staweek_int > 0 {
			one_map := map[string]string{}
			mon_one := TimeStampToString(start_stamp - (staweek_int-1)*86400)
			sun_one := TimeStampToString(start_stamp + (7-staweek_int)*86400)
			one_map["mon"] = mon_one
			one_map["sun"] = sun_one
			week_list = append(week_list, one_map)
			return week_list
		}
		one_map := map[string]string{}
		mon_one := TimeStampToString(start_stamp - (staweek_int-1)*86400)
		sun_one := TimeStampToString(start_stamp + (7-staweek_int)*86400)
		one_map["mon"] = mon_one
		one_map["sun"] = sun_one
		week_list = append(week_list, one_map)
		tow_map := map[string]string{}
		mon_tow := TimeStampToString(end_stamp - (endweek_int-1)*86400)
		sun_tow := TimeStampToString(end_stamp + (7-endweek_int)*86400)
		tow_map["mon"] = mon_tow
		tow_map["sun"] = sun_tow
		week_list = append(week_list, tow_map)
		return week_list
	}
	week_n := (end_stamp - start_stamp) / 604800
	one_map := map[string]string{}
	mon_one := TimeStampToString(start_stamp - (staweek_int-1)*86400)
	sun_one := TimeStampToString(start_stamp + (7-staweek_int)*86400)
	one_map["mon"] = mon_one
	one_map["sun"] = sun_one
	week_list = append(week_list, one_map)
	for i := 1; i <= int(week_n); i++ {
		week_map := map[string]string{}
		mon_day := TimeStampToString(start_stamp - (staweek_int-1)*86400 + int64(i)*604800)
		sun_day := TimeStampToString(start_stamp + (7-staweek_int)*86400 + int64(i)*604800)
		week_map["mon"] = mon_day
		week_map["sun"] = sun_day
		week_list = append(week_list, week_map)
	}
	if endweek_int-staweek_int >= 0 {
		return week_list
	}
	tow_map := map[string]string{}
	mon_tow := TimeStampToString(end_stamp - (endweek_int-1)*86400)
	sun_tow := TimeStampToString(end_stamp + (7-endweek_int)*86400)
	tow_map["mon"] = mon_tow
	tow_map["sun"] = sun_tow
	week_list = append(week_list, tow_map)
	return week_list
}
