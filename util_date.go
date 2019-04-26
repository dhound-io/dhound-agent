package main

import (
	"strings"
	"time"
)

/*
	Formats:
	M    - month (1)
	MM   - month (01)
	MMM  - month (Jan)
	MMMM - month (January)
	D    - day (2)
	DD   - day (02)
	DDD  - day (Mon)
	DDDD - day (Monday)
	YY   - year (06)
	YYYY - year (2006)
  hh   - hours (15)
	mm   - minutes (04)
	ss   - seconds (05)
	AM/PM hours: 'h' followed by optional 'mm' and 'ss' followed by 'pm', e.g.
  hpm        - hours (03PM)
  h:mmpm     - hours:minutes (03:04PM)
  h:mm:sspm  - hours:minutes:seconds (03:04:05PM)
  Time zones: a time format followed by 'ZZZZ', 'ZZZ' or 'ZZ', e.g.
  hh:mm:ss ZZZZ (16:05:06 +0100)
  hh:mm:ss ZZZ  (16:05:06 CET)
	hh:mm:ss ZZ   (16:05:06 +01:00)
*/

type p struct{ find, subst string }

var Placeholder = []p{
	{"hh", "15"},
	{"h", "03"},
	{"mm", "04"},
	{"ss", "05"},
	{"MMMM", "January"},
	{"MMM", "Jan"},
	{"MM", "01"},
	{"M", "1"},
	{"pm", "PM"},
	{"ZZZZ", "-0700"},
	{"ZZZ", "MST"},
	{"ZZ", "Z07:00"},
	{"YYYY", "2006"},
	{"YY", "06"},
	{"DDDD", "Monday"},
	{"DDD", "Mon"},
	{"DD", "02"},
	{"D", "2"},
}

func init() {
}

func DateToCustomLong(time time.Time) (out int64) {
	result := time.UTC().Unix()
	return result
}

func CustomLongToTime(number int64) (out time.Time) {
	result := time.Unix(number, 0)
	return result
}

func replace(in string) (out string) {
	out = in
	for _, ph := range Placeholder {
		out = strings.Replace(out, ph.find, ph.subst, -1)
	}
	return
}

// smart format date, if year is not specified it will be calculated automatically
func ExYearParseDate(format string, value string, loc *time.Location) (time.Time, error) {
	if strings.Contains(format, "Y") == false {
		format = "YYYY " + format
		timeUtc := time.Now().UTC()
		value = Itoa(timeUtc.Year()) + " " + value
	}

	return time.ParseInLocation(replace(format), value, loc)
}

var (
	DefaultTimeFormat     = "hh:mm:ss"
	DefaultDateFormat     = "YYYY-MM-DD"
	DefaultDateTimeFormat = "YYYY-MM-DD hh:mm:ss"
)
