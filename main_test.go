package main

import (
	"strings"
	"testing"
	"time"
)

func testLocation(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestProcessUpdateCommand(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	response, handled, err := processStaffCommand(
		"/อัพเดท แพรว แพรวา 9/5/2569 13:00-15:00",
		store,
		loc,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected update command to be handled")
	}
	if !strings.Contains(response, "🔄 อัพเดทเวลาเรียน") {
		t.Fatalf("expected update response, got %q", response)
	}
	if !strings.Contains(response, "⏳ แพรว") {
		t.Fatalf("expected waiting confirmation emoji in response, got %q", response)
	}
	if !strings.Contains(response, "ชม.7-8") {
		t.Fatalf("expected compact hour range, got %q", response)
	}

	lesson := findLesson(t, store.ListLessons(), "แพรว", "English Foundation")
	if lesson.Confirmed {
		t.Fatal("update should reset confirmation status")
	}
	if lesson.SessionHours != 2 {
		t.Fatalf("expected 2 session hours, got %d", lesson.SessionHours)
	}
	if lesson.NextStart.Year() != 2026 || lesson.NextStart.Month() != time.May || lesson.NextStart.Day() != 9 {
		t.Fatalf("unexpected parsed start date: %s", lesson.NextStart)
	}
}

func TestProcessConfirmCommand(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	response, handled, err := processStaffCommand("/คอนเฟิร์ม แพรว แพรวา", store, loc)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected confirm command to be handled")
	}
	if !strings.Contains(response, "✅ คอนเฟิร์มเวลาเรียน") {
		t.Fatalf("expected confirm response, got %q", response)
	}

	lesson := findLesson(t, store.ListLessons(), "แพรว", "English Foundation")
	if !lesson.Confirmed {
		t.Fatal("confirm command should mark lesson as confirmed")
	}
}

func TestProcessUnconfirmCommand(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	_, _, err := processStaffCommand("/คอนเฟิร์ม แพรว แพรวา", store, loc)
	if err != nil {
		t.Fatal(err)
	}

	response, handled, err := processStaffCommand("/ไม่คอนเฟิร์ม แพรว แพรวา", store, loc)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected unconfirm command to be handled")
	}
	if !strings.Contains(response, "⏳ ไม่คอนเฟิร์มเวลาเรียน") {
		t.Fatalf("expected unconfirm response, got %q", response)
	}

	lesson := findLesson(t, store.ListLessons(), "แพรว", "English Foundation")
	if lesson.Confirmed {
		t.Fatal("unconfirm command should mark lesson as not confirmed")
	}
}

func TestAddStudentCommandIsRemoved(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	response, handled, err := processStaffCommand("/เพิ่มนักเรียน พลอย พลอยลดา/Little 3D รุ่นที่ 1/8", store, loc)
	if err != nil {
		t.Fatal(err)
	}
	if handled {
		t.Fatalf("expected add student command to be removed, got response %q", response)
	}
}

func TestScheduleYearParsing(t *testing.T) {
	loc := testLocation(t)

	start, _, ok := parseSchedule("9/5 13:00-15:00", loc)
	if !ok {
		t.Fatal("expected short schedule to parse")
	}
	if start.Year() != time.Now().In(loc).Year() {
		t.Fatalf("expected current year for short date, got %d", start.Year())
	}

	start, _, ok = parseSchedule("9/5/2570 13:00-15:00", loc)
	if !ok {
		t.Fatal("expected explicit year schedule to parse")
	}
	if start.Year() != 2027 {
		t.Fatalf("expected Buddhist year 2570 to become 2027, got %d", start.Year())
	}
}

func TestParseLineGroupIDs(t *testing.T) {
	groupIDs := parseLineGroupIDs(
		"C111,C222",
		"C222 C333",
		"your_line_group_id",
		"UuserShouldNotBeRoutineTarget",
	)

	want := []string{"C111", "C222", "C333"}
	if len(groupIDs) != len(want) {
		t.Fatalf("expected %d group IDs, got %d: %#v", len(want), len(groupIDs), groupIDs)
	}
	for i := range want {
		if groupIDs[i] != want[i] {
			t.Fatalf("expected groupIDs[%d]=%s, got %s", i, want[i], groupIDs[i])
		}
	}
}

