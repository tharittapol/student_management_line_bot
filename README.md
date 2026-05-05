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

POSTGRES_DB=linebot
POSTGRES_USER=linebot
POSTGRES_PASSWORD=linebot_password
POSTGRES_PORT=5432

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

- `postgres`: database จริง พร้อม schema และ seed จากไฟล์ CSV ที่แนบมา
- `line-bot`: Go app
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

ถ้าต้องการล้าง database volume แล้ว seed ใหม่จาก `db/*.sql`:

```powershell
docker compose down -v
docker compose up --build
```

## Database

Schema อยู่ที่:

```text
db/schema.sql
```

Seed จาก CSV อยู่ที่:

```text
db/seed_kmutt_smart_kid_students.sql
```

โครงสร้างหลัก:

- `courses`: ข้อมูลคอร์ส
- `students`: นักเรียนและข้อมูลผู้ปกครอง
- `enrollments`: นักเรียนเรียนคอร์สไหน ชั่วโมงรวม/ชั่วโมงที่เรียนไปแล้ว
- `lesson_sessions`: ตารางเรียนแต่ละครั้งและสถานะ `confirmed` / `unconfirmed`
- `line_groups`: LINE group ที่ bot เคยเห็น เพื่อส่ง routine notification ทุก group

ความสัมพันธ์:

```text
students 1 -- * enrollments * -- 1 courses
enrollments 1 -- * lesson_sessions
line_groups แยกสำหรับปลายทางแจ้งเตือน
```

## คำสั่งใน LINE Group

ขอตารางเรียน 7 วันข้างหน้า:

```text
/ตารางเรียน
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

ถ้าไม่ได้ตั้ง `DATABASE_URL` app จะ fallback เป็น mock in-memory store:

```powershell
go test .
go run .
```

Webhook path:

```text
http://localhost:8080/line/webhook
```
