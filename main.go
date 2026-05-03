package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
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
)

const (
	linePushURL  = "https://api.line.me/v2/bot/message/push"
	lineReplyURL = "https://api.line.me/v2/bot/message/reply"
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
	httpClient     *http.Client
}

type lineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewLineClient() *LineClient {
	return &LineClient{
		accessToken:    os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
		defaultGroupID: os.Getenv("LINE_STAFF_GROUP_ID"),
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *LineClient) SendText(to string, text string) error {
	if !isLikelyLineTargetID(to) {
		return errors.New("missing or invalid LINE target ID")
	}
	return c.send(linePushURL, map[string]any{
		"to":       to,
		"messages": []lineTextMessage{{Type: "text", Text: text}},
	})
}

func (c *LineClient) ReplyText(replyToken string, text string) error {
	if strings.TrimSpace(replyToken) == "" {
		return errors.New("missing LINE reply token")
	}
	return c.send(lineReplyURL, map[string]any{
		"replyToken": replyToken,
		"messages":   []lineTextMessage{{Type: "text", Text: text}},
	})
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
	FullName       string
	Course         string
	TotalHours     int
	CompletedHours int
	SessionHours   int
	NextStart      time.Time
	NextEnd        time.Time
	ScheduleText   string
	Confirmed      bool
	UpdatedAt      time.Time
}

type LessonStore interface {
	ListLessons() []StudentLesson
	UpdateLesson(nickname, fullName, course, scheduleText string) (StudentLesson, error)
	ConfirmLesson(nickname, fullName, course, scheduleText string) (StudentLesson, error)
	FindLessonByNickname(nickname string) (StudentLesson, error)
}

type MockLessonStore struct {
	mu      sync.RWMutex
	lessons map[string]StudentLesson
	loc     *time.Location
}

func NewMockLessonStore(loc *time.Location) *MockLessonStore {
	store := &MockLessonStore{
		lessons: map[string]StudentLesson{},
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
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[lessonKey(lesson.Nickname, lesson.FullName, lesson.Course)] = lesson
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

func (s *MockLessonStore) UpdateLesson(nickname, fullName, course, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := lessonKey(nickname, fullName, course)
	lesson, ok := s.lessons[key]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s / %s", nickname, fullName, course)
	}

	lesson = applySchedule(lesson, scheduleText, s.loc)
	lesson.Confirmed = false
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[key] = lesson
	return lesson, nil
}

func (s *MockLessonStore) ConfirmLesson(nickname, fullName, course, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := lessonKey(nickname, fullName, course)
	lesson, ok := s.lessons[key]
	if !ok {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนใน mock database: %s / %s / %s", nickname, fullName, course)
	}

	if strings.TrimSpace(scheduleText) != "" {
		lesson = applySchedule(lesson, scheduleText, s.loc)
	}
	lesson.Confirmed = true
	lesson.UpdatedAt = time.Now().In(s.loc)
	s.lessons[key] = lesson
	return lesson, nil
}

func (s *MockLessonStore) FindLessonByNickname(nickname string) (StudentLesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return StudentLesson{}, errors.New("กรุณาระบุชื่อเล่นนักเรียน")
	}

	var matches []StudentLesson
	for _, lesson := range s.lessons {
		if strings.EqualFold(lesson.Nickname, nickname) {
			matches = append(matches, lesson)
		}
	}
	if len(matches) == 0 {
		return StudentLesson{}, fmt.Errorf("ไม่พบนักเรียนชื่อเล่น %s ใน mock database", nickname)
	}
	if len(matches) > 1 {
		return StudentLesson{}, fmt.Errorf("พบชื่อเล่น %s มากกว่า 1 คน กรุณาใช้คำสั่งแบบเต็ม", nickname)
	}
	return matches[0], nil
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

func lessonKey(nickname, fullName, course string) string {
	return strings.ToLower(strings.Join([]string{
		strings.TrimSpace(nickname),
		strings.TrimSpace(fullName),
		strings.TrimSpace(course),
	}, "|"))
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
	if strings.HasPrefix(normalized, "/") {
		if response, handled, err := processCompactSlashCommand(normalized, store); handled {
			return response, handled, err
		}
	} else {
		return "", false, nil
	}

	parts := splitCommandParts(normalized)
	if len(parts) == 0 {
		return "", false, nil
	}

	action := normalizeAction(parts[0])
	switch action {
	case "update":
		if len(parts) < 5 {
			return "", true, errors.New("คำสั่งอัพเดทต้องขึ้นต้นด้วย / เช่น /อัพเดท แพรว 9/5 13:00-15:00")
		}
		lesson, err := store.UpdateLesson(parts[1], parts[2], parts[3], parts[4])
		if err != nil {
			return "", true, err
		}
		return formatUpdateNotification(lesson), true, nil
	case "confirm":
		if len(parts) < 5 {
			return "", true, errors.New("คำสั่งคอนเฟิร์มต้องขึ้นต้นด้วย / เช่น /คอนเฟิร์ม แพรว")
		}
		if len(parts) >= 6 && !isConfirmWord(parts[5]) {
			return "", true, errors.New("ท้ายคำสั่งคอนเฟิร์มควรเป็นคำว่า คอนเฟิร์ม หรือ ยืนยัน")
		}
		lesson, err := store.ConfirmLesson(parts[1], parts[2], parts[3], parts[4])
		if err != nil {
			return "", true, err
		}
		return formatConfirmNotification(lesson), true, nil
	default:
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
			return "", true, errors.New("คำสั่งอัพเดทต้องเป็น: /อัพเดท ชื่อเล่น วันที่ เวลา เช่น /อัพเดท แพรว 9/5 13:00-15:00")
		}
		lesson, err := store.FindLessonByNickname(command.Nickname)
		if err != nil {
			return "", true, err
		}
		lesson, err = store.UpdateLesson(lesson.Nickname, lesson.FullName, lesson.Course, command.ScheduleText)
		if err != nil {
			return "", true, err
		}
		return formatUpdateNotification(lesson), true, nil
	case "confirm":
		lesson, err := store.FindLessonByNickname(command.Nickname)
		if err != nil {
			return "", true, err
		}
		lesson, err = store.ConfirmLesson(lesson.Nickname, lesson.FullName, lesson.Course, command.ScheduleText)
		if err != nil {
			return "", true, err
		}
		return formatConfirmNotification(lesson), true, nil
	default:
		return "", false, nil
	}
}

type compactSlashCommand struct {
	Action       string
	Nickname     string
	ScheduleText string
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
		return compactSlashCommand{}, errors.New("กรุณาระบุชื่อเล่นนักเรียน")
	}

	var nickname string
	var scheduleText string
	if usesSlashSeparator {
		parts := splitCommandParts(body)
		if len(parts) >= 4 {
			nickname = parts[0]
			scheduleText = parts[3]
		} else {
			nickname, scheduleText, _ = strings.Cut(body, "/")
		}
	} else {
		fields := strings.Fields(body)
		if len(fields) > 0 {
			nickname = fields[0]
			scheduleText = strings.TrimSpace(strings.TrimPrefix(body, nickname))
		}
	}

	return compactSlashCommand{
		Action:       action,
		Nickname:     strings.TrimSpace(nickname),
		ScheduleText: strings.TrimSpace(scheduleText),
	}, nil
}

