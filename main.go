package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const (
	linePushURL           = "https://api.line.me/v2/bot/message/push"
	lineReplyURL          = "https://api.line.me/v2/bot/message/reply"
	lineTextMaxLength     = 4500
	lineMessageBatchLimit = 5
)

type LineWebhook struct {
	Events []LineEvent `json:"events"`
}

type LineEvent struct {
	Type       string `json:"type"`
	ReplyToken string `json:"replyToken"`
	Source     struct {
		Type    string `json:"type"`
		UserID  string `json:"userId"`
		GroupID string `json:"groupId"`
	} `json:"source"`
	Message struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"message"`
}

type LineClient struct {
	accessToken    string
	defaultGroupID string
	targetGroupIDs []string
	httpClient     *http.Client
}

type lineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewLineClient() *LineClient {
	defaultGroupID := strings.TrimSpace(os.Getenv("LINE_STAFF_GROUP_ID"))
	return &LineClient{
		accessToken:    os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		defaultGroupID: defaultGroupID,
		targetGroupIDs: parseLineGroupIDs(os.Getenv("LINE_GROUP_IDS"), defaultGroupID),
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

func parseLineGroupIDs(values ...string) []string {
	seen := map[string]bool{}
	var groupIDs []string
	for _, value := range values {
		for _, groupID := range strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
		}) {
			groupID = strings.TrimSpace(groupID)
			if !strings.HasPrefix(groupID, "C") || !isLikelyLineTargetID(groupID) || seen[groupID] {
				continue
			}
			seen[groupID] = true
			groupIDs = append(groupIDs, groupID)
		}
	}
	return groupIDs
}

func (c *LineClient) TargetGroupIDs() []string {
	groupIDs := make([]string, len(c.targetGroupIDs))
	copy(groupIDs, c.targetGroupIDs)
	return groupIDs
}

func (c *LineClient) AllowsGroup(groupID string) bool {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return true
	}
	if len(c.targetGroupIDs) == 0 {
		return true
	}
	for _, targetGroupID := range c.targetGroupIDs {
		if targetGroupID == groupID {
			return true
		}
	}
	return false
}

func (c *LineClient) FirstTargetGroupID() string {
	if len(c.targetGroupIDs) == 0 {
		return ""
	}
	return c.targetGroupIDs[0]
}

func (c *LineClient) SendText(to string, text string) error {
	return c.SendTextParts(to, splitLongLineMessage(text, lineTextMaxLength))
}

func (c *LineClient) SendTextParts(to string, parts []string) error {
	if !isLikelyLineTargetID(to) {
		return errors.New("missing or invalid LINE target ID")
	}
	parts = cleanLineTextParts(parts)
	if len(parts) == 0 {
		return nil
	}
	for _, batch := range batchLineTextParts(parts, lineMessageBatchLimit) {
		if err := c.send(linePushURL, map[string]any{
			"to":       to,
			"messages": lineTextMessages(batch),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (c *LineClient) ReplyText(replyToken string, text string) error {
	return c.ReplyTextParts(replyToken, splitLongLineMessage(text, lineTextMaxLength))
}

func (c *LineClient) ReplyTextParts(replyToken string, parts []string) error {
	if strings.TrimSpace(replyToken) == "" {
		return errors.New("missing LINE reply token")
	}
	parts = cleanLineTextParts(parts)
	if len(parts) == 0 {
		return nil
	}
	if len(parts) > lineMessageBatchLimit {
		return fmt.Errorf("LINE reply can contain at most %d messages", lineMessageBatchLimit)
	}
	return c.send(lineReplyURL, map[string]any{
		"replyToken": replyToken,
		"messages":   lineTextMessages(parts),
	})
}

func lineTextMessages(parts []string) []lineTextMessage {
	messages := make([]lineTextMessage, 0, len(parts))
	for _, part := range parts {
		messages = append(messages, lineTextMessage{Type: "text", Text: part})
	}
	return messages
}

func cleanLineTextParts(parts []string) []string {
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return cleaned
}

func batchLineTextParts(parts []string, size int) [][]string {
	if size <= 0 {
		size = lineMessageBatchLimit
	}
	var batches [][]string
	for len(parts) > 0 {
		n := size
		if len(parts) < n {
			n = len(parts)
		}
		batches = append(batches, parts[:n])
		parts = parts[n:]
	}
	return batches
}

func (c *LineClient) send(url string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if strings.TrimSpace(c.accessToken) == "" {
		log.Printf("[LINE dry-run] POST %s %s", url, string(body))
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE API returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

type StudentLesson struct {
	ID             string
	Nickname       string
	FirstName      string
	FullName       string
	Course         string
	TotalHours     int
	CompletedHours int
	SessionHours   int
	NextStart      time.Time
	NextEnd        time.Time
	ScheduleText   string
	Confirmed      bool
	LearningStatus string
	UpdatedAt      time.Time
}

type StudentScheduleSummary struct {
	Nickname        string
	FirstName       string
	FullName        string
	Course          string
	TotalHours      int
	CompletedHours  int
	DefaultSchedule string
	PastLessons     string
	NextLessons     string
	ScheduleNotes   string
}

type LessonStore interface {
	ListLessons() []StudentLesson
	ListStudentSchedules() []StudentScheduleSummary
	FindStudentSchedules(nickname, firstName string) ([]StudentScheduleSummary, error)
	AddStudent(nickname, firstName, course string, totalHours int, scheduleText string) (StudentLesson, error)
	UpdateLesson(nickname, firstName, scheduleText string) (StudentLesson, error)
	ConfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error)
	UnconfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error)
	UpdateLearningStatus(nickname, firstName, status string) (StudentLesson, error)
	FindLessonByStudentName(nickname, firstName string) (StudentLesson, error)
	RegisterLineGroup(groupID string) error
	ListLineGroupIDs() ([]string, error)
}

type MockLessonStore struct {
	mu      sync.RWMutex
	lessons map[string]StudentLesson
	groups  map[string]bool
	loc     *time.Location
}

func NewMockLessonStore(loc *time.Location) *MockLessonStore {
	store := &MockLessonStore{
		lessons: map[string]StudentLesson{},
		groups:  map[string]bool{},
		loc:     loc,
	}

	store.seed(StudentLesson{
		ID:             "stu-001",
		Nickname:       "แพรว",
		FullName:       "แพรวา ศิริพงษ์",
		Course:         "English Foundation",
		TotalHours:     20,
		CompletedHours: 6,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Monday, 18, 0),
		Confirmed:      false,
	})
	store.seed(StudentLesson{
		ID:             "stu-002",
		Nickname:       "บอส",
		FullName:       "ธนากร ใจดี",
		Course:         "คณิตศาสตร์ ม.2",
		TotalHours:     24,
		CompletedHours: 4,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Saturday, 13, 0),
		Confirmed:      true,
	})
	store.seed(StudentLesson{
		ID:             "stu-003",
		Nickname:       "มิน",
		FullName:       "ณัฐธิดา วงศ์วาน",
		Course:         "IELTS Speaking",
		TotalHours:     12,
		CompletedHours: 1,
		SessionHours:   1,
		NextStart:      nextWeekdayAt(loc, time.Wednesday, 19, 0),
		Confirmed:      false,
	})
	store.seed(StudentLesson{
		ID:             "stu-004",
		Nickname:       "ต้น",
		FullName:       "กฤติน ภักดี",
		Course:         "Physics ม.4",
		TotalHours:     30,
		CompletedHours: 10,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Tuesday, 17, 30),
		Confirmed:      true,
	})
	store.seed(StudentLesson{
		ID:             "stu-005",
		Nickname:       "ฟ้า",
		FullName:       "ฟ้าลดา เกียรติไกร",
		Course:         "Chemistry ม.5",
		TotalHours:     28,
		CompletedHours: 12,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Thursday, 18, 0),
		Confirmed:      false,
	})
	store.seed(StudentLesson{
		ID:             "stu-006",
		Nickname:       "เจ",
		FullName:       "จิรายุ ตั้งมั่น",
		Course:         "English Conversation",
		TotalHours:     16,
		CompletedHours: 8,
		SessionHours:   1,
		NextStart:      nextWeekdayAt(loc, time.Friday, 20, 0),
		Confirmed:      true,
	})
	store.seed(StudentLesson{
		ID:             "stu-007",
		Nickname:       "ออม",
		FullName:       "อรอุมา แสงทอง",
		Course:         "คณิตศาสตร์ ม.6",
		TotalHours:     40,
		CompletedHours: 18,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Sunday, 10, 0),
		Confirmed:      false,
	})
	store.seed(StudentLesson{
		ID:             "stu-008",
		Nickname:       "พีท",
		FullName:       "พีรวิชญ์ สุขสวัสดิ์",
		Course:         "SAT Math",
		TotalHours:     20,
		CompletedHours: 14,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Saturday, 15, 30),
		Confirmed:      true,
	})
	store.seed(StudentLesson{
		ID:             "stu-009",
		Nickname:       "ข้าว",
		FullName:       "ขวัญชนก มีสุข",
		Course:         "ภาษาไทย ป.6",
		TotalHours:     18,
		CompletedHours: 5,
		SessionHours:   1,
		NextStart:      nextWeekdayAt(loc, time.Monday, 16, 0),
		Confirmed:      false,
	})
	store.seed(StudentLesson{
		ID:             "stu-010",
		Nickname:       "นิว",
		FullName:       "นวพล จันทร์เจ้า",
		Course:         "Biology ม.5",
		TotalHours:     26,
		CompletedHours: 20,
		SessionHours:   2,
		NextStart:      nextWeekdayAt(loc, time.Wednesday, 17, 0),
		Confirmed:      true,
	})

	return store
}

func (s *MockLessonStore) seed(lesson StudentLesson) {
	if lesson.SessionHours <= 0 {
		lesson.SessionHours = 1
	}
	if lesson.NextEnd.IsZero() {
		lesson.NextEnd = lesson.NextStart.Add(time.Duration(lesson.SessionHours) * time.Hour)
	}
	if strings.TrimSpace(lesson.ScheduleText) == "" {
		lesson.ScheduleText = formatThaiSchedule(lesson.NextStart, lesson.NextEnd)
	}
	if strings.TrimSpace(lesson.FirstName) == "" {
		lesson.FirstName = firstWord(lesson.FullName)
	}
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[studentNameKey(lesson.Nickname, lesson.FirstName)] = lesson
}

func (s *MockLessonStore) ListLessons() []StudentLesson {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lessons := make([]StudentLesson, 0, len(s.lessons))
	for _, lesson := range s.lessons {
		lessons = append(lessons, lesson)
	}
	sort.Slice(lessons, func(i, j int) bool {
		return lessons[i].NextStart.Before(lessons[j].NextStart)
	})
	return lessons
}

func (s *MockLessonStore) ListStudentSchedules() []StudentScheduleSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]StudentScheduleSummary, 0, len(s.lessons))
	for _, lesson := range s.lessons {
		summaries = append(summaries, lessonToScheduleSummary(lesson))
	}
	sortStudentScheduleSummaries(summaries)
	return summaries
}

func (s *MockLessonStore) FindStudentSchedules(nickname, firstName string) ([]StudentScheduleSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var summaries []StudentScheduleSummary
	for _, lesson := range s.lessons {
		if strings.EqualFold(lesson.Nickname, strings.TrimSpace(nickname)) && strings.EqualFold(lesson.FirstName, strings.TrimSpace(firstName)) {
			summaries = append(summaries, lessonToScheduleSummary(lesson))
		}
	}
	if len(summaries) == 0 {
		return nil, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s", nickname, firstName)
	}
	sortStudentScheduleSummaries(summaries)
	return summaries, nil
}

func (s *MockLessonStore) AddStudent(nickname, firstName, course string, totalHours int, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nickname = strings.TrimSpace(nickname)
	firstName = strings.TrimSpace(firstName)
	course = strings.TrimSpace(course)
	if nickname == "" || firstName == "" || course == "" {
		return StudentLesson{}, errors.New("กรุณาระบุชื่อเล่น ชื่อจริง และคอร์ส")
	}
	if totalHours <= 0 {
		totalHours = 8
	}

	lesson := StudentLesson{
		ID:             fmt.Sprintf("mock-%d", len(s.lessons)+1),
		Nickname:       nickname,
		FirstName:      firstName,
		FullName:       firstName,
		Course:         course,
		TotalHours:     totalHours,
		CompletedHours: 0,
		SessionHours:   2,
		Confirmed:      false,
		UpdatedAt:      time.Now().In(s.loc),
	}
	if strings.TrimSpace(scheduleText) != "" {
		lesson = applySchedule(lesson, scheduleText, s.loc)
	}
	s.lessons[studentNameKey(lesson.Nickname, lesson.FirstName)] = lesson
	return lesson, nil
}

func (s *MockLessonStore) UpdateLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := studentNameKey(nickname, firstName)
	lesson, ok := s.lessons[key]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s", nickname, firstName)
	}

	lesson = applySchedule(lesson, scheduleText, s.loc)
	lesson.Confirmed = false
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[key] = lesson
	return lesson, nil
}

func (s *MockLessonStore) ConfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := studentNameKey(nickname, firstName)
	lesson, ok := s.lessons[key]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s", nickname, firstName)
	}

	if strings.TrimSpace(scheduleText) != "" {
		lesson = applySchedule(lesson, scheduleText, s.loc)
	}
	lesson.Confirmed = true
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[key] = lesson
	return lesson, nil
}

func (s *MockLessonStore) UnconfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := studentNameKey(nickname, firstName)
	lesson, ok := s.lessons[key]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s", nickname, firstName)
	}

	if strings.TrimSpace(scheduleText) != "" {
		lesson = applySchedule(lesson, scheduleText, s.loc)
	}
	lesson.Confirmed = false
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[key] = lesson
	return lesson, nil
}

func (s *MockLessonStore) UpdateLearningStatus(nickname, firstName, status string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := studentNameKey(nickname, firstName)
	lesson, ok := s.lessons[key]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s", nickname, firstName)
	}
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[key] = lesson
	return lesson, nil
}

func (s *MockLessonStore) FindLessonByStudentName(nickname, firstName string) (StudentLesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lesson, ok := s.lessons[studentNameKey(nickname, firstName)]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s", nickname, firstName)
	}
	return lesson, nil
}

func (s *MockLessonStore) RegisterLineGroup(groupID string) error {
	if !isLikelyLineTargetID(groupID) || !strings.HasPrefix(strings.TrimSpace(groupID), "C") {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups[strings.TrimSpace(groupID)] = true
	return nil
}

func (s *MockLessonStore) ListLineGroupIDs() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groupIDs := make([]string, 0, len(s.groups))
	for groupID := range s.groups {
		groupIDs = append(groupIDs, groupID)
	}
	sort.Strings(groupIDs)
	return groupIDs, nil
}

type PostgresLessonStore struct {
	db  *sql.DB
	loc *time.Location
}

func NewPostgresLessonStore(databaseURL string, loc *time.Location) (*PostgresLessonStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	store := &PostgresLessonStore{db: db, loc: loc}
	if err := store.waitForDatabase(30 * time.Second); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresLessonStore) waitForDatabase(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var err error
	for time.Now().Before(deadline) {
		if err = s.db.Ping(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func (s *PostgresLessonStore) Migrate(schemaPath string) error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(string(schema))
	return err
}

func (s *PostgresLessonStore) ListLessons() []StudentLesson {
	rows, err := s.db.Query(`
		SELECT
			ls.id::text,
			st.nickname,
			st.first_name,
			COALESCE(st.full_name_th, ''),
			c.name,
			e.total_hours,
			e.completed_hours,
			GREATEST(1, CEIL(EXTRACT(EPOCH FROM (ls.end_at - ls.start_at)) / 3600.0)::int),
			ls.start_at,
			ls.end_at,
			ls.status = 'confirmed',
			ls.updated_at
		FROM lesson_sessions ls
		JOIN enrollments e ON e.id = ls.enrollment_id
		JOIN students st ON st.id = e.student_id
		JOIN courses c ON c.id = e.course_id
		WHERE e.active = true
		  AND ls.status <> 'cancelled'
		  AND ls.start_at >= (NOW() - INTERVAL '30 days')
		  AND ls.start_at < (NOW() + INTERVAL '365 days')
		ORDER BY ls.start_at, st.nickname
	`)
	if err != nil {
		log.Println("list lessons query error:", err)
		return nil
	}
	defer rows.Close()

	var lessons []StudentLesson
	for rows.Next() {
		var lesson StudentLesson
		if err := rows.Scan(
			&lesson.ID,
			&lesson.Nickname,
			&lesson.FirstName,
			&lesson.FullName,
			&lesson.Course,
			&lesson.TotalHours,
			&lesson.CompletedHours,
			&lesson.SessionHours,
			&lesson.NextStart,
			&lesson.NextEnd,
			&lesson.Confirmed,
			&lesson.UpdatedAt,
		); err != nil {
			log.Println("scan lesson error:", err)
			continue
		}
		lesson.NextStart = lesson.NextStart.In(s.loc)
		lesson.NextEnd = lesson.NextEnd.In(s.loc)
		lesson.UpdatedAt = lesson.UpdatedAt.In(s.loc)
		lesson.ScheduleText = formatThaiSchedule(lesson.NextStart, lesson.NextEnd)
		lessons = append(lessons, lesson)
	}
	return lessons
}

func (s *PostgresLessonStore) ListStudentSchedules() []StudentScheduleSummary {
	rows, err := s.db.Query(studentScheduleSummaryQuery(""))
	if err != nil {
		log.Println("list student schedules query error:", err)
		return nil
	}
	defer rows.Close()

	summaries, err := scanStudentScheduleSummaries(rows)
	if err != nil {
		log.Println("scan student schedules error:", err)
		return nil
	}
	return summaries
}

func (s *PostgresLessonStore) FindStudentSchedules(nickname, firstName string) ([]StudentScheduleSummary, error) {
	rows, err := s.db.Query(
		studentScheduleSummaryQuery(`
		  AND lower(st.nickname) = lower($1)
		  AND lower(st.first_name) = lower($2)
		`),
		strings.TrimSpace(nickname),
		strings.TrimSpace(firstName),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries, err := scanStudentScheduleSummaries(rows)
	if err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return nil, fmt.Errorf("ไม่พบนักเรียน: %s / %s", nickname, firstName)
	}
	return summaries, nil
}

func studentScheduleSummaryQuery(extraWhere string) string {
	return `
		SELECT
			st.nickname,
			st.first_name,
			COALESCE(st.full_name_th, ''),
			c.name,
			e.total_hours,
			e.completed_hours,
			COALESCE((
				SELECT string_agg(
					eds.weekday_text || ' ' || to_char(eds.start_time, 'HH24:MI') || '-' || to_char(eds.end_time, 'HH24:MI'),
					', ' ORDER BY eds.weekday_text, eds.start_time
				)
				FROM enrollment_default_schedules eds
				WHERE eds.enrollment_id = e.id
				  AND eds.active = true
			), (
				SELECT concat_ws(' | ', NULLIF(default_day_text, ''), string_agg(slot_text, ', ' ORDER BY sort_time))
				FROM (
					SELECT DISTINCT
						COALESCE(cds.default_day_text, '') AS default_day_text,
						to_char(cds.start_time, 'HH24:MI') || '-' || to_char(cds.end_time, 'HH24:MI') AS slot_text,
						cds.start_time AS sort_time
					FROM course_default_schedules cds
					WHERE cds.course_id = c.id
					  AND cds.active = true
				) default_slots
				GROUP BY default_day_text
				LIMIT 1
			), '') AS default_schedule,
			COALESCE((
				SELECT string_agg(item, ', ' ORDER BY start_at DESC)
				FROM (
					SELECT
						ls.start_at,
						to_char(ls.start_at AT TIME ZONE 'Asia/Bangkok', 'DD/MM/YYYY HH24:MI') || '-' ||
						to_char(ls.end_at AT TIME ZONE 'Asia/Bangkok', 'HH24:MI') AS item
					FROM lesson_sessions ls
					WHERE ls.enrollment_id = e.id
					  AND ls.status = 'completed'
					ORDER BY ls.start_at DESC
					LIMIT 3
				) past_sessions
			), '') AS past_lessons,
			COALESCE((
				SELECT string_agg(item, ', ' ORDER BY start_at)
				FROM (
					SELECT
						ls.start_at,
						to_char(ls.start_at AT TIME ZONE 'Asia/Bangkok', 'DD/MM/YYYY HH24:MI') || '-' ||
						to_char(ls.end_at AT TIME ZONE 'Asia/Bangkok', 'HH24:MI') AS item
					FROM lesson_sessions ls
					WHERE ls.enrollment_id = e.id
					  AND ls.status NOT IN ('completed', 'cancelled')
					  AND ls.start_at >= NOW()
					ORDER BY ls.start_at
					LIMIT 3
				) next_sessions
			), '') AS next_lessons,
			COALESCE((
				SELECT string_agg(source_text, ', ' ORDER BY sort_status, sequence_no)
				FROM (
					SELECT
						esn.sequence_no,
						esn.source_text,
						CASE esn.status
							WHEN 'pending_confirm' THEN 0
							WHEN 'pending_date' THEN 1
							ELSE 2
						END AS sort_status
					FROM enrollment_schedule_notes esn
					WHERE esn.enrollment_id = e.id
					ORDER BY sort_status, esn.sequence_no
					LIMIT 3
				) notes
			), '') AS schedule_notes
		FROM enrollments e
		JOIN students st ON st.id = e.student_id
		JOIN courses c ON c.id = e.course_id
		WHERE e.active = true
	` + extraWhere + `
		ORDER BY c.name, st.nickname, st.first_name
	`
}

func scanStudentScheduleSummaries(rows *sql.Rows) ([]StudentScheduleSummary, error) {
	var summaries []StudentScheduleSummary
	for rows.Next() {
		var summary StudentScheduleSummary
		if err := rows.Scan(
			&summary.Nickname,
			&summary.FirstName,
			&summary.FullName,
			&summary.Course,
			&summary.TotalHours,
			&summary.CompletedHours,
			&summary.DefaultSchedule,
			&summary.PastLessons,
			&summary.NextLessons,
			&summary.ScheduleNotes,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

func (s *PostgresLessonStore) AddStudent(nickname, firstName, course string, totalHours int, scheduleText string) (StudentLesson, error) {
	nickname = strings.TrimSpace(nickname)
	firstName = strings.TrimSpace(firstName)
	course = strings.TrimSpace(course)
	scheduleText = strings.TrimSpace(scheduleText)
	if nickname == "" || firstName == "" || course == "" {
		return StudentLesson{}, errors.New("กรุณาระบุชื่อเล่น ชื่อจริง และคอร์ส")
	}
	if totalHours <= 0 {
		totalHours = 8
	}

	tx, err := s.db.Begin()
	if err != nil {
		return StudentLesson{}, err
	}
	defer tx.Rollback()

	var studentID int64
	if err := tx.QueryRow(`
		INSERT INTO students (nickname, first_name, full_name_th)
		VALUES ($1, $2, $2)
		RETURNING id
	`, nickname, firstName).Scan(&studentID); err != nil {
		return StudentLesson{}, err
	}

	var courseID int64
	if err := tx.QueryRow(`
		INSERT INTO courses (name, default_total_hours, default_session_hours)
		VALUES ($1, $2, 2)
		ON CONFLICT (name) DO UPDATE SET
			default_total_hours = EXCLUDED.default_total_hours,
			updated_at = NOW()
		RETURNING id
	`, course, totalHours).Scan(&courseID); err != nil {
		return StudentLesson{}, err
	}

	var enrollmentID int64
	if err := tx.QueryRow(`
		INSERT INTO enrollments (student_id, course_id, total_hours, completed_hours, default_session_hours, active)
		VALUES ($1, $2, $3, 0, 2, true)
		ON CONFLICT (student_id, course_id) DO UPDATE SET
			total_hours = EXCLUDED.total_hours,
			active = true,
			updated_at = NOW()
		RETURNING id
	`, studentID, courseID, totalHours).Scan(&enrollmentID); err != nil {
		return StudentLesson{}, err
	}

	var sessionID int64
	if scheduleText != "" {
		start, end, ok := parseSchedule(scheduleText, s.loc)
		if !ok {
			return StudentLesson{}, fmt.Errorf("อ่านวันเวลาไม่ได้: %s", scheduleText)
		}
		if err := tx.QueryRow(`
			WITH next_sequence AS (
				SELECT COALESCE(MAX(sequence_no), 0) + 1 AS value
				FROM lesson_sessions
				WHERE enrollment_id = $1
			)
			INSERT INTO lesson_sessions (enrollment_id, sequence_no, start_at, end_at, status)
			SELECT $1, value, $2, $3, 'unconfirmed' FROM next_sequence
			RETURNING id
		`, enrollmentID, start, end).Scan(&sessionID); err != nil {
			return StudentLesson{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return StudentLesson{}, err
	}
	if sessionID != 0 {
		return s.lessonBySessionID(sessionID)
	}
	return StudentLesson{
		ID:             fmt.Sprintf("enrollment-%d", enrollmentID),
		Nickname:       nickname,
		FirstName:      firstName,
		FullName:       firstName,
		Course:         course,
		TotalHours:     totalHours,
		CompletedHours: 0,
		SessionHours:   2,
		Confirmed:      false,
		UpdatedAt:      time.Now().In(s.loc),
	}, nil
}

func (s *PostgresLessonStore) UpdateLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, "unconfirmed", true)
}

func (s *PostgresLessonStore) ConfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, "confirmed", false)
}

func (s *PostgresLessonStore) UnconfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, "unconfirmed", false)
}

func (s *PostgresLessonStore) UpdateLearningStatus(nickname, firstName, status string) (StudentLesson, error) {
	return s.FindLessonByStudentName(nickname, firstName)
}

func (s *PostgresLessonStore) FindLessonByStudentName(nickname, firstName string) (StudentLesson, error) {
	enrollmentID, err := s.findEnrollmentID(nickname, firstName)
	if err != nil {
		return StudentLesson{}, err
	}
	sessionID, err := s.findEditableSessionID(enrollmentID)
	if err != nil {
		return StudentLesson{}, err
	}
	return s.lessonBySessionID(sessionID)
}

func (s *PostgresLessonStore) RegisterLineGroup(groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if !isLikelyLineTargetID(groupID) || !strings.HasPrefix(groupID, "C") {
		return nil
	}

	_, err := s.db.Exec(`
		INSERT INTO line_groups (group_id, active, first_seen_at, last_seen_at)
		VALUES ($1, true, NOW(), NOW())
		ON CONFLICT (group_id)
		DO UPDATE SET active = true, last_seen_at = NOW()
	`, groupID)
	return err
}

func (s *PostgresLessonStore) ListLineGroupIDs() ([]string, error) {
	rows, err := s.db.Query(`SELECT group_id FROM line_groups WHERE active = true ORDER BY group_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}
	return groupIDs, rows.Err()
}

