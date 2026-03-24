package valueobject

import (
	"fmt"
	"time"
)

// TimeSlot 时间段值对象
// 表示一个连续的时间段，包含起始时间和结束时间
type TimeSlot struct {
	StartTime time.Time `json:"start_time"` // 起始时间
	EndTime   time.Time `json:"end_time"`   // 结束时间
}

// NewTimeSlot 创建新的时间段
func NewTimeSlot(startTime, endTime time.Time) (*TimeSlot, error) {
	if startTime.After(endTime) {
		return nil, fmt.Errorf("start time cannot be after end time: %v > %v", startTime, endTime)
	}
	if startTime.Equal(endTime) {
		return nil, fmt.Errorf("start time and end time cannot be equal")
	}
	return &TimeSlot{
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// Duration 返回时间段的持续时间
func (ts *TimeSlot) Duration() time.Duration {
	return ts.EndTime.Sub(ts.StartTime)
}

// DurationMinutes 返回时间段的持续时间（分钟）
func (ts *TimeSlot) DurationMinutes() int {
	return int(ts.Duration().Minutes())
}

// DurationHours 返回时间段的持续时间（小时）
func (ts *TimeSlot) DurationHours() float64 {
	return ts.Duration().Hours()
}

// Equals 判断两个时间段是否相等
func (ts *TimeSlot) Equals(other *TimeSlot) bool {
	if other == nil {
		return false
	}
	return ts.StartTime.Equal(other.StartTime) && ts.EndTime.Equal(other.EndTime)
}

// Contains 检查给定的时间点是否在时间段内
func (ts *TimeSlot) Contains(t time.Time) bool {
	return !t.Before(ts.StartTime) && !t.After(ts.EndTime)
}

// ContainsAll 检查给定的时间段是否完全包含在当前时间段内
func (ts *TimeSlot) ContainsAll(other *TimeSlot) bool {
	if other == nil {
		return false
	}
	return !other.StartTime.Before(ts.StartTime) && !other.EndTime.After(ts.EndTime)
}

// Overlaps 检查两个时间段是否有重叠
func (ts *TimeSlot) Overlaps(other *TimeSlot) bool {
	if other == nil {
		return false
	}
	return ts.StartTime.Before(other.EndTime) && other.StartTime.Before(ts.EndTime)
}

// OverlapDuration 计算两个时间段重叠的持续时间
func (ts *TimeSlot) OverlapDuration(other *TimeSlot) time.Duration {
	if !ts.Overlaps(other) {
		return 0
	}

	var overlapStart, overlapEnd time.Time

	if ts.StartTime.After(other.StartTime) {
		overlapStart = ts.StartTime
	} else {
		overlapStart = other.StartTime
	}

	if ts.EndTime.Before(other.EndTime) {
		overlapEnd = ts.EndTime
	} else {
		overlapEnd = other.EndTime
	}

	return overlapEnd.Sub(overlapStart)
}

// Adjacent 检查两个时间段是否相邻（一个的结束时间等于另一个的开始时间）
func (ts *TimeSlot) Adjacent(other *TimeSlot) bool {
	if other == nil {
		return false
	}
	return ts.EndTime.Equal(other.StartTime) || other.EndTime.Equal(ts.StartTime)
}

// Before 检查当前时间段是否在另一个时间段之前
func (ts *TimeSlot) Before(other *TimeSlot) bool {
	if other == nil {
		return false
	}
	return ts.EndTime.Before(other.StartTime)
}

// After 检查当前时间段是否在另一个时间段之后
func (ts *TimeSlot) After(other *TimeSlot) bool {
	if other == nil {
		return false
	}
	return ts.StartTime.After(other.EndTime)
}

// Merge 合并两个相邻或重叠的时间段，如果不相邻或重叠则返回错误
func (ts *TimeSlot) Merge(other *TimeSlot) (*TimeSlot, error) {
	if other == nil {
		return ts, nil
	}

	if !ts.Overlaps(other) && !ts.Adjacent(other) {
		return nil, fmt.Errorf("time slots do not overlap or adjacent")
	}

	var startTime, endTime time.Time

	if ts.StartTime.Before(other.StartTime) {
		startTime = ts.StartTime
	} else {
		startTime = other.StartTime
	}

	if ts.EndTime.After(other.EndTime) {
		endTime = ts.EndTime
	} else {
		endTime = other.EndTime
	}

	return NewTimeSlot(startTime, endTime)
}

// Intersect 计算两个时间段的交集
func (ts *TimeSlot) Intersect(other *TimeSlot) (*TimeSlot, error) {
	if !ts.Overlaps(other) {
		return nil, fmt.Errorf("time slots do not overlap")
	}

	var startTime, endTime time.Time

	if ts.StartTime.After(other.StartTime) {
		startTime = ts.StartTime
	} else {
		startTime = other.StartTime
	}

	if ts.EndTime.Before(other.EndTime) {
		endTime = ts.EndTime
	} else {
		endTime = other.EndTime
	}

	return NewTimeSlot(startTime, endTime)
}

// Split 将时间段分割成两部分
func (ts *TimeSlot) Split(t time.Time) (*TimeSlot, *TimeSlot, error) {
	if !ts.Contains(t) {
		return nil, nil, fmt.Errorf("time point is not within the time slot")
	}

	if t.Equal(ts.StartTime) || t.Equal(ts.EndTime) {
		return nil, nil, fmt.Errorf("cannot split at start or end time")
	}

	first, _ := NewTimeSlot(ts.StartTime, t)
	second, _ := NewTimeSlot(t, ts.EndTime)

	return first, second, nil
}

// Shift 将时间段平移指定的时长
func (ts *TimeSlot) Shift(duration time.Duration) *TimeSlot {
	return &TimeSlot{
		StartTime: ts.StartTime.Add(duration),
		EndTime:   ts.EndTime.Add(duration),
	}
}

// Extend 延长时间段的起始或结束时间
func (ts *TimeSlot) Extend(startDuration, endDuration time.Duration) *TimeSlot {
	return &TimeSlot{
		StartTime: ts.StartTime.Add(-startDuration),
		EndTime:   ts.EndTime.Add(endDuration),
	}
}

// String 返回时间段的字符串表示
func (ts *TimeSlot) String() string {
	return fmt.Sprintf("TimeSlot(%s - %s, duration: %v)",
		ts.StartTime.Format("2006-01-02 15:04:05"),
		ts.EndTime.Format("2006-01-02 15:04:05"),
		ts.Duration())
}

// IsToday 检查时间段是否在今天
func (ts *TimeSlot) IsToday() bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	return !ts.EndTime.Before(today) && ts.StartTime.Before(tomorrow)
}

// IsPast 检查时间段是否已过期
func (ts *TimeSlot) IsPast() bool {
	return ts.EndTime.Before(time.Now())
}

// IsFuture 检查时间段是否在未来
func (ts *TimeSlot) IsFuture() bool {
	return ts.StartTime.After(time.Now())
}

// IsHappening 检查时间段是否正在进行中
func (ts *TimeSlot) IsHappening() bool {
	now := time.Now()
	return ts.Contains(now)
}
