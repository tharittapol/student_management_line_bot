package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	classOverviewSheet  = "Overview"
	classTimetableSheet = "ตารางเรียน"
	classWeeklySheet    = "สัปดาห์นี้"
)

type ClassScheduleSheetsLessonStore struct {
	client *GoogleSheetsClient
	loc    *time.Location
	mu     sync.Mutex
}

type overviewRecord struct {
	rowNumber      int
	FullName       string
	Nickname       string
	Course         string
	Level          string
	Teacher        string
	Weekday        string
	TimeText       string
	StartDateText  string
	TotalHours     int
	CompletedHours int
	RemainingHours int
	Status         string
	ParentName     string
	ParentPhone    string
}

type weeklyRecord struct {
	rowNumber      int
	DateText       string
	DayText        string
	TimeText       string
	DurationText   string
	FullName       string
	Nickname       string
	Course         string
	Level          string
	Teacher        string
	ParentName     string
	ParentPhone    string
	TotalHours     int
	CompletedHours int
	Confirmed      bool
	LearningStatus string
	AvailableSlots []string
}

func NewClassScheduleSheetsLessonStore(spreadsheetID string, loc *time.Location) (*ClassScheduleSheetsLessonStore, error) {
	client, err := NewGoogleSheetsClientFromEnv(spreadsheetID)
	if err != nil {
		return nil, err
	}
	store := &ClassScheduleSheetsLessonStore{client: client, loc: loc}
	if !strings.EqualFold(os.Getenv("GOOGLE_SHEETS_INIT_SCHEMA"), "false") {
		if err := store.EnsureSchema(context.Background()); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *ClassScheduleSheetsLessonStore) EnsureSchema(ctx context.Context) error {
	titles, err := s.client.spreadsheetTitles(ctx)
	if err != nil {
		return err
	}
	for _, title := range []string{classOverviewSheet, classTimetableSheet, classWeeklySheet} {
		if !titles[title] {
			return fmt.Errorf("missing required Google Sheet tab %q; sync the Class schedule sheet before starting the bot", title)
		}
	}
	return nil
}

func (s *ClassScheduleSheetsLessonStore) ListLessons() []StudentLesson {
	weekly, err := s.loadWeekly(context.Background())
	if err != nil {
		log.Println("list class schedule lessons error:", err)
		return nil
	}
	var lessons []StudentLesson
	for _, record := range weekly {
		start, end, ok := weeklyRecordTime(record, s.loc)
		if !ok {
			continue
		}
		lessons = append(lessons, record.toStudentLesson(start, end, s.loc))
	}
	sort.Slice(lessons, func(i, j int) bool {
		return lessons[i].NextStart.Before(lessons[j].NextStart)
	})
	return lessons
}

func (s *ClassScheduleSheetsLessonStore) ListStudentSchedules() []StudentScheduleSummary {
	summaries, err := s.findStudentSchedules(context.Background(), "", "")
	if err != nil {
		log.Println("list class schedule summaries error:", err)
		return nil
	}
	return summaries
}

func (s *ClassScheduleSheetsLessonStore) FindStudentSchedules(nickname, firstName string) ([]StudentScheduleSummary, error) {
	summaries, err := s.findStudentSchedules(context.Background(), nickname, firstName)
	if err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return nil, fmt.Errorf("ไม่พบนักเรียน: %s / %s", nickname, firstName)
	}
	return summaries, nil
}