func (s *PostgresLessonStore) changeLesson(nickname, firstName, scheduleText, status string, requireSchedule bool) (StudentLesson, error) {
	enrollmentID, err := s.findEnrollmentID(nickname, firstName)
	if err != nil {
		return StudentLesson{}, err
	}

	scheduleText = strings.TrimSpace(scheduleText)
	var start time.Time
	var end time.Time
	hasSchedule := false
	if scheduleText != "" {
		var ok bool
		start, end, ok = parseSchedule(scheduleText, s.loc)
		if !ok {
			return StudentLesson{}, fmt.Errorf("อ่านวันเวลาไม่ได้: %s", scheduleText)
		}
		hasSchedule = true
	}
	if requireSchedule && !hasSchedule {
		return StudentLesson{}, errors.New("กรุณาระบุวันที่และเวลา เช่น 9/5 13:00-15:00")
	}

	sessionID, err := s.findEditableSessionID(enrollmentID)
	if err != nil && !hasSchedule {
		return StudentLesson{}, err
	}
	if err != nil && hasSchedule {
		sessionID, err = s.insertLessonSession(enrollmentID, start, end, status)
		if err != nil {
			return StudentLesson{}, err
		}
		return s.lessonBySessionID(sessionID)
	}

	if hasSchedule {
		_, err = s.db.Exec(`
			UPDATE lesson_sessions
			SET start_at = $1, end_at = $2, status = $3, updated_at = NOW()
			WHERE id = $4
		`, start, end, status, sessionID)
	} else {
		_, err = s.db.Exec(`
			UPDATE lesson_sessions
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`, status, sessionID)
	}
	if err != nil {
		return StudentLesson{}, err
	}
	return s.lessonBySessionID(sessionID)
}

