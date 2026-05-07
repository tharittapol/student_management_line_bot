package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const googleSheetsScope = "https://www.googleapis.com/auth/spreadsheets"

var googleSheetHeaders = map[string][]string{
	"courses": {
		"id", "name", "description", "default_total_hours", "default_session_hours", "created_at", "updated_at",
	},
	"students": {
		"id", "source_key", "nickname", "first_name", "full_name_th", "full_name_en", "gender", "age", "nationality", "parent_phone", "parent_name", "line_name", "email", "note", "created_at", "updated_at",
	},
	"enrollments": {
		"id", "student_id", "course_id", "total_hours", "completed_hours", "default_session_hours", "teacher", "started_on", "external_status", "active", "note", "created_at", "updated_at",
	},
	"course_default_schedules": {
		"id", "course_id", "sequence_no", "default_day_text", "session_label", "date_label", "scheduled_date", "slot_label", "start_time", "end_time", "active", "note", "created_at", "updated_at",
	},
	"enrollment_default_schedules": {
		"id", "enrollment_id", "weekday_text", "start_time", "end_time", "active", "note", "created_at", "updated_at",
	},
	"lesson_sessions": {
		"id", "enrollment_id", "sequence_no", "start_at", "end_at", "status", "note", "created_at", "updated_at",
	},
	"enrollment_schedule_notes": {
		"id", "enrollment_id", "course_default_schedule_id", "sequence_no", "source_text", "status", "note", "created_at", "updated_at",
	},
	"line_groups": {
		"id", "group_id", "display_name", "active", "first_seen_at", "last_seen_at", "created_at", "updated_at",
	},
}

var googleSheetTableOrder = []string{
	"courses",
	"students",
	"enrollments",
	"course_default_schedules",
	"enrollment_default_schedules",
	"lesson_sessions",
	"enrollment_schedule_notes",
	"line_groups",
}

type googleServiceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

type GoogleSheetsClient struct {
	spreadsheetID string
	service       googleServiceAccount
	privateKey    *rsa.PrivateKey
	httpClient    *http.Client
	mu            sync.Mutex
	accessToken   string
	tokenExpiry   time.Time
}

func NewGoogleSheetsClientFromEnv(spreadsheetID string) (*GoogleSheetsClient, error) {
	serviceJSON, err := loadGoogleServiceAccountJSON()
	if err != nil {
		return nil, err
	}

	var service googleServiceAccount
	if err := json.Unmarshal(serviceJSON, &service); err != nil {
		return nil, fmt.Errorf("parse Google service account JSON: %w", err)
	}
	if strings.TrimSpace(service.ClientEmail) == "" || strings.TrimSpace(service.PrivateKey) == "" {
		return nil, errors.New("Google service account JSON must include client_email and private_key")
	}
	if strings.TrimSpace(service.TokenURI) == "" {
		service.TokenURI = "https://oauth2.googleapis.com/token"
	}

	privateKey, err := parseRSAPrivateKey(service.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &GoogleSheetsClient{
		spreadsheetID: strings.TrimSpace(spreadsheetID),
		service:       service,
		privateKey:    privateKey,
		httpClient:    &http.Client{Timeout: 20 * time.Second},
	}, nil
}

func loadGoogleServiceAccountJSON() ([]byte, error) {
	if raw := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON")); raw != "" {
		return []byte(raw), nil
	}
	if raw := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON_BASE64")); raw != "" {
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("decode GOOGLE_SERVICE_ACCOUNT_JSON_BASE64: %w", err)
		}
		return data, nil
	}
	path := strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON_PATH"))
	if path == "" {
		return nil, errors.New("set GOOGLE_SERVICE_ACCOUNT_JSON_BASE64 or GOOGLE_SERVICE_ACCOUNT_JSON_PATH for Google Sheets")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Google service account JSON: %w", err)
	}
	return data, nil
}

func parseRSAPrivateKey(value string) (*rsa.PrivateKey, error) {
	value = strings.ReplaceAll(value, `\n`, "\n")
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("parse Google private key: PEM block not found")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if privateKey, ok := key.(*rsa.PrivateKey); ok {
			return privateKey, nil
		}
		return nil, errors.New("Google private key is not RSA")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse Google private key: %w", err)
	}
	return privateKey, nil
}

