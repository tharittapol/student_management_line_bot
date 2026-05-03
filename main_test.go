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
		"/อัพเดท แพรว 9/5/2569 13:00-15:00",
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

	response, handled, err := processStaffCommand("/คอนเฟิร์ม แพรว", store, loc)
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

func TestFormatWeeklyLessons(t *testing.T) {
	loc := testLocation(t)
	now := time.Date(2026, time.May, 4, 9, 0, 0, 0, loc)
	lessons := []StudentLesson{
		{
			Nickname:       "แพรว",
			FullName:       "แพรวา ศิริพงษ์",
			Course:         "English Foundation",
			TotalHours:     20,
			CompletedHours: 6,
			SessionHours:   2,
			NextStart:      time.Date(2026, time.May, 4, 18, 0, 0, 0, loc),
			NextEnd:        time.Date(2026, time.May, 4, 20, 0, 0, 0, loc),
			Confirmed:      false,
		},
		{
			Nickname:       "บอส",
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
			Nickname:  "นอกสัปดาห์",
			FullName:  "อย่าแสดง",
			Course:    "Mock",
			NextStart: time.Date(2026, time.May, 11, 9, 0, 0, 0, loc),
			NextEnd:   time.Date(2026, time.May, 11, 10, 0, 0, 0, loc),
		},
	}

	message := formatWeeklyLessons(lessons, now)
	if !strings.Contains(message, "📚 ตารางเรียนสัปดาห์นี้") {
		t.Fatalf("expected weekly header, got %q", message)
	}
	if !strings.Contains(message, "4-10 พ.ค. 2569") {
		t.Fatalf("expected week range, got %q", message)
	}
	if !strings.Contains(message, "⏳ แพรว") || !strings.Contains(message, "✅ บอส") {
		t.Fatalf("expected confirmation emojis per student, got %q", message)
	}
	if strings.Contains(message, "นอกสัปดาห์") {
		t.Fatalf("expected lessons outside the week to be hidden, got %q", message)
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
	if !strings.Contains(response, "📚 ตารางเรียนสัปดาห์นี้") {
		t.Fatalf("expected weekly lesson schedule response, got %q", response)
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
