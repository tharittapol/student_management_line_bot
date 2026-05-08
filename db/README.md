# Google Sheets Database

Runtime หลักใช้ Google Sheets เป็น source of truth โดยยึดโครงสร้างจาก Class schedule 3 tab:

- `Overview`: master enrollment/student ตาม `Class schedule - Overview.csv`
- `ตารางเรียน`: ตาราง grid หลายสัปดาห์ ตาม `Class schedule - ตารางเรียน.csv`
- `สัปดาห์นี้`: ตาราง session รายครั้ง ตาม `Class schedule - สัปดาห์นี้.csv`

ไม่มี local seed แล้ว ข้อมูลจริงต้องอยู่ใน Google Spreadsheet นั้นก่อนรัน bot ถ้า `GOOGLE_SHEETS_INIT_SCHEMA=true` bot จะตรวจว่า tab `Overview`, `ตารางเรียน`, `สัปดาห์นี้` มีอยู่จริง

การใช้งานใน bot:

- `/ตารางเรียน` อ่านจาก tab `สัปดาห์นี้`
- `/ข้อมูลนักเรียน` อ่านจาก `Overview` แล้วแสดงเฉพาะนักเรียนที่ยังเรียนไม่จบ โดยรวมเด็กคนเดียวกันเป็น 1 คน
- `/อัพเดท` เขียนวันที่, วัน, เวลา, ระยะเวลา กลับ tab `สัปดาห์นี้`
- `/คอนเฟิร์ม`, `/ไม่คอนเฟิร์ม` เขียนค่า boolean `TRUE`/`FALSE` ที่ช่อง `สถานะคอนเฟิร์ม` ใน tab `สัปดาห์นี้`
- `/ลา`, `/เข้าเรียน` เขียนช่อง `สถานะการเรียน` ใน tab `สัปดาห์นี้`

เวลา bot เขียนข้อมูลกลับ tab `สัปดาห์นี้` จะ copy format และ data validation จากแถวข้างเคียง เพื่อให้ checkbox, ตัวหนา, สีพื้นหลัง และ format หลักตาม pattern เดิมของ sheet