func (c *GoogleSheetsClient) token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-1*time.Minute)) {
		return c.accessToken, nil
	}

	now := time.Now()
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	claims := map[string]any{
		"iss":   c.service.ClientEmail,
		"scope": googleSheetsScope,
		"aud":   c.service.TokenURI,
		"exp":   now.Add(time.Hour).Unix(),
		"iat":   now.Unix(),
	}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	assertion := unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.service.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Google OAuth returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.AccessToken == "" {
		return "", errors.New("Google OAuth did not return access_token")
	}
	c.accessToken = parsed.AccessToken
	c.tokenExpiry = now.Add(time.Duration(parsed.ExpiresIn) * time.Second)
	return c.accessToken, nil
}

func (c *GoogleSheetsClient) doJSON(ctx context.Context, method string, apiURL string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, body)
	if err != nil {
		return err
	}
	token, err := c.token(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Google Sheets API returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(responseBody, out)
}

func (c *GoogleSheetsClient) spreadsheetURL(path string) string {
	return "https://sheets.googleapis.com/v4/spreadsheets/" + url.PathEscape(c.spreadsheetID) + path
}

func (c *GoogleSheetsClient) spreadsheetTitles(ctx context.Context) (map[string]bool, error) {
	var out struct {
		Sheets []struct {
			Properties struct {
				Title string `json:"title"`
			} `json:"properties"`
		} `json:"sheets"`
	}
	err := c.doJSON(ctx, http.MethodGet, c.spreadsheetURL("?fields=sheets.properties.title"), nil, &out)
	if err != nil {
		return nil, err
	}
	titles := map[string]bool{}
	for _, sheet := range out.Sheets {
		titles[sheet.Properties.Title] = true
	}
	return titles, nil
}

func (c *GoogleSheetsClient) addSheet(ctx context.Context, title string) error {
	payload := map[string]any{
		"requests": []map[string]any{
			{
				"addSheet": map[string]any{
					"properties": map[string]any{"title": title},
				},
			},
		},
	}
	return c.doJSON(ctx, http.MethodPost, c.spreadsheetURL(":batchUpdate"), payload, nil)
}

func (c *GoogleSheetsClient) valuesGet(ctx context.Context, rangeName string) ([][]string, error) {
	var out struct {
		Values [][]string `json:"values"`
	}
	apiURL := c.spreadsheetURL("/values/" + url.PathEscape(rangeName) + "?majorDimension=ROWS")
	err := c.doJSON(ctx, http.MethodGet, apiURL, nil, &out)
	return out.Values, err
}

func (c *GoogleSheetsClient) valuesUpdate(ctx context.Context, rangeName string, values [][]string) error {
	payload := map[string]any{"majorDimension": "ROWS", "values": values}
	apiURL := c.spreadsheetURL("/values/" + url.PathEscape(rangeName) + "?valueInputOption=RAW")
	return c.doJSON(ctx, http.MethodPut, apiURL, payload, nil)
}

func (c *GoogleSheetsClient) valuesAppend(ctx context.Context, rangeName string, values [][]string) error {
	if len(values) == 0 {
		return nil
	}
	payload := map[string]any{"majorDimension": "ROWS", "values": values}
	apiURL := c.spreadsheetURL("/values/" + url.PathEscape(rangeName) + ":append?valueInputOption=RAW&insertDataOption=INSERT_ROWS")
	return c.doJSON(ctx, http.MethodPost, apiURL, payload, nil)
}

type sheetRecord struct {
	rowNumber int
	values    map[string]string
}

type GoogleSheetsLessonStore struct {
	client *GoogleSheetsClient
	loc    *time.Location
	mu     sync.Mutex
}

