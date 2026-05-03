# Student Management LINE Bot

LINE bot สำหรับแจ้งเวลาเรียน, อัพเดทเวลาเรียน และคอนเฟิร์มเวลาเรียนจากข้อมูล mock database ใน `main.go`

## Env

สร้างไฟล์ `.env` จากตัวอย่าง:

```powershell
Copy-Item .env.example .env
```

แล้วแก้ค่าใน `.env`:

```env
LINE_CHANNEL_SECRET=your_line_channel_secret
LINE_CHANNEL_ACCESS_TOKEN=your_line_channel_access_token
LINE_STAFF_GROUP_ID=
PORT=8080
RUN_DAILY_ON_START=false
HOST_PORT=8080
NGROK_AUTHTOKEN=your_ngrok_authtoken
NGROK_INSPECTOR_PORT=4040
```

ถ้ายังไม่รู้ `LINE_STAFF_GROUP_ID` ให้เว้นว่างไว้ก่อน เพิ่ม LINE OA เข้า group แล้วส่งข้อความในกลุ่ม 1 ครั้ง จากนั้นดู log บรรทัด `STAFF GROUP ID = ...`

## Run แบบ Go local

```powershell
go test .
go run .
```

Server จะรันที่:

```text
http://localhost:8080/line/webhook
```

ถ้าจะทดสอบ webhook จาก LINE จริง ให้ expose เครื่อง local ผ่าน tunnel เช่น ngrok แล้วตั้ง Webhook URL ใน LINE Developers เป็น:

```text
https://your-domain.example/line/webhook
```

## Run แบบ Docker

Build image:

```powershell
docker build -t student-management-line-bot .
```

Run container:

```powershell
docker run --rm --env-file .env -p 8080:8080 student-management-line-bot
```

## Run แบบ Docker Compose

```powershell
docker compose up --build
```

Compose จะรันทั้ง `line-bot` และ `ngrok` ให้พร้อมกัน โดย ngrok จะ tunnel เข้า service ภายใน Docker ที่ `http://line-bot:8080`

ดู public HTTPS URL ของ ngrok:

```powershell
docker compose logs -f ngrok
```

มองหาบรรทัดที่มี `url=https://...ngrok...` แล้วเอา URL นั้นไปต่อ path webhook:

```text
https://your-ngrok-url.ngrok-free.app/line/webhook
```

หรือเปิด ngrok inspector ใน browser:

```text
http://localhost:4040
```

แล้วตั้งค่าใน LINE Developers Console:

1. เปิด Messaging API channel ของ bot
2. ไปที่แท็บ `Messaging API`
3. ตั้ง `Webhook URL` เป็น `https://your-ngrok-url.ngrok-free.app/line/webhook`
4. กด `Verify`
5. เปิด `Use webhook`
6. เปิด `Allow bot to join group chats`

หลังจากนั้นส่งข้อความอะไรก็ได้ใน LINE group แล้วดู log:

```powershell
docker compose logs -f line-bot
```

ถ้าเห็น `STAFF GROUP ID = C...` ให้เอาค่านั้นไปใส่ `LINE_STAFF_GROUP_ID` ใน `.env` แล้ว recreate container เพราะ app อ่าน env ตอน start เท่านั้น:

```powershell
docker compose up -d --build --force-recreate
```

ถ้าต้องการรันเบื้องหลัง:

```powershell
docker compose up -d --build
```

ดู log:

```powershell
docker compose logs -f line-bot
```

หยุด service:

```powershell
docker compose down
```

## คำสั่งใน LINE Group

ขอตารางเรียนล่าสุด:

```text
/ตารางเรียน
```

Bot จะตอบตารางเดียวกับ daily notification โดยใช้ข้อมูลล่าสุดใน mock database ที่ถูกอัพเดท/คอนเฟิร์มระหว่างที่ app กำลังรัน

```text
อัพเดทเวลาเรียน/แพรว/แพรวา ศิริพงษ์/English Foundation/วันเสาร์ 9 พฤษภาคม 2569 เวลา 13.00-15.00 น.
```

```text
คอนเฟิร์มเวลาเรียน/แพรว/แพรวา ศิริพงษ์/English Foundation/วันเสาร์ 9 พฤษภาคม 2569 เวลา 13.00-15.00 น./คอนเฟิร์ม
```

ระบบจะแจ้งเวลาเรียนรายวันทุกวันเวลา 09:00 ตามเวลา `Asia/Bangkok`
