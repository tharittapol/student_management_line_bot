CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    default_total_hours INTEGER NOT NULL DEFAULT 8 CHECK (default_total_hours >= 0),
    default_session_hours INTEGER NOT NULL DEFAULT 2 CHECK (default_session_hours > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS students (
    id BIGSERIAL PRIMARY KEY,
    source_key TEXT,
    nickname TEXT NOT NULL,
    first_name TEXT NOT NULL,
    full_name_th TEXT,
    full_name_en TEXT,
    gender TEXT,
    age INTEGER CHECK (age IS NULL OR age >= 0),
    nationality TEXT,
    parent_phone TEXT,
    parent_name TEXT,
    line_name TEXT,
    email TEXT,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_nickname_first_name_key;
ALTER TABLE students ADD COLUMN IF NOT EXISTS source_key TEXT;

CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    total_hours INTEGER NOT NULL DEFAULT 8 CHECK (total_hours >= 0),
    completed_hours INTEGER NOT NULL DEFAULT 0 CHECK (completed_hours >= 0),
    default_session_hours INTEGER NOT NULL DEFAULT 2 CHECK (default_session_hours > 0),
    teacher TEXT,
    started_on DATE,
    external_status TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, course_id)
);

ALTER TABLE enrollments ADD COLUMN IF NOT EXISTS teacher TEXT;
ALTER TABLE enrollments ADD COLUMN IF NOT EXISTS started_on DATE;
ALTER TABLE enrollments ADD COLUMN IF NOT EXISTS external_status TEXT;

CREATE TABLE IF NOT EXISTS course_default_schedules (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    sequence_no INTEGER NOT NULL CHECK (sequence_no > 0),
    default_day_text TEXT,
    session_label TEXT,
    date_label TEXT,
    scheduled_date DATE,
    slot_label TEXT NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time),
    UNIQUE (course_id, sequence_no)
);

CREATE TABLE IF NOT EXISTS enrollment_default_schedules (
    id BIGSERIAL PRIMARY KEY,
    enrollment_id BIGINT NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    weekday_text TEXT NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time),
    UNIQUE (enrollment_id, weekday_text, start_time, end_time)
);

CREATE TABLE IF NOT EXISTS lesson_sessions (
    id BIGSERIAL PRIMARY KEY,
    enrollment_id BIGINT NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    sequence_no INTEGER NOT NULL CHECK (sequence_no > 0),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'unconfirmed'
        CHECK (status IN ('scheduled', 'confirmed', 'unconfirmed', 'completed', 'cancelled')),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_at > start_at),
    UNIQUE (enrollment_id, sequence_no)
);

CREATE TABLE IF NOT EXISTS enrollment_schedule_notes (
    id BIGSERIAL PRIMARY KEY,
    enrollment_id BIGINT NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    course_default_schedule_id BIGINT REFERENCES course_default_schedules(id) ON DELETE SET NULL,
    sequence_no INTEGER NOT NULL CHECK (sequence_no > 0),
    source_text TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'note'
        CHECK (status IN ('pending_date', 'pending_confirm', 'note')),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (enrollment_id, sequence_no, source_text)
);

CREATE TABLE IF NOT EXISTS line_groups (
    id BIGSERIAL PRIMARY KEY,
    group_id TEXT NOT NULL UNIQUE,
    display_name TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lesson_sessions_start_at ON lesson_sessions(start_at);
CREATE INDEX IF NOT EXISTS idx_lesson_sessions_enrollment_status ON lesson_sessions(enrollment_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_source_key_unique ON students(source_key) WHERE source_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_course_default_schedules_course ON course_default_schedules(course_id, active);
CREATE INDEX IF NOT EXISTS idx_enrollment_default_schedules_enrollment ON enrollment_default_schedules(enrollment_id, active);
CREATE INDEX IF NOT EXISTS idx_enrollment_schedule_notes_enrollment ON enrollment_schedule_notes(enrollment_id, status);
CREATE INDEX IF NOT EXISTS idx_enrollments_active ON enrollments(active);
CREATE INDEX IF NOT EXISTS idx_line_groups_active ON line_groups(active);