func NewGoogleSheetsLessonStore(spreadsheetID string, loc *time.Location) (*GoogleSheetsLessonStore, error) {
	client, err := NewGoogleSheetsClientFromEnv(spreadsheetID)
	if err != nil {
		return nil, err
	}
	store := &GoogleSheetsLessonStore{client: client, loc: loc}
	if !strings.EqualFold(os.Getenv("GOOGLE_SHEETS_INIT_SCHEMA"), "false") {
		if err := store.EnsureSchema(context.Background()); err != nil {
			return nil, err
		}
	}
	if seedDir := strings.TrimSpace(os.Getenv("GOOGLE_SHEETS_SEED_DIR")); seedDir != "" {
		if err := store.SeedFromCSVDir(context.Background(), seedDir); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *GoogleSheetsLessonStore) EnsureSchema(ctx context.Context) error {
	titles, err := s.client.spreadsheetTitles(ctx)
	if err != nil {
		return err
	}
	for _, table := range googleSheetTableOrder {
		if !titles[table] {
			if err := s.client.addSheet(ctx, table); err != nil {
				return fmt.Errorf("create sheet %s: %w", table, err)
			}
		}
		headers := googleSheetHeaders[table]
		headerRange := fmt.Sprintf("%s!A1:%s1", quoteSheetName(table), columnName(len(headers)))
		current, err := s.client.valuesGet(ctx, headerRange)
		if err != nil {
			return fmt.Errorf("read header %s: %w", table, err)
		}
		if len(current) == 0 || !sameStringSlice(current[0], headers) {
			if err := s.client.valuesUpdate(ctx, headerRange, [][]string{headers}); err != nil {
				return fmt.Errorf("write header %s: %w", table, err)
			}
		}
	}
	return nil
}

func (s *GoogleSheetsLessonStore) SeedFromCSVDir(ctx context.Context, dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, table := range googleSheetTableOrder {
		existing, err := s.loadTable(ctx, table)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		path := filepath.Join(dir, table+".csv")
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1
		rows, readErr := reader.ReadAll()
		closeErr := file.Close()
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		if closeErr != nil {
			return closeErr
		}
		if len(rows) <= 1 {
			continue
		}
		headers := googleSheetHeaders[table]
		if !sameStringSlice(rows[0], headers) {
			return fmt.Errorf("seed %s header mismatch", path)
		}
		if err := s.client.valuesAppend(ctx, quoteSheetName(table)+"!A1", rows[1:]); err != nil {
			return fmt.Errorf("seed %s: %w", table, err)
		}
	}
	return nil
}

func (s *GoogleSheetsLessonStore) loadTable(ctx context.Context, table string) ([]sheetRecord, error) {
	values, err := s.client.valuesGet(ctx, quoteSheetName(table)+"!A:"+columnName(len(googleSheetHeaders[table])))
	if err != nil {
		return nil, fmt.Errorf("read sheet %s: %w", table, err)
	}
	if len(values) == 0 {
		return nil, nil
	}
	headers := values[0]
	var records []sheetRecord
	for rowIndex, row := range values[1:] {
		empty := true
		mapped := map[string]string{}
		for i, header := range headers {
			if strings.TrimSpace(header) == "" {
				continue
			}
			if i < len(row) {
				mapped[header] = strings.TrimSpace(row[i])
				if strings.TrimSpace(row[i]) != "" {
					empty = false
				}
			}
		}
		if empty {
			continue
		}
		records = append(records, sheetRecord{rowNumber: rowIndex + 2, values: mapped})
	}
	return records, nil
}

func (s *GoogleSheetsLessonStore) appendRecord(ctx context.Context, table string, values map[string]string) error {
	row := valuesForHeaders(googleSheetHeaders[table], values)
	return s.client.valuesAppend(ctx, quoteSheetName(table)+"!A1", [][]string{row})
}

func (s *GoogleSheetsLessonStore) updateRecord(ctx context.Context, table string, record sheetRecord) error {
	headers := googleSheetHeaders[table]
	rowRange := fmt.Sprintf("%s!A%d:%s%d", quoteSheetName(table), record.rowNumber, columnName(len(headers)), record.rowNumber)
	return s.client.valuesUpdate(ctx, rowRange, [][]string{valuesForHeaders(headers, record.values)})
}

func (s *GoogleSheetsLessonStore) loadCore(ctx context.Context) (map[string][]sheetRecord, error) {
	tables := map[string][]sheetRecord{}
	for _, table := range googleSheetTableOrder {
		records, err := s.loadTable(ctx, table)
		if err != nil {
			return nil, err
		}
		tables[table] = records
	}
	return tables, nil
}

func (s *GoogleSheetsLessonStore) ListLessons() []StudentLesson {
	ctx := context.Background()
	tables, err := s.loadCore(ctx)
	if err != nil {
		log.Println("list Google Sheets lessons error:", err)
		return nil
	}

	students := recordsByID(tables["students"])
	courses := recordsByID(tables["courses"])
	enrollments := recordsByID(tables["enrollments"])
	now := time.Now().In(s.loc)
	from := now.AddDate(0, 0, -30)
	to := now.AddDate(0, 0, 365)

	var lessons []StudentLesson
	for _, session := range tables["lesson_sessions"] {
		status := rowValue(session, "status")
		if status == "cancelled" {
			continue
		}
		enrollment, ok := enrollments[rowValue(session, "enrollment_id")]
		if !ok || !parseSheetBool(rowValue(enrollment, "active"), true) {
			continue
		}
		start, end, ok := parseSheetDateTimeRange(rowValue(session, "start_at"), rowValue(session, "end_at"), s.loc)
		if !ok || start.Before(from) || !start.Before(to) {
			continue
		}
		student := students[rowValue(enrollment, "student_id")]
		course := courses[rowValue(enrollment, "course_id")]
		lessons = append(lessons, buildSheetLesson(session, enrollment, student, course, start, end, s.loc))
	}
	sort.Slice(lessons, func(i, j int) bool {
		return lessons[i].NextStart.Before(lessons[j].NextStart)
	})
	return lessons
}

func (s *GoogleSheetsLessonStore) ListStudentSchedules() []StudentScheduleSummary {
	summaries, err := s.listStudentSchedules(context.Background(), "", "")
	if err != nil {
		log.Println("list Google Sheets student schedules error:", err)
		return nil
	}
	return summaries
}

func (s *GoogleSheetsLessonStore) FindStudentSchedules(nickname, firstName string) ([]StudentScheduleSummary, error) {
	summaries, err := s.listStudentSchedules(context.Background(), nickname, firstName)
	if err != nil {
		return nil, err
	}
	if len(summaries) == 0 {
		return nil, fmt.Errorf("ไม่พบนักเรียน: %s / %s", nickname, firstName)
	}
	return summaries, nil
}

func (s *GoogleSheetsLessonStore) listStudentSchedules(ctx context.Context, nickname, firstName string) ([]StudentScheduleSummary, error) {
	tables, err := s.loadCore(ctx)
	if err != nil {
		return nil, err
	}
	students := recordsByID(tables["students"])
	courses := recordsByID(tables["courses"])
	now := time.Now().In(s.loc)

	var summaries []StudentScheduleSummary
	for _, enrollment := range tables["enrollments"] {
		if !parseSheetBool(rowValue(enrollment, "active"), true) {
			continue
		}
		student := students[rowValue(enrollment, "student_id")]
		if strings.TrimSpace(nickname) != "" && !strings.EqualFold(rowValue(student, "nickname"), strings.TrimSpace(nickname)) {
			continue
		}
		if strings.TrimSpace(firstName) != "" && !strings.EqualFold(rowValue(student, "first_name"), strings.TrimSpace(firstName)) {
			continue
		}
		course := courses[rowValue(enrollment, "course_id")]
		summaries = append(summaries, StudentScheduleSummary{
			Nickname:        rowValue(student, "nickname"),
			FirstName:       rowValue(student, "first_name"),
			FullName:        firstNonEmpty(rowValue(student, "full_name_th"), rowValue(student, "first_name")),
			Course:          rowValue(course, "name"),
			TotalHours:      parseSheetInt(rowValue(enrollment, "total_hours"), 0),
			CompletedHours:  parseSheetInt(rowValue(enrollment, "completed_hours"), 0),
			DefaultSchedule: defaultScheduleText(enrollment, course, tables),
			PastLessons:     lessonListText(rowValue(enrollment, "id"), tables["lesson_sessions"], s.loc, now, true),
			NextLessons:     lessonListText(rowValue(enrollment, "id"), tables["lesson_sessions"], s.loc, now, false),
			ScheduleNotes:   scheduleNotesText(rowValue(enrollment, "id"), tables["enrollment_schedule_notes"]),
		})
	}
	sortStudentScheduleSummaries(summaries)
	return summaries, nil
}

func (s *GoogleSheetsLessonStore) AddStudent(nickname, firstName, course string, totalHours int, scheduleText string) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
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

	tables, err := s.loadCore(ctx)
	if err != nil {
		return StudentLesson{}, err
	}
	now := formatSheetDateTime(time.Now().In(s.loc))
	studentID := newSheetID("student")
	courseID, err := s.findOrCreateCourse(ctx, tables["courses"], course, totalHours)
	if err != nil {
		return StudentLesson{}, err
	}

	if err := s.appendRecord(ctx, "students", map[string]string{
		"id":           studentID,
		"nickname":     nickname,
		"first_name":   firstName,
		"full_name_th": firstName,
		"created_at":   now,
		"updated_at":   now,
	}); err != nil {
		return StudentLesson{}, err
	}

	enrollmentID := newSheetID("enrollment")
	if err := s.appendRecord(ctx, "enrollments", map[string]string{
		"id":                    enrollmentID,
		"student_id":            studentID,
		"course_id":             courseID,
		"total_hours":           strconv.Itoa(totalHours),
		"completed_hours":       "0",
		"default_session_hours": "2",
		"active":                "true",
		"created_at":            now,
		"updated_at":            now,
	}); err != nil {
		return StudentLesson{}, err
	}

	if scheduleText == "" {
		return StudentLesson{
			ID:             "enrollment-" + enrollmentID,
			Nickname:       nickname,
			FirstName:      firstName,
			FullName:       firstName,
			Course:         course,
			TotalHours:     totalHours,
			CompletedHours: 0,
			SessionHours:   2,
			UpdatedAt:      time.Now().In(s.loc),
		}, nil
	}

	start, end, ok := parseSchedule(scheduleText, s.loc)
	if !ok {
		return StudentLesson{}, fmt.Errorf("อ่านวันเวลาไม่ได้: %s", scheduleText)
	}
	sessionID, err := s.insertLessonSession(ctx, enrollmentID, start, end, "unconfirmed")
	if err != nil {
		return StudentLesson{}, err
	}
	return s.lessonBySessionID(ctx, sessionID)
}