func (s *ClassScheduleSheetsLessonStore) findStudentSchedules(ctx context.Context, nickname, firstName string) ([]StudentScheduleSummary, error) {
	overview, err := s.loadOverview(ctx)
	if err != nil {
		return nil, err
	}
	weekly, err := s.loadWeekly(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().In(s.loc)

	var summaries []StudentScheduleSummary
	for _, record := range overview {
		if !isUnfinishedOverviewRecord(record) {
			continue
		}
		if strings.TrimSpace(nickname) != "" && !strings.EqualFold(cleanClassText(record.Nickname), cleanClassText(nickname)) {
			continue
		}
		if strings.TrimSpace(firstName) != "" && !strings.EqualFold(classFirstName(record.FullName), cleanClassText(firstName)) {
			continue
		}
		summaries = append(summaries, StudentScheduleSummary{
			Nickname:        record.Nickname,
			FirstName:       classFirstName(record.FullName),
			FullName:        record.FullName,
			Course:          classCourseName(record.Course, record.Level),
			TotalHours:      record.TotalHours,
			CompletedHours:  record.CompletedHours,
			DefaultSchedule: cleanClassText(record.Weekday + " " + record.TimeText),
			PastLessons:     recordPastLessonsText(record, weekly, s.loc, now),
			NextLessons:     recordNextLessonsText(record, weekly, s.loc, now),
			ScheduleNotes:   record.Status,
			ParentName:      record.ParentName,
			ParentPhone:     record.ParentPhone,
		})
	}
	sortStudentScheduleSummaries(summaries)
	return summaries, nil
}

func (s *ClassScheduleSheetsLessonStore) AddStudent(nickname, firstName, course string, totalHours int, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname = cleanClassText(nickname)
	firstName = cleanClassText(firstName)
	course = cleanClassText(course)
	scheduleText = cleanClassText(scheduleText)
	if nickname == "" || firstName == "" || course == "" {
		return StudentLesson{}, errors.New("กรุณาระบุชื่อเล่น ชื่อจริง และคอร์ส")
	}
	if totalHours <= 0 {
		totalHours = 8
	}

	ctx := context.Background()
	if err := s.client.valuesAppend(ctx, quoteSheetName(classOverviewSheet)+"!A1", [][]string{{
		firstName,
		nickname,
		"",
		course,
		"",
		"",
		"",
		"",
		strconv.Itoa(totalHours),
		"0",
		strconv.Itoa(totalHours),
		"กำลังเรียน",
		"",
		"",
	}}); err != nil {
		return StudentLesson{}, err
	}

	lesson := StudentLesson{
		ID:             "overview-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Nickname:       nickname,
		FirstName:      firstName,
		FullName:       firstName,
		Course:         course,
		TotalHours:     totalHours,
		CompletedHours: 0,
		SessionHours:   1,
		UpdatedAt:      time.Now().In(s.loc),
	}
	if scheduleText == "" {
		return lesson, nil
	}
	start, end, ok := parseSchedule(scheduleText, s.loc)
	if !ok {
		return StudentLesson{}, fmt.Errorf("อ่านวันเวลาไม่ได้: %s", scheduleText)
	}
	if err := s.appendWeeklyLesson(ctx, overviewRecord{
		FullName:       firstName,
		Nickname:       nickname,
		Level:          course,
		TotalHours:     totalHours,
		CompletedHours: 0,
	}, start, end, false); err != nil {
		return StudentLesson{}, err
	}
	lesson.NextStart = start
	lesson.NextEnd = end
	lesson.SessionHours = lessonSessionHours(start, end)
	lesson.ScheduleText = formatThaiSchedule(start, end)
	return lesson, nil
}

func (s *ClassScheduleSheetsLessonStore) UpdateLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, false, true)
}

func (s *ClassScheduleSheetsLessonStore) ConfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, true, false)
}

func (s *ClassScheduleSheetsLessonStore) UnconfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, false, false)
}

func (s *ClassScheduleSheetsLessonStore) UpdateLearningStatus(nickname, firstName, status string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	weekly, err := s.loadWeekly(ctx)
	if err != nil {
		return StudentLesson{}, err
	}
	record, err := findEditableWeeklyRecord(weekly, nickname, firstName, s.loc)
	if err != nil {
		return StudentLesson{}, err
	}
	start, end, ok := weeklyRecordTime(record, s.loc)
	if !ok {
		return StudentLesson{}, errors.New("อ่านวันเวลาในตารางสัปดาห์นี้ไม่ได้")
	}
	record.LearningStatus = cleanClassText(status)
	if err := s.updateWeeklyRecord(ctx, record); err != nil {
		return StudentLesson{}, err
	}
	return record.toStudentLesson(start, end, s.loc), nil
}

