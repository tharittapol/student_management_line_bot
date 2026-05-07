# Student Management LINE Bot

LINE bot สำหรับแจ้งตารางเรียน 7 วันข้างหน้า, อัพเดทเวลาเรียน, คอนเฟิร์ม/ไม่คอนเฟิร์มเวลาเรียน และส่ง routine notification ไปยังหลาย LINE group

## Env

สร้างไฟล์ `.env` จากตัวอย่าง:

```powershell
Copy-Item .env.example .env
```

ค่าหลักที่ต้องแก้:

```env
LINE_CHANNEL_SECRET=your_line_channel_secret
LINE_CHANNEL_ACCESS_TOKEN=your_line_channel_access_token
LINE_GROUP_IDS=Cxxxxxxxxxxxxxxxx,Cyyyyyyyyyyyyyyyy
LINE_STAFF_GROUP_ID=

GOOGLE_SHEET_ID=your_google_spreadsheet_id
GOOGLE_SERVICE_ACCOUNT_JSON_BASE64=base64_encoded_service_account_json
GOOGLE_SHEETS_INIT_SCHEMA=true
GOOGLE_SHEETS_SEED_DIR=db/google_sheets

RUN_DAILY_ON_START=false
HOST_PORT=8080
NGROK_AUTHTOKEN=your_ngrok_authtoken
NGROK_INSPECTOR_PORT=4040
```

ใส่ group ที่ต้องการให้ bot ทำงานใน `LINE_GROUP_IDS` คั่นด้วย comma เช่น:

```env
LINE_GROUP_IDS=Cgroup1,Cgroup2,Cgroup3
```

ถ้ามี `LINE_GROUP_IDS` แล้ว bot จะอ่านคำสั่งเฉพาะ group ที่อยู่ใน env และ routine 09:00 จะส่งตารางไปเฉพาะ group เหล่านั้น ส่วน `LINE_STAFF_GROUP_ID` ยังรองรับของเดิมไว้เป็น fallback แต่แนะนำให้ใช้ `LINE_GROUP_IDS`

## Run ด้วย Docker Compose

```powershell
docker compose up --build
```

Compose จะรัน:

- `line-bot`: Go app ที่ใช้ Google Sheets เป็น database
- `ngrok`: HTTPS tunnel สำหรับ LINE webhook ตอนรัน local

ดู URL ของ ngrok:

```powershell
docker compose logs -f ngrok
```

เอา URL ที่ได้ไปตั้งใน LINE Developers:

```text
https://your-ngrok-url.ngrok-free.app/line/webhook
```

แล้วเปิด:

- `Use webhook`
- `Allow bot to join group chats`

ดู log bot:

```powershell
docker compose logs -f line-bot
```

ถ้าต้อง rebuild หลังแก้โค้ด:

```powershell
docker compose up -d --build --force-recreate
```

หยุด service:

```powershell
docker compose down
```

## Database

ระบบใช้ Google Spreadsheet เป็น database โดย 1 worksheet/tab = 1 table:

```text
courses
students
enrollments
course_default_schedules
enrollment_default_schedules
lesson_sessions
enrollment_schedule_notes
line_groups
```

ไฟล์ seed สำหรับ Google Sheets อยู่ที่:

```text
db/google_sheets/*.csv
```

ตอน start ถ้า `GOOGLE_SHEETS_INIT_SCHEMA=true` bot จะสร้าง worksheet/header ให้เอง และถ้า sheet นั้นยังว่าง จะ seed จาก `GOOGLE_SHEETS_SEED_DIR`

การสร้าง credential:

1. สร้าง Google Cloud service account และเปิด Google Sheets API
2. ดาวน์โหลด service account JSON
3. Share Google Spreadsheet ให้ `client_email` ใน JSON เป็น Editor
4. แปลง JSON เป็น base64 แล้วใส่ใน `.env`

```powershell
$json = Get-Content .\service-account.json -Raw
$bytes = [System.Text.Encoding]::UTF8.GetBytes($json)
[Convert]::ToBase64String($bytes)
```