func (s *GoogleSheetsLessonStore) findOrCreateCourse(ctx context.Context, courses []sheetRecord, name string, totalHours int) (string, error) {
	for _, course := range courses {
		if strings.EqualFold(rowValue(course, "name"), name) {
			return rowValue(course, "id"), nil
		}
	}
	now := formatSheetDateTime(time.Now().In(s.loc))
	id := newSheetID("course")
	err := s.appendRecord(ctx, "courses", map[string]string{
		"id":                    id,
		"name":                  name,
		"default_total_hours":   strconv.Itoa(totalHours),
		"default_session_hours": "2",
		"created_at":            now,
		"updated_at":            now,
	})
	return id, err
}

func (s *GoogleSheetsLessonStore) UpdateLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, "unconfirmed", true)
}

func (s *GoogleSheetsLessonStore) ConfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, "confirmed", false)
}

func (s *GoogleSheetsLessonStore) UnconfirmLesson(nickname, firstName, scheduleText string) (StudentLesson, error) {
	return s.changeLesson(nickname, firstName, scheduleText, "unconfirmed", false)
}

func (s *GoogleSheetsLessonStore) FindLessonByStudentName(nickname, firstName string) (StudentLesson, error) {
	ctx := context.Background()
	tables, err := s.loadCore(ctx)
	if err != nil {
		return StudentLesson{}, err
	}
	enrollment, err := findSheetEnrollmentByStudentName(tables, nickname, firstName)
	if err != nil {
		return StudentLesson{}, err
	}
	session, err := findEditableSheetSession(tables["lesson_sessions"], rowValue(enrollment, "id"), s.loc)
	if err != nil {
		return StudentLesson{}, err
	}
	return s.lessonBySessionID(ctx, rowValue(session, "id"))
}

