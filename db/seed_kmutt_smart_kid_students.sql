-- Generated from Class schedule CSV files in C:\Users\Thari\Downloads\student
-- Source files: Class schedule - Overview.csv, Class schedule - timetable.csv, Class schedule - current-week.csv
-- Generated date: 2026-05-07
-- Courses: 7, unique students: 13, enrollments: 21, default schedules: 21, seeded sessions: 94

UPDATE enrollments
SET active = false,
    updated_at = NOW()
FROM courses
WHERE enrollments.course_id = courses.id
  AND courses.description LIKE 'Seeded from%CSV export%';

UPDATE course_default_schedules
SET active = false,
    updated_at = NOW()
FROM courses
WHERE course_default_schedules.course_id = courses.id
  AND courses.description LIKE 'Seeded from%CSV export%';

DELETE FROM enrollment_default_schedules
USING enrollments
JOIN courses ON courses.id = enrollments.course_id
WHERE enrollment_default_schedules.enrollment_id = enrollments.id
  AND courses.description LIKE 'Seeded from%CSV export%';

DELETE FROM lesson_sessions
USING enrollments
JOIN courses ON courses.id = enrollments.course_id
WHERE lesson_sessions.enrollment_id = enrollments.id
  AND courses.description LIKE 'Seeded from%CSV export%'
  AND lesson_sessions.note LIKE 'Seeded%Class schedule%';

WITH course_seed (name, description, default_total_hours, default_session_hours) AS (
    VALUES
        ('Basic Create (Design 3D)', 'Seeded from class schedule CSV export | Robogenesis', 8, 1),
        ('INTERMEDIATE CONTROL (Programming)', 'Seeded from class schedule CSV export | Robogenesis', 8, 1),
        ('Level #1 – Junior Robotics Discovery (Lego Wedo)', 'Seeded from class schedule CSV export | RoboJourney: From Discovery to Mission', 20, 1),
        ('Level #1 – Little 3D Inventors (Basic)', 'Seeded from class schedule CSV export | Little 3D Inventors Series', 8, 1),
        ('Level #2 – Little 3D Inventors (Pro)', 'Seeded from class schedule CSV export | Little 3D Inventors Series', 8, 1),
        ('Level #3 – Little 3D Inventors (Advanced)', 'Seeded from class schedule CSV export | Little 3D Inventors Series', 16, 1),
        ('PRE-INTERMEDIATE POWER (CIRCUITS)', 'Seeded from class schedule CSV export | Robogenesis', 16, 1)
)
INSERT INTO courses (name, description, default_total_hours, default_session_hours)
SELECT name, description, default_total_hours, default_session_hours FROM course_seed
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    default_total_hours = EXCLUDED.default_total_hours,
    default_session_hours = EXCLUDED.default_session_hours,
    updated_at = NOW();