func (s *ClassScheduleSheetsLessonStore) FindLessonByStudentName(nickname, firstName string) (StudentLesson, error) {
	weekly, err := s.loadWeekly(context.Background())
	if err != nil {
		return StudentLesson{}, err
	}
	record, err := findEditableWeeklyRecord(weekly, nickname, firstName, s.loc)
	if err != nil {
		return StudentLesson{}, err
	}
	start, end, ok := weeklyRecordTime(record, s.loc)
	if !ok {
		return StudentLesson{}, errors.New("อ่านวันเวลาในตารางสัปดาห์นี้ไม่ได้")
	}
	return record.toStudentLesson(start, end, s.loc), nil
}

func (s *ClassScheduleSheetsLessonStore) changeLesson(nickname, firstName, scheduleText string, confirmed bool, requireSchedule bool) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	overview, err := s.loadOverview(ctx)
	if err != nil {
		return StudentLesson{}, err
	}
	weekly, err := s.loadWeekly(ctx)
	if err != nil {
		return StudentLesson{}, err
	}

	scheduleText = cleanClassText(scheduleText)
	var start time.Time
	var end time.Time
	hasSchedule := scheduleText != ""
	if hasSchedule {
		var ok bool
		start, end, ok = parseSchedule(scheduleText, s.loc)
		if !ok {
			return StudentLesson{}, fmt.Errorf("อ่านวันเวลาไม่ได้: %s", scheduleText)
		}
	}
	if requireSchedule && !hasSchedule {
		return StudentLesson{}, errors.New("กรุณาระบุวันที่และเวลา เช่น 9/5 13:00-15:00")
	}

	record, err := findEditableWeeklyRecord(weekly, nickname, firstName, s.loc)
	if err != nil && !hasSchedule {
		return StudentLesson{}, err
	}
	if err != nil && hasSchedule {
		base, baseErr := findSingleOverviewRecord(overview, nickname, firstName)
		if baseErr != nil {
			return StudentLesson{}, baseErr
		}
		if err := s.appendWeeklyLesson(ctx, base, start, end, confirmed); err != nil {
			return StudentLesson{}, err
		}
		return base.toStudentLesson(start, end, confirmed, s.loc), nil
	}

	if hasSchedule {
		record.DateText = formatClassDate(start)
		record.DayText = "วัน" + thaiWeekdays[start.Weekday()]
		record.TimeText = formatClassTimeRange(start, end)
		record.DurationText = fmt.Sprintf("%d ชม.", lessonSessionHours(start, end))
	} else {
		var ok bool
		start, end, ok = weeklyRecordTime(record, s.loc)
		if !ok {
			return StudentLesson{}, errors.New("อ่านวันเวลาในตารางสัปดาห์นี้ไม่ได้")
		}
	}
	record.Confirmed = confirmed
	if err := s.updateWeeklyRecord(ctx, record); err != nil {
		return StudentLesson{}, err
	}
	return record.toStudentLesson(start, end, s.loc), nil
}

func (s *ClassScheduleSheetsLessonStore) RegisterLineGroup(groupID string) error {
	return nil
}

func (s *ClassScheduleSheetsLessonStore) ListLineGroupIDs() ([]string, error) {
	return nil, nil
}

