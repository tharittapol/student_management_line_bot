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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (nickname, first_name)
);

CREATE TABLE IF NOT EXISTS enrollments (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    total_hours INTEGER NOT NULL DEFAULT 8 CHECK (total_hours >= 0),
    completed_hours INTEGER NOT NULL DEFAULT 0 CHECK (completed_hours >= 0),
    default_session_hours INTEGER NOT NULL DEFAULT 2 CHECK (default_session_hours > 0),
    active BOOLEAN NOT NULL DEFAULT true,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, course_id)
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
CREATE INDEX IF NOT EXISTS idx_enrollments_active ON enrollments(active);
CREATE INDEX IF NOT EXISTS idx_line_groups_active ON line_groups(active);
