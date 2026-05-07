# Google Sheets Database

ใช้ Google Spreadsheet เป็น database นอกโค้ด โดย 1 worksheet/tab = 1 table:

- `courses`: ข้อมูลคอร์ส
- `students`: ข้อมูลนักเรียนและผู้ปกครอง โดย `nickname` + `first_name` ไม่ unique แล้ว และใช้ `source_key` สำหรับข้อมูล import จาก CSV
- `enrollments`: นักเรียนคนไหนเรียนคอร์สไหน พร้อมชั่วโมงรวม/ชั่วโมงที่เรียนไปแล้ว, ครู, วันเริ่มเรียน, สถานะจากระบบต้นทาง
- `course_default_schedules`: ตาราง default/slot เวลาเรียนของคอร์ส เช่น วันเรียนปกติ, ครั้งที่, date label, เวลาเริ่ม/จบ
- `enrollment_default_schedules`: ตาราง default ราย enrollment เช่น เด็กคนนี้เรียนปกติวันไหน เวลาไหน
- `lesson_sessions`: ตารางเรียนจริงของ enrollment นั้น ทั้งที่เรียนแล้วและครั้งถัดไป พร้อมสถานะ `confirmed`, `unconfirmed`, `completed`
- `enrollment_schedule_notes`: note ราย enrollment จากช่องตารางที่ยังไม่มีวันเวลาชัดเจน เช่น `ใส่วันที่`, `รอ CF`, ข้อความเลื่อน/ชดเชย
- `line_groups`: LINE group ที่ bot เคยเห็นและต้องส่ง routine notification ไปหา

ความสัมพันธ์:

```text
students 1 -- * enrollments * -- 1 courses
courses 1 -- * course_default_schedules
enrollments 1 -- * enrollment_default_schedules
enrollments 1 -- * lesson_sessions
enrollments 1 -- * enrollment_schedule_notes
line_groups แยกสำหรับปลายทางแจ้งเตือน LINE
```

ไฟล์ seed สำหรับ Google Sheets อยู่ที่ `db/google_sheets/*.csv` และ seed นักเรียนปัจจุบันจาก Class schedule CSV (7/5/2569) โดยเก็บ:

- รายชื่อนักเรียน/ผู้ปกครอง
- enrollment ของแต่ละคอร์สจาก `Class schedule - Overview.csv`
- default schedule รายคนจาก `วันเรียน` + `เวลา`
- session สัปดาห์นี้จาก `Class schedule - สัปดาห์นี้.csv`
- session ถัดไปที่ generate จาก default schedule จนถึง horizon ใน `Class schedule - ตารางเรียน.csv`

หมายเหตุ: `schema.sql` ยังเก็บไว้เป็น reference/legacy สำหรับ PostgreSQL แต่ runtime หลักตอนนี้ใช้ Google Sheets แล้ว