func (s *ClassScheduleSheetsLessonStore) loadOverview(ctx context.Context) ([]overviewRecord, error) {
	rows, err := s.client.valuesGet(ctx, quoteSheetName(classOverviewSheet)+"!A:N")
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, nil
	}
	var records []overviewRecord
	for i, row := range rows[1:] {
		if cleanClassText(valueAt(row, 0)) == "" {
			continue
		}
		records = append(records, overviewRecord{
			rowNumber:      i + 2,
			FullName:       cleanClassText(valueAt(row, 0)),
			Nickname:       cleanClassText(valueAt(row, 1)),
			Course:         cleanClassText(valueAt(row, 2)),
			Level:          cleanClassText(valueAt(row, 3)),
			Teacher:        cleanClassText(valueAt(row, 4)),
			Weekday:        cleanClassText(valueAt(row, 5)),
			TimeText:       cleanClassText(valueAt(row, 6)),
			StartDateText:  cleanClassText(valueAt(row, 7)),
			TotalHours:     parseClassInt(valueAt(row, 8), 0),
			CompletedHours: parseClassInt(valueAt(row, 9), 0),
			RemainingHours: parseClassInt(valueAt(row, 10), 0),
			Status:         cleanClassText(valueAt(row, 11)),
			ParentName:     cleanClassText(valueAt(row, 12)),
			ParentPhone:    cleanClassText(valueAt(row, 13)),
		})
	}
	return records, nil
}

func (s *ClassScheduleSheetsLessonStore) loadWeekly(ctx context.Context) ([]weeklyRecord, error) {
	rows, err := s.client.valuesGet(ctx, quoteSheetName(classWeeklySheet)+"!A:U")
	if err != nil {
		return nil, err
	}
	if len(rows) <= 2 {
		return nil, nil
	}
	var records []weeklyRecord
	for i, row := range rows[2:] {
		if cleanClassText(valueAt(row, 0)) == "" || cleanClassText(valueAt(row, 4)) == "" {
			continue
		}
		var slots []string
		for col := 15; col < len(row); col++ {
			if slot := cleanClassText(valueAt(row, col)); slot != "" {
				slots = append(slots, slot)
			}
		}
		records = append(records, weeklyRecord{
			rowNumber:      i + 3,
			DateText:       cleanClassText(valueAt(row, 0)),
			DayText:        cleanClassText(valueAt(row, 1)),
			TimeText:       cleanClassText(valueAt(row, 2)),
			DurationText:   cleanClassText(valueAt(row, 3)),
			FullName:       cleanClassText(valueAt(row, 4)),
			Nickname:       cleanClassText(valueAt(row, 5)),
			Course:         cleanClassText(valueAt(row, 6)),
			Level:          cleanClassText(valueAt(row, 7)),
			Teacher:        cleanClassText(valueAt(row, 8)),
			ParentName:     cleanClassText(valueAt(row, 9)),
			ParentPhone:    cleanClassText(valueAt(row, 10)),
			TotalHours:     parseClassInt(valueAt(row, 11), 0),
			CompletedHours: parseClassInt(valueAt(row, 12), 0),
			Confirmed:      strings.EqualFold(cleanClassText(valueAt(row, 13)), "true"),
			LearningStatus: cleanClassText(valueAt(row, 14)),
			AvailableSlots: slots,
		})
	}
	return records, nil
}

func (s *ClassScheduleSheetsLessonStore) updateWeeklyRecord(ctx context.Context, record weeklyRecord) error {
	if err := s.copyWeeklyRowPattern(ctx, record.rowNumber); err != nil {
		return err
	}
	row := record.toSheetRowValues()
	if err := s.client.valuesUpdateAny(ctx, fmt.Sprintf("%s!A%d:U%d", quoteSheetName(classWeeklySheet), record.rowNumber, record.rowNumber), [][]any{row}); err != nil {
		return err
	}
	return s.copyWeeklyRowPattern(ctx, record.rowNumber)
}