WITH student_seed (source_key, nickname, first_name, full_name_th, full_name_en, gender, age, nationality, parent_phone, parent_name, line_name, email, note) AS (
    VALUES
        ('class-schedule:01d6c3649be334c6', 'Daniel', 'Jaeyong', 'Jaeyong Shin', 'Jaeyong Shin', NULL, NULL, NULL, '0930462797', 'Juyoung Park', NULL, NULL, NULL),
        ('class-schedule:0e11e0e791f1d472', 'กิสโม่', 'อัศจรรย์', 'อัศจรรย์ ธรรมสังคีติ', NULL, NULL, NULL, NULL, '0839098777', 'Patcha Bank', NULL, NULL, NULL),
        ('class-schedule:108e707e83f3d78a', 'ฌาณ', 'Chatchon', 'Chatchon Chayathab', 'Chatchon Chayathab', NULL, NULL, NULL, '0927426546', 'ธนธนิก บุญมี', NULL, NULL, NULL),
        ('class-schedule:235c3d4f2c83ee03', 'เกรท', 'กรพจน์', 'กรพจน์ สันติเจริญเลิศ', NULL, NULL, NULL, NULL, '0946456951', 'ไอรินทร์ เสรีกัญญาลักษณ์', NULL, NULL, NULL),
        ('class-schedule:2f89b74f54b8c9a0', 'ไพร์ม', 'Phonlaphat', 'Phonlaphat Kritsadavorakul', 'Phonlaphat Kritsadavorakul', NULL, NULL, NULL, '0994651749', 'สราวุธ กฤษฎาวรกุล', NULL, NULL, NULL),
        ('class-schedule:34defae8b4b54a32', 'ชาลี', 'Yik', 'Yik Hei Sung', 'Yik Hei Sung', NULL, NULL, NULL, '0637203986', 'Ms.ELSA', NULL, NULL, NULL),
        ('class-schedule:49b0b1435f12cb50', 'อเล็กซ์', 'Sucheep', 'Sucheep Tangpanwong', 'Sucheep Tangpanwong', NULL, NULL, NULL, '0865164060', 'สมฤดี ตั้งพรรณ', NULL, NULL, NULL),
        ('class-schedule:5f7183a79fb0f918', 'ไททัน', 'ภูมิณภัทร์', 'ดช ภูมิณภัทร์ โอศิริ', NULL, NULL, NULL, NULL, '0896146541', 'ไลล่า โกดารี', NULL, NULL, NULL),
        ('class-schedule:720f12c43df8ba28', 'TJ', 'Bira', 'Bira Ningsanond', 'Bira Ningsanond', NULL, NULL, NULL, '0894960090', 'T Nathee Ningsanond', NULL, NULL, NULL),
        ('class-schedule:a1a6659ef36406f5', 'เกรซ', 'เกศรินทร์', 'เกศรินทร์ มหาศิริ', NULL, NULL, NULL, NULL, '0819209338', 'Patamaporn Sermsaksasitorn', NULL, NULL, NULL),
        ('class-schedule:aecf236fa29f672f', 'นีน่า', 'นีน่า', 'นีน่า', NULL, NULL, NULL, NULL, '0826550006', 'ศุภรา สุขสันติสวัสดิ์', NULL, NULL, NULL),
        ('class-schedule:b47276be4ba4b650', 'ปราชญ์', 'Prach', 'Prach Raungjutiphophan', 'Prach Raungjutiphophan', NULL, NULL, NULL, '0841591381', 'วันเพ็ญ', NULL, NULL, NULL),
        ('class-schedule:d636489fd37c9e28', 'wind wind', 'wind', 'wind wind', 'wind wind', NULL, NULL, NULL, '0894571747', 'suchada chertkiatwong', NULL, NULL, NULL)
)
INSERT INTO students (source_key, nickname, first_name, full_name_th, full_name_en, gender, age, nationality, parent_phone, parent_name, line_name, email, note)
SELECT source_key, nickname, first_name, full_name_th, full_name_en, gender, age, nationality, parent_phone, parent_name, line_name, email, note FROM student_seed
ON CONFLICT (source_key) WHERE source_key IS NOT NULL DO UPDATE SET
    nickname = EXCLUDED.nickname,
    first_name = EXCLUDED.first_name,
    full_name_th = EXCLUDED.full_name_th,
    full_name_en = EXCLUDED.full_name_en,
    parent_phone = EXCLUDED.parent_phone,
    parent_name = EXCLUDED.parent_name,
    note = EXCLUDED.note,
    updated_at = NOW();