func (s *GoogleSheetsLessonStore) changeLesson(nickname, firstName, scheduleText, status string, requireSchedule bool) (StudentLesson, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	tables, err := s.loadCore(ctx)
	if err != nil {
		return StudentLesson{}, err
	}
	enrollment, err := findSheetEnrollmentByStudentName(tables, nickname, firstName)
	if err != nil {
		return StudentLesson{}, err
	}

	scheduleText = strings.TrimSpace(scheduleText)
	hasSchedule := scheduleText != ""
	var start time.Time
	var end time.Time
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

	session, err := findEditableSheetSession(tables["lesson_sessions"], rowValue(enrollment, "id"), s.loc)
	if err != nil && !hasSchedule {
		return StudentLesson{}, err
	}
	if err != nil && hasSchedule {
		sessionID, err := s.insertLessonSession(ctx, rowValue(enrollment, "id"), start, end, status)
		if err != nil {
			return StudentLesson{}, err
		}
		return s.lessonBySessionID(ctx, sessionID)
	}

	session.values["status"] = status
	session.values["updated_at"] = formatSheetDateTime(time.Now().In(s.loc))
	if hasSchedule {
		session.values["start_at"] = formatSheetDateTime(start)
		session.values["end_at"] = formatSheetDateTime(end)
	}
	if err := s.updateRecord(ctx, "lesson_sessions", session); err != nil {
		return StudentLesson{}, err
	}
	return s.lessonBySessionID(ctx, rowValue(session, "id"))
}

