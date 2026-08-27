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
    verify_mode TEXT NOT NULL DEFAULT 'internal',
    verify_url TEXT NOT NULL DEFAULT '',
    verify_token TEXT NOT NULL DEFAULT '',
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

CREATE TABLE IF NOT EXISTS certificate_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    assignment_id UUID REFERENCES assignments(id),
    status TEXT NOT NULL DEFAULT 'pending',
    admin_note TEXT NOT NULL DEFAULT '',
    certificate_id UUID REFERENCES certificates(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_certificate_requests_status ON certificate_requests (status);
CREATE INDEX IF NOT EXISTS idx_certificate_requests_volunteer ON certificate_requests (volunteer_id, created_at DESC);

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
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '';
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS gender TEXT NOT NULL DEFAULT '';
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS occupation TEXT NOT NULL DEFAULT '';
ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS occupation_other TEXT NOT NULL DEFAULT '';
UPDATE volunteers SET
    first_name = CASE WHEN first_name = '' THEN split_part(btrim(full_name), ' ', 1) ELSE first_name END,
    last_name = CASE WHEN last_name = '' THEN COALESCE(NULLIF(btrim(substr(btrim(full_name), length(split_part(btrim(full_name), ' ', 1)) + 1)), ''), '') ELSE last_name END
WHERE coalesce(full_name, '') <> '';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS required_skill_ids UUID[] NOT NULL DEFAULT '{}';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS work_mode TEXT NOT NULL DEFAULT 'onsite';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS delivery_hint TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'one_off';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS series_id UUID;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS weekday INT;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurrence_slots JSONB NOT NULL DEFAULT '[]';
CREATE INDEX IF NOT EXISTS idx_tasks_series ON tasks (series_id);
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_note TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_object_key TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivery_mime TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS check_in_at TIMESTAMPTZ;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS check_out_at TIMESTAMPTZ;
ALTER TABLE certificate_requests ADD COLUMN IF NOT EXISTS delivery_method TEXT NOT NULL DEFAULT '';
ALTER TABLE certificate_requests ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ;

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS requires_training BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS training_kind TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS training_location TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS training_at TIMESTAMPTZ;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'notice';
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS remind_at TIMESTAMPTZ;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS fired_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_notifications_remind ON notifications (user_id, remind_at) WHERE remind_at IS NOT NULL;

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

ALTER TABLE missions ADD COLUMN IF NOT EXISTS verify_mode TEXT NOT NULL DEFAULT 'internal';
ALTER TABLE missions ADD COLUMN IF NOT EXISTS verify_url TEXT NOT NULL DEFAULT '';
ALTER TABLE missions ADD COLUMN IF NOT EXISTS verify_token TEXT NOT NULL DEFAULT '';

UPDATE missions SET verify_mode = 'inbound'
WHERE kind IN ('invite_users', 'webhook') AND verify_mode = 'internal' AND coalesce(verify_url, '') = '';

UPDATE missions SET verify_token = replace(gen_random_uuid()::text || gen_random_uuid()::text, '-', '')
WHERE verify_token = '' AND verify_mode IN ('inbound', 'outbound');

CREATE INDEX IF NOT EXISTS idx_missions_verify_token ON missions (verify_token) WHERE verify_token <> '';

CREATE TABLE IF NOT EXISTS volunteer_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role TEXT NOT NULL DEFAULT '',
    event_type TEXT NOT NULL,
    from_status TEXT NOT NULL DEFAULT '',
    to_status TEXT NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_volunteer_events_volunteer ON volunteer_events (volunteer_id, created_at DESC);

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets (status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_volunteer ON tickets (volunteer_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    author_role TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket ON ticket_messages (ticket_id, created_at);

CREATE TABLE IF NOT EXISTS volunteer_trainings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volunteer_id UUID NOT NULL REFERENCES volunteers(id) ON DELETE CASCADE,
    series_id UUID,
    training_kind TEXT NOT NULL DEFAULT '',
    training_location TEXT NOT NULL DEFAULT '',
    training_at TIMESTAMPTZ,
    source_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    assignment_id UUID REFERENCES assignments(id) ON DELETE SET NULL,
    confirmed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    confirmed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_volunteer_trainings_vol ON volunteer_trainings (volunteer_id, confirmed_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_volunteer_trainings_assignment ON volunteer_trainings (assignment_id) WHERE assignment_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS assignment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    actor_role TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_assignment_events_asg ON assignment_events (assignment_id, created_at);

CREATE TABLE IF NOT EXISTS assignment_event_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES assignment_events(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL DEFAULT '',
    object_key TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_assignment_event_files_event ON assignment_event_files (event_id);

CREATE TABLE IF NOT EXISTS schema_patches (
    name TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM schema_patches WHERE name = 'reopen_self_reported_missions') THEN
        UPDATE volunteers v SET
            total_hours = GREATEST(0, total_hours - sub.hours),
            average_score = GREATEST(0, average_score - sub.score)
        FROM (
            SELECT p.volunteer_id, COALESCE(SUM(m.hour_weight), 0) AS hours, COUNT(*) * 5 AS score
            FROM mission_progress p
            JOIN missions m ON m.id = p.mission_id
            WHERE p.status = 'completed'
            GROUP BY p.volunteer_id
        ) sub
        WHERE v.id = sub.volunteer_id;

        UPDATE mission_progress SET status = 'in_progress', progress = 0, completed_at = NULL
        WHERE status = 'completed';

        INSERT INTO schema_patches (name) VALUES ('reopen_self_reported_missions');
    END IF;
END $$;

DO $$
DECLARE r RECORD;
    eid UUID;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM schema_patches WHERE name = 'backfill_assignment_delivery_events') THEN
        FOR r IN
            SELECT id, COALESCE(delivery_note, '') AS note, COALESCE(delivery_file_name, '') AS file_name,
                COALESCE(delivery_object_key, '') AS object_key, COALESCE(delivery_mime, '') AS mime,
                COALESCE(delivered_at, created_at) AS at
            FROM assignments
            WHERE COALESCE(delivery_note, '') <> '' OR COALESCE(delivery_file_name, '') <> ''
        LOOP
            eid := gen_random_uuid();
            INSERT INTO assignment_events (id, assignment_id, kind, note, actor_role, created_at)
            VALUES (eid, r.id, 'delivery', r.note, 'volunteer', r.at);
            IF r.object_key <> '' THEN
                INSERT INTO assignment_event_files (id, event_id, file_name, object_key, mime_type)
                VALUES (gen_random_uuid(), eid, r.file_name, r.object_key, r.mime);
            END IF;
        END LOOP;
        INSERT INTO schema_patches (name) VALUES ('backfill_assignment_delivery_events');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS training_courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'in_person',
    location TEXT NOT NULL DEFAULT '',
    training_at TIMESTAMPTZ,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_training_courses_title ON training_courses (lower(btrim(title)));

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS training_course_id UUID REFERENCES training_courses(id);
ALTER TABLE volunteer_trainings ADD COLUMN IF NOT EXISTS course_id UUID REFERENCES training_courses(id);
ALTER TABLE volunteer_trainings ADD COLUMN IF NOT EXISTS course_title TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM schema_patches WHERE name = 'backfill_training_courses') THEN
        INSERT INTO training_courses (id, title, kind, location, training_at, status, created_at, updated_at)
        SELECT gen_random_uuid(),
            COALESCE(NULLIF(btrim(MAX(COALESCE(training_location, ''))), ''), 'دوره آموزشی')
                || ' — ' || COALESCE(NULLIF(MAX(training_kind), ''), 'in_person'),
            COALESCE(NULLIF(MAX(training_kind), ''), 'in_person'),
            COALESCE(MAX(training_location), ''),
            MIN(training_at),
            'active',
            now(),
            now()
        FROM tasks
        WHERE requires_training
        GROUP BY lower(btrim(COALESCE(training_kind, ''))), lower(btrim(COALESCE(training_location, '')));

        UPDATE tasks t SET training_course_id = c.id
        FROM training_courses c
        WHERE t.requires_training AND t.training_course_id IS NULL
          AND lower(btrim(COALESCE(t.training_kind, ''))) = lower(btrim(c.kind))
          AND lower(btrim(COALESCE(t.training_location, ''))) = lower(btrim(c.location));

        UPDATE volunteer_trainings vt SET course_id = c.id, course_title = c.title
        FROM training_courses c
        WHERE vt.course_id IS NULL
          AND lower(btrim(COALESCE(vt.training_kind, ''))) = lower(btrim(c.kind))
          AND lower(btrim(COALESCE(vt.training_location, ''))) = lower(btrim(c.location));

        UPDATE volunteer_trainings vt SET course_id = t.training_course_id,
            course_title = COALESCE(NULLIF(vt.course_title, ''), c.title, '')
        FROM tasks t
        LEFT JOIN training_courses c ON c.id = t.training_course_id
        WHERE vt.course_id IS NULL AND vt.source_task_id = t.id AND t.training_course_id IS NOT NULL;

        INSERT INTO schema_patches (name) VALUES ('backfill_training_courses');
    END IF;
END $$;


