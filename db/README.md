# Google Sheets Database

Runtime หลักใช้ Google Sheets เป็น source of truth โดยยึดโครงสร้างจาก Class schedule 3 tab:

- `Overview`: master enrollment/student ตาม `Class schedule - Overview.csv`
- `ตารางเรียน`: ตาราง grid หลายสัปดาห์ ตาม `Class schedule - ตารางเรียน.csv`
- `สัปดาห์นี้`: ตาราง session รายครั้ง ตาม `Class schedule - สัปดาห์นี้.csv`

ไม่มี local seed แล้ว ข้อมูลจริงต้องอยู่ใน Google Spreadsheet นั้นก่อนรัน bot ถ้า `GOOGLE_SHEETS_INIT_SCHEMA=true` bot จะตรวจว่า tab `Overview`, `ตารางเรียน`, `สัปดาห์นี้` มีอยู่จริง

การใช้งานใน bot:

- `/ตารางเรียน` อ่านจาก tab `สัปดาห์นี้`
- `/ข้อมูลนักเรียน` อ่านจาก `Overview` แล้วแสดงเฉพาะนักเรียนที่ยังเรียนไม่จบ
- `/อัพเดท` เขียนวันที่, วัน, เวลา, ระยะเวลา กลับ tab `สัปดาห์นี้`
- `/คอนเฟิร์ม`, `/ไม่คอนเฟิร์ม` เขียนค่า `true`/`false` ที่ช่อง `สถานะคอนเฟิร์ม` ใน tab `สัปดาห์นี้`