func sendImmediateResponse(lineClient *LineClient, event LineEvent, response string) error {
	if strings.TrimSpace(response) == "" {
		return nil
	}
	if strings.TrimSpace(event.ReplyToken) != "" {
		return lineClient.ReplyText(event.ReplyToken, response)
	}
	if strings.TrimSpace(event.Source.GroupID) != "" {
		return lineClient.SendText(event.Source.GroupID, response)
	}
	return lineClient.SendText(lineClient.defaultGroupID, response)
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
	return text == "help" || text == "/help" || text == "/วิธีใช้" || text == "วิธีใช้" || text == "ตัวอย่างคำสั่ง"
}

func isScheduleRequestCommand(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, " ", "")
	return text == "/ตารางเรียน"
}

func isConfirmWord(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "์", "")
	return strings.Contains(text, "คอนเฟ") || strings.Contains(text, "confirm") || strings.Contains(text, "ยืนยัน")
}

func commandHelpText() string {
	return strings.Join([]string{
		"ตัวอย่างคำสั่ง",
		"/ตารางเรียน",
		"/อัพเดท แพรว 9/5 13:00-15:00",
		"/คอนเฟิร์ม แพรว",
		"/คอนเฟิร์ม แพรว 9/5 13:00-15:00",
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
	targetGroupID := strings.TrimSpace(lineClient.defaultGroupID)
	if !isLikelyLineTargetID(targetGroupID) {
		return errors.New("missing or invalid LINE_STAFF_GROUP_ID for weekly notification")
	}

	message := formatWeeklyLessons(store.ListLessons(), time.Now().In(loc))
	for _, part := range splitLongLineMessage(message, 4500) {
		if err := lineClient.SendText(targetGroupID, part); err != nil {
			return err
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
	if len([]rune(text)) <= maxLength {
		return []string{text}
	}

	lines := strings.Split(text, "\n")
	var messages []string
	var current strings.Builder

	for _, line := range lines {
		nextLen := len([]rune(current.String())) + len([]rune(line)) + 1
		if current.Len() > 0 && nextLen > maxLength {
			messages = append(messages, strings.TrimSpace(current.String()))
			current.Reset()
		}
		current.WriteString(line)
		current.WriteString("\n")
	}

	if strings.TrimSpace(current.String()) != "" {
		messages = append(messages, strings.TrimSpace(current.String()))
	}
	return messages
}

func weekRange(now time.Time) (time.Time, time.Time) {
	now = now.In(now.Location())
	weekdayOffset := (int(now.Weekday()) + 6) % 7
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -weekdayOffset)
	return start, start.AddDate(0, 0, 7)
}

func formatWeeklyLessons(lessons []StudentLesson, now time.Time) string {
	weekStart, weekEnd := weekRange(now)
	weeklyLessons := filterLessonsInRange(lessons, weekStart, weekEnd)

	var b strings.Builder
	b.WriteString("📚 ตารางเรียนสัปดาห์นี้\n")
	b.WriteString(formatThaiDateRange(weekStart, weekEnd.AddDate(0, 0, -1)))

	if len(weeklyLessons) == 0 {
		b.WriteString("\n\nยังไม่มีตารางเรียนในสัปดาห์นี้")
		return b.String()
	}

	for _, lesson := range weeklyLessons {
		b.WriteString("\n\n")
		b.WriteString(formatCompactLessonLine(lesson))
	}
	return b.String()
}

func filterLessonsInRange(lessons []StudentLesson, start time.Time, end time.Time) []StudentLesson {
	filtered := make([]StudentLesson, 0, len(lessons))
	for _, lesson := range lessons {
		if !lesson.NextStart.Before(start) && lesson.NextStart.Before(end) {
			filtered = append(filtered, lesson)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].NextStart.Before(filtered[j].NextStart)
	})
	return filtered
}

func formatCompactLessonLine(lesson StudentLesson) string {
	return fmt.Sprintf(
		"%s %s (%s) | %s\n%s | %s | เหลือ %d ชม.",
		confirmEmoji(lesson),
		lesson.Nickname,
		lesson.FullName,
		lesson.Course,
		formatShortLessonTime(lesson.NextStart, lesson.NextEnd),
		shortHourLabel(lesson),
		remainingHours(lesson),
	)
}

func formatUpdateNotification(lesson StudentLesson) string {
	return "🔄 อัพเดทเวลาเรียน\n" + formatCompactLessonLine(lesson)
}

func formatConfirmNotification(lesson StudentLesson) string {
	return "✅ คอนเฟิร์มเวลาเรียน\n" + formatCompactLessonLine(lesson)
}

func confirmEmoji(lesson StudentLesson) string {
	if lesson.Confirmed {
		return "✅"
	}
	return "⏳"
}

func formatShortLessonTime(start time.Time, end time.Time) string {
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
	if strings.TrimSpace(matches[3]) == "" && start.Before(time.Now().In(loc).Add(-24*time.Hour)) {
		return buildSchedule(year+1, month, day, startHour, startMinute, endHour, endMinute, loc)
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
	if strings.TrimSpace(matches[3]) == "" && start.Before(time.Now().In(loc).Add(-24*time.Hour)) {
		return buildSchedule(year+1, month, day, startHour, startMinute, endHour, endMinute, loc)
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

func main() {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}

	store := NewMockLessonStore(loc)
	lineClient := NewLineClient()

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