func (s *ClassScheduleSheetsLessonStore) appendWeeklyLesson(ctx context.Context, base overviewRecord, start, end time.Time, confirmed bool) error {
	record := weeklyRecord{
		DateText:       formatClassDate(start),
		DayText:        "วัน" + thaiWeekdays[start.Weekday()],
		TimeText:       formatClassTimeRange(start, end),
		DurationText:   fmt.Sprintf("%d ชม.", lessonSessionHours(start, end)),
		FullName:       base.FullName,
		Nickname:       base.Nickname,
		Course:         base.Course,
		Level:          base.Level,
		Teacher:        base.Teacher,
		ParentName:     base.ParentName,
		ParentPhone:    base.ParentPhone,
		TotalHours:     base.TotalHours,
		CompletedHours: base.CompletedHours,
		Confirmed:      confirmed,
		LearningStatus: "เข้าเรียนปกติ",
	}
	updatedRange, err := s.client.valuesAppendAny(ctx, quoteSheetName(classWeeklySheet)+"!A1", [][]any{record.toSheetRowValues()})
	if err != nil {
		return err
	}
	if rowNumber := rowNumberFromUpdatedRange(updatedRange); rowNumber > 0 {
		return s.copyWeeklyRowPattern(ctx, rowNumber)
	}
	return nil
}

func (r weeklyRecord) toSheetRowValues() []any {
	row := []any{
		r.DateText,
		r.DayText,
		r.TimeText,
		r.DurationText,
		r.FullName,
		r.Nickname,
		r.Course,
		r.Level,
		r.Teacher,
		r.ParentName,
		r.ParentPhone,
		strconv.Itoa(r.TotalHours),
		strconv.Itoa(r.CompletedHours),
		r.Confirmed,
		r.LearningStatus,
	}
	for _, slot := range r.AvailableSlots {
		row = append(row, slot)
	}
	for len(row) < 21 {
		row = append(row, "")
	}
	return row[:21]
}

func (s *ClassScheduleSheetsLessonStore) copyWeeklyRowPattern(ctx context.Context, targetRow int) error {
	if targetRow <= 3 {
		return s.copyWeeklyRowPatternFrom(ctx, targetRow+1, targetRow)
	}
	return s.copyWeeklyRowPatternFrom(ctx, targetRow-1, targetRow)
}

func (s *ClassScheduleSheetsLessonStore) copyWeeklyRowPatternFrom(ctx context.Context, sourceRow int, targetRow int) error {
	if sourceRow <= 2 || targetRow <= 2 || sourceRow == targetRow {
		return nil
	}
	rows, err := s.client.valuesGet(ctx, quoteSheetName(classWeeklySheet)+"!A:A")
	if err != nil {
		return err
	}
	if sourceRow > len(rows) {
		return nil
	}
	return s.client.copyPasteRow(ctx, classWeeklySheet, sourceRow, targetRow, 0, 21, "PASTE_FORMAT", "PASTE_DATA_VALIDATION")
}

func (r weeklyRecord) toStudentLesson(start, end time.Time, loc *time.Location) StudentLesson {
	start = start.In(loc)
	end = end.In(loc)
	return StudentLesson{
		ID:             fmt.Sprintf("weekly-row-%d", r.rowNumber),
		Nickname:       r.Nickname,
		FirstName:      classFirstName(r.FullName),
		FullName:       r.FullName,
		Course:         classCourseName(r.Course, r.Level),
		TotalHours:     r.TotalHours,
		CompletedHours: r.CompletedHours,
		SessionHours:   lessonSessionHours(start, end),
		NextStart:      start,
		NextEnd:        end,
		ScheduleText:   formatThaiSchedule(start, end),
		Confirmed:      r.Confirmed,
		LearningStatus: r.LearningStatus,
		ParentPhone:    r.ParentPhone,
		UpdatedAt:      time.Now().In(loc),
	}
}

func (r overviewRecord) toStudentLesson(start, end time.Time, confirmed bool, loc *time.Location) StudentLesson {
	return StudentLesson{
		ID:             fmt.Sprintf("overview-row-%d", r.rowNumber),
		Nickname:       r.Nickname,
		FirstName:      classFirstName(r.FullName),
		FullName:       r.FullName,
		Course:         classCourseName(r.Course, r.Level),
		TotalHours:     r.TotalHours,
		CompletedHours: r.CompletedHours,
		SessionHours:   lessonSessionHours(start, end),
		NextStart:      start,
		NextEnd:        end,
		ScheduleText:   formatThaiSchedule(start, end),
		Confirmed:      confirmed,
		LearningStatus: "เข้าเรียนปกติ",
		UpdatedAt:      time.Now().In(loc),
	}
}