func (s *GoogleSheetsLessonStore) insertLessonSession(ctx context.Context, enrollmentID string, start time.Time, end time.Time, status string) (string, error) {
	records, err := s.loadTable(ctx, "lesson_sessions")
	if err != nil {
		return "", err
	}
	maxSequence := 0
	for _, record := range records {
		if rowValue(record, "enrollment_id") == enrollmentID {
			if sequence := parseSheetInt(rowValue(record, "sequence_no"), 0); sequence > maxSequence {
				maxSequence = sequence
			}
		}
	}
	now := formatSheetDateTime(time.Now().In(s.loc))
	sessionID := newSheetID("session")
	err = s.appendRecord(ctx, "lesson_sessions", map[string]string{
		"id":            sessionID,
		"enrollment_id": enrollmentID,
		"sequence_no":   strconv.Itoa(maxSequence + 1),
		"start_at":      formatSheetDateTime(start),
		"end_at":        formatSheetDateTime(end),
		"status":        status,
		"created_at":    now,
		"updated_at":    now,
	})
	return sessionID, err
}

func (s *GoogleSheetsLessonStore) lessonBySessionID(ctx context.Context, sessionID string) (StudentLesson, error) {
	tables, err := s.loadCore(ctx)
	if err != nil {
		return StudentLesson{}, err
	}
	students := recordsByID(tables["students"])
	courses := recordsByID(tables["courses"])
	enrollments := recordsByID(tables["enrollments"])
	for _, session := range tables["lesson_sessions"] {
		if rowValue(session, "id") != sessionID {
			continue
		}
		enrollment := enrollments[rowValue(session, "enrollment_id")]
		student := students[rowValue(enrollment, "student_id")]
		course := courses[rowValue(enrollment, "course_id")]
		start, end, ok := parseSheetDateTimeRange(rowValue(session, "start_at"), rowValue(session, "end_at"), s.loc)
		if !ok {
			return StudentLesson{}, errors.New("อ่านวันเวลาใน Google Sheet ไม่ได้")
		}
		return buildSheetLesson(session, enrollment, student, course, start, end, s.loc), nil
	}
	return StudentLesson{}, errors.New("ไม่พบ session ใน Google Sheet")
}

