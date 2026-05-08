# Student Management LINE Bot

LINE bot สำหรับแจ้งตารางเรียนจาก tab `สัปดาห์นี้`, อัพเดทเวลาเรียน, คอนเฟิร์ม/ไม่คอนเฟิร์มเวลาเรียน และส่ง routine notification ไปยังหลาย LINE group

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
Overview
ตารางเรียน
สัปดาห์นี้
```

ให้ Google Sheet เป็น source of truth โดย sync/import ข้อมูลให้มี 3 tab หลักก่อนรัน:

- `Overview`
- `ตารางเรียน`
- `สัปดาห์นี้`

ถ้า `GOOGLE_SHEETS_INIT_SCHEMA=true` bot จะตรวจว่า 3 tab หลักมีอยู่จริง ถ้า tab ใดหาย app จะ error เพื่อให้ sync Google Sheet ให้ครบก่อน

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

โครงสร้างหลักยึดตามไฟล์ CSV:

- `Overview`: master enrollment/student จาก `Class schedule - Overview.csv`
- `ตารางเรียน`: grid ตารางหลายสัปดาห์จาก `Class schedule - ตารางเรียน.csv`
- `สัปดาห์นี้`: session รายครั้งจาก `Class schedule - สัปดาห์นี้.csv`

การอ่าน/เขียนของ bot:

- `/ตารางเรียน` อ่านจาก `สัปดาห์นี้`
- `/ข้อมูลนักเรียน` อ่านจาก `Overview` แล้วแสดงชื่อเล่น, ชื่อจริง, คอร์ส ของนักเรียนที่ยังเรียนไม่จบ โดยรวมเด็กคนเดียวกันเป็น 1 คน
- `/อัพเดท` เขียนวันที่, วัน, เวลา, ระยะเวลา กลับ `สัปดาห์นี้`
- `/คอนเฟิร์ม`, `/ไม่คอนเฟิร์ม` เขียนค่า boolean `TRUE`/`FALSE` ที่ช่อง `สถานะคอนเฟิร์ม` ใน `สัปดาห์นี้`
- `/ลา`, `/เข้าเรียน` เขียนช่อง `สถานะการเรียน` ใน `สัปดาห์นี้`

เวลา bot เขียนข้อมูลกลับ `สัปดาห์นี้` จะ copy format และ data validation จากแถวข้างเคียง เพื่อให้ checkbox, ตัวหนา, สีพื้นหลัง และ format หลักตาม pattern เดิมของ sheet

ถ้า LINE response ยาวเกิน limit ระบบจะแบ่งข้อความส่งให้อัตโนมัติ

## คำสั่งใน LINE Group

📌 คำสั่งใน LINE Group

🧭 วิธีใช้งาน

```text
/วิธีใช้งาน
```

📚 ตารางเรียนจากแท็บสัปดาห์นี้

```text
/ตารางเรียน
```

👥 นักเรียนที่ยังเรียนไม่จบ (`ชื่อเล่น ชื่อจริง` ดูได้จากคำสั่งนี้)

```text
/ข้อมูลนักเรียน
```

🔄 อัพเดทเวลาเรียน

```text
/อัพเดท ชื่อเล่น ชื่อจริง วัน/เดือน เวลาเริ่ม-เวลาจบ
ตัวอย่าง: "/อัพเดท โบ โบรอท 9/5 13:00-15:00"
```

✅ คอนเฟิร์ม

```text
/คอนเฟิร์ม ชื่อเล่น ชื่อจริง
ตัวอย่าง: "/คอนเฟิร์ม โบ โบรอท"
```

⏳ ไม่คอนเฟิร์ม

```text
/ไม่คอนเฟิร์ม ชื่อเล่น ชื่อจริง
ตัวอย่าง: "/ไม่คอนเฟิร์ม โบ โบรอท"
```

📝 สถานะการเรียน

```text
/ลา ชื่อเล่น ชื่อจริง
/เข้าเรียน ชื่อเล่น ชื่อจริง
```

🗓 คอนเฟิร์มหรือไม่คอนเฟิร์มพร้อมเปลี่ยนเวลา

```text
/คอนเฟิร์ม ชื่อเล่น ชื่อจริง 9/5 13:00-15:00
/ไม่คอนเฟิร์ม ชื่อเล่น ชื่อจริง 9/5 13:00-15:00
```

📅 กติกาปี

- ใส่แค่ `วัน/เดือน` เช่น `9/5` = ปีปัจจุบัน
- ใส่ปีด้วย เช่น `9/5/2570` = ใช้ปีที่ระบุ

⏰ ระบบจะแจ้งตารางเรียนจากแท็บสัปดาห์นี้ทุกวัน 09:00 เวลา `Asia/Bangkok`

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