โครงสร้างหลัก:

- `courses`: ข้อมูลคอร์ส
- `students`: นักเรียนและข้อมูลผู้ปกครอง
- `students.source_key`: key สำหรับข้อมูล import จาก CSV ทำให้ `nickname` + `first_name` ไม่ต้อง unique
- `enrollments`: นักเรียนเรียนคอร์สไหน ชั่วโมงรวม/ชั่วโมงที่เรียนไปแล้ว, ครู, วันเริ่มเรียน, สถานะจากระบบต้นทาง
- `course_default_schedules`: ตาราง default/slot เวลาเรียนของแต่ละคอร์สจากหัวตาราง CSV
- `enrollment_default_schedules`: ตาราง default ราย enrollment ว่านักเรียนคนนี้เรียนปกติวันไหน/เวลาไหน
- `lesson_sessions`: ตารางเรียนแต่ละครั้งและสถานะ `confirmed` / `unconfirmed`
- `enrollment_schedule_notes`: note ตารางรายนักเรียนที่ยังไม่ใช่วันเวลาชัดเจน เช่น `ใส่วันที่`, `รอ CF`
- `line_groups`: LINE group ที่ bot เคยเห็น เพื่อส่ง routine notification ทุก group

ความสัมพันธ์:

```text
students 1 -- * enrollments * -- 1 courses
courses 1 -- * course_default_schedules
enrollments 1 -- * enrollment_default_schedules
enrollments 1 -- * lesson_sessions
enrollments 1 -- * enrollment_schedule_notes
line_groups แยกสำหรับปลายทางแจ้งเตือน
```

Seed ปัจจุบันสร้างจากไฟล์ `Class schedule - Overview.csv`, `Class schedule - ตารางเรียน.csv`, และ `Class schedule - สัปดาห์นี้.csv`

## คำสั่งใน LINE Group

ขอตารางเรียน 7 วันข้างหน้า:

```text
/ตารางเรียน
```

ดูข้อมูลตารางรายนักเรียน:

```text
/ข้อมูลนักเรียน
/ข้อมูลนักเรียน แพรว แพรวา
```

อัพเดทเวลาเรียน:

```text
/อัพเดท ชื่อเล่น ชื่อจริง วัน/เดือน เวลาเริ่ม-เวลาจบ
/อัพเดท แพรว แพรวา 9/5 13:00-15:00
```

เพิ่มนักเรียน:

```text
/เพิ่มนักเรียน ชื่อเล่น ชื่อจริง/คอร์ส/ชั่วโมงรวม
/เพิ่มนักเรียน แพรว แพรวา/Little 3D รุ่นที่ 1/8
```

คอนเฟิร์ม:

```text
/คอนเฟิร์ม แพรว แพรวา
```

ไม่คอนเฟิร์ม:

```text
/ไม่คอนเฟิร์ม แพรว แพรวา
```

คอนเฟิร์มหรือไม่คอนเฟิร์มพร้อมเปลี่ยนเวลา:

```text
/คอนเฟิร์ม แพรว แพรวา 9/5/2570 13:00-15:00
/ไม่คอนเฟิร์ม แพรว แพรวา 9/5 13:00-15:00
```

กติกาปี:

- ใส่แค่ `วัน/เดือน` เช่น `9/5` = ปีปัจจุบัน
- ใส่ปีด้วย เช่น `9/5/2570` = ใช้ปีที่ระบุ

ระบบจะแจ้งตารางเรียน 7 วันข้างหน้าทุกวันเวลา 09:00 ตามเวลา `Asia/Bangkok`

## Run แบบ Go Local

ตั้ง `GOOGLE_SHEET_ID` และ credential ใน `.env` ก่อน แล้วรัน:

```powershell
go test .
go run .
```

ถ้าไม่ได้ตั้ง `GOOGLE_SHEET_ID` และไม่ได้ตั้ง `DATABASE_URL` app จะ fallback เป็น mock in-memory store สำหรับทดสอบ parser/command

Webhook path:

```text
http://localhost:8080/line/webhook
```