WITH enrollment_seed (source_key, course_name, total_hours, completed_hours, default_session_hours, teacher, started_on, external_status, active, note) AS (
    VALUES
        ('class-schedule:01d6c3649be334c6', 'Level #2 – Little 3D Inventors (Pro)', 8, 8, 1, 'Kung', DATE '2026-03-22', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 16, 0, 1, 'Kung', DATE '2026-04-07', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 8, 0, 1, 'Test', DATE '2026-04-26', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:108e707e83f3d78a', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 20, 16, 1, 'Kung', DATE '2026-02-28', 'กำลังเรียน', true, 'RoboJourney: From Discovery to Mission'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 8, 4, 1, 'Test', DATE '2026-03-24', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:2f89b74f54b8c9a0', 'Basic Create (Design 3D)', 8, 0, 1, 'Test', DATE '2026-01-18', 'เรียนจบแล้ว', false, 'Robogenesis'),
        ('class-schedule:2f89b74f54b8c9a0', 'INTERMEDIATE CONTROL (Programming)', 8, 8, 1, 'Test', DATE '2026-04-04', 'เรียนจบแล้ว', false, 'Robogenesis'),
        ('class-schedule:2f89b74f54b8c9a0', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 8, 0, 1, 'Test', DATE '2026-02-22', 'เรียนจบแล้ว', false, 'Robogenesis'),
        ('class-schedule:34defae8b4b54a32', 'Basic Create (Design 3D)', 8, 0, 1, 'Kung', DATE '2026-01-18', 'เรียนจบแล้ว', false, 'Robogenesis'),
        ('class-schedule:34defae8b4b54a32', 'INTERMEDIATE CONTROL (Programming)', 8, 0, 1, 'Kung', DATE '2026-03-29', 'เรียนจบแล้ว', false, 'Robogenesis'),
        ('class-schedule:34defae8b4b54a32', 'Level #1 – Little 3D Inventors (Basic)', 8, 0, 1, 'Kung', DATE '2026-01-24', 'เรียนจบแล้ว', false, 'Little 3D Inventors Series'),
        ('class-schedule:34defae8b4b54a32', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 8, 0, 1, 'Kung', DATE '2026-02-22', 'เรียนจบแล้ว', false, 'Robogenesis'),
        ('class-schedule:49b0b1435f12cb50', 'Basic Create (Design 3D)', 8, 0, 1, 'Kung', DATE '2026-01-18', 'ออกแล้ว', false, 'Robogenesis'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 20, 14, 1, 'Test', DATE '2026-02-08', 'กำลังเรียน', true, 'RoboJourney: From Discovery to Mission'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 20, 2, 1, 'Kung', DATE '2026-04-26', 'กำลังเรียน', true, 'RoboJourney: From Discovery to Mission'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 8, 0, 1, 'Kawin Yamtuan', DATE '2026-04-26', 'รอเริ่ม', true, 'Little 3D Inventors Series'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 8, 0, 1, 'NookNick', DATE '2026-03-28', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 8, 0, 1, 'Test', DATE '2026-04-10', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 8, 0, 1, 'Kung', DATE '2026-05-02', 'กำลังเรียน', true, 'Little 3D Inventors Series'),
        ('class-schedule:d636489fd37c9e28', 'Level #1 – Little 3D Inventors (Basic)', 8, 0, 1, 'Boss', DATE '2026-03-29', 'เรียนจบแล้ว', false, 'Little 3D Inventors Series'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 16, 0, 1, 'Test', DATE '2026-04-25', 'กำลังเรียน', true, 'Robogenesis')
), target AS (
    SELECT students.id AS student_id, courses.id AS course_id, enrollment_seed.*
    FROM enrollment_seed
    JOIN students ON students.source_key = enrollment_seed.source_key
    JOIN courses ON courses.name = enrollment_seed.course_name
)
INSERT INTO enrollments (student_id, course_id, total_hours, completed_hours, default_session_hours, teacher, started_on, external_status, active, note)
SELECT student_id, course_id, total_hours, completed_hours, default_session_hours, teacher, started_on, external_status, active, note FROM target
ON CONFLICT (student_id, course_id) DO UPDATE SET
    total_hours = EXCLUDED.total_hours,
    completed_hours = EXCLUDED.completed_hours,
    default_session_hours = EXCLUDED.default_session_hours,
    teacher = EXCLUDED.teacher,
    started_on = EXCLUDED.started_on,
    external_status = EXCLUDED.external_status,
    active = EXCLUDED.active,
    note = EXCLUDED.note,
    updated_at = NOW();

