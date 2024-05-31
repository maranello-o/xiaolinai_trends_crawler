package main

import (
	"fmt"
	"strings"
	"time"
)

// 解析文章割裂的日期和时间到秒时间戳，日期格式为 2024-05-25，时间格式为 21:41:09
func parseSeparateTime(dateStr, timeStr string) (int32, error) {
	timeStr = strings.TrimSpace(timeStr)
	dateTimeStr := fmt.Sprintf("%s %s", dateStr, timeStr)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", dateTimeStr, time.Local)
	if err != nil {
		return 0, err
	}
	return int32(t.Unix()), nil
}

// 解析时间字符串到秒时间戳，时间格式为 2024-05-25 21:41
func parseTime(timeStr string) (int32, error) {
	t, err := time.ParseInLocation("2006-01-02 15:04", timeStr, time.Local)
	if err != nil {
		return 0, err
	}
	return int32(t.Unix()), nil
}