func (s *PostgresLessonStore) findEnrollmentID(nickname, firstName string) (int64, error) {
	rows, err := s.db.Query(`
		SELECT e.id
		FROM enrollments e
		JOIN students st ON st.id = e.student_id
		WHERE e.active = true
		  AND lower(st.nickname) = lower($1)
		  AND lower(st.first_name) = lower($2)
		ORDER BY e.id
	`, strings.TrimSpace(nickname), strings.TrimSpace(firstName))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var matches []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		matches = append(matches, id)
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("ไม่พบนักเรียน: %s / %s", nickname, firstName)
	}
	if len(matches) > 1 {
		return 0, fmt.Errorf("พบ %s / %s มากกว่า 1 คอร์ส กรุณาเพิ่มคำสั่งระบุคอร์สในอนาคต", nickname, firstName)
	}
	return matches[0], nil
}

func (s *PostgresLessonStore) findEditableSessionID(enrollmentID int64) (int64, error) {
	var sessionID int64
	err := s.db.QueryRow(`
		SELECT id
		FROM lesson_sessions
		WHERE enrollment_id = $1
		  AND status NOT IN ('completed', 'cancelled')
		ORDER BY
		  CASE WHEN start_at >= NOW() THEN 0 ELSE 1 END,
		  start_at
		LIMIT 1
	`, enrollmentID).Scan(&sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("ยังไม่มี session ที่แก้ไขได้สำหรับนักเรียนคนนี้")
		}
		return 0, err
	}
	return sessionID, nil
}