WITH course_default_schedule_seed (course_name, sequence_no, default_day_text, session_label, date_label, scheduled_date, slot_label, start_time, end_time, note) AS (
    VALUES
        ('Basic Create (Design 3D)', 1, 'อาทิตย์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'Robogenesis'),
        ('Basic Create (Design 3D)', 2, 'อาทิตย์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Robogenesis'),
        ('INTERMEDIATE CONTROL (Programming)', 1, 'อาทิตย์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Robogenesis'),
        ('INTERMEDIATE CONTROL (Programming)', 2, 'เสาร์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Robogenesis'),
        ('Level #1 – Junior Robotics Discovery (Lego Wedo)', 1, 'อาทิตย์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'RoboJourney: From Discovery to Mission'),
        ('Level #1 – Junior Robotics Discovery (Lego Wedo)', 2, 'อาทิตย์', NULL, NULL, NULL, '16:00-17:00', TIME '16:00:00', TIME '17:00:00', 'RoboJourney: From Discovery to Mission'),
        ('Level #1 – Junior Robotics Discovery (Lego Wedo)', 3, 'เสาร์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'RoboJourney: From Discovery to Mission'),
        ('Level #1 – Little 3D Inventors (Basic)', 1, 'ศุกร์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'Little 3D Inventors Series'),
        ('Level #1 – Little 3D Inventors (Basic)', 2, 'อังคาร', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'Little 3D Inventors Series'),
        ('Level #1 – Little 3D Inventors (Basic)', 3, 'อาทิตย์', NULL, NULL, NULL, '09:00-10:00', TIME '09:00:00', TIME '10:00:00', 'Little 3D Inventors Series'),
        ('Level #1 – Little 3D Inventors (Basic)', 4, 'อาทิตย์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'Little 3D Inventors Series'),
        ('Level #1 – Little 3D Inventors (Basic)', 5, 'อาทิตย์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Little 3D Inventors Series'),
        ('Level #1 – Little 3D Inventors (Basic)', 6, 'เสาร์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Little 3D Inventors Series'),
        ('Level #2 – Little 3D Inventors (Pro)', 1, 'อาทิตย์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'Little 3D Inventors Series'),
        ('Level #2 – Little 3D Inventors (Pro)', 2, 'เสาร์', NULL, NULL, NULL, '15:00-16:00', TIME '15:00:00', TIME '16:00:00', 'Little 3D Inventors Series'),
        ('Level #3 – Little 3D Inventors (Advanced)', 1, 'อาทิตย์', NULL, NULL, NULL, '10:00-11:00', TIME '10:00:00', TIME '11:00:00', 'Little 3D Inventors Series'),
        ('PRE-INTERMEDIATE POWER (CIRCUITS)', 1, 'อาทิตย์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Robogenesis'),
        ('PRE-INTERMEDIATE POWER (CIRCUITS)', 2, 'เสาร์', NULL, NULL, NULL, '13:00-14:00', TIME '13:00:00', TIME '14:00:00', 'Robogenesis')
), target AS (
    SELECT courses.id AS course_id, seed.*
    FROM course_default_schedule_seed seed
    JOIN courses ON courses.name = seed.course_name
)
INSERT INTO course_default_schedules (course_id, sequence_no, default_day_text, session_label, date_label, scheduled_date, slot_label, start_time, end_time, active, note)
SELECT course_id, sequence_no, default_day_text, session_label, date_label, scheduled_date, slot_label, start_time, end_time, true, note FROM target
ON CONFLICT (course_id, sequence_no) DO UPDATE SET
    default_day_text = EXCLUDED.default_day_text,
    session_label = EXCLUDED.session_label,
    date_label = EXCLUDED.date_label,
    scheduled_date = EXCLUDED.scheduled_date,
    slot_label = EXCLUDED.slot_label,
    start_time = EXCLUDED.start_time,
    end_time = EXCLUDED.end_time,
    active = true,
    note = EXCLUDED.note,
    updated_at = NOW();

WITH enrollment_default_schedule_seed (source_key, course_name, weekday_text, start_time, end_time, note) AS (
    VALUES
        ('class-schedule:01d6c3649be334c6', 'Level #2 – Little 3D Inventors (Pro)', 'อาทิตย์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 'อาทิตย์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 'อาทิตย์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:108e707e83f3d78a', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 'เสาร์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 'อังคาร', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:2f89b74f54b8c9a0', 'Basic Create (Design 3D)', 'อาทิตย์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:2f89b74f54b8c9a0', 'INTERMEDIATE CONTROL (Programming)', 'เสาร์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:2f89b74f54b8c9a0', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 'อาทิตย์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:34defae8b4b54a32', 'Basic Create (Design 3D)', 'อาทิตย์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:34defae8b4b54a32', 'INTERMEDIATE CONTROL (Programming)', 'อาทิตย์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:34defae8b4b54a32', 'Level #1 – Little 3D Inventors (Basic)', 'เสาร์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:34defae8b4b54a32', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 'อาทิตย์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:49b0b1435f12cb50', 'Basic Create (Design 3D)', 'อาทิตย์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 'อาทิตย์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 'อาทิตย์', TIME '16:00:00', TIME '17:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 'อาทิตย์', TIME '09:00:00', TIME '10:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 'เสาร์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 'ศุกร์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 'เสาร์', TIME '15:00:00', TIME '16:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:d636489fd37c9e28', 'Level #1 – Little 3D Inventors (Basic)', 'อาทิตย์', TIME '10:00:00', TIME '11:00:00', 'Seeded from Class schedule - Overview.csv'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 'เสาร์', TIME '13:00:00', TIME '14:00:00', 'Seeded from Class schedule - Overview.csv')
), target AS (
    SELECT enrollments.id AS enrollment_id, seed.*
    FROM enrollment_default_schedule_seed seed
    JOIN students ON students.source_key = seed.source_key
    JOIN courses ON courses.name = seed.course_name
    JOIN enrollments ON enrollments.student_id = students.id AND enrollments.course_id = courses.id
)
INSERT INTO enrollment_default_schedules (enrollment_id, weekday_text, start_time, end_time, active, note)
SELECT enrollment_id, weekday_text, start_time, end_time, true, note FROM target
ON CONFLICT (enrollment_id, weekday_text, start_time, end_time) DO UPDATE SET
    active = true,
    note = EXCLUDED.note,
    updated_at = NOW();

WITH session_seed (source_key, course_name, sequence_no, start_at, end_at, status, note) AS (
    VALUES
        ('class-schedule:108e707e83f3d78a', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 17, TIMESTAMPTZ '2026-05-09 10:00:00+07', TIMESTAMPTZ '2026-05-09 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:108e707e83f3d78a', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 18, TIMESTAMPTZ '2026-05-16 10:00:00+07', TIMESTAMPTZ '2026-05-16 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:108e707e83f3d78a', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 19, TIMESTAMPTZ '2026-05-23 10:00:00+07', TIMESTAMPTZ '2026-05-23 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:108e707e83f3d78a', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 20, TIMESTAMPTZ '2026-05-30 10:00:00+07', TIMESTAMPTZ '2026-05-30 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 15, TIMESTAMPTZ '2026-05-10 10:00:00+07', TIMESTAMPTZ '2026-05-10 11:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 16, TIMESTAMPTZ '2026-05-17 10:00:00+07', TIMESTAMPTZ '2026-05-17 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 17, TIMESTAMPTZ '2026-05-24 10:00:00+07', TIMESTAMPTZ '2026-05-24 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 18, TIMESTAMPTZ '2026-05-31 10:00:00+07', TIMESTAMPTZ '2026-05-31 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 19, TIMESTAMPTZ '2026-06-07 10:00:00+07', TIMESTAMPTZ '2026-06-07 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:49b0b1435f12cb50', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 20, TIMESTAMPTZ '2026-06-14 10:00:00+07', TIMESTAMPTZ '2026-06-14 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 3, TIMESTAMPTZ '2026-05-10 16:00:00+07', TIMESTAMPTZ '2026-05-10 17:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 4, TIMESTAMPTZ '2026-05-17 16:00:00+07', TIMESTAMPTZ '2026-05-17 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 5, TIMESTAMPTZ '2026-05-24 16:00:00+07', TIMESTAMPTZ '2026-05-24 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 6, TIMESTAMPTZ '2026-05-31 16:00:00+07', TIMESTAMPTZ '2026-05-31 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 7, TIMESTAMPTZ '2026-06-07 16:00:00+07', TIMESTAMPTZ '2026-06-07 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 8, TIMESTAMPTZ '2026-06-14 16:00:00+07', TIMESTAMPTZ '2026-06-14 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 9, TIMESTAMPTZ '2026-06-21 16:00:00+07', TIMESTAMPTZ '2026-06-21 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 10, TIMESTAMPTZ '2026-06-28 16:00:00+07', TIMESTAMPTZ '2026-06-28 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 11, TIMESTAMPTZ '2026-07-05 16:00:00+07', TIMESTAMPTZ '2026-07-05 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 12, TIMESTAMPTZ '2026-07-12 16:00:00+07', TIMESTAMPTZ '2026-07-12 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 13, TIMESTAMPTZ '2026-07-19 16:00:00+07', TIMESTAMPTZ '2026-07-19 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:5f7183a79fb0f918', 'Level #1 – Junior Robotics Discovery (Lego Wedo)', 14, TIMESTAMPTZ '2026-07-26 16:00:00+07', TIMESTAMPTZ '2026-07-26 17:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-05-10 13:00:00+07', TIMESTAMPTZ '2026-05-10 14:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 2, TIMESTAMPTZ '2026-05-17 13:00:00+07', TIMESTAMPTZ '2026-05-17 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 3, TIMESTAMPTZ '2026-05-24 13:00:00+07', TIMESTAMPTZ '2026-05-24 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 4, TIMESTAMPTZ '2026-05-31 13:00:00+07', TIMESTAMPTZ '2026-05-31 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 5, TIMESTAMPTZ '2026-06-07 13:00:00+07', TIMESTAMPTZ '2026-06-07 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 6, TIMESTAMPTZ '2026-06-14 13:00:00+07', TIMESTAMPTZ '2026-06-14 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 7, TIMESTAMPTZ '2026-06-21 13:00:00+07', TIMESTAMPTZ '2026-06-21 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:0e11e0e791f1d472', 'Level #1 – Little 3D Inventors (Basic)', 8, TIMESTAMPTZ '2026-06-28 13:00:00+07', TIMESTAMPTZ '2026-06-28 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 5, TIMESTAMPTZ '2026-05-05 10:00:00+07', TIMESTAMPTZ '2026-05-05 11:00:00+07', 'completed', 'Seeded from Class schedule CSV'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 6, TIMESTAMPTZ '2026-05-12 10:00:00+07', TIMESTAMPTZ '2026-05-12 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 7, TIMESTAMPTZ '2026-05-19 10:00:00+07', TIMESTAMPTZ '2026-05-19 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 8, TIMESTAMPTZ '2026-05-26 10:00:00+07', TIMESTAMPTZ '2026-05-26 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:235c3d4f2c83ee03', 'Level #1 – Little 3D Inventors (Basic)', 9, TIMESTAMPTZ '2026-06-02 10:00:00+07', TIMESTAMPTZ '2026-06-02 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-05-10 09:00:00+07', TIMESTAMPTZ '2026-05-10 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 2, TIMESTAMPTZ '2026-05-10 15:00:00+07', TIMESTAMPTZ '2026-05-10 16:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 3, TIMESTAMPTZ '2026-05-17 09:00:00+07', TIMESTAMPTZ '2026-05-17 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 4, TIMESTAMPTZ '2026-05-24 09:00:00+07', TIMESTAMPTZ '2026-05-24 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 5, TIMESTAMPTZ '2026-05-31 09:00:00+07', TIMESTAMPTZ '2026-05-31 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 6, TIMESTAMPTZ '2026-06-07 09:00:00+07', TIMESTAMPTZ '2026-06-07 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 7, TIMESTAMPTZ '2026-06-14 09:00:00+07', TIMESTAMPTZ '2026-06-14 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 8, TIMESTAMPTZ '2026-06-21 09:00:00+07', TIMESTAMPTZ '2026-06-21 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:720f12c43df8ba28', 'Level #1 – Little 3D Inventors (Basic)', 9, TIMESTAMPTZ '2026-06-28 09:00:00+07', TIMESTAMPTZ '2026-06-28 10:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-05-09 13:00:00+07', TIMESTAMPTZ '2026-05-09 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 2, TIMESTAMPTZ '2026-05-16 13:00:00+07', TIMESTAMPTZ '2026-05-16 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 3, TIMESTAMPTZ '2026-05-23 13:00:00+07', TIMESTAMPTZ '2026-05-23 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 4, TIMESTAMPTZ '2026-05-30 13:00:00+07', TIMESTAMPTZ '2026-05-30 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 5, TIMESTAMPTZ '2026-06-06 13:00:00+07', TIMESTAMPTZ '2026-06-06 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 6, TIMESTAMPTZ '2026-06-13 13:00:00+07', TIMESTAMPTZ '2026-06-13 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 7, TIMESTAMPTZ '2026-06-20 13:00:00+07', TIMESTAMPTZ '2026-06-20 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:a1a6659ef36406f5', 'Level #1 – Little 3D Inventors (Basic)', 8, TIMESTAMPTZ '2026-06-27 13:00:00+07', TIMESTAMPTZ '2026-06-27 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 1, TIMESTAMPTZ '2026-05-08 10:00:00+07', TIMESTAMPTZ '2026-05-08 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 2, TIMESTAMPTZ '2026-05-15 10:00:00+07', TIMESTAMPTZ '2026-05-15 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 3, TIMESTAMPTZ '2026-05-22 10:00:00+07', TIMESTAMPTZ '2026-05-22 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 4, TIMESTAMPTZ '2026-05-29 10:00:00+07', TIMESTAMPTZ '2026-05-29 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 5, TIMESTAMPTZ '2026-06-05 10:00:00+07', TIMESTAMPTZ '2026-06-05 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 6, TIMESTAMPTZ '2026-06-12 10:00:00+07', TIMESTAMPTZ '2026-06-12 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 7, TIMESTAMPTZ '2026-06-19 10:00:00+07', TIMESTAMPTZ '2026-06-19 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:aecf236fa29f672f', 'Level #1 – Little 3D Inventors (Basic)', 8, TIMESTAMPTZ '2026-06-26 10:00:00+07', TIMESTAMPTZ '2026-06-26 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 1, TIMESTAMPTZ '2026-05-09 15:00:00+07', TIMESTAMPTZ '2026-05-09 16:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 2, TIMESTAMPTZ '2026-05-16 15:00:00+07', TIMESTAMPTZ '2026-05-16 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 3, TIMESTAMPTZ '2026-05-23 15:00:00+07', TIMESTAMPTZ '2026-05-23 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 4, TIMESTAMPTZ '2026-05-30 15:00:00+07', TIMESTAMPTZ '2026-05-30 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 5, TIMESTAMPTZ '2026-06-06 15:00:00+07', TIMESTAMPTZ '2026-06-06 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 6, TIMESTAMPTZ '2026-06-13 15:00:00+07', TIMESTAMPTZ '2026-06-13 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 7, TIMESTAMPTZ '2026-06-20 15:00:00+07', TIMESTAMPTZ '2026-06-20 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:b47276be4ba4b650', 'Level #2 – Little 3D Inventors (Pro)', 8, TIMESTAMPTZ '2026-06-27 15:00:00+07', TIMESTAMPTZ '2026-06-27 16:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 1, TIMESTAMPTZ '2026-05-05 10:00:00+07', TIMESTAMPTZ '2026-05-05 11:00:00+07', 'completed', 'Seeded from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 2, TIMESTAMPTZ '2026-05-10 10:00:00+07', TIMESTAMPTZ '2026-05-10 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 3, TIMESTAMPTZ '2026-05-17 10:00:00+07', TIMESTAMPTZ '2026-05-17 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 4, TIMESTAMPTZ '2026-05-24 10:00:00+07', TIMESTAMPTZ '2026-05-24 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 5, TIMESTAMPTZ '2026-05-31 10:00:00+07', TIMESTAMPTZ '2026-05-31 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 6, TIMESTAMPTZ '2026-06-07 10:00:00+07', TIMESTAMPTZ '2026-06-07 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 7, TIMESTAMPTZ '2026-06-14 10:00:00+07', TIMESTAMPTZ '2026-06-14 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 8, TIMESTAMPTZ '2026-06-21 10:00:00+07', TIMESTAMPTZ '2026-06-21 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 9, TIMESTAMPTZ '2026-06-28 10:00:00+07', TIMESTAMPTZ '2026-06-28 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 10, TIMESTAMPTZ '2026-07-05 10:00:00+07', TIMESTAMPTZ '2026-07-05 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 11, TIMESTAMPTZ '2026-07-12 10:00:00+07', TIMESTAMPTZ '2026-07-12 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 12, TIMESTAMPTZ '2026-07-19 10:00:00+07', TIMESTAMPTZ '2026-07-19 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:01d6c3649be334c6', 'Level #3 – Little 3D Inventors (Advanced)', 13, TIMESTAMPTZ '2026-07-26 10:00:00+07', TIMESTAMPTZ '2026-07-26 11:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 1, TIMESTAMPTZ '2026-05-09 13:00:00+07', TIMESTAMPTZ '2026-05-09 14:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 2, TIMESTAMPTZ '2026-05-10 13:00:00+07', TIMESTAMPTZ '2026-05-10 14:00:00+07', 'unconfirmed', 'Seeded from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 3, TIMESTAMPTZ '2026-05-16 13:00:00+07', TIMESTAMPTZ '2026-05-16 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 4, TIMESTAMPTZ '2026-05-23 13:00:00+07', TIMESTAMPTZ '2026-05-23 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 5, TIMESTAMPTZ '2026-05-30 13:00:00+07', TIMESTAMPTZ '2026-05-30 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 6, TIMESTAMPTZ '2026-06-06 13:00:00+07', TIMESTAMPTZ '2026-06-06 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 7, TIMESTAMPTZ '2026-06-13 13:00:00+07', TIMESTAMPTZ '2026-06-13 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 8, TIMESTAMPTZ '2026-06-20 13:00:00+07', TIMESTAMPTZ '2026-06-20 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 9, TIMESTAMPTZ '2026-06-27 13:00:00+07', TIMESTAMPTZ '2026-06-27 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 10, TIMESTAMPTZ '2026-07-04 13:00:00+07', TIMESTAMPTZ '2026-07-04 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 11, TIMESTAMPTZ '2026-07-11 13:00:00+07', TIMESTAMPTZ '2026-07-11 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 12, TIMESTAMPTZ '2026-07-18 13:00:00+07', TIMESTAMPTZ '2026-07-18 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV'),
        ('class-schedule:d636489fd37c9e28', 'PRE-INTERMEDIATE POWER (CIRCUITS)', 13, TIMESTAMPTZ '2026-07-25 13:00:00+07', TIMESTAMPTZ '2026-07-25 14:00:00+07', 'unconfirmed', 'Seeded recurring from Class schedule CSV')
), target AS (
    SELECT enrollments.id AS enrollment_id, session_seed.*
    FROM session_seed
    JOIN students ON students.source_key = session_seed.source_key
    JOIN courses ON courses.name = session_seed.course_name
    JOIN enrollments ON enrollments.student_id = students.id AND enrollments.course_id = courses.id
)
INSERT INTO lesson_sessions (enrollment_id, sequence_no, start_at, end_at, status, note)
SELECT enrollment_id, sequence_no, start_at, end_at, status, note FROM target
ON CONFLICT (enrollment_id, sequence_no) DO UPDATE SET
    start_at = EXCLUDED.start_at,
    end_at = EXCLUDED.end_at,
    status = EXCLUDED.status,
    note = EXCLUDED.note,
    updated_at = NOW();
