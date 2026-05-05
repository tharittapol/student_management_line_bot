-- Generated from CSV files in C:\Users\Thari\Downloads\student
-- Course name rule: use the filename text after "KMUTT Smart Kid - "
-- Courses: 8, unique students: 19, enrollments: 28, seeded sessions: 7

WITH course_seed (name, description, default_total_hours, default_session_hours) AS (
    VALUES
        ('Basic (3D) รุ่นที่ 1', 'Seeded from CSV export', 8, 2),
        ('Intermediate (Programming) รุ่นที่ 1', 'Seeded from CSV export', 14, 2),
        ('Junior Robotics Challenge (Lego Wedo)', 'Seeded from CSV export', 20, 2),
        ('Junior Robotics Discovery (Lego Wedo)', 'Seeded from CSV export', 20, 2),
        ('Little 3D Inventors (Basic)', 'Seeded from CSV export', 8, 2),
        ('Little 3D Inventors (Pro)', 'Seeded from CSV export', 8, 2),
        ('Little 3D รุ่นที่ 1', 'Seeded from CSV export', 10, 2),
        ('Pre-intermediate (Circuits) รุ่นที่ 1', 'Seeded from CSV export', 10, 2)
),
upsert_courses AS (
    INSERT INTO courses (name, description, default_total_hours, default_session_hours)
    SELECT name, description, default_total_hours, default_session_hours FROM course_seed
    ON CONFLICT (name) DO UPDATE SET
        description = EXCLUDED.description,
        default_total_hours = EXCLUDED.default_total_hours,
        default_session_hours = EXCLUDED.default_session_hours,
        updated_at = NOW()
    RETURNING id, name
),
student_seed (nickname, first_name, full_name_th, full_name_en, gender, age, nationality, parent_phone, parent_name, line_name, email, note) AS (
    VALUES
        ('Daniel', 'Daniel', 'Daniel', 'Jaeyong Shin', 'ชาย', NULL, 'เกาหลี', '0930462797', 'Juyong Park', NULL, 'wd521@naver.com', 'เพิ่มการสอนแบบภาษาอังกฤษ (Instruction in English.)'),
        ('ชาลี', 'ชาลี', 'ชาลี', 'Yik Hei Sung', NULL, NULL, 'ต่างชาติ', '0637203986', 'Ms.ELSA', NULL, 'elsayanyan86@gmail.com', 'ชดเชย ส. 28 ก.พ.; 13.00 - 15.00 น.'),
        ('ฌาณ', 'ฌาณ', 'ฌาณ', 'Chatchon Chayathab', NULL, NULL, NULL, '0927426546', 'ธนธนิก บุญมี', NULL, 'thisissine@gmail.com', NULL),
        ('ทาวิน', 'ทาวิน', 'ทาวิน', 'Tavin Humbert Pitisant', NULL, NULL, NULL, '0824938899', 'อลีนา ปิติสันต์', NULL, 'Apitisant@yahoo.com', NULL),
        ('ทีเจ', 'ทีเจ', 'ทีเจ', 'Chawanakorn Sornprasit', NULL, NULL, NULL, '0616216444', 'ฉัตรฐพิชญาภรณ์', NULL, 'sornprasitjenny@gmail.com', '31 ม.ค. ขอเรียน; 10.00-12.00 น.'),
        ('นีน่า', 'น้องนีน่า', 'น้องนีน่า', NULL, NULL, NULL, NULL, '0826550066', 'กฤตฏา จินดถานนท์', NULL, 'nickchefnick@gmail.com', 'วันที่สมัครเรียน 11 เม.ย. 69'),
        ('มีริน', 'มีริน', 'มีริน', NULL, 'หญิง', 6, 'ไทย', '0823829968', 'ปิยนันท์ เตชถาวรกุล', NULL, NULL, NULL),
        ('ลินลี่', 'ลินลี่', 'ลินลี่', 'lily django', 'หญิง', 5, NULL, '0814274005', 'alisa kulladis', NULL, 'alisa.kulladis@gmail.com', NULL),
        ('วินวิน', 'น้องวินวิน', 'น้องวินวิน', NULL, NULL, NULL, NULL, '0894571747', 'Suchada chertkintwong', NULL, 'inmdany@hotmail.com', NULL),
        ('อลิซ่า', 'น้องอลิซ่า', 'น้องอลิซ่า', NULL, NULL, NULL, NULL, '0644212060', 'วรรณี สมพงษ์', NULL, NULL, 'สะดวกหลังวันที่ 27'),
        ('อะตอม', 'น้องอะตอม', 'น้องอะตอม', NULL, NULL, NULL, NULL, NULL, 'วันเพ็ญ ขนาบแก้ว', 'J_Pen', NULL, NULL),
        ('อะตอม', 'อะตอม', 'อะตอม', 'Prach Raungjutiphophan', NULL, NULL, NULL, '0841591381', 'วันเพ็ญ', NULL, 'wanphen.kharapkaeo@gmail.com', NULL),
        ('อเล็กซ์', 'อเล็กซ์', 'อเล็กซ์', 'Sucheep Tangpanwong', NULL, NULL, NULL, '0865164060', 'สมฤดี ตั้งพรรณ', NULL, 'anne_a26@hotmail.com', 'กิจกรรมวันตรุษจีน; *หาวันชดเชย*'),
        ('เกรซ', 'เกศรินทร์', 'เกศรินทร์ มหาศิริ (เกรซ)', 'Ketsarin Mahasiri', 'หญิง', NULL, 'ไทย', '0819209338', 'Patamaporn Sermsaksasitorn', 'JanePat', 's.patamaporn@gmail.com', NULL),
        ('เกรท', 'กรพจน์', 'กรพจน์ สันติเจริญเลิศ (เกรท)', 'Gornpoj Santicharoenlert', 'ชาย', 5, 'ไทย', '0946456951', 'ไอรินทร์ เสรีกัญญาลักษณ์', 'B-E-S-T', 'iryn.sr@gmail.com', NULL),
        ('เทวา', 'เทวา', 'เทวา', 'Teiva Pitisant Humbert', NULL, NULL, NULL, '0824938899', 'อลีนา ปิติสันต์', NULL, 'Apitisant@yahoo.com', NULL),
        ('เรย์', 'เรย์', 'เรย์', 'Jitas Tunlayadechanont', NULL, NULL, NULL, '0909736912', 'ทรายทอง ตุลยาเดชานนท์', NULL, 's.phadungthong@gmail.com', NULL),
        ('เวรอน', 'น้องเวรอน', 'น้องเวรอน', NULL, NULL, NULL, NULL, '0875444888', 'ณัฐพงษ์ สุขเจริญไกรศรี', 'Manywong', NULL, NULL),
        ('ไพร์ม', 'ไพร์ม', 'ไพร์ม', 'Phonlaphat Kritsadavorakul', NULL, NULL, NULL, '0994651749 0956963652', 'สราวุธ กฤษฎาวรกุล', NULL, 'sarawut.kri@gmail.com', 'กิจกรรมวันตรุษจีน; *หาวันชดเชย*')
),
upsert_students AS (
    INSERT INTO students (nickname, first_name, full_name_th, full_name_en, gender, age, nationality, parent_phone, parent_name, line_name, email, note)
    SELECT nickname, first_name, full_name_th, full_name_en, gender, age, nationality, parent_phone, parent_name, line_name, email, note FROM student_seed
    ON CONFLICT (nickname, first_name) DO UPDATE SET
        full_name_th = EXCLUDED.full_name_th,
        full_name_en = EXCLUDED.full_name_en,
        gender = EXCLUDED.gender,
        age = EXCLUDED.age,
        nationality = EXCLUDED.nationality,
        parent_phone = EXCLUDED.parent_phone,
        parent_name = EXCLUDED.parent_name,
        line_name = EXCLUDED.line_name,
        email = EXCLUDED.email,
        note = EXCLUDED.note,
        updated_at = NOW()
    RETURNING id, nickname, first_name
),
enrollment_seed (nickname, first_name, course_name, total_hours, completed_hours, default_session_hours, note) AS (
    VALUES
        ('ชาลี', 'ชาลี', 'Basic (3D) รุ่นที่ 1', 8, 0, 2, NULL),
        ('อเล็กซ์', 'อเล็กซ์', 'Basic (3D) รุ่นที่ 1', 8, 0, 2, NULL),
        ('ไพร์ม', 'ไพร์ม', 'Basic (3D) รุ่นที่ 1', 8, 0, 2, NULL),
        ('ชาลี', 'ชาลี', 'Intermediate (Programming) รุ่นที่ 1', 14, 0, 2, NULL),
        ('อเล็กซ์', 'อเล็กซ์', 'Intermediate (Programming) รุ่นที่ 1', 14, 0, 2, NULL),
        ('ไพร์ม', 'ไพร์ม', 'Intermediate (Programming) รุ่นที่ 1', 14, 0, 2, NULL),
        ('ฌาณ', 'ฌาณ', 'Junior Robotics Challenge (Lego Wedo)', 20, 0, 2, NULL),
        ('ฌาณ', 'ฌาณ', 'Junior Robotics Discovery (Lego Wedo)', 20, 0, 2, NULL),
        ('ลินลี่', 'ลินลี่', 'Junior Robotics Discovery (Lego Wedo)', 20, 0, 2, NULL),
        ('นีน่า', 'น้องนีน่า', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('มีริน', 'มีริน', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('วินวิน', 'น้องวินวิน', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('อลิซ่า', 'น้องอลิซ่า', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('เกรซ', 'เกศรินทร์', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('เกรท', 'กรพจน์', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('เวรอน', 'น้องเวรอน', 'Little 3D Inventors (Basic)', 8, 0, 2, NULL),
        ('Daniel', 'Daniel', 'Little 3D Inventors (Pro)', 8, 0, 2, NULL),
        ('อะตอม', 'น้องอะตอม', 'Little 3D Inventors (Pro)', 8, 0, 2, NULL),
        ('ชาลี', 'ชาลี', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('ฌาณ', 'ฌาณ', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('ทาวิน', 'ทาวิน', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('ทีเจ', 'ทีเจ', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('อะตอม', 'อะตอม', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('เทวา', 'เทวา', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('เรย์', 'เรย์', 'Little 3D รุ่นที่ 1', 10, 0, 2, NULL),
        ('ชาลี', 'ชาลี', 'Pre-intermediate (Circuits) รุ่นที่ 1', 10, 0, 2, NULL),
        ('อเล็กซ์', 'อเล็กซ์', 'Pre-intermediate (Circuits) รุ่นที่ 1', 10, 0, 2, NULL),
        ('ไพร์ม', 'ไพร์ม', 'Pre-intermediate (Circuits) รุ่นที่ 1', 10, 0, 2, NULL)
)
INSERT INTO enrollments (student_id, course_id, total_hours, completed_hours, default_session_hours, active, note)
SELECT upsert_students.id, upsert_courses.id, enrollment_seed.total_hours, enrollment_seed.completed_hours, enrollment_seed.default_session_hours, true, enrollment_seed.note
FROM enrollment_seed
JOIN upsert_students USING (nickname, first_name)
JOIN upsert_courses ON upsert_courses.name = enrollment_seed.course_name
ON CONFLICT (student_id, course_id) DO UPDATE SET
    total_hours = EXCLUDED.total_hours,
    completed_hours = EXCLUDED.completed_hours,
    default_session_hours = EXCLUDED.default_session_hours,
    active = true,
    note = EXCLUDED.note,
    updated_at = NOW();

WITH session_seed (nickname, first_name, course_name, sequence_no, start_at, end_at, status, note) AS (
    VALUES
        ('อลิซ่า', 'น้องอลิซ่า', 'Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-04-12 10:00:00+07', TIMESTAMPTZ '2026-04-12 12:00:00+07', 'unconfirmed', NULL),
        ('เกรซ', 'เกศรินทร์', 'Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-03-28 13:00:00+07', TIMESTAMPTZ '2026-03-28 15:00:00+07', 'unconfirmed', NULL),
        ('เกรซ', 'เกศรินทร์', 'Little 3D Inventors (Basic)', 2, TIMESTAMPTZ '2026-04-25 10:00:00+07', TIMESTAMPTZ '2026-04-25 12:00:00+07', 'unconfirmed', NULL),
        ('เกรท', 'กรพจน์', 'Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-03-24 10:00:00+07', TIMESTAMPTZ '2026-03-24 12:00:00+07', 'unconfirmed', NULL),
        ('เกรท', 'กรพจน์', 'Little 3D Inventors (Basic)', 2, TIMESTAMPTZ '2026-03-31 10:00:00+07', TIMESTAMPTZ '2026-03-31 12:00:00+07', 'unconfirmed', NULL),
        ('Daniel', 'Daniel', 'Little 3D Inventors (Pro)', 1, TIMESTAMPTZ '2026-03-22 15:30:00+07', TIMESTAMPTZ '2026-03-22 17:30:00+07', 'unconfirmed', NULL),
        ('Daniel', 'Daniel', 'Little 3D Inventors (Pro)', 2, TIMESTAMPTZ '2026-04-04 15:30:00+07', TIMESTAMPTZ '2026-04-04 17:30:00+07', 'unconfirmed', NULL)
),
target_enrollments AS (
    SELECT enrollments.id AS enrollment_id, students.nickname, students.first_name, courses.name AS course_name
    FROM enrollments
    JOIN students ON students.id = enrollments.student_id
    JOIN courses ON courses.id = enrollments.course_id
)
INSERT INTO lesson_sessions (enrollment_id, sequence_no, start_at, end_at, status, note)
SELECT target_enrollments.enrollment_id, session_seed.sequence_no, session_seed.start_at, session_seed.end_at, session_seed.status, session_seed.note
FROM session_seed
JOIN target_enrollments USING (nickname, first_name, course_name)
ON CONFLICT (enrollment_id, sequence_no) DO UPDATE SET
    start_at = EXCLUDED.start_at,
    end_at = EXCLUDED.end_at,
    status = EXCLUDED.status,
    note = EXCLUDED.note,
    updated_at = NOW();
