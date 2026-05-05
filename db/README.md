# Database Schema

ใช้ PostgreSQL เป็น database นอกโค้ด โดยตารางหลักเชื่อมกันแบบนี้:

- `courses`: ข้อมูลคอร์ส
- `students`: ข้อมูลนักเรียนและผู้ปกครอง
- `enrollments`: นักเรียนคนไหนเรียนคอร์สไหน พร้อมชั่วโมงรวม/ชั่วโมงที่เรียนไปแล้ว
- `lesson_sessions`: ตารางเรียนแต่ละครั้งของ enrollment นั้น พร้อมสถานะ `confirmed` หรือ `unconfirmed`
- `line_groups`: LINE group ที่ bot เคยเห็นและต้องส่ง routine notification ไปหา

ความสัมพันธ์:

```text
students 1 -- * enrollments * -- 1 courses
enrollments 1 -- * lesson_sessions
line_groups แยกสำหรับปลายทางแจ้งเตือน LINE
```

ไฟล์ `seed_kmutt_smart_kid_students.sql` seed นักเรียนปัจจุบัน (6/5/2569)