func (s *GoogleSheetsLessonStore) RegisterLineGroup(groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if !isLikelyLineTargetID(groupID) || !strings.HasPrefix(groupID, "C") {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	records, err := s.loadTable(ctx, "line_groups")
	if err != nil {
		return err
	}
	now := formatSheetDateTime(time.Now().In(s.loc))
	for _, record := range records {
		if rowValue(record, "group_id") == groupID {
			record.values["active"] = "true"
			record.values["last_seen_at"] = now
			record.values["updated_at"] = now
			return s.updateRecord(ctx, "line_groups", record)
		}
	}
	return s.appendRecord(ctx, "line_groups", map[string]string{
		"id":            newSheetID("line_group"),
		"group_id":      groupID,
		"active":        "true",
		"first_seen_at": now,
		"last_seen_at":  now,
		"created_at":    now,
		"updated_at":    now,
	})
}

func (s *GoogleSheetsLessonStore) ListLineGroupIDs() ([]string, error) {
	records, err := s.loadTable(context.Background(), "line_groups")
	if err != nil {
		return nil, err
	}
	var groupIDs []string
	for _, record := range records {
		if parseSheetBool(rowValue(record, "active"), true) && strings.TrimSpace(rowValue(record, "group_id")) != "" {
			groupIDs = append(groupIDs, rowValue(record, "group_id"))
		}
	}
	sort.Strings(groupIDs)
	return groupIDs, nil
}

func findSheetEnrollmentByStudentName(tables map[string][]sheetRecord, nickname, firstName string) (sheetRecord, error) {
	students := recordsByID(tables["students"])
	var matches []sheetRecord
	for _, enrollment := range tables["enrollments"] {
		if !parseSheetBool(rowValue(enrollment, "active"), true) {
			continue
		}
		student := students[rowValue(enrollment, "student_id")]
		if strings.EqualFold(rowValue(student, "nickname"), strings.TrimSpace(nickname)) &&
			strings.EqualFold(rowValue(student, "first_name"), strings.TrimSpace(firstName)) {
			matches = append(matches, enrollment)
		}
	}
	if len(matches) == 0 {
		return sheetRecord{}, fmt.Errorf("ไม่พบนักเรียน: %s / %s", nickname, firstName)
	}
	if len(matches) > 1 {
		return sheetRecord{}, fmt.Errorf("พบ %s / %s มากกว่า 1 คอร์ส กรุณาเพิ่มคำสั่งระบุคอร์สในอนาคต", nickname, firstName)
	}
	return matches[0], nil
}

func findEditableSheetSession(records []sheetRecord, enrollmentID string, loc *time.Location) (sheetRecord, error) {
	now := time.Now().In(loc)
	type candidate struct {
		record sheetRecord
		start  time.Time
	}
	var candidates []candidate
	for _, record := range records {
		if rowValue(record, "enrollment_id") != enrollmentID {
			continue
		}
		status := rowValue(record, "status")
		if status == "completed" || status == "cancelled" {
			continue
		}
		start, ok := parseSheetDateTime(rowValue(record, "start_at"), loc)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate{record: record, start: start})
	}
	if len(candidates) == 0 {
		return sheetRecord{}, errors.New("ยังไม่มี session ที่แก้ไขได้สำหรับนักเรียนคนนี้")
	}
	sort.Slice(candidates, func(i, j int) bool {
		iFuture := !candidates[i].start.Before(now)
		jFuture := !candidates[j].start.Before(now)
		if iFuture != jFuture {
			return iFuture
		}
		return candidates[i].start.Before(candidates[j].start)
	})
	return candidates[0].record, nil
}

func buildSheetLesson(session, enrollment, student, course sheetRecord, start time.Time, end time.Time, loc *time.Location) StudentLesson {
	start = start.In(loc)
	end = end.In(loc)
	updatedAt, ok := parseSheetDateTime(rowValue(session, "updated_at"), loc)
	if !ok {
		updatedAt = time.Now().In(loc)
	}
	sessionHours := int(end.Sub(start).Hours())
	if sessionHours <= 0 {
		sessionHours = 1
	}
	lesson := StudentLesson{
		ID:             rowValue(session, "id"),
		Nickname:       rowValue(student, "nickname"),
		FirstName:      rowValue(student, "first_name"),
		FullName:       firstNonEmpty(rowValue(student, "full_name_th"), rowValue(student, "first_name")),
		Course:         rowValue(course, "name"),
		TotalHours:     parseSheetInt(rowValue(enrollment, "total_hours"), 0),
		CompletedHours: parseSheetInt(rowValue(enrollment, "completed_hours"), 0),
		SessionHours:   sessionHours,
		NextStart:      start,
		NextEnd:        end,
		Confirmed:      rowValue(session, "status") == "confirmed",
		UpdatedAt:      updatedAt,
	}
	lesson.ScheduleText = formatThaiSchedule(start, end)
	return lesson
}

func recordsByID(records []sheetRecord) map[string]sheetRecord {
	mapped := map[string]sheetRecord{}
	for _, record := range records {
		if id := rowValue(record, "id"); id != "" {
			mapped[id] = record
		}
	}
	return mapped
}

func rowValue(record sheetRecord, key string) string {
	return strings.TrimSpace(record.values[key])
}