func TestFormatWeeklyLessons(t *testing.T) {
	loc := testLocation(t)
	now := time.Date(2026, time.May, 6, 9, 0, 0, 0, loc)
	lessons := []StudentLesson{
		{
			Nickname:       "แพรว",
			FirstName:      "แพรวา",
			FullName:       "แพรวา ศิริพงษ์",
			Course:         "English Foundation",
			TotalHours:     20,
			CompletedHours: 6,
			SessionHours:   2,
			NextStart:      time.Date(2026, time.May, 6, 18, 0, 0, 0, loc),
			NextEnd:        time.Date(2026, time.May, 6, 20, 0, 0, 0, loc),
			Confirmed:      false,
		},
		{
			Nickname:       "บอส",
			FirstName:      "ธนากร",
			FullName:       "ธนากร ใจดี",
			Course:         "คณิตศาสตร์ ม.2",
			TotalHours:     24,
			CompletedHours: 4,
			SessionHours:   2,
			NextStart:      time.Date(2026, time.May, 9, 13, 0, 0, 0, loc),
			NextEnd:        time.Date(2026, time.May, 9, 15, 0, 0, 0, loc),
			Confirmed:      true,
		},
		{
			Nickname:  "นอกช่วง",
			FullName:  "อย่าแสดง",
			Course:    "Mock",
			NextStart: time.Date(2026, time.May, 13, 9, 0, 0, 0, loc),
			NextEnd:   time.Date(2026, time.May, 13, 10, 0, 0, 0, loc),
		},
	}

	message := formatWeeklyLessons(lessons, now)
	if !strings.Contains(message, "📚 ตารางเรียน 7 วันข้างหน้า") {
		t.Fatalf("expected rolling 7-day header, got %q", message)
	}
	if !strings.Contains(message, "6-12 พ.ค. 2569") {
		t.Fatalf("expected 7-day range from today, got %q", message)
	}
	if !strings.Contains(message, "⏳ แพรว") || !strings.Contains(message, "✅ บอส") {
		t.Fatalf("expected confirmation emojis per student, got %q", message)
	}
	if strings.Contains(message, "นอกช่วง") {
		t.Fatalf("expected lessons outside the 7-day range to be hidden, got %q", message)
	}
}

func TestProcessScheduleRequestCommand(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	if got := len(store.ListLessons()); got != 10 {
		t.Fatalf("expected 10 mock lessons, got %d", got)
	}

	response, handled, err := processStaffCommand("/ตารางเรียน", store, loc)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected schedule request command to be handled")
	}
	if !strings.Contains(response, "📚 ตารางเรียน 7 วันข้างหน้า") {
		t.Fatalf("expected rolling 7-day lesson schedule response, got %q", response)
	}
}

func TestProcessStudentScheduleRequestCommand(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	response, handled, err := processStaffCommand("/ข้อมูลนักเรียน", store, loc)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected student schedule request command to be handled")
	}
	if !strings.Contains(response, "👥 นักเรียนที่ยังเรียนไม่จบ") {
		t.Fatalf("expected active student list header, got %q", response)
	}
	if !strings.Contains(response, "แพรว - แพรวา") || !strings.Contains(response, "English Foundation") {
		t.Fatalf("expected compact student list, got %q", response)
	}
	if strings.Contains(response, "เรียนแล้ว:") || strings.Contains(response, "ถัดไป:") || strings.Contains(response, "ปกติ:") {
		t.Fatalf("expected individual student details to be removed, got %q", response)
	}
}

func TestSplitLongLineMessage(t *testing.T) {
	message := strings.Join([]string{
		strings.Repeat("ก", 12),
		strings.Repeat("ข", 12),
		strings.Repeat("ค", 12),
	}, "\n")

	parts := splitLongLineMessage(message, 20)
	if len(parts) != 3 {
		t.Fatalf("expected 3 split messages, got %d: %#v", len(parts), parts)
	}
	for _, part := range parts {
		if len([]rune(part)) > 20 {
			t.Fatalf("message part exceeds max length: %q", part)
		}
	}
}

func findLesson(t *testing.T, lessons []StudentLesson, nickname string, course string) StudentLesson {
	t.Helper()
	for _, lesson := range lessons {
		if lesson.Nickname == nickname && lesson.Course == course {
			return lesson
		}
	}
	t.Fatalf("lesson not found: %s / %s", nickname, course)
	return StudentLesson{}
}
