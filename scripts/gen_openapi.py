#!/usr/bin/env python3
"""Generate docs/openapi.yaml for the Mahak volunteer API handover package."""
from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "openapi.yaml"

SPEC = r'''openapi: 3.0.3
info:
  title: Mahak Volunteer Management Platform API
  description: |
    وب‌سرویس REST سامانه مدیریت داوطلبان محک (`mahak-volunteer-api`).

    این مشخصات برای پیاده‌سازی کلاینت توسط تیم دیگر است. منبع مسیرها با
    `GET /api/v1/` و `backend/internal/adapter/http/catalog.go` هم‌تراز است.

    **نقش‌ها**
    - `volunteer` داوطلب
    - `operator` بهره‌بردار / واحد پشتیبانی — همان دسترسی `/admin/*`
    - `admin` ادمین سامانه — همان دسترسی `/admin/*`

    **احراز**
    - JWT: `Authorization: Bearer <token>`
    - یکپارچگی داخلی: `X-Internal-Token: <INTERNAL_API_TOKEN>`
    - وب‌هوک ماموریت: توکن ماموریت در بدنه (`token`) یا هدر `X-Mission-Token`؛ با توکن داخلی قاطی نشود

    خطاها همیشه `{"error":"<پیام فارسی یا انگلیسی>"}` هستند.
  version: "1.1.0"
  contact:
    name: Mahak Volunteer Platform
servers:
  - url: http://localhost:8080
    description: Local
  - url: https://{host}
    description: Deployed
    variables:
      host:
        default: volunteers.example.ir
tags:
  - name: Health
  - name: Auth
  - name: Session
  - name: Tickets
  - name: Volunteer
  - name: Tasks
  - name: Missions
  - name: Certificates
  - name: StaffVolunteers
  - name: StaffTasks
  - name: StaffAssignments
  - name: StaffMissions
  - name: StaffTickets
  - name: StaffCertificates
  - name: StaffReports
  - name: StaffSkills
  - name: Integrations
security: []
paths:
'''

# Compact path helpers
B = [{"BearerAuth": []}]
I = [{"InternalToken": []}]


def op(summary, tags, security=None, params=None, body=None, responses=None, extra=None):
    lines = [f"      summary: {summary}", f"      tags: [{tags}]"]
    if security is not None:
        if security:
            sch, = security
            key = list(sch.keys())[0]
            lines.append(f"      security: [{{ {key}: [] }}]")
        else:
            lines.append("      security: []")
    if params:
        lines.append("      parameters:")
        for p in params:
            lines.append(f"        - {p}")
    if body:
        lines.append("      requestBody:")
        lines.append("        required: true")
        lines.append("        content:")
        if body.get("multipart"):
            lines.append("          multipart/form-data:")
            lines.append("            schema:")
            lines.append("              $ref: '#/components/schemas/%s'" % body["schema"])
        else:
            lines.append("          application/json:")
            lines.append("            schema:")
            lines.append("              $ref: '#/components/schemas/%s'" % body["schema"])
    lines.append("      responses:")
    for code, desc in (responses or {"200": "OK"}).items():
        if desc.startswith("$"):
            lines.append(f'        "{code}":')
            lines.append(f"          description: OK")
            lines.append("          content:")
            lines.append("            application/json:")
            lines.append("              schema:")
            lines.append(f"                $ref: '{desc}'")
        elif desc == "pdf":
            lines.append(f'        "{code}":')
            lines.append("          description: PDF")
            lines.append("          content:")
            lines.append("            application/pdf:")
            lines.append("              schema: { type: string, format: binary }")
        elif desc == "file":
            lines.append(f'        "{code}":')
            lines.append("          description: Binary file")
            lines.append("          content:")
            lines.append("            application/octet-stream:")
            lines.append("              schema: { type: string, format: binary }")
        elif desc == "csv":
            lines.append(f'        "{code}":')
            lines.append("          description: CSV or JSON")
        else:
            lines.append(f'        "{code}": {{ description: {desc} }}')
    lines.append('        "400": { $ref: "#/components/responses/Error" }')
    if extra:
        lines.extend(extra)
    return "\n".join(lines)


PID = '$ref: "#/components/parameters/Id"'
PCODE = '{ name: code, in: path, required: true, schema: { type: string, format: uuid } }'