func valuesForHeaders(headers []string, values map[string]string) []string {
	row := make([]string, len(headers))
	for i, header := range headers {
		row[i] = values[header]
	}
	return row
}

func quoteSheetName(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}

func columnName(index int) string {
	if index <= 0 {
		return "A"
	}
	var name []byte
	for index > 0 {
		index--
		name = append([]byte{byte('A' + index%26)}, name...)
		index /= 26
	}
	return string(name)
}

func sameStringSlice(a []string, b []string) bool {
	if len(a) < len(b) {
		return false
	}
	for i := range b {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

func parseSheetBool(value string, defaultValue bool) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes" || value == "y"
}

func parseSheetInt(value string, defaultValue int) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return number
}

func parseSheetDateTime(value string, loc *time.Location) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"02/01/2006 15:04",
		"1/2/2006 15:04:05",
		"1/2/2006 15:04",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, loc)
		if err == nil {
			return parsed.In(loc), true
		}
	}
	return time.Time{}, false
}

func parseSheetDateTimeRange(startText string, endText string, loc *time.Location) (time.Time, time.Time, bool) {
	start, ok := parseSheetDateTime(startText, loc)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	end, ok := parseSheetDateTime(endText, loc)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func formatSheetDateTime(value time.Time) string {
	return value.Format(time.RFC3339)
}

func formatSheetTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func newSheetID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func defaultScheduleText(enrollment sheetRecord, course sheetRecord, tables map[string][]sheetRecord) string {
	enrollmentID := rowValue(enrollment, "id")
	var parts []string
	for _, record := range tables["enrollment_default_schedules"] {
		if rowValue(record, "enrollment_id") == enrollmentID && parseSheetBool(rowValue(record, "active"), true) {
			parts = append(parts, strings.TrimSpace(rowValue(record, "weekday_text")+" "+formatSheetTime(rowValue(record, "start_time"))+"-"+formatSheetTime(rowValue(record, "end_time"))))
		}
	}
	if len(parts) > 0 {
		sort.Strings(parts)
		return strings.Join(parts, ", ")
	}

	courseID := rowValue(course, "id")
	slotSeen := map[string]bool{}
	dayText := ""
	for _, record := range tables["course_default_schedules"] {
		if rowValue(record, "course_id") != courseID || !parseSheetBool(rowValue(record, "active"), true) {
			continue
		}
		if dayText == "" {
			dayText = rowValue(record, "default_day_text")
		}
		slot := formatSheetTime(rowValue(record, "start_time")) + "-" + formatSheetTime(rowValue(record, "end_time"))
		if !slotSeen[slot] {
			slotSeen[slot] = true
			parts = append(parts, slot)
		}
	}
	sort.Strings(parts)
	return strings.TrimSpace(strings.Join([]string{dayText, strings.Join(parts, ", ")}, " "))
}

func lessonListText(enrollmentID string, sessions []sheetRecord, loc *time.Location, now time.Time, past bool) string {
	type item struct {
		start time.Time
		text  string
	}
	var items []item
	for _, session := range sessions {
		if rowValue(session, "enrollment_id") != enrollmentID {
			continue
		}
		status := rowValue(session, "status")
		start, end, ok := parseSheetDateTimeRange(rowValue(session, "start_at"), rowValue(session, "end_at"), loc)
		if !ok {
			continue
		}
		if past {
			if status != "completed" {
				continue
			}
		} else if status == "completed" || status == "cancelled" || start.Before(now) {
			continue
		}
		items = append(items, item{start: start, text: formatShortLessonTime(start.In(loc), end.In(loc))})
	}
	sort.Slice(items, func(i, j int) bool {
		if past {
			return items[i].start.After(items[j].start)
		}
		return items[i].start.Before(items[j].start)
	})
	if len(items) > 3 {
		items = items[:3]
	}
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = item.text
	}
	return strings.Join(parts, ", ")
}

func scheduleNotesText(enrollmentID string, notes []sheetRecord) string {
	var parts []string
	for _, note := range notes {
		if rowValue(note, "enrollment_id") == enrollmentID && rowValue(note, "source_text") != "" {
			parts = append(parts, rowValue(note, "source_text"))
		}
	}
	sort.Strings(parts)
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, ", ")
}