func (s *PostgresLessonStore) insertLessonSession(enrollmentID int64, start time.Time, end time.Time, status string) (int64, error) {
	var sessionID int64
	err := s.db.QueryRow(`
		WITH next_sequence AS (
			SELECT COALESCE(MAX(sequence_no), 0) + 1 AS value
			FROM lesson_sessions
			WHERE enrollment_id = $1
		)
		INSERT INTO lesson_sessions (enrollment_id, sequence_no, start_at, end_at, status)
		SELECT $1, value, $2, $3, $4 FROM next_sequence
		RETURNING id
	`, enrollmentID, start, end, status).Scan(&sessionID)
	return sessionID, err
}

func (s *PostgresLessonStore) lessonBySessionID(sessionID int64) (StudentLesson, error) {
	var lesson StudentLesson
	err := s.db.QueryRow(`
		SELECT
			ls.id::text,
			st.nickname,
			st.first_name,
			COALESCE(st.full_name_th, ''),
			c.name,
			e.total_hours,
			e.completed_hours,
			GREATEST(1, CEIL(EXTRACT(EPOCH FROM (ls.end_at - ls.start_at)) / 3600.0)::int),
			ls.start_at,
			ls.end_at,
			ls.status = 'confirmed',
			ls.updated_at
		FROM lesson_sessions ls
		JOIN enrollments e ON e.id = ls.enrollment_id
		JOIN students st ON st.id = e.student_id
		JOIN courses c ON c.id = e.course_id
		WHERE ls.id = $1
	`, sessionID).Scan(
		&lesson.ID,
		&lesson.Nickname,
		&lesson.FirstName,
		&lesson.FullName,
		&lesson.Course,
		&lesson.TotalHours,
		&lesson.CompletedHours,
		&lesson.SessionHours,
		&lesson.NextStart,
		&lesson.NextEnd,
		&lesson.Confirmed,
		&lesson.UpdatedAt,
	)
	if err != nil {
		return StudentLesson{}, err
	}
	lesson.NextStart = lesson.NextStart.In(s.loc)
	lesson.NextEnd = lesson.NextEnd.In(s.loc)
	lesson.UpdatedAt = lesson.UpdatedAt.In(s.loc)
	lesson.ScheduleText = formatThaiSchedule(lesson.NextStart, lesson.NextEnd)
	return lesson, nil
}

func applySchedule(lesson StudentLesson, scheduleText string, loc *time.Location) StudentLesson {
	if start, end, ok := parseSchedule(scheduleText, loc); ok {
		lesson.NextStart = start
		lesson.NextEnd = end
		lesson.SessionHours = lessonHours(start, end)
		lesson.ScheduleText = formatThaiSchedule(start, end)
		return lesson
	}

	lesson.ScheduleText = strings.TrimSpace(scheduleText)
	return lesson
}