func findEditableWeeklyRecord(records []weeklyRecord, nickname, firstName string, loc *time.Location) (weeklyRecord, error) {
	var matches []weeklyRecord
	for _, record := range records {
		if matchClassStudent(record.Nickname, record.FullName, nickname, firstName) {
			matches = append(matches, record)
		}
	}
	if len(matches) == 0 {
		return weeklyRecord{}, fmt.Errorf("ไม่พบนักเรียนในสัปดาห์นี้: %s / %s", nickname, firstName)
	}
	now := time.Now().In(loc)
	sort.Slice(matches, func(i, j int) bool {
		iStart, _, iOK := weeklyRecordTime(matches[i], loc)
		jStart, _, jOK := weeklyRecordTime(matches[j], loc)
		if !iOK {
			return false
		}
		if !jOK {
			return true
		}
		iFuture := !iStart.Before(now)
		jFuture := !jStart.Before(now)
		if iFuture != jFuture {
			return iFuture
		}
		return iStart.Before(jStart)
	})
	return matches[0], nil
}

func findSingleOverviewRecord(records []overviewRecord, nickname, firstName string) (overviewRecord, error) {
	var matches []overviewRecord
	for _, record := range records {
		if isUnfinishedOverviewRecord(record) && matchClassStudent(record.Nickname, record.FullName, nickname, firstName) {
			matches = append(matches, record)
		}
	}
	if len(matches) == 0 {
		return overviewRecord{}, fmt.Errorf("ไม่พบนักเรียน: %s / %s", nickname, firstName)
	}
	active := matches[:0]
	for _, record := range matches {
		if record.Status == "กำลังเรียน" || record.Status == "รอเริ่ม" {
			active = append(active, record)
		}
	}
	if len(active) == 1 {
		return active[0], nil
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return overviewRecord{}, fmt.Errorf("พบ %s / %s มากกว่า 1 คอร์ส กรุณาระบุให้ชัดขึ้นในอนาคต", nickname, firstName)
}

func isUnfinishedOverviewRecord(record overviewRecord) bool {
	status := cleanClassText(record.Status)
	if strings.Contains(status, "เรียนจบ") || strings.Contains(status, "จบแล้ว") || strings.Contains(status, "ออก") {
		return false
	}
	if record.TotalHours > 0 && record.CompletedHours >= record.TotalHours {
		return false
	}
	if record.TotalHours > 0 && record.RemainingHours == 0 {
		return false
	}
	return true
}

func recordPastLessonsText(base overviewRecord, weekly []weeklyRecord, loc *time.Location, now time.Time) string {
	var parts []string
	for _, record := range weekly {
		if !sameClassEnrollment(base, record) {
			continue
		}
		start, end, ok := weeklyRecordTime(record, loc)
		if ok && start.Before(now) {
			parts = append(parts, formatShortLessonTime(start, end))
		}
	}
	if len(parts) == 0 && base.CompletedHours > 0 {
		return fmt.Sprintf("เรียนแล้ว %d ชม.", base.CompletedHours)
	}
	if len(parts) > 3 {
		parts = parts[len(parts)-3:]
	}
	return strings.Join(parts, ", ")
}

func recordNextLessonsText(base overviewRecord, weekly []weeklyRecord, loc *time.Location, now time.Time) string {
	type item struct {
		start time.Time
		text  string
	}
	var items []item
	for _, record := range weekly {
		if !sameClassEnrollment(base, record) {
			continue
		}
		start, end, ok := weeklyRecordTime(record, loc)
		if ok && !start.Before(now) {
			items = append(items, item{start: start, text: formatShortLessonTime(start, end)})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].start.Before(items[j].start) })
	if len(items) > 3 {
		items = items[:3]
	}
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = item.text
	}
	return strings.Join(parts, ", ")
}

