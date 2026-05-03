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
		"อัพเดทเวลาเรียน/แพรว/แพรวา ศิริพงษ์/English Foundation/วันเสาร์ 9 พฤษภาคม 2569 เวลา 13.00-15.00 น.",
		store,
		loc,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected update command to be handled")
	}
	if !strings.Contains(response, "อัพเดทเวลาเรียน") {
		t.Fatalf("expected update response, got %q", response)
	}
	if !strings.Contains(response, "แพรว / แพรวา ศิริพงษ์ / English Foundation") {
		t.Fatalf("expected student details in response, got %q", response)
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

	response, handled, err := processStaffCommand(
		"คอนเฟิร์มเวลาเรียน/แพรว/แพรวา ศิริพงษ์/English Foundation/2026-05-09 13:00-15:00/คอนเฟิร์ม",
		store,
		loc,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected confirm command to be handled")
	}
	if !strings.Contains(response, "/ คอนเฟิร์ม") {
		t.Fatalf("expected confirm response, got %q", response)
	}

	lesson := findLesson(t, store.ListLessons(), "แพรว", "English Foundation")
	if !lesson.Confirmed {
		t.Fatal("confirm command should mark lesson as confirmed")
	}
}

func TestFormatDailyLessons(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	if got := len(store.ListLessons()); got != 10 {
		t.Fatalf("expected 10 mock lessons, got %d", got)
	}

	message := formatDailyLessons(store.ListLessons(), time.Date(2026, time.May, 3, 9, 0, 0, 0, loc))
	if !strings.Contains(message, "แจ้งเวลาเรียนประจำวันที่ วันอาทิตย์ 3 พฤษภาคม 2569") {
		t.Fatalf("expected Thai daily header, got %q", message)
	}
	if !strings.Contains(message, "ชั่วโมงที่ 7-8") {
		t.Fatalf("expected next hour range, got %q", message)
	}
	if !strings.Contains(message, "รอคอนเฟิร์ม") {
		t.Fatalf("expected confirmation status, got %q", message)
	}
}

func TestProcessScheduleRequestCommand(t *testing.T) {
	loc := testLocation(t)
	store := NewMockLessonStore(loc)

	response, handled, err := processStaffCommand("/ตารางเรียน", store, loc)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("expected schedule request command to be handled")
	}
	if !strings.Contains(response, "แจ้งเวลาเรียนประจำวันที่") {
		t.Fatalf("expected daily lesson schedule response, got %q", response)
	}
	if !strings.Contains(response, "แพรว / แพรวา ศิริพงษ์ / English Foundation") {
		t.Fatalf("expected current lesson data in response, got %q", response)
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