func studentNameKey(nickname, firstName string) string {
	return strings.ToLower(strings.Join([]string{
		strings.TrimSpace(nickname),
		strings.TrimSpace(firstName),
	}, "|"))
}

func firstWord(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return strings.TrimSpace(value)
	}
	return fields[0]
}

func nextWeekdayAt(loc *time.Location, weekday time.Weekday, hour int, minute int) time.Time {
	now := time.Now().In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	for next.Weekday() != weekday || !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func lessonHours(start time.Time, end time.Time) int {
	minutes := int(end.Sub(start).Minutes())
	if minutes <= 0 {
		return 1
	}
	return (minutes + 59) / 60
}

func verifyLineSignature(channelSecret string, body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func lineWebhookHandler(store LessonStore, lineClient *LineClient, loc *time.Location) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		signature := r.Header.Get("X-Line-Signature")
		channelSecret := os.Getenv("LINE_CHANNEL_SECRET")

		if channelSecret != "" {
			if !verifyLineSignature(channelSecret, body, signature) {
				log.Println("Invalid LINE signature")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		var payload LineWebhook
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Println("JSON parse error:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, event := range payload.Events {
			log.Printf("event=%s source=%s userId=%s groupId=%s text=%q", event.Type, event.Source.Type, event.Source.UserID, event.Source.GroupID, event.Message.Text)
			if event.Source.GroupID != "" {
				log.Println("STAFF GROUP ID =", event.Source.GroupID)
				if !lineClient.AllowsGroup(event.Source.GroupID) {
					log.Println("ignored LINE group not listed in LINE_GROUP_IDS:", event.Source.GroupID)
					continue
				}
				if err := store.RegisterLineGroup(event.Source.GroupID); err != nil {
					log.Println("register LINE group error:", err)
				}
			}
			if event.Type != "message" || event.Message.Type != "text" {
				continue
			}

			response, handled, err := processStaffCommand(event.Message.Text, store, loc)
			if !handled {
				continue
			}
			if err != nil {
				response = err.Error() + "\n\n" + commandHelpText()
			}

			if sendErr := sendImmediateResponse(lineClient, event, response); sendErr != nil {
				log.Println("send LINE response error:", sendErr)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func processStaffCommand(text string, store LessonStore, loc *time.Location) (string, bool, error) {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return "", false, nil
	}

	if isHelpCommand(normalized) {
		return commandHelpText(), true, nil
	}
	if isScheduleRequestCommand(normalized) {
		return formatWeeklyLessons(store.ListLessons(), time.Now().In(loc)), true, nil
	}
	if isStudentScheduleRequestCommand(normalized) {
		return processStudentScheduleRequest(normalized, store)
	}
	if strings.HasPrefix(normalized, "/") {
		if response, handled, err := processCompactSlashCommand(normalized, store); handled {
			return response, handled, err
		}
		return "", false, nil
	} else {
		return "", false, nil
	}
}

func processCompactSlashCommand(text string, store LessonStore) (string, bool, error) {
	command, err := parseCompactSlashCommand(text)
	if err != nil {
		return "", true, err
	}

	switch command.Action {
	case "update":
		if command.ScheduleText == "" {
			return "", true, errors.New("คำสั่งอัพเดทต้องเป็น: /อัพเดท ชื่อเล่น ชื่อจริง วันที่ เวลา เช่น /อัพเดท แพรว แพรวา 9/5 13:00-15:00")
		}
		lesson, err := store.UpdateLesson(command.Nickname, command.FirstName, command.ScheduleText)
		if err != nil {
			return "", true, err
		}
		return formatUpdateNotification(lesson, command.ScheduleHasYear), true, nil
	case "confirm":
		lesson, err := store.ConfirmLesson(command.Nickname, command.FirstName, command.ScheduleText)
		if err != nil {
			return "", true, err
		}
		return formatConfirmNotification(lesson, command.ScheduleHasYear), true, nil
	case "unconfirm":
		lesson, err := store.UnconfirmLesson(command.Nickname, command.FirstName, command.ScheduleText)
		if err != nil {
			return "", true, err
		}
		return formatUnconfirmNotification(lesson, command.ScheduleHasYear), true, nil
	case "leave":
		lesson, err := store.UpdateLearningStatus(command.Nickname, command.FirstName, "ลา")
		if err != nil {
			return "", true, err
		}
		return formatLearningStatusNotification(lesson, "ลา"), true, nil
	case "attend":
		lesson, err := store.UpdateLearningStatus(command.Nickname, command.FirstName, "เข้าเรียนปกติ")
		if err != nil {
			return "", true, err
		}
		return formatLearningStatusNotification(lesson, "เข้าเรียนปกติ"), true, nil
	default:
		return "", false, nil
	}
}

type compactSlashCommand struct {
	Action          string
	Nickname        string
	FirstName       string
	ScheduleText    string
	ScheduleHasYear bool
}

func parseCompactSlashCommand(text string) (compactSlashCommand, error) {
	text = strings.TrimSpace(strings.TrimPrefix(text, "/"))
	if text == "" {
		return compactSlashCommand{}, errors.New("คำสั่งต้องขึ้นต้นด้วย / เช่น /ตารางเรียน")
	}

	actionText, body, _ := strings.Cut(text, " ")
	usesSlashSeparator := false
	if strings.Contains(actionText, "/") {
		actionText, body, _ = strings.Cut(text, "/")
		usesSlashSeparator = true
	}

	action := normalizeAction(actionText)
	if action == "" {
		return compactSlashCommand{}, nil
	}

	body = strings.TrimSpace(strings.TrimLeft(body, "/"))
	if body == "" {
		return compactSlashCommand{}, errors.New("กรุณาระบุชื่อเล่นและชื่อจริงนักเรียน")
	}

	if usesSlashSeparator {
		body = strings.Replace(body, "/", " ", 2)
	}

	fields := strings.Fields(body)
	if len(fields) < 2 {
		return compactSlashCommand{}, errors.New("กรุณาระบุเป็น ชื่อเล่น ชื่อจริง เช่น /คอนเฟิร์ม แพรว แพรวา")
	}

	nickname := fields[0]
	firstName := fields[1]
	scheduleText := ""
	if len(fields) > 2 {
		scheduleText = strings.Join(fields[2:], " ")
	}

	return compactSlashCommand{
		Action:          action,
		Nickname:        strings.TrimSpace(nickname),
		FirstName:       strings.TrimSpace(firstName),
		ScheduleText:    strings.TrimSpace(scheduleText),
		ScheduleHasYear: scheduleTextHasExplicitYear(scheduleText),
	}, nil
}

func sendImmediateResponse(lineClient *LineClient, event LineEvent, response string) error {
	if strings.TrimSpace(response) == "" {
		return nil
	}
	parts := splitLongLineMessage(response, lineTextMaxLength)
	if strings.TrimSpace(event.ReplyToken) != "" {
		replyCount := len(parts)
		if replyCount > lineMessageBatchLimit {
			replyCount = lineMessageBatchLimit
		}
		if err := lineClient.ReplyTextParts(event.ReplyToken, parts[:replyCount]); err != nil {
			return err
		}
		parts = parts[replyCount:]
		if len(parts) == 0 {
			return nil
		}
	}
	targetID := strings.TrimSpace(event.Source.GroupID)
	if targetID == "" {
		targetID = strings.TrimSpace(event.Source.UserID)
	}
	if targetID != "" {
		return lineClient.SendTextParts(targetID, parts)
	}
	return lineClient.SendTextParts(lineClient.FirstTargetGroupID(), parts)
}

func splitCommandParts(text string) []string {
	rawParts := strings.Split(text, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func normalizeAction(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	action = strings.TrimPrefix(action, "/")
	action = strings.ReplaceAll(action, " ", "")
	action = strings.ReplaceAll(action, "์", "")

	switch {
	case strings.Contains(action, "ไม่คอนเฟ") || strings.Contains(action, "unconfirm") || strings.Contains(action, "notconfirm"):
		return "unconfirm"
	case action == "ลา" || strings.Contains(action, "leave") || strings.Contains(action, "absent"):
		return "leave"
	case strings.Contains(action, "เข้าเรียน") || strings.Contains(action, "attend") || strings.Contains(action, "present"):
		return "attend"
	case strings.Contains(action, "อัพ") || strings.Contains(action, "อัป") || strings.Contains(action, "update") || strings.Contains(action, "เลื่อน"):
		return "update"
	case strings.Contains(action, "คอนเฟ") || strings.Contains(action, "confirm") || strings.Contains(action, "ยืนยัน"):
		return "confirm"
	default:
		return ""
	}
}

func isHelpCommand(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, " ", "")
	return text == "help" ||
		text == "/help" ||
		text == "/วิธีใช้" ||
		text == "วิธีใช้" ||
		text == "/วิธีใช้งาน" ||
		text == "วิธีใช้งาน" ||
		text == "ตัวอย่างคำสั่ง"
}

func isScheduleRequestCommand(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, " ", "")
	return text == "/ตารางเรียน"
}

func isStudentScheduleRequestCommand(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	return strings.HasPrefix(text, "/ข้อมูลนักเรียน") || strings.HasPrefix(text, "/นักเรียน")
}

func processStudentScheduleRequest(text string, store LessonStore) (string, bool, error) {
	return formatStudentScheduleSummaries(store.ListStudentSchedules()), true, nil
}

func isConfirmWord(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "์", "")
	return strings.Contains(text, "คอนเฟ") || strings.Contains(text, "confirm") || strings.Contains(text, "ยืนยัน")
}

func commandHelpText() string {
	return strings.Join([]string{
		"วิธีใช้งาน",
		"/ตารางเรียน - ดูตารางในแท็บสัปดาห์นี้",
		"/ข้อมูลนักเรียน - ดูนักเรียนที่ยังเรียนไม่จบ",
		"/อัพเดท ชื่อเล่น ชื่อจริง 9/5 13:00-15:00",
		"/คอนเฟิร์ม ชื่อเล่น ชื่อจริง",
		"/ไม่คอนเฟิร์ม ชื่อเล่น ชื่อจริง",
		"/ลา ชื่อเล่น ชื่อจริง",
		"/เข้าเรียน ชื่อเล่น ชื่อจริง",
	}, "\n")
}

func startDailyNotifier(store LessonStore, lineClient *LineClient, loc *time.Location) {
	go func() {
		for {
			now := time.Now().In(loc)
			nextRun := nextDailyRun(now, 9, 0)
			wait := time.Until(nextRun)
			log.Printf("daily lesson notifier scheduled at %s", nextRun.Format(time.RFC3339))

			timer := time.NewTimer(wait)
			<-timer.C

			if err := notifyDailyLessons(store, lineClient, loc); err != nil {
				log.Println("daily lesson notification error:", err)
			}
		}
	}()
}

func nextDailyRun(now time.Time, hour int, minute int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func notifyDailyLessons(store LessonStore, lineClient *LineClient, loc *time.Location) error {
	groupIDs := lineClient.TargetGroupIDs()
	if len(groupIDs) == 0 {
		var err error
		groupIDs, err = store.ListLineGroupIDs()
		if err != nil {
			return err
		}
	}
	if len(groupIDs) == 0 {
		return errors.New("ยังไม่มี LINE group ที่ลงทะเบียนสำหรับ weekly notification")
	}

	message := formatWeeklyLessons(store.ListLessons(), time.Now().In(loc))
	for _, groupID := range groupIDs {
		if err := lineClient.SendText(groupID, message); err != nil {
			return fmt.Errorf("send to group %s: %w", groupID, err)
		}
	}
	return nil
}

func isLikelyLineTargetID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	lowered := strings.ToLower(value)
	if strings.Contains(lowered, "your_") || strings.Contains(lowered, "example") || strings.Contains(lowered, "placeholder") {
		return false
	}

	return strings.HasPrefix(value, "C") || strings.HasPrefix(value, "R") || strings.HasPrefix(value, "U")
}

func splitLongLineMessage(text string, maxLength int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if maxLength <= 0 {
		maxLength = lineTextMaxLength
	}
	if len([]rune(text)) <= maxLength {
		return []string{text}
	}

	lines := strings.Split(text, "\n")
	var messages []string
	var current strings.Builder

	for _, line := range lines {
		segments := splitLongLine(line, maxLength)
		if len(segments) == 0 {
			segments = []string{""}
		}
		for _, segment := range segments {
			nextLen := len([]rune(current.String())) + len([]rune(segment))
			if current.Len() > 0 {
				nextLen++
			}
			if current.Len() > 0 && nextLen > maxLength {
				messages = append(messages, strings.TrimSpace(current.String()))
				current.Reset()
			}
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			current.WriteString(segment)
		}
	}

	if strings.TrimSpace(current.String()) != "" {
		messages = append(messages, strings.TrimSpace(current.String()))
	}
	return messages
}

func splitLongLine(line string, maxLength int) []string {
	runes := []rune(line)
	if len(runes) <= maxLength {
		return []string{line}
	}
	segments := make([]string, 0, (len(runes)/maxLength)+1)
	for len(runes) > 0 {
		n := maxLength
		if len(runes) < n {
			n = len(runes)
		}
		segments = append(segments, string(runes[:n]))
		runes = runes[n:]
	}
	return segments
}

func formatWeeklyLessons(lessons []StudentLesson, now time.Time) string {
	_ = now
	weeklyLessons := make([]StudentLesson, len(lessons))
	copy(weeklyLessons, lessons)
	sort.Slice(weeklyLessons, func(i, j int) bool {
		return weeklyLessons[i].NextStart.Before(weeklyLessons[j].NextStart)
	})

	var b strings.Builder
	b.WriteString("📚 ตารางเรียนในแท็บสัปดาห์นี้")
	if dateRange := lessonDateRangeText(weeklyLessons); dateRange != "" {
		b.WriteString("\n")
		b.WriteString(dateRange)
	}

	if len(weeklyLessons) == 0 {
		b.WriteString("\n\nยังไม่มีตารางเรียนที่มีวันที่ชัดเจนในแท็บสัปดาห์นี้")
		return b.String()
	}

	for _, lesson := range weeklyLessons {
		b.WriteString("\n\n")
		b.WriteString(formatCompactLessonLine(lesson))
	}
	return b.String()
}

func lessonDateRangeText(lessons []StudentLesson) string {
	var dated []StudentLesson
	for _, lesson := range lessons {
		if !lesson.NextStart.IsZero() {
			dated = append(dated, lesson)
		}
	}
	if len(dated) == 0 {
		return ""
	}
	start := time.Date(dated[0].NextStart.Year(), dated[0].NextStart.Month(), dated[0].NextStart.Day(), 0, 0, 0, 0, dated[0].NextStart.Location())
	end := time.Date(dated[len(dated)-1].NextStart.Year(), dated[len(dated)-1].NextStart.Month(), dated[len(dated)-1].NextStart.Day(), 0, 0, 0, 0, dated[len(dated)-1].NextStart.Location())
	return formatThaiDateRange(start, end)
}

func formatCompactLessonLine(lesson StudentLesson) string {
	return formatCompactLessonLineWithYear(lesson, false)
}

func formatCompactLessonLineWithYear(lesson StudentLesson, showYear bool) string {
	statusPart := ""
	if strings.TrimSpace(lesson.LearningStatus) != "" {
		statusPart = " | " + strings.TrimSpace(lesson.LearningStatus)
	}
	return fmt.Sprintf(
		"%s %s (%s) | %s\n%s | %s%s | เหลือ %d ชม.",
		confirmEmoji(lesson),
		lesson.Nickname,
		lesson.FullName,
		lesson.Course,
		formatShortLessonTimeWithYear(lesson.NextStart, lesson.NextEnd, showYear),
		shortHourLabel(lesson),
		statusPart,
		remainingHours(lesson),
	)
}

func lessonToScheduleSummary(lesson StudentLesson) StudentScheduleSummary {
	nextLessons := ""
	if !lesson.NextStart.IsZero() {
		nextLessons = formatShortLessonTime(lesson.NextStart, lesson.NextEnd)
	}
	defaultSchedule := strings.TrimSpace(lesson.ScheduleText)
	if defaultSchedule == "" {
		defaultSchedule = nextLessons
	}
	return StudentScheduleSummary{
		Nickname:        lesson.Nickname,
		FirstName:       lesson.FirstName,
		FullName:        lesson.FullName,
		Course:          lesson.Course,
		TotalHours:      lesson.TotalHours,
		CompletedHours:  lesson.CompletedHours,
		DefaultSchedule: defaultSchedule,
		NextLessons:     nextLessons,
	}
}

func sortStudentScheduleSummaries(summaries []StudentScheduleSummary) {
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Course != summaries[j].Course {
			return summaries[i].Course < summaries[j].Course
		}
		if summaries[i].Nickname != summaries[j].Nickname {
			return summaries[i].Nickname < summaries[j].Nickname
		}
		return summaries[i].FirstName < summaries[j].FirstName
	})
}

func formatStudentScheduleSummaries(summaries []StudentScheduleSummary) string {
	if len(summaries) == 0 {
		return "ยังไม่มีนักเรียนที่กำลังเรียนอยู่"
	}

	studentGroups := groupStudentScheduleSummaries(summaries)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("👥 นักเรียนที่ยังเรียนไม่จบ (%d คน)", len(studentGroups)))
	for i, summary := range studentGroups {
		b.WriteString("\n")
		b.WriteString(formatStudentListLine(i+1, summary))
	}
	return b.String()
}

func groupStudentScheduleSummaries(summaries []StudentScheduleSummary) []StudentScheduleSummary {
	sort.Slice(summaries, func(i, j int) bool {
		if cleanClassText(summaries[i].Nickname) != cleanClassText(summaries[j].Nickname) {
			return cleanClassText(summaries[i].Nickname) < cleanClassText(summaries[j].Nickname)
		}
		if cleanClassText(summaries[i].FirstName) != cleanClassText(summaries[j].FirstName) {
			return cleanClassText(summaries[i].FirstName) < cleanClassText(summaries[j].FirstName)
		}
		return cleanClassText(summaries[i].Course) < cleanClassText(summaries[j].Course)
	})

	var grouped []StudentScheduleSummary
	groupIndexByKey := map[string]int{}
	courseSeenByKey := map[string]map[string]bool{}
	for _, summary := range summaries {
		key := strings.ToLower(cleanClassText(summary.Nickname)) + "\x00" + strings.ToLower(cleanClassText(summary.FirstName))
		if key == "\x00" {
			key = strings.ToLower(cleanClassText(summary.FullName))
		}
		index, ok := groupIndexByKey[key]
		if !ok {
			groupIndexByKey[key] = len(grouped)
			courseSeenByKey[key] = map[string]bool{}
			grouped = append(grouped, summary)
			index = len(grouped) - 1
			grouped[index].Course = ""
		}
		course := cleanClassText(summary.Course)
		if course != "" && !courseSeenByKey[key][course] {
			if grouped[index].Course == "" {
				grouped[index].Course = course
			} else {
				grouped[index].Course += ", " + course
			}
			courseSeenByKey[key][course] = true
		}
	}
	return grouped
}

func formatStudentListLine(index int, summary StudentScheduleSummary) string {
	return fmt.Sprintf(
		"%d. %s - %s | %s",
		index,
		fallbackText(summary.Nickname, "-"),
		fallbackText(summary.FirstName, fallbackText(summary.FullName, "-")),
		fallbackText(summary.Course, "-"),
	)
}

func fallbackText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func formatUpdateNotification(lesson StudentLesson, showYear bool) string {
	return "🔄 อัพเดทเวลาเรียน\n" + formatCompactLessonLineWithYear(lesson, showYear)
}

func formatConfirmNotification(lesson StudentLesson, showYear bool) string {
	return "✅ คอนเฟิร์มเวลาเรียน\n" + formatCompactLessonLineWithYear(lesson, showYear)
}

func formatUnconfirmNotification(lesson StudentLesson, showYear bool) string {
	return "⏳ ไม่คอนเฟิร์มเวลาเรียน\n" + formatCompactLessonLineWithYear(lesson, showYear)
}

func formatLearningStatusNotification(lesson StudentLesson, status string) string {
	return "📝 อัพเดทสถานะการเรียน: " + status + "\n" + formatCompactLessonLine(lesson)
}

func confirmEmoji(lesson StudentLesson) string {
	if lesson.Confirmed {
		return "✅"
	}
	return "⏳"
}

func formatShortLessonTime(start time.Time, end time.Time) string {
	return formatShortLessonTimeWithYear(start, end, false)
}

func formatShortLessonTimeWithYear(start time.Time, end time.Time, showYear bool) string {
	if showYear {
		return fmt.Sprintf(
			"%s %d %s %d %02d:%02d-%02d:%02d",
			thaiShortWeekdays[start.Weekday()],
			start.Day(),
			thaiShortMonths[start.Month()-1],
			start.Year()+543,
			start.Hour(),
			start.Minute(),
			end.Hour(),
			end.Minute(),
		)
	}
	return fmt.Sprintf(
		"%s %d %s %02d:%02d-%02d:%02d",
		thaiShortWeekdays[start.Weekday()],
		start.Day(),
		thaiShortMonths[start.Month()-1],
		start.Hour(),
		start.Minute(),
		end.Hour(),
		end.Minute(),
	)
}

func shortHourLabel(lesson StudentLesson) string {
	if lesson.CompletedHours >= lesson.TotalHours {
		return "ครบแล้ว"
	}

	start := lesson.CompletedHours + 1
	end := start + lesson.SessionHours - 1
	if end > lesson.TotalHours {
		end = lesson.TotalHours
	}
	if start == end {
		return fmt.Sprintf("ชม.%d", start)
	}
	return fmt.Sprintf("ชม.%d-%d", start, end)
}

func nextHourLabel(lesson StudentLesson) string {
	if lesson.CompletedHours >= lesson.TotalHours {
		return "เรียนครบแล้ว"
	}

	start := lesson.CompletedHours + 1
	end := start + lesson.SessionHours - 1
	if end > lesson.TotalHours {
		end = lesson.TotalHours
	}
	if start == end {
		return fmt.Sprintf("ชั่วโมงที่ %d", start)
	}
	return fmt.Sprintf("ชั่วโมงที่ %d-%d", start, end)
}

func remainingHours(lesson StudentLesson) int {
	remaining := lesson.TotalHours - lesson.CompletedHours
	if remaining < 0 {
		return 0
	}
	return remaining
}

func confirmStatus(lesson StudentLesson) string {
	if lesson.Confirmed {
		return "คอนเฟิร์มแล้ว"
	}
	return "รอคอนเฟิร์ม"
}

var thaiWeekdays = []string{
	"อาทิตย์",
	"จันทร์",
	"อังคาร",
	"พุธ",
	"พฤหัสบดี",
	"ศุกร์",
	"เสาร์",
}

var thaiMonths = []string{
	"มกราคม",
	"กุมภาพันธ์",
	"มีนาคม",
	"เมษายน",
	"พฤษภาคม",
	"มิถุนายน",
	"กรกฎาคม",
	"สิงหาคม",
	"กันยายน",
	"ตุลาคม",
	"พฤศจิกายน",
	"ธันวาคม",
}

var thaiShortWeekdays = []string{
	"อา.",
	"จ.",
	"อ.",
	"พ.",
	"พฤ.",
	"ศ.",
	"ส.",
}

var thaiShortMonths = []string{
	"ม.ค.",
	"ก.พ.",
	"มี.ค.",
	"เม.ย.",
	"พ.ค.",
	"มิ.ย.",
	"ก.ค.",
	"ส.ค.",
	"ก.ย.",
	"ต.ค.",
	"พ.ย.",
	"ธ.ค.",
}

func formatThaiDate(t time.Time) string {
	t = t.In(t.Location())
	return fmt.Sprintf("วัน%s %d %s %d", thaiWeekdays[t.Weekday()], t.Day(), thaiMonths[t.Month()-1], t.Year()+543)
}

func formatThaiDateRange(start time.Time, end time.Time) string {
	start = start.In(start.Location())
	end = end.In(start.Location())

	if start.Year() == end.Year() && start.Month() == end.Month() {
		return fmt.Sprintf("%d-%d %s %d", start.Day(), end.Day(), thaiShortMonths[start.Month()-1], start.Year()+543)
	}
	if start.Year() == end.Year() {
		return fmt.Sprintf("%d %s-%d %s %d", start.Day(), thaiShortMonths[start.Month()-1], end.Day(), thaiShortMonths[end.Month()-1], start.Year()+543)
	}
	return fmt.Sprintf("%d %s %d-%d %s %d", start.Day(), thaiShortMonths[start.Month()-1], start.Year()+543, end.Day(), thaiShortMonths[end.Month()-1], end.Year()+543)
}

func formatThaiSchedule(start time.Time, end time.Time) string {
	return fmt.Sprintf(
		"%s เวลา %02d.%02d-%02d.%02d น.",
		formatThaiDate(start),
		start.Hour(),
		start.Minute(),
		end.Hour(),
		end.Minute(),
	)
}

func parseSchedule(text string, loc *time.Location) (time.Time, time.Time, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, time.Time{}, false
	}

	if start, end, ok := parseISODateSchedule(text, loc); ok {
		return start, end, true
	}
	if start, end, ok := parseSlashDateSchedule(text, loc); ok {
		return start, end, true
	}
	if start, end, ok := parseThaiDateSchedule(text, loc); ok {
		return start, end, true
	}
	return time.Time{}, time.Time{}, false
}

