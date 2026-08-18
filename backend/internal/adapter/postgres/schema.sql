-- MAHAK volunteer module schema
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('volunteer', 'admin', 'operator')),
    external_user_id TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS volunteers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    full_name TEXT NOT NULL,
    national_id TEXT,
    phone TEXT,
    city TEXT,
    bio TEXT,
    skill_categories TEXT[] NOT NULL DEFAULT '{}',
    education_field TEXT,
    medical_license TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    rejection_reason TEXT,
    average_score NUMERIC(6,2) NOT NULL DEFAULT 0,
    total_hours NUMERIC(10,2) NOT NULL DEFAULT 0,
    completed_tasks INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS volunteer_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    weekday INT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    object_key TEXT NOT NULL,
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    capacity INT NOT NULL,
    reserved_count INT NOT NULL DEFAULT 0,
    hour_weight NUMERIC(6,2) NOT NULL,
    required_skills TEXT[] NOT NULL DEFAULT '{}',
    min_score NUMERIC(6,2) NOT NULL DEFAULT 0,
    required_education TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'reserved',
    volunteer_rating INT,
    volunteer_comment TEXT,
    admin_discipline INT,
    admin_expertise INT,
    admin_ethics INT,
    admin_comment TEXT,
    composite_score NUMERIC(6,2),
    hours_awarded NUMERIC(6,2) NOT NULL DEFAULT 0,
    attended_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, volunteer_id)
);

CREATE TABLE IF NOT EXISTS missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    kind TEXT NOT NULL,
    hour_weight NUMERIC(6,2) NOT NULL,
    deadline_hours INT,
    webhook_event TEXT,
    target_count INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mission_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'in_progress',
    progress INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    UNIQUE (mission_id, volunteer_id)
);

CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    verification_code UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    assignment_id UUID REFERENCES assignments(id),
    title TEXT NOT NULL,
    hours NUMERIC(10,2) NOT NULL,
    period_start DATE,
    period_end DATE,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_volunteers_status ON volunteers(status);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_assignments_volunteer ON assignments(volunteer_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);

ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS province TEXT;
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS plaque TEXT;
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS unit TEXT;
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS phone2 TEXT;
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS education_level TEXT;
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS birth_date DATE;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS required_skill_ids UUID[] NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS work_mode TEXT NOT NULL DEFAULT 'onsite';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS delivery_hint TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_note TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_object_key TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_mime TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS skill_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES skill_groups(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS volunteer_skills (
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    PRIMARY KEY (volunteer_id, skill_id)
);

CREATE TABLE IF NOT EXISTS skill_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES skill_groups(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    admin_note TEXT,
    created_skill_id UUID REFERENCES skills(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_skill_proposals_status ON skill_proposals(status);

ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_unique ON users (phone) WHERE phone <> '';