PATHS = [
    ("/healthz", {"get": ("Liveness", "Health", None, None, None, {"200": "$#/components/schemas/Health"})}),
    ("/readyz", {"get": ("Readiness", "Health", None, None, None, {"200": "$#/components/schemas/Health", "503": "Database unavailable"})}),
    ("/api/v1/", {"get": ("Full route catalog", "Health", None, None, None, {"200": "Catalog JSON"})}),
    ("/api/v1/auth/otp/send", {"post": ("Send OTP", "Auth", [], None, {"schema": "OTPSendRequest"}, {"200": "$#/components/schemas/OTPSendResult"})}),
    ("/api/v1/auth/otp/verify", {"post": ("Verify OTP", "Auth", [], None, {"schema": "OTPVerifyRequest"}, {"200": "$#/components/schemas/AuthResult", "201": "New user"})}),
    ("/api/v1/auth/login", {"post": ("Login", "Auth", [], None, {"schema": "LoginRequest"}, {"200": "$#/components/schemas/AuthResult", "401": "Unauthorized"})}),
    ("/api/v1/auth/register", {"post": ("Register volunteer (role ignored)", "Auth", [], None, {"schema": "RegisterRequest"}, {"201": "$#/components/schemas/AuthResult", "409": "Email exists"})}),
    ("/api/v1/auth/external", {"post": ("Map Mahak Auth user", "Auth", I, None, {"schema": "ExternalAuthRequest"}, {"200": "$#/components/schemas/AuthResult"})}),
    ("/api/v1/me", {"get": ("Current session", "Session", B, None, None, {"200": "$#/components/schemas/MeResponse"})}),
    ("/api/v1/notifications", {"get": ("Notifications", "Session", B, None, None, {"200": "$#/components/schemas/NotificationList"})}),
    ("/api/v1/notifications/{id}/read", {"post": ("Mark one read", "Session", B, [PID], None, {"200": "$#/components/schemas/StatusOK"})}),
    ("/api/v1/notifications/read-all", {"post": ("Mark all read", "Session", B, None, None, {"200": "$#/components/schemas/StatusOK"})}),
    ("/api/v1/tickets/me", {"get": ("My tickets", "Tickets", B, None, None, {"200": "$#/components/schemas/TicketList"})}),
    ("/api/v1/tickets", {"post": ("Create ticket", "Tickets", B, None, {"schema": "TicketCreate"}, {"201": "$#/components/schemas/Ticket"})}),
    ("/api/v1/tickets/{id}", {"get": ("Ticket thread", "Tickets", B, [PID], None, {"200": "$#/components/schemas/Ticket"})}),
    ("/api/v1/tickets/{id}/messages", {"post": ("Reply to ticket", "Tickets", B, [PID], {"schema": "TicketReply"}, {"200": "$#/components/schemas/Ticket"})}),
    ("/api/v1/skills", {"get": ("Public skill catalog", "Volunteer", B, None, None, {"200": "$#/components/schemas/SkillGroupList"})}),
    ("/api/v1/volunteers/me", {
        "get": ("My profile", "Volunteer", B, None, None, {"200": "$#/components/schemas/Volunteer"}),
        "put": ("Update profile", "Volunteer", B, None, {"schema": "ProfileInput"}, {"200": "$#/components/schemas/Volunteer"}),
    }),
    ("/api/v1/volunteers/me/submit", {"post": ("Submit for review", "Volunteer", B, None, None, {"200": "$#/components/schemas/Volunteer"})}),
    ("/api/v1/volunteers/me/availability", {
        "get": ("Availability slots", "Volunteer", B, None, None, {"200": "$#/components/schemas/AvailabilityList"}),
        "put": ("Replace availability", "Volunteer", B, None, {"schema": "AvailabilityInput"}, {"200": "$#/components/schemas/StatusOK"}),
    }),
    ("/api/v1/volunteers/me/documents", {
        "get": ("My documents", "Volunteer", B, None, None, {"200": "$#/components/schemas/DocumentList"}),
        "post": ("Upload document", "Volunteer", B, None, {"schema": "DocumentUpload", "multipart": True}, {"201": "$#/components/schemas/Document"}),
    }),
    ("/api/v1/volunteers/me/documents/{id}", {"delete": ("Delete document", "Volunteer", B, [PID], None, {"200": "$#/components/schemas/StatusOK"})}),
    ("/api/v1/volunteers/me/skill-proposals", {
        "get": ("My skill proposals", "Volunteer", B, None, None, {"200": "$#/components/schemas/SkillProposalList"}),
        "post": ("Propose skill", "Volunteer", B, None, {"schema": "SkillPropose"}, {"201": "$#/components/schemas/SkillProposal"}),
    }),
    ("/api/v1/tasks", {"get": ("Eligible tasks", "Tasks", B, ['{ name: q, in: query, schema: { type: string } }', '{ name: skill, in: query, schema: { type: string } }', '{ name: limit, in: query, schema: { type: integer, default: 50 } }', '{ name: offset, in: query, schema: { type: integer, default: 0 } }'], None, {"200": "$#/components/schemas/TaskPage"})}),
    ("/api/v1/tasks/{id}", {"get": ("Task detail", "Tasks", B, [PID], None, {"200": "$#/components/schemas/Task"})}),
    ("/api/v1/tasks/{id}/accept", {"post": ("Apply / reserve", "Tasks", B, [PID], None, {"201": "$#/components/schemas/Assignment", "409": "Already assigned", "422": "Not eligible or full"})}),
    ("/api/v1/assignments/me", {"get": ("My assignments", "Tasks", B, None, None, {"200": "$#/components/schemas/AssignmentList"})}),
    ("/api/v1/assignments/{id}/start", {"post": ("Start remote work", "Tasks", B, [PID], None, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/assignments/{id}/deliver", {"post": ("Deliver remote work", "Tasks", B, [PID], {"schema": "DeliveryUpload", "multipart": True}, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/assignments/{id}/rate", {"post": ("Rate organization", "Tasks", B, [PID], {"schema": "VolunteerRating"}, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/assignments/{id}/cancel", {"post": ("Cancel my request", "Tasks", B, [PID], None, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/missions", {"get": ("Active missions", "Missions", B, None, None, {"200": "$#/components/schemas/MissionList"})}),
    ("/api/v1/missions/{id}/start", {"post": ("Start mission", "Missions", B, [PID], None, {"200": "$#/components/schemas/MissionProgress"})}),
    ("/api/v1/missions/{id}/progress", {"post": ("Verify mission completion", "Missions", B, [PID], None, {"200": "$#/components/schemas/MissionProgress", "422": "Not verified"})}),
    ("/api/v1/missions/me", {"get": ("My mission progress", "Missions", B, None, None, {"200": "$#/components/schemas/MissionProgressList"})}),
    ("/api/v1/certificates/me", {"get": ("My certificates", "Certificates", B, None, None, {"200": "$#/components/schemas/CertificateList"})}),
    ("/api/v1/certificates/requests", {
        "get": ("My certificate requests", "Certificates", B, None, None, {"200": "$#/components/schemas/CertificateRequestList"}),
        "post": ("Create certificate request", "Certificates", B, None, {"schema": "CertificateRequestInput"}, {"201": "$#/components/schemas/CertificateRequest"}),
    }),
    ("/api/v1/certificates/verify/{code}", {"get": ("Public verify", "Certificates", None, [PCODE], None, {"200": "$#/components/schemas/Certificate"})}),
    ("/api/v1/certificates/{code}/pdf", {"get": ("Public PDF", "Certificates", None, [PCODE], None, {"200": "pdf"})}),
    ("/api/v1/admin/dashboard", {"get": ("Staff dashboard", "StaffVolunteers", B, None, None, {"200": "$#/components/schemas/Dashboard"})}),
    ("/api/v1/admin/volunteers", {"get": ("List volunteers", "StaffVolunteers", B, ['{ name: status, in: query, schema: { type: string } }', '{ name: skill, in: query, schema: { type: string } }', '{ name: q, in: query, schema: { type: string } }', '{ name: attention, in: query, schema: { type: string } }', '{ name: limit, in: query, schema: { type: integer } }', '{ name: offset, in: query, schema: { type: integer } }'], None, {"200": "$#/components/schemas/VolunteerPage"})}),
    ("/api/v1/admin/volunteers/{id}", {
        "get": ("Volunteer dossier", "StaffVolunteers", B, [PID], None, {"200": "$#/components/schemas/VolunteerDossier"}),
        "put": ("Edit profile", "StaffVolunteers", B, [PID], {"schema": "ProfileInput"}, {"200": "$#/components/schemas/Volunteer"}),
    }),
    ("/api/v1/admin/volunteers/{id}/review", {"post": ("Review membership", "StaffVolunteers", B, [PID], {"schema": "ReviewInput"}, {"200": "$#/components/schemas/Volunteer"})}),
    ("/api/v1/admin/volunteers/{id}/status", {"post": ("Set status", "StaffVolunteers", B, [PID], {"schema": "StatusInput"}, {"200": "$#/components/schemas/Volunteer"})}),
    ("/api/v1/admin/volunteers/{id}/comments", {"post": ("File comment", "StaffVolunteers", B, [PID], {"schema": "CommentInput"}, {"200": "$#/components/schemas/Volunteer"})}),
    ("/api/v1/admin/volunteers/{id}/documents", {"get": ("Documents", "StaffVolunteers", B, [PID], None, {"200": "$#/components/schemas/DocumentList"})}),
    ("/api/v1/admin/volunteers/{id}/availability", {"get": ("Availability", "StaffVolunteers", B, [PID], None, {"200": "$#/components/schemas/AvailabilityList"})}),
    ("/api/v1/admin/documents/{id}", {"get": ("Stream document", "StaffVolunteers", B, [PID], None, {"200": "file"})}),
    ("/api/v1/admin/tasks", {
        "get": ("List activities", "StaffTasks", B, ['{ name: q, in: query, schema: { type: string } }', '{ name: status, in: query, schema: { type: string } }', '{ name: series_id, in: query, schema: { type: string, format: uuid } }', '{ name: limit, in: query, schema: { type: integer } }', '{ name: offset, in: query, schema: { type: integer } }'], None, {"200": "$#/components/schemas/TaskPage"}),
        "post": ("Create activity", "StaffTasks", B, None, {"schema": "TaskInput"}, {"201": "$#/components/schemas/Task"}),
    }),
    ("/api/v1/admin/tasks/{id}", {
        "put": ("Update activity", "StaffTasks", B, [PID], {"schema": "TaskInput"}, {"200": "$#/components/schemas/Task"}),
        "delete": ("Delete activity", "StaffTasks", B, [PID], None, {"200": "$#/components/schemas/StatusOK"}),
    }),
    ("/api/v1/admin/tasks/{id}/status", {"post": ("Set activity status", "StaffTasks", B, [PID], {"schema": "TaskStatusInput"}, {"200": "$#/components/schemas/Task"})}),
    ("/api/v1/admin/tasks/{id}/assign", {"post": ("Manual assign", "StaffTasks", B, [PID], {"schema": "AssignInput"}, {"201": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/tasks/{id}/assignments", {"get": ("Assignments of activity", "StaffTasks", B, [PID], None, {"200": "$#/components/schemas/AssignmentPage"})}),
    ("/api/v1/admin/assignments", {"get": ("List assignments", "StaffAssignments", B, ['{ name: status, in: query, schema: { type: string } }', '{ name: volunteer_id, in: query, schema: { type: string, format: uuid } }', '{ name: task_id, in: query, schema: { type: string, format: uuid } }', '{ name: series_id, in: query, schema: { type: string, format: uuid } }', '{ name: limit, in: query, schema: { type: integer } }', '{ name: offset, in: query, schema: { type: integer } }'], None, {"200": "$#/components/schemas/AssignmentPage"})}),
    ("/api/v1/admin/assignments/{id}/approve", {"post": ("Approve application", "StaffAssignments", B, [PID], None, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/reject", {"post": ("Reject application", "StaffAssignments", B, [PID], {"schema": "CommentInput"}, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/revision", {"post": ("Request remote revision", "StaffAssignments", B, [PID], {"schema": "CommentInput"}, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/message", {"post": ("Message volunteer", "StaffAssignments", B, [PID], {"schema": "MessageInput"}, {"200": "$#/components/schemas/StatusOK"})}),
    ("/api/v1/admin/assignments/{id}/attendance", {"post": ("Confirm attendance", "StaffAssignments", B, [PID], {"schema": "AttendanceInput"}, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/absent", {"post": ("Mark absent", "StaffAssignments", B, [PID], None, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/complete", {"post": ("Complete with scores 1-5", "StaffAssignments", B, [PID], {"schema": "CompleteInput"}, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/cancel", {"post": ("Staff cancel", "StaffAssignments", B, [PID], None, {"200": "$#/components/schemas/Assignment"})}),
    ("/api/v1/admin/assignments/{id}/certificate", {"post": ("Issue task appreciation", "StaffAssignments", B, [PID], None, {"201": "$#/components/schemas/Certificate"})}),
    ("/api/v1/admin/assignments/{id}/delivery", {"get": ("Download delivery", "StaffAssignments", B, [PID], None, {"200": "file"})}),
    ("/api/v1/admin/missions", {
        "get": ("List missions (includes verify_token)", "StaffMissions", B, None, None, {"200": "$#/components/schemas/MissionList"}),
        "post": ("Create mission", "StaffMissions", B, None, {"schema": "MissionInput"}, {"201": "$#/components/schemas/Mission"}),
    }),
    ("/api/v1/admin/missions/{id}", {"put": ("Update mission", "StaffMissions", B, [PID], {"schema": "MissionInput"}, {"200": "$#/components/schemas/Mission"})}),
    ("/api/v1/admin/tickets", {"get": ("List tickets", "StaffTickets", B, ['{ name: status, in: query, schema: { type: string, enum: [open, answered, closed] } }'], None, {"200": "$#/components/schemas/TicketList"})}),
    ("/api/v1/admin/tickets/{id}", {"get": ("Ticket", "StaffTickets", B, [PID], None, {"200": "$#/components/schemas/Ticket"})}),
    ("/api/v1/admin/tickets/{id}/messages", {"post": ("Staff reply", "StaffTickets", B, [PID], {"schema": "TicketReply"}, {"200": "$#/components/schemas/Ticket"})}),
    ("/api/v1/admin/tickets/{id}/status", {"post": ("Set ticket status", "StaffTickets", B, [PID], {"schema": "TicketStatusInput"}, {"200": "$#/components/schemas/Ticket"})}),
    ("/api/v1/admin/volunteers/{id}/certificates/aggregated", {"post": ("Issue 12-month appreciation", "StaffCertificates", B, [PID], None, {"201": "$#/components/schemas/Certificate"})}),
    ("/api/v1/admin/certificate-requests", {"get": ("Certificate request queue", "StaffCertificates", B, ['{ name: status, in: query, schema: { type: string } }'], None, {"200": "$#/components/schemas/CertificateRequestList"})}),
    ("/api/v1/admin/certificate-requests/{id}/review", {"post": ("Review certificate request", "StaffCertificates", B, [PID], {"schema": "CertReviewInput"}, {"200": "$#/components/schemas/CertificateRequest"})}),
    ("/api/v1/admin/reports/ranking", {"get": ("Ranking", "StaffReports", B, ['{ name: format, in: query, schema: { type: string, enum: [json, csv] } }', '{ name: limit, in: query, schema: { type: integer } }'], None, {"200": "csv"})}),
    ("/api/v1/admin/reports/skills", {"get": ("Skill histogram", "StaffReports", B, None, None, {"200": "map of skill to count"})}),
    ("/api/v1/admin/reports/overview", {"get": ("Overview report", "StaffReports", B, None, None, {"200": "$#/components/schemas/ReportOverview"})}),
    ("/api/v1/admin/skills/", {"get": ("Staff skill catalog", "StaffSkills", B, None, None, {"200": "$#/components/schemas/SkillGroupList"})}),
    ("/api/v1/admin/skills/groups", {"post": ("Create group", "StaffSkills", B, None, {"schema": "SkillGroupInput"}, {"201": "$#/components/schemas/SkillGroup"})}),
    ("/api/v1/admin/skills/groups/{id}", {
        "put": ("Update group", "StaffSkills", B, [PID], {"schema": "SkillGroupInput"}, {"200": "$#/components/schemas/SkillGroup"}),
        "delete": ("Delete group", "StaffSkills", B, [PID], None, {"200": "$#/components/schemas/StatusOK"}),
    }),
    ("/api/v1/admin/skills/", {"post": ("Create skill", "StaffSkills", B, None, {"schema": "SkillInput"}, {"201": "$#/components/schemas/Skill"})}),
    ("/api/v1/admin/skills/{id}", {
        "put": ("Update skill", "StaffSkills", B, [PID], {"schema": "SkillUpdate"}, {"200": "$#/components/schemas/Skill"}),
        "delete": ("Delete skill", "StaffSkills", B, [PID], None, {"200": "$#/components/schemas/StatusOK"}),
    }),
    ("/api/v1/admin/skills/proposals", {"get": ("Skill proposals", "StaffSkills", B, ['{ name: status, in: query, schema: { type: string } }'], None, {"200": "$#/components/schemas/SkillProposalList"})}),
    ("/api/v1/admin/skills/proposals/{id}/review", {"post": ("Review proposal", "StaffSkills", B, [PID], {"schema": "ProposalReview"}, {"200": "$#/components/schemas/SkillProposal"})}),
    ("/api/v1/webhooks/events", {"post": ("Inbound mission award. Send X-Internal-Token plus body.token or X-Mission-Token.", "Integrations", I, None, {"schema": "WebhookEvent"}, {"202": "$#/components/schemas/StatusOK"})}),
]


COMPONENTS = r'''
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
    InternalToken:
      type: apiKey
      in: header
      name: X-Internal-Token
  parameters:
    Id:
      name: id
      in: path
      required: true
      schema: { type: string, format: uuid }
  responses:
    Error:
      description: Error
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
  schemas:
    Error:
      type: object
      required: [error]
      properties:
        error: { type: string, example: ظرفیت این فعالیت تکمیل است }
    StatusOK:
      type: object
      properties:
        status: { type: string, example: ok }
    Health:
      type: object
      properties:
        status: { type: string }
        service: { type: string }
    User:
      type: object
      properties:
        id: { type: string, format: uuid }
        email: { type: string }
        phone: { type: string }
        role: { type: string, enum: [volunteer, admin, operator] }
        external_user_id: { type: string }
        created_at: { type: string, format: date-time }
    AuthResult:
      type: object
      properties:
        token: { type: string }
        user: { $ref: "#/components/schemas/User" }
        is_new: { type: boolean }
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email: { type: string }
        password: { type: string }
    RegisterRequest:
      type: object
      required: [email, password, full_name]
      properties:
        email: { type: string }
        password: { type: string, minLength: 8 }
        full_name: { type: string }
        role: { type: string, description: "Ignored; always volunteer" }
    OTPSendRequest:
      type: object
      required: [phone]
      properties:
        phone: { type: string, example: "09121234567" }
    OTPSendResult:
      type: object
      properties:
        phone: { type: string }
        ttl_seconds: { type: integer }
        resend_after: { type: integer }
        is_new: { type: boolean }
        dev_code: { type: string, description: "Only when OTP_REVEAL=true" }
    OTPVerifyRequest:
      type: object
      required: [phone, code]
      properties:
        phone: { type: string }
        code: { type: string }
        full_name: { type: string }
    ExternalAuthRequest:
      type: object
      required: [external_user_id]
      properties:
        external_user_id: { type: string }
        email: { type: string }
        full_name: { type: string }
        role: { type: string, enum: [volunteer, admin, operator] }
    MeResponse:
      type: object
      properties:
        user: { $ref: "#/components/schemas/User" }
        volunteer: { $ref: "#/components/schemas/Volunteer" }
    Volunteer:
      type: object
      properties:
        id: { type: string, format: uuid }
        user_id: { type: string, format: uuid }
        full_name: { type: string }
        first_name: { type: string }
        last_name: { type: string }
        national_id: { type: string }
        phone: { type: string }
        phone2: { type: string }
        province: { type: string }
        city: { type: string }
        address: { type: string }
        plaque: { type: string }
        unit: { type: string }
        bio: { type: string }
        skill_categories: { type: array, items: { type: string } }
        skill_ids: { type: array, items: { type: string, format: uuid } }
        skills: { type: array, items: { $ref: "#/components/schemas/VolunteerSkill" } }
        proposals: { type: array, items: { $ref: "#/components/schemas/SkillProposal" } }
        education_level: { type: string }
        education_field: { type: string }
        medical_license: { type: string }
        birth_date: { type: string, description: "YYYY-MM-DD; volunteer must be 18+" }
        gender: { type: string, enum: [male, female] }
        occupation: { type: string }
        occupation_other: { type: string }
        email: { type: string }
        status: { type: string, enum: [draft, pending, approved, rejected, suspended] }
        rejection_reason: { type: string }
        history: { type: array, items: { $ref: "#/components/schemas/VolunteerEvent" } }
        average_score: { type: number }
        total_hours: { type: number }
        completed_tasks: { type: integer }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
    VolunteerEvent:
      type: object
      properties:
        id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        actor_user_id: { type: string, format: uuid }
        actor_role: { type: string }
        event_type: { type: string }
        from_status: { type: string }
        to_status: { type: string }
        comment: { type: string }
        created_at: { type: string, format: date-time }
    VolunteerSkill:
      type: object
      properties:
        skill_id: { type: string, format: uuid }
        title: { type: string }
        group_id: { type: string, format: uuid }
        group_slug: { type: string }
        group_title: { type: string }
    ProfileInput:
      type: object
      properties:
        full_name: { type: string }
        first_name: { type: string }
        last_name: { type: string }
        national_id: { type: string }
        phone: { type: string }
        phone2: { type: string }
        province: { type: string }
        city: { type: string }
        address: { type: string }
        plaque: { type: string }
        unit: { type: string }
        bio: { type: string }
        skill_ids: { type: array, items: { type: string, format: uuid } }
        skill_categories: { type: array, items: { type: string } }
        education_level: { type: string }
        education_field: { type: string }
        medical_license: { type: string }
        birth_date: { type: string }
        gender: { type: string, enum: [male, female] }
        occupation: { type: string }
        occupation_other: { type: string }
    VolunteerPage:
      type: object
      properties:
        items: { type: array, items: { $ref: "#/components/schemas/Volunteer" } }
        total: { type: integer }
    VolunteerDossier:
      type: object
      properties:
        volunteer: { $ref: "#/components/schemas/Volunteer" }
        documents: { type: array, items: { $ref: "#/components/schemas/Document" } }
        availability: { type: array, items: { $ref: "#/components/schemas/AvailabilitySlot" } }
        assignments: { type: array, items: { $ref: "#/components/schemas/Assignment" } }
        missions: { type: array, items: { $ref: "#/components/schemas/MissionProgress" } }
    AvailabilitySlot:
      type: object
      properties:
        id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        weekday: { type: integer, minimum: 0, maximum: 6, description: "0=Sunday" }
        start_time: { type: string, example: "09:00" }
        end_time: { type: string, example: "13:00" }
    AvailabilityList:
      type: array
      items: { $ref: "#/components/schemas/AvailabilitySlot" }
    AvailabilityInput:
      type: object
      properties:
        slots:
          type: array
          items:
            type: object
            properties:
              weekday: { type: integer }
              start_time: { type: string }
              end_time: { type: string }
    Document:
      type: object
      properties:
        id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        kind: { type: string, enum: [national_id, driving_license, medical_license, education, other] }
        file_name: { type: string }
        mime_type: { type: string }
        size_bytes: { type: integer }
        created_at: { type: string, format: date-time }
    DocumentList:
      type: array
      items: { $ref: "#/components/schemas/Document" }
    DocumentUpload:
      type: object
      properties:
        kind: { type: string }
        file: { type: string, format: binary }
    Task:
      type: object
      properties:
        id: { type: string, format: uuid }
        title: { type: string }
        description: { type: string }
        location: { type: string }
        starts_at: { type: string, format: date-time }
        ends_at: { type: string, format: date-time }
        capacity: { type: integer }
        reserved_count: { type: integer }
        hour_weight: { type: number }
        required_skills: { type: array, items: { type: string } }
        required_skill_ids: { type: array, items: { type: string, format: uuid } }
        min_score: { type: number }
        required_education: { type: string }
        work_mode: { type: string, enum: [onsite, remote] }
        delivery_hint: { type: string }
        requires_training: { type: boolean }
        training_kind: { type: string, enum: [in_person, online, hybrid, workshop, other] }
        training_location: { type: string }
        training_at: { type: string, format: date-time }
        kind: { type: string, enum: [one_off, recurring, occurrence] }
        series_id: { type: string, format: uuid }
        weekday: { type: integer }
        slots: { type: array, items: { $ref: "#/components/schemas/TaskSlot" } }
        status: { type: string, enum: [open, closed, cancelled, inactive] }
        created_at: { type: string, format: date-time }
    TaskSlot:
      type: object
      properties:
        weekday: { type: integer }
        capacity: { type: integer }
        start_time: { type: string }
        end_time: { type: string }
    TaskInput:
      type: object
      required: [title, starts_at, ends_at]
      properties:
        title: { type: string }
        description: { type: string }
        location: { type: string }
        starts_at: { type: string, format: date-time }
        ends_at: { type: string, format: date-time }
        capacity: { type: integer }
        hour_weight: { type: number }
        required_skills: { type: array, items: { type: string } }
        required_skill_ids: { type: array, items: { type: string } }
        min_score: { type: number }
        required_education: { type: string }
        work_mode: { type: string, enum: [onsite, remote] }
        delivery_hint: { type: string }
        requires_training: { type: boolean }
        training_kind: { type: string }
        training_location: { type: string }
        training_at: { type: string, format: date-time }
        status: { type: string }
        kind: { type: string, enum: [one_off, recurring] }
        slots: { type: array, items: { $ref: "#/components/schemas/TaskSlot" } }
    TaskPage:
      type: object
      properties:
        items: { type: array, items: { $ref: "#/components/schemas/Task" } }
        total: { type: integer }
    TaskStatusInput:
      type: object
      required: [status]
      properties:
        status: { type: string, enum: [open, closed, cancelled, inactive] }
    AssignInput:
      type: object
      required: [volunteer_id]
      properties:
        volunteer_id: { type: string, format: uuid }
    Assignment:
      type: object
      properties:
        id: { type: string, format: uuid }
        task_id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        status: { type: string, enum: [requested, reserved, in_progress, attended, submitted, completed, cancelled, rejected, absent, revision_requested] }
        volunteer_rating: { type: integer }
        volunteer_comment: { type: string }
        admin_discipline: { type: integer }
        admin_expertise: { type: integer }
        admin_ethics: { type: integer }
        admin_comment: { type: string }
        composite_score: { type: number }
        hours_awarded: { type: number }
        attended_at: { type: string, format: date-time }
        check_in_at: { type: string, format: date-time }
        check_out_at: { type: string, format: date-time }
        completed_at: { type: string, format: date-time }
        delivery_note: { type: string }
        delivery_file_name: { type: string }
        delivered_at: { type: string, format: date-time }
        created_at: { type: string, format: date-time }
        task: { $ref: "#/components/schemas/Task" }
        volunteer: { $ref: "#/components/schemas/Volunteer" }
    AssignmentList:
      type: array
      items: { $ref: "#/components/schemas/Assignment" }
    AssignmentPage:
      type: object
      properties:
        items: { type: array, items: { $ref: "#/components/schemas/Assignment" } }
        total: { type: integer }
    VolunteerRating:
      type: object
      properties:
        rating: { type: integer, minimum: 1, maximum: 5 }
        comment: { type: string }
    DeliveryUpload:
      type: object
      properties:
        note: { type: string }
        file: { type: string, format: binary }
    AttendanceInput:
      type: object
      properties:
        check_in_at: { type: string, format: date-time, description: "RFC3339; empty = now" }
        check_out_at: { type: string, format: date-time }
    CompleteInput:
      type: object
      required: [discipline, expertise, ethics]
      properties:
        discipline: { type: integer, minimum: 1, maximum: 5 }
        expertise: { type: integer, minimum: 1, maximum: 5 }
        ethics: { type: integer, minimum: 1, maximum: 5 }
        comment: { type: string }
    Mission:
      type: object
      properties:
        id: { type: string, format: uuid }
        title: { type: string }
        description: { type: string }
        kind: { type: string, enum: [complete_profile, invite_users, custom, webhook] }
        hour_weight: { type: number }
        deadline_hours: { type: integer }
        webhook_event: { type: string }
        target_count: { type: integer }
        verify_mode: { type: string, enum: [internal, outbound, inbound] }
        verify_url: { type: string }
        verify_token: { type: string, description: "Staff only; stripped from volunteer list" }
        can_check: { type: boolean }
        status: { type: string, enum: [active, archived] }
    MissionList:
      type: array
      items: { $ref: "#/components/schemas/Mission" }
    MissionInput:
      type: object
      properties:
        title: { type: string }
        description: { type: string }
        kind: { type: string }
        hour_weight: { type: number }
        deadline_hours: { type: integer }
        webhook_event: { type: string }
        target_count: { type: integer }
        verify_mode: { type: string, enum: [internal, outbound, inbound] }
        verify_url: { type: string }
        verify_token: { type: string }
        status: { type: string }
    MissionProgress:
      type: object
      properties:
        id: { type: string, format: uuid }
        mission_id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        status: { type: string, enum: [in_progress, completed, expired] }
        progress: { type: integer }
        started_at: { type: string, format: date-time }
        due_at: { type: string, format: date-time }
        completed_at: { type: string, format: date-time }
        mission: { $ref: "#/components/schemas/Mission" }
    MissionProgressList:
      type: array
      items: { $ref: "#/components/schemas/MissionProgress" }
    Certificate:
      type: object
      properties:
        id: { type: string, format: uuid }
        verification_code: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        volunteer_name: { type: string }
        kind: { type: string, enum: [task, aggregated, official] }
        title: { type: string }
        hours: { type: number }
        period_start: { type: string, format: date-time }
        period_end: { type: string, format: date-time }
        issued_at: { type: string, format: date-time }
        authentic: { type: boolean }
    CertificateList:
      type: array
      items: { $ref: "#/components/schemas/Certificate" }
    CertificateRequest:
      type: object
      properties:
        id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        volunteer_name: { type: string }
        kind: { type: string, enum: [task, aggregated, official] }
        assignment_id: { type: string, format: uuid }
        assignment_title: { type: string }
        status: { type: string, enum: [pending, preparing, ready, delivered, approved, rejected] }
        admin_note: { type: string }
        certificate_id: { type: string, format: uuid }
        delivery_method: { type: string, enum: [send, in_person] }
        delivered_at: { type: string, format: date-time }
        created_at: { type: string, format: date-time }
        reviewed_at: { type: string, format: date-time }
    CertificateRequestList:
      type: array
      items: { $ref: "#/components/schemas/CertificateRequest" }
    CertificateRequestInput:
      type: object
      required: [kind]
      properties:
        kind: { type: string, enum: [task, aggregated, official] }
        assignment_id: { type: string, format: uuid }
    CertReviewInput:
      type: object
      required: [action]
      properties:
        action: { type: string, enum: [approve, reject, deliver] }
        admin_note: { type: string }
        delivery_method: { type: string, enum: [send, in_person] }
    Notification:
      type: object
      properties:
        id: { type: string, format: uuid }
        user_id: { type: string, format: uuid }
        title: { type: string }
        body: { type: string }
        read: { type: boolean }
        kind: { type: string, enum: [notice, reminder] }
        remind_at: { type: string, format: date-time }
        fired_at: { type: string, format: date-time }
        created_at: { type: string, format: date-time }
    NotificationList:
      type: array
      items: { $ref: "#/components/schemas/Notification" }
    Ticket:
      type: object
      properties:
        id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        volunteer_name: { type: string }
        subject: { type: string }
        status: { type: string, enum: [open, answered, closed] }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
        messages: { type: array, items: { $ref: "#/components/schemas/TicketMessage" } }
    TicketList:
      type: array
      items: { $ref: "#/components/schemas/Ticket" }
    TicketMessage:
      type: object
      properties:
        id: { type: string, format: uuid }
        ticket_id: { type: string, format: uuid }
        author_user_id: { type: string, format: uuid }
        author_role: { type: string }
        body: { type: string }
        created_at: { type: string, format: date-time }
    TicketCreate:
      type: object
      required: [subject, body]
      properties:
        subject: { type: string }
        body: { type: string }
    TicketReply:
      type: object
      required: [body]
      properties:
        body: { type: string }
    TicketStatusInput:
      type: object
      required: [status]
      properties:
        status: { type: string, enum: [open, answered, closed] }
    SkillGroup:
      type: object
      properties:
        id: { type: string, format: uuid }
        slug: { type: string }
        title: { type: string }
        sort_order: { type: integer }
        skills: { type: array, items: { $ref: "#/components/schemas/Skill" } }
    SkillGroupList:
      type: array
      items: { $ref: "#/components/schemas/SkillGroup" }
    Skill:
      type: object
      properties:
        id: { type: string, format: uuid }
        group_id: { type: string, format: uuid }
        title: { type: string }
        status: { type: string }
        group_title: { type: string }
    SkillGroupInput:
      type: object
      properties:
        slug: { type: string }
        title: { type: string }
        sort_order: { type: integer }
    SkillInput:
      type: object
      properties:
        group_id: { type: string, format: uuid }
        title: { type: string }
    SkillUpdate:
      type: object
      properties:
        title: { type: string }
        status: { type: string }
        group_id: { type: string, format: uuid }
    SkillProposal:
      type: object
      properties:
        id: { type: string, format: uuid }
        volunteer_id: { type: string, format: uuid }
        volunteer_name: { type: string }
        group_id: { type: string, format: uuid }
        group_title: { type: string }
        title: { type: string }
        status: { type: string, enum: [pending, approved, rejected] }
        admin_note: { type: string }
        created_skill_id: { type: string, format: uuid }
        created_at: { type: string, format: date-time }
    SkillProposalList:
      type: array
      items: { $ref: "#/components/schemas/SkillProposal" }
    SkillPropose:
      type: object
      required: [group_id, title]
      properties:
        group_id: { type: string, format: uuid }
        title: { type: string }
    ProposalReview:
      type: object
      required: [action]
      properties:
        action: { type: string, enum: [approve, reject, edit, edit_approve] }
        title: { type: string }
        group_id: { type: string, format: uuid }
        admin_note: { type: string }
    ReviewInput:
      type: object
      required: [action]
      properties:
        action: { type: string, enum: [approve, reject, request_documents, suspend, unsuspend] }
        reason: { type: string }
    StatusInput:
      type: object
      required: [status]
      properties:
        status: { type: string, enum: [draft, pending, approved, rejected, suspended] }
        reason: { type: string }
    CommentInput:
      type: object
      properties:
        comment: { type: string }
        body: { type: string }
    MessageInput:
      type: object
      required: [body]
      properties:
        body: { type: string }
    Dashboard:
      type: object
      properties:
        total_volunteers: { type: integer }
        pending_volunteers: { type: integer }
        approved_volunteers: { type: integer }
        online_estimate: { type: integer }
        open_tasks: { type: integer }
        active_assignments: { type: integer }
        completed_this_month: { type: integer }
        participation_rate: { type: number }
        total_hours: { type: number }
        pending_task_requests: { type: integer }
        pending_deliveries: { type: integer }
        pending_skill_proposals: { type: integer }
        pending_certificates: { type: integer }
        open_tickets: { type: integer }
        resubmitted_documents: { type: integer }
        skill_distribution:
          type: object
          additionalProperties: { type: integer }
    ReportOverview:
      allOf:
        - $ref: "#/components/schemas/Dashboard"
        - type: object
          properties:
            volunteers_by_status:
              type: object
              additionalProperties: { type: integer }
            assignments_by_status:
              type: object
              additionalProperties: { type: integer }
            tasks_by_status:
              type: object
              additionalProperties: { type: integer }
            tasks_by_kind:
              type: object
              additionalProperties: { type: integer }
            hours_this_month: { type: number }
            certificates_issued: { type: integer }
            top_cities:
              type: array
              items:
                type: object
                properties:
                  city: { type: string }
                  count: { type: integer }
    WebhookEvent:
      type: object
      properties:
        event: { type: string, example: user.invited }
        volunteer_id: { type: string, format: uuid }
        phone: { type: string }
        increment: { type: integer, default: 1 }
        token: { type: string, description: "Mission verify_token" }
'''


def render_op(method, spec):
    summary, tags, security, params, body, responses = spec
    return f"    {method}:\n" + op(summary, tags, security, params, body, responses)


def main():
    chunks = [SPEC]
    # Merge duplicate path keys (GET + POST on /admin/skills/)
    merged = {}
    for path, methods in PATHS:
        merged.setdefault(path, {}).update(methods)
    for path, methods in merged.items():
        chunks.append(f"  {path}:\n")
        for method, spec in methods.items():
            chunks.append(render_op(method, spec) + "\n")
    chunks.append(COMPONENTS)
    OUT.write_text("".join(chunks), encoding="utf-8")
    print(f"wrote {OUT} ({OUT.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