func scheduleTextHasExplicitYear(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\d{4}-\d{1,2}-\d{1,2}`),
		regexp.MustCompile(`\d{1,2}[/-]\d{1,2}[/-]\d{2,4}`),
		regexp.MustCompile(`\d{1,2}\s*[ก-๙.]+\s+\d{2,4}(?:\s|$)`),
	}
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func parseISODateSchedule(text string, loc *time.Location) (time.Time, time.Time, bool) {
	re := regexp.MustCompile(`(\d{4})-(\d{1,2})-(\d{1,2}).*?(\d{1,2})[:.](\d{2})\s*[-–]\s*(\d{1,2})[:.](\d{2})`)
	matches := re.FindStringSubmatch(text)
	if len(matches) != 8 {
		return time.Time{}, time.Time{}, false
	}

	year := normalizeYear(mustAtoi(matches[1]))
	month := time.Month(mustAtoi(matches[2]))
	day := mustAtoi(matches[3])
	startHour := mustAtoi(matches[4])
	startMinute := mustAtoi(matches[5])
	endHour := mustAtoi(matches[6])
	endMinute := mustAtoi(matches[7])
	return buildSchedule(year, month, day, startHour, startMinute, endHour, endMinute, loc)
}

func parseSlashDateSchedule(text string, loc *time.Location) (time.Time, time.Time, bool) {
	re := regexp.MustCompile(`(\d{1,2})[/-](\d{1,2})(?:[/-](\d{2,4}))?.*?(\d{1,2})[:.](\d{2})\s*[-–]\s*(\d{1,2})[:.](\d{2})`)
	matches := re.FindStringSubmatch(text)
	if len(matches) != 8 {
		return time.Time{}, time.Time{}, false
	}

	day := mustAtoi(matches[1])
	month := time.Month(mustAtoi(matches[2]))
	year := time.Now().In(loc).Year()
	if strings.TrimSpace(matches[3]) != "" {
		year = normalizeYear(mustAtoi(matches[3]))
	}
	startHour := mustAtoi(matches[4])
	startMinute := mustAtoi(matches[5])
	endHour := mustAtoi(matches[6])
	endMinute := mustAtoi(matches[7])
	start, end, ok := buildSchedule(year, month, day, startHour, startMinute, endHour, endMinute, loc)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func parseThaiDateSchedule(text string, loc *time.Location) (time.Time, time.Time, bool) {
	re := regexp.MustCompile(`(\d{1,2})\s*([ก-๙.]+)\s*(\d{2,4})?.*?(\d{1,2})[:.](\d{2})\s*[-–]\s*(\d{1,2})[:.](\d{2})`)
	matches := re.FindStringSubmatch(text)
	if len(matches) != 8 {
		return time.Time{}, time.Time{}, false
	}

	month, ok := parseThaiMonth(matches[2])
	if !ok {
		return time.Time{}, time.Time{}, false
	}

	day := mustAtoi(matches[1])
	year := time.Now().In(loc).Year()
	if strings.TrimSpace(matches[3]) != "" {
		year = normalizeYear(mustAtoi(matches[3]))
	}
	startHour := mustAtoi(matches[4])
	startMinute := mustAtoi(matches[5])
	endHour := mustAtoi(matches[6])
	endMinute := mustAtoi(matches[7])

	start, end, ok := buildSchedule(year, month, day, startHour, startMinute, endHour, endMinute, loc)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func buildSchedule(year int, month time.Month, day int, startHour int, startMinute int, endHour int, endMinute int, loc *time.Location) (time.Time, time.Time, bool) {
	start := time.Date(year, month, day, startHour, startMinute, 0, 0, loc)
	end := time.Date(year, month, day, endHour, endMinute, 0, 0, loc)
	if end.Before(start) || end.Equal(start) {
		end = end.AddDate(0, 0, 1)
	}

	if start.Year() != year || start.Month() != month || start.Day() != day {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func normalizeYear(year int) int {
	if year < 100 {
		return 2000 + year
	}
	if year > 2400 {
		return year - 543
	}
	return year
}

func parseThaiMonth(value string) (time.Month, bool) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), ".", "")
	normalized = strings.ReplaceAll(normalized, " ", "")
	months := map[string]time.Month{
		"มกราคม":     time.January,
		"มค":         time.January,
		"กุมภาพันธ์": time.February,
		"กพ":         time.February,
		"มีนาคม":     time.March,
		"มีค":        time.March,
		"เมษายน":     time.April,
		"เมย":        time.April,
		"พฤษภาคม":    time.May,
		"พค":         time.May,
		"มิถุนายน":   time.June,
		"มิย":        time.June,
		"กรกฎาคม":    time.July,
		"กค":         time.July,
		"สิงหาคม":    time.August,
		"สค":         time.August,
		"กันยายน":    time.September,
		"กย":         time.September,
		"ตุลาคม":     time.October,
		"ตค":         time.October,
		"พฤศจิกายน":  time.November,
		"พย":         time.November,
		"ธันวาคม":    time.December,
		"ธค":         time.December,
	}
	month, ok := months[normalized]
	return month, ok
}

func mustAtoi(value string) int {
	number, _ := strconv.Atoi(value)
	return number
}

func newLessonStore(loc *time.Location, lineClient *LineClient) LessonStore {
	googleSheetID := strings.TrimSpace(os.Getenv("GOOGLE_SHEET_ID"))
	if googleSheetID != "" {
		store, err := NewClassScheduleSheetsLessonStore(googleSheetID, loc)
		if err != nil {
			log.Fatal("connect Google Sheets error:", err)
		}
		registerConfiguredLineGroups(store, lineClient)
		log.Println("Using Class Schedule Google Sheets lesson store")
		return store
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Println("DATABASE_URL is empty; using in-memory mock lesson store")
		store := NewMockLessonStore(loc)
		registerConfiguredLineGroups(store, lineClient)
		return store
	}

	store, err := NewPostgresLessonStore(databaseURL, loc)
	if err != nil {
		log.Fatal("connect database error:", err)
	}

	if !strings.EqualFold(os.Getenv("AUTO_MIGRATE"), "false") {
		schemaPath := strings.TrimSpace(os.Getenv("DB_SCHEMA_PATH"))
		if schemaPath == "" {
			schemaPath = "db/schema.sql"
		}
		if err := store.Migrate(schemaPath); err != nil {
			log.Fatal("database migration error:", err)
		}
	}

	registerConfiguredLineGroups(store, lineClient)
	log.Println("Using PostgreSQL lesson store")
	return store
}

func registerConfiguredLineGroups(store LessonStore, lineClient *LineClient) {
	for _, groupID := range lineClient.TargetGroupIDs() {
		if err := store.RegisterLineGroup(groupID); err != nil {
			log.Println("register configured LINE group error:", err)
		}
	}
}

func main() {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}

	lineClient := NewLineClient()
	store := newLessonStore(loc, lineClient)

	startDailyNotifier(store, lineClient, loc)

	if strings.EqualFold(os.Getenv("RUN_DAILY_ON_START"), "true") {
		if err := notifyDailyLessons(store, lineClient, loc); err != nil {
			log.Println("daily lesson notification error:", err)
		}
	}

	http.HandleFunc("/line/webhook", lineWebhookHandler(store, lineClient, loc))

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	log.Println("Server started on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
