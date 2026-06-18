package stdlib

import (
	"fmt"
	"time"
)

// DateTime utilities for Kylix

// TDateTime wraps Go's time.Time for Pascal-style date/time operations
type TDateTime struct {
	t time.Time
}

// Now returns the current date and time
func Now() *TDateTime {
	return &TDateTime{t: time.Now()}
}

// Today returns today's date (time set to midnight)
func Today() *TDateTime {
	now := time.Now()
	return &TDateTime{t: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())}
}

// MakeDate creates a date from year, month, day
func MakeDate(year, month, day int) *TDateTime {
	return &TDateTime{t: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)}
}

// MakeTime creates a datetime from components
func MakeTime(year, month, day, hour, minute, second int) *TDateTime {
	return &TDateTime{t: time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)}
}

// ParseDate parses a date string in common formats
func ParseDate(s string) (*TDateTime, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"Jan 2, 2006",
		"2 Jan 2006",
	}
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return &TDateTime{t: t}, nil
		}
	}
	return nil, fmt.Errorf("ParseDate: cannot parse '%s'", s)
}

// ParseDateTime parses a datetime string
func ParseDateTime(s string) (*TDateTime, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"01/02/2006 15:04:05",
		time.RFC3339,
	}
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return &TDateTime{t: t}, nil
		}
	}
	return nil, fmt.Errorf("ParseDateTime: cannot parse '%s'", s)
}

// Accessor methods

func (dt *TDateTime) Year() int      { return dt.t.Year() }
func (dt *TDateTime) Month() int     { return int(dt.t.Month()) }
func (dt *TDateTime) Day() int       { return dt.t.Day() }
func (dt *TDateTime) Hour() int      { return dt.t.Hour() }
func (dt *TDateTime) Minute() int    { return dt.t.Minute() }
func (dt *TDateTime) Second() int    { return dt.t.Second() }
func (dt *TDateTime) DayOfWeek() int { return int(dt.t.Weekday()) } // 0=Sunday
func (dt *TDateTime) DayOfYear() int { return dt.t.YearDay() }
func (dt *TDateTime) Unix() int64    { return dt.t.Unix() }

// Formatting methods

// Format returns the datetime formatted with the given layout
func (dt *TDateTime) Format(layout string) string {
	return dt.t.Format(layout)
}

// FormatDate returns the date in YYYY-MM-DD format
func (dt *TDateTime) FormatDate() string {
	return dt.t.Format("2006-01-02")
}

// FormatTime returns the time in HH:MM:SS format
func (dt *TDateTime) FormatTime() string {
	return dt.t.Format("15:04:05")
}

// FormatDateTime returns the datetime in YYYY-MM-DD HH:MM:SS format
func (dt *TDateTime) FormatDateTime() string {
	return dt.t.Format("2006-01-02 15:04:05")
}

// String returns the default string representation
func (dt *TDateTime) String() string {
	return dt.FormatDateTime()
}

// Arithmetic methods

// AddDays returns a new datetime with days added
func (dt *TDateTime) AddDays(days int) *TDateTime {
	return &TDateTime{t: dt.t.AddDate(0, 0, days)}
}

// AddMonths returns a new datetime with months added
func (dt *TDateTime) AddMonths(months int) *TDateTime {
	return &TDateTime{t: dt.t.AddDate(0, months, 0)}
}

// AddYears returns a new datetime with years added
func (dt *TDateTime) AddYears(years int) *TDateTime {
	return &TDateTime{t: dt.t.AddDate(years, 0, 0)}
}

// AddHours returns a new datetime with hours added
func (dt *TDateTime) AddHours(hours int) *TDateTime {
	return &TDateTime{t: dt.t.Add(time.Duration(hours) * time.Hour)}
}

// AddMinutes returns a new datetime with minutes added
func (dt *TDateTime) AddMinutes(minutes int) *TDateTime {
	return &TDateTime{t: dt.t.Add(time.Duration(minutes) * time.Minute)}
}

// AddSeconds returns a new datetime with seconds added
func (dt *TDateTime) AddSeconds(seconds int) *TDateTime {
	return &TDateTime{t: dt.t.Add(time.Duration(seconds) * time.Second)}
}

// Comparison methods

// DiffDays returns the difference in days between two datetimes
func (dt *TDateTime) DiffDays(other *TDateTime) int {
	diff := dt.t.Sub(other.t)
	return int(diff.Hours() / 24)
}

// DiffHours returns the difference in hours
func (dt *TDateTime) DiffHours(other *TDateTime) int {
	return int(dt.t.Sub(other.t).Hours())
}

// DiffSeconds returns the difference in seconds
func (dt *TDateTime) DiffSeconds(other *TDateTime) int64 {
	return int64(dt.t.Sub(other.t).Seconds())
}

// Before returns true if dt is before other
func (dt *TDateTime) Before(other *TDateTime) bool {
	return dt.t.Before(other.t)
}

// After returns true if dt is after other
func (dt *TDateTime) After(other *TDateTime) bool {
	return dt.t.After(other.t)
}

// Equal returns true if dt equals other
func (dt *TDateTime) Equal(other *TDateTime) bool {
	return dt.t.Equal(other.t)
}

// Utility methods

// IsWeekend returns true if the date is Saturday or Sunday
func (dt *TDateTime) IsWeekend() bool {
	day := dt.t.Weekday()
	return day == time.Saturday || day == time.Sunday
}

// IsLeapYear returns true if the year is a leap year
func (dt *TDateTime) IsLeapYear() bool {
	year := dt.t.Year()
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in the current month
func (dt *TDateTime) DaysInMonth() int {
	year, month, _ := dt.t.Date()
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// StartOfDay returns the start of the day (midnight)
func (dt *TDateTime) StartOfDay() *TDateTime {
	t := dt.t
	return &TDateTime{t: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())}
}

// EndOfDay returns the end of the day (23:59:59)
func (dt *TDateTime) EndOfDay() *TDateTime {
	t := dt.t
	return &TDateTime{t: time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())}
}

// DayName returns the name of the day (Monday, Tuesday, etc.)
func (dt *TDateTime) DayName() string {
	return dt.t.Weekday().String()
}

// MonthName returns the name of the month
func (dt *TDateTime) MonthName() string {
	return dt.t.Month().String()
}

// Sleep pauses execution for the given number of milliseconds
func Sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// GetTimestamp returns the current Unix timestamp in seconds
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// GetTimestampMs returns the current Unix timestamp in milliseconds
func GetTimestampMs() int64 {
	return time.Now().UnixMilli()
}