func sameClassEnrollment(base overviewRecord, weekly weeklyRecord) bool {
	return strings.EqualFold(cleanClassText(base.Nickname), cleanClassText(weekly.Nickname)) &&
		strings.EqualFold(cleanClassText(base.Level), cleanClassText(weekly.Level))
}

func matchClassStudent(actualNickname, actualFullName, nickname, firstName string) bool {
	return strings.EqualFold(cleanClassText(actualNickname), cleanClassText(nickname)) &&
		strings.EqualFold(classFirstName(actualFullName), cleanClassText(firstName))
}

func weeklyRecordTime(record weeklyRecord, loc *time.Location) (time.Time, time.Time, bool) {
	date, ok := parseClassDate(record.DateText)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	startHour, startMinute, endHour, endMinute, ok := parseClassTimeRange(record.TimeText)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	start := time.Date(date.Year(), date.Month(), date.Day(), startHour, startMinute, 0, 0, loc)
	end := time.Date(date.Year(), date.Month(), date.Day(), endHour, endMinute, 0, 0, loc)
	if !end.After(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func parseClassDate(text string) (time.Time, bool) {
	parts := strings.Split(cleanClassText(text), "/")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	day := parseClassInt(parts[0], 0)
	month := parseClassInt(parts[1], 0)
	year := parseClassInt(parts[2], 0)
	if year > 2400 {
		year -= 543
	}
	if day <= 0 || month <= 0 || year <= 0 {
		return time.Time{}, false
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), true
}

func parseClassTimeRange(text string) (int, int, int, int, bool) {
	parts := strings.FieldsFunc(cleanClassText(text), func(r rune) bool {
		return r == '-' || r == '–'
	})
	if len(parts) != 2 {
		return 0, 0, 0, 0, false
	}
	start := strings.Split(parts[0], ":")
	end := strings.Split(parts[1], ":")
	if len(start) != 2 || len(end) != 2 {
		return 0, 0, 0, 0, false
	}
	return parseClassInt(start[0], 0), parseClassInt(start[1], 0), parseClassInt(end[0], 0), parseClassInt(end[1], 0), true
}

func formatClassDate(t time.Time) string {
	return fmt.Sprintf("%02d/%02d/%04d", t.Day(), int(t.Month()), t.Year()+543)
}

func formatClassTimeRange(start, end time.Time) string {
	return fmt.Sprintf("%02d:%02d-%02d:%02d", start.Hour(), start.Minute(), end.Hour(), end.Minute())
}

func lessonSessionHours(start, end time.Time) int {
	hours := int(end.Sub(start).Hours())
	if hours <= 0 {
		return 1
	}
	return hours
}

func classFirstName(fullName string) string {
	text := cleanClassText(fullName)
	for _, prefix := range []string{"ด.ช.", "ดช", "เด็กชาย", "ด.ญ.", "ดญ", "เด็กหญิง", "น้อง"} {
		if strings.HasPrefix(text, prefix) {
			text = cleanClassText(strings.TrimPrefix(text, prefix))
			break
		}
	}
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return text
	}
	return parts[0]
}

func classCourseName(course, level string) string {
	if cleanClassText(level) != "" {
		return cleanClassText(level)
	}
	return cleanClassText(course)
}

func cleanClassText(text string) string {
	text = strings.ReplaceAll(text, "\u200b", "")
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func parseClassInt(text string, fallback int) int {
	value, err := strconv.Atoi(cleanClassText(text))
	if err != nil {
		return fallback
	}
	return value
}

func valueAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func rowNumberFromUpdatedRange(updatedRange string) int {
	value := strings.TrimSpace(updatedRange)
	if value == "" {
		return 0
	}
	if bangIndex := strings.LastIndex(value, "!"); bangIndex >= 0 {
		value = value[bangIndex+1:]
	}
	if colonIndex := strings.Index(value, ":"); colonIndex >= 0 {
		value = value[:colonIndex]
	}
	var digits strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else if digits.Len() > 0 {
			break
		}
	}
	return parseClassInt(digits.String(), 0)
}
