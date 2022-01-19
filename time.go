package tools

import "time"

func FromWeekOfYearToMondayAndSunday(weeks int, timeLoc *time.Location) (monday, sunday time.Time) {
	// 计算原理：
	// 根据当年的日期计算出当年的1月1日，根据1月1日所在的周数看看是否是每年的第一周
	// 如果不是每年的第一周，计算下一周的周一。下一周的周一即每年的第一周。再用总周数求出总周数开始的周一和周日
	now := time.Now()
	baseDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, timeLoc)
	year, week := baseDate.ISOWeek()
	wd := int(baseDate.Weekday())
	if year != now.Year() && week > 1 {
		if wd == 0 {
			baseDate = baseDate.AddDate(0, 0, 1)
		} else {
			baseDate = baseDate.AddDate(0, 0, 7-wd+1)
		}
	} else if wd != 1 {
		baseDate = baseDate.AddDate(0, 0, -wd+1)
	}

	monday = baseDate.Add(time.Duration(weeks-1) * 7 * 24 * time.Hour)
	sunday = monday.AddDate(0, 0, 6)
	return
}

func CalcWeeks(ts time.Time, timeLoc *time.Location) (year, weeks int) {
	return ts.In(timeLoc).ISOWeek()
}
