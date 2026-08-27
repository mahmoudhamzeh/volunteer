package httpserver

// CatalogRoute is the contract handed to other teams: method, path, auth, and grouping.
// Auth: public | jwt | staff | internal
type CatalogRoute struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Auth      string `json:"auth"`
	Group     string `json:"group"`
	Summary   string `json:"summary"`
	SummaryFA string `json:"summary_fa"`
	Alias     bool   `json:"alias,omitempty"`
}

func CatalogRoutes() []CatalogRoute {
	return []CatalogRoute{
		{Method: "GET", Path: "/healthz", Auth: "public", Group: "health", Summary: "Liveness", SummaryFA: "سلامت سرویس"},
		{Method: "GET", Path: "/readyz", Auth: "public", Group: "health", Summary: "Readiness (Postgres ping)", SummaryFA: "آمادگی دیتابیس"},
		{Method: "GET", Path: "/api/v1/", Auth: "public", Group: "health", Summary: "API catalog", SummaryFA: "فهرست مسیرهای وب‌سرویس"},

		{Method: "POST", Path: "/api/v1/auth/otp/send", Auth: "public", Group: "auth", Summary: "Send SMS OTP", SummaryFA: "ارسال کد یک‌بارمصرف موبایل"},
		{Method: "POST", Path: "/api/v1/auth/otp/verify", Auth: "public", Group: "auth", Summary: "Verify OTP and issue JWT", SummaryFA: "تایید کد و صدور توکن"},
		{Method: "POST", Path: "/api/v1/auth/login", Auth: "public", Group: "auth", Summary: "Email/password login", SummaryFA: "ورود با ایمیل و رمز"},
		{Method: "POST", Path: "/api/v1/auth/register", Auth: "public", Group: "auth", Summary: "Register volunteer (role always volunteer)", SummaryFA: "ثبت‌نام داوطلب با ایمیل"},
		{Method: "POST", Path: "/api/v1/auth/external", Auth: "internal", Group: "auth", Summary: "Map Mahak Auth subject to local user", SummaryFA: "نگاشت کاربر سرویس احراز محک"},

		{Method: "GET", Path: "/api/v1/me", Auth: "jwt", Group: "session", Summary: "Current user and volunteer profile", SummaryFA: "کاربر جاری"},
		{Method: "GET", Path: "/api/v1/notifications", Auth: "jwt", Group: "session", Summary: "List notifications", SummaryFA: "فهرست اعلان‌ها"},
		{Method: "POST", Path: "/api/v1/notifications/{id}/read", Auth: "jwt", Group: "session", Summary: "Mark one notification read", SummaryFA: "خوانده شدن یک اعلان"},
		{Method: "POST", Path: "/api/v1/notifications/read-all", Auth: "jwt", Group: "session", Summary: "Mark all notifications read", SummaryFA: "خوانده شدن همه اعلان‌ها"},

		{Method: "GET", Path: "/api/v1/tickets/me", Auth: "jwt", Group: "tickets", Summary: "My support tickets", SummaryFA: "تیکت‌های من"},
		{Method: "POST", Path: "/api/v1/tickets", Auth: "jwt", Group: "tickets", Summary: "Open a ticket", SummaryFA: "ثبت تیکت پشتیبانی"},
		{Method: "GET", Path: "/api/v1/tickets/{id}", Auth: "jwt", Group: "tickets", Summary: "Ticket thread (owner)", SummaryFA: "جزئیات تیکت"},
		{Method: "POST", Path: "/api/v1/tickets/{id}/messages", Auth: "jwt", Group: "tickets", Summary: "Reply to my ticket", SummaryFA: "پاسخ داوطلب به تیکت"},

		{Method: "GET", Path: "/api/v1/skills", Auth: "jwt", Group: "volunteer_profile", Summary: "Skill catalog", SummaryFA: "کاتالوگ مهارت"},
		{Method: "GET", Path: "/api/v1/volunteers/me", Auth: "jwt", Group: "volunteer_profile", Summary: "My profile", SummaryFA: "پروفایل داوطلب"},
		{Method: "PUT", Path: "/api/v1/volunteers/me", Auth: "jwt", Group: "volunteer_profile", Summary: "Update profile", SummaryFA: "ویرایش پروفایل"},
		{Method: "POST", Path: "/api/v1/volunteers/me/submit", Auth: "jwt", Group: "volunteer_profile", Summary: "Submit for review", SummaryFA: "ارسال پرونده برای بررسی"},
		{Method: "GET", Path: "/api/v1/volunteers/me/availability", Auth: "jwt", Group: "volunteer_profile", Summary: "My availability", SummaryFA: "زمان آزاد"},
		{Method: "PUT", Path: "/api/v1/volunteers/me/availability", Auth: "jwt", Group: "volunteer_profile", Summary: "Replace availability slots", SummaryFA: "ثبت زمان آزاد"},
		{Method: "GET", Path: "/api/v1/volunteers/me/documents", Auth: "jwt", Group: "volunteer_profile", Summary: "My documents", SummaryFA: "مدارک من"},
		{Method: "POST", Path: "/api/v1/volunteers/me/documents", Auth: "jwt", Group: "volunteer_profile", Summary: "Upload document", SummaryFA: "آپلود مدرک"},
		{Method: "DELETE", Path: "/api/v1/volunteers/me/documents/{id}", Auth: "jwt", Group: "volunteer_profile", Summary: "Delete document before approval", SummaryFA: "حذف مدرک"},
		{Method: "POST", Path: "/api/v1/volunteers/me/skill-proposals", Auth: "jwt", Group: "volunteer_profile", Summary: "Propose a skill", SummaryFA: "پیشنهاد مهارت جدید"},
		{Method: "GET", Path: "/api/v1/volunteers/me/skill-proposals", Auth: "jwt", Group: "volunteer_profile", Summary: "My skill proposals", SummaryFA: "پیشنهادهای مهارت من"},

		{Method: "GET", Path: "/api/v1/tasks", Auth: "jwt", Group: "volunteer_work", Summary: "Eligible open tasks", SummaryFA: "فعالیت‌های واجد شرایط"},
		{Method: "GET", Path: "/api/v1/tasks/{id}", Auth: "jwt", Group: "volunteer_work", Summary: "Task detail", SummaryFA: "جزئیات فعالیت"},
		{Method: "POST", Path: "/api/v1/tasks/{id}/accept", Auth: "jwt", Group: "volunteer_work", Summary: "Apply / reserve capacity", SummaryFA: "درخواست فعالیت"},
		{Method: "GET", Path: "/api/v1/assignments/me", Auth: "jwt", Group: "volunteer_work", Summary: "My assignments", SummaryFA: "کارهای من"},
		{Method: "GET", Path: "/api/v1/volunteers/me/trainings", Auth: "jwt", Group: "volunteer_work", Summary: "My completed training courses", SummaryFA: "دوره‌های آموزشی گذرانده‌شده"},
		{Method: "POST", Path: "/api/v1/assignments/{id}/start", Auth: "jwt", Group: "volunteer_work", Summary: "Start remote assignment", SummaryFA: "شروع کار دورکار"},
		{Method: "POST", Path: "/api/v1/assignments/{id}/deliver", Auth: "jwt", Group: "volunteer_work", Summary: "Upload remote result", SummaryFA: "تحویل نتیجه دورکار"},
		{Method: "GET", Path: "/api/v1/assignments/{id}/files/{fileId}", Auth: "jwt", Group: "volunteer_work", Summary: "Download my delivery file", SummaryFA: "دانلود پیوست نتیجه داوطلب"},
		{Method: "POST", Path: "/api/v1/assignments/{id}/rate", Auth: "jwt", Group: "volunteer_work", Summary: "Rate organization after completion", SummaryFA: "امتیاز داوطلب به سازماندهی"},
		{Method: "POST", Path: "/api/v1/assignments/{id}/cancel", Auth: "jwt", Group: "volunteer_work", Summary: "Volunteer cancels request", SummaryFA: "انصراف داوطلب"},

		{Method: "GET", Path: "/api/v1/missions", Auth: "jwt", Group: "missions", Summary: "Active missions (public fields)", SummaryFA: "ماموریت‌های فعال"},
		{Method: "POST", Path: "/api/v1/missions/{id}/start", Auth: "jwt", Group: "missions", Summary: "Start mission", SummaryFA: "شروع ماموریت"},
		{Method: "POST", Path: "/api/v1/missions/{id}/progress", Auth: "jwt", Group: "missions", Summary: "Verify mission (not manual increment)", SummaryFA: "بررسی تأیید ماموریت"},
		{Method: "GET", Path: "/api/v1/missions/me", Auth: "jwt", Group: "missions", Summary: "My mission progress", SummaryFA: "پیشرفت ماموریت‌های من"},

		{Method: "GET", Path: "/api/v1/certificates/me", Auth: "jwt", Group: "certificates", Summary: "My certificates / appreciations", SummaryFA: "تقدیرنامه‌ها و گواهی‌های من"},
		{Method: "GET", Path: "/api/v1/certificates/requests", Auth: "jwt", Group: "certificates", Summary: "My certificate requests", SummaryFA: "درخواست‌های گواهی من"},
		{Method: "POST", Path: "/api/v1/certificates/requests", Auth: "jwt", Group: "certificates", Summary: "Request certificate", SummaryFA: "ثبت درخواست تقدیرنامه یا گواهی‌نامه"},
		{Method: "GET", Path: "/api/v1/certificates/verify/{code}", Auth: "public", Group: "certificates", Summary: "Public authenticity check", SummaryFA: "استعلام اصالت"},
		{Method: "GET", Path: "/api/v1/certificates/{code}/pdf", Auth: "public", Group: "certificates", Summary: "Download certificate PDF", SummaryFA: "دانلود PDF تقدیرنامه"},

		{Method: "GET", Path: "/api/v1/admin/dashboard", Auth: "staff", Group: "staff_ops", Summary: "Operations dashboard", SummaryFA: "داشبورد بهره‌بردار"},
		{Method: "GET", Path: "/api/v1/admin/volunteers", Auth: "staff", Group: "staff_volunteers", Summary: "List volunteers", SummaryFA: "فهرست داوطلبان"},
		{Method: "GET", Path: "/api/v1/admin/volunteers/{id}", Auth: "staff", Group: "staff_volunteers", Summary: "Volunteer dossier", SummaryFA: "پرونده داوطلب"},
		{Method: "PUT", Path: "/api/v1/admin/volunteers/{id}", Auth: "staff", Group: "staff_volunteers", Summary: "Edit volunteer profile", SummaryFA: "ویرایش پرونده"},
		{Method: "POST", Path: "/api/v1/admin/volunteers/{id}/review", Auth: "staff", Group: "staff_volunteers", Summary: "Approve / reject / request documents / suspend", SummaryFA: "بررسی عضویت"},
		{Method: "POST", Path: "/api/v1/admin/volunteers/{id}/status", Auth: "staff", Group: "staff_volunteers", Summary: "Set membership status", SummaryFA: "تغییر وضعیت عضویت"},
		{Method: "POST", Path: "/api/v1/admin/volunteers/{id}/comments", Auth: "staff", Group: "staff_volunteers", Summary: "Add file comment", SummaryFA: "ثبت نظر در پرونده"},
		{Method: "GET", Path: "/api/v1/admin/volunteers/{id}/documents", Auth: "staff", Group: "staff_volunteers", Summary: "Volunteer documents", SummaryFA: "مدارک داوطلب"},
		{Method: "GET", Path: "/api/v1/admin/volunteers/{id}/availability", Auth: "staff", Group: "staff_volunteers", Summary: "Volunteer availability", SummaryFA: "زمان آزاد داوطلب"},
		{Method: "GET", Path: "/api/v1/admin/documents/{id}", Auth: "staff", Group: "staff_volunteers", Summary: "Stream document file", SummaryFA: "دانلود مدرک"},

		{Method: "GET", Path: "/api/v1/admin/tasks", Auth: "staff", Group: "staff_tasks", Summary: "List activities", SummaryFA: "فهرست فعالیت‌ها"},
		{Method: "POST", Path: "/api/v1/admin/tasks", Auth: "staff", Group: "staff_tasks", Summary: "Create activity", SummaryFA: "تعریف فعالیت"},
		{Method: "PUT", Path: "/api/v1/admin/tasks/{id}", Auth: "staff", Group: "staff_tasks", Summary: "Update activity", SummaryFA: "ویرایش فعالیت"},
		{Method: "POST", Path: "/api/v1/admin/tasks/{id}/status", Auth: "staff", Group: "staff_tasks", Summary: "Set activity status", SummaryFA: "تغییر وضعیت فعالیت"},
		{Method: "POST", Path: "/api/v1/admin/tasks/{id}/assign", Auth: "staff", Group: "staff_tasks", Summary: "Assign volunteer to activity", SummaryFA: "تخصیص دستی داوطلب"},
		{Method: "DELETE", Path: "/api/v1/admin/tasks/{id}", Auth: "staff", Group: "staff_tasks", Summary: "Delete activity", SummaryFA: "حذف فعالیت"},
		{Method: "GET", Path: "/api/v1/admin/tasks/{id}/assignments", Auth: "staff", Group: "staff_tasks", Summary: "Assignments of one activity", SummaryFA: "درخواست‌های یک فعالیت"},

		{Method: "GET", Path: "/api/v1/admin/assignments", Auth: "staff", Group: "staff_assignments", Summary: "List assignments", SummaryFA: "فهرست تخصیص‌ها"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/approve", Auth: "staff", Group: "staff_assignments", Summary: "Approve application", SummaryFA: "تایید درخواست فعالیت"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/confirm-training", Auth: "staff", Group: "staff_assignments", Summary: "Confirm volunteer attended required training", SummaryFA: "تایید حضور داوطلب در آموزش"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/reject", Auth: "staff", Group: "staff_assignments", Summary: "Reject application", SummaryFA: "رد درخواست فعالیت"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/revision", Auth: "staff", Group: "staff_assignments", Summary: "Request remote-work revision", SummaryFA: "درخواست اصلاح نتیجه دورکار"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/message", Auth: "staff", Group: "staff_assignments", Summary: "Message volunteer about assignment", SummaryFA: "پیام به داوطلب درباره فعالیت"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/attendance", Auth: "staff", Group: "staff_assignments", Summary: "Confirm attendance (optional manual times)", SummaryFA: "ثبت حضور"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/absent", Auth: "staff", Group: "staff_assignments", Summary: "Mark absent", SummaryFA: "ثبت غیبت"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/complete", Auth: "staff", Group: "staff_assignments", Summary: "Score and complete", SummaryFA: "امتیاز و تکمیل"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/cancel", Auth: "staff", Group: "staff_assignments", Summary: "Staff cancel assignment", SummaryFA: "لغو تخصیص توسط بهره‌بردار"},
		{Method: "POST", Path: "/api/v1/admin/assignments/{id}/certificate", Auth: "staff", Group: "staff_assignments", Summary: "Issue task appreciation", SummaryFA: "صدور تقدیرنامه موردی"},
		{Method: "GET", Path: "/api/v1/admin/assignments/{id}/delivery", Auth: "staff", Group: "staff_assignments", Summary: "Download remote delivery file", SummaryFA: "دانلود فایل نتیجه دورکار"},
		{Method: "GET", Path: "/api/v1/admin/assignments/{id}/files/{fileId}", Auth: "staff", Group: "staff_assignments", Summary: "Download one delivery file from history", SummaryFA: "دانلود یک پیوست از تاریخچه نتیجه"},

		{Method: "GET", Path: "/api/v1/admin/missions", Auth: "staff", Group: "staff_missions", Summary: "List missions (includes verify secrets)", SummaryFA: "فهرست ماموریت‌ها"},
		{Method: "POST", Path: "/api/v1/admin/missions", Auth: "staff", Group: "staff_missions", Summary: "Create mission", SummaryFA: "تعریف ماموریت"},
		{Method: "PUT", Path: "/api/v1/admin/missions/{id}", Auth: "staff", Group: "staff_missions", Summary: "Update mission", SummaryFA: "ویرایش ماموریت"},

		{Method: "GET", Path: "/api/v1/admin/tickets", Auth: "staff", Group: "staff_tickets", Summary: "List tickets", SummaryFA: "فهرست تیکت‌ها"},
		{Method: "GET", Path: "/api/v1/admin/tickets/{id}", Auth: "staff", Group: "staff_tickets", Summary: "Ticket thread", SummaryFA: "جزئیات تیکت"},
		{Method: "POST", Path: "/api/v1/admin/tickets/{id}/messages", Auth: "staff", Group: "staff_tickets", Summary: "Staff reply", SummaryFA: "پاسخ بهره‌بردار"},
		{Method: "POST", Path: "/api/v1/admin/tickets/{id}/status", Auth: "staff", Group: "staff_tickets", Summary: "Set ticket status", SummaryFA: "تغییر وضعیت تیکت"},

		{Method: "POST", Path: "/api/v1/admin/volunteers/{id}/certificates/aggregated", Auth: "staff", Group: "staff_certificates", Summary: "Issue aggregated appreciation (last 12 months)", SummaryFA: "صدور تقدیرنامه تجمیعی"},
		{Method: "GET", Path: "/api/v1/admin/certificate-requests", Auth: "staff", Group: "staff_certificates", Summary: "Certificate requests queue", SummaryFA: "صف درخواست تقدیرنامه و گواهی‌نامه"},
		{Method: "POST", Path: "/api/v1/admin/certificate-requests/{id}/review", Auth: "staff", Group: "staff_certificates", Summary: "Review certificate request", SummaryFA: "بررسی درخواست گواهی"},

		{Method: "GET", Path: "/api/v1/admin/reports/ranking", Auth: "staff", Group: "staff_reports", Summary: "Volunteer ranking (json or csv)", SummaryFA: "رتبه‌بندی داوطلبان"},
		{Method: "GET", Path: "/api/v1/admin/reports/skills", Auth: "staff", Group: "staff_reports", Summary: "Skill histogram", SummaryFA: "توزیع مهارت"},
		{Method: "GET", Path: "/api/v1/admin/reports/overview", Auth: "staff", Group: "staff_reports", Summary: "Operations overview report", SummaryFA: "گزارش جامع"},

		{Method: "GET", Path: "/api/v1/admin/skills/", Auth: "staff", Group: "staff_skills", Summary: "Skill catalog (staff)", SummaryFA: "کاتالوگ مهارت بهره‌بردار"},
		{Method: "GET", Path: "/api/v1/admin/skills/catalog", Auth: "staff", Group: "staff_skills", Summary: "Skill catalog alias", SummaryFA: "کاتالوگ مهارت (مسیر جایگزین)", Alias: true},
		{Method: "POST", Path: "/api/v1/admin/skills/groups", Auth: "staff", Group: "staff_skills", Summary: "Create skill group", SummaryFA: "افزودن گروه مهارت"},
		{Method: "PUT", Path: "/api/v1/admin/skills/groups/{id}", Auth: "staff", Group: "staff_skills", Summary: "Update skill group", SummaryFA: "ویرایش گروه مهارت"},
		{Method: "DELETE", Path: "/api/v1/admin/skills/groups/{id}", Auth: "staff", Group: "staff_skills", Summary: "Delete skill group", SummaryFA: "حذف گروه مهارت"},
		{Method: "POST", Path: "/api/v1/admin/skills/", Auth: "staff", Group: "staff_skills", Summary: "Create skill", SummaryFA: "افزودن مهارت"},
		{Method: "PUT", Path: "/api/v1/admin/skills/{id}", Auth: "staff", Group: "staff_skills", Summary: "Update skill", SummaryFA: "ویرایش مهارت"},
		{Method: "DELETE", Path: "/api/v1/admin/skills/{id}", Auth: "staff", Group: "staff_skills", Summary: "Delete skill", SummaryFA: "حذف مهارت"},
		{Method: "GET", Path: "/api/v1/admin/skills/proposals", Auth: "staff", Group: "staff_skills", Summary: "Skill proposals", SummaryFA: "پیشنهادهای مهارت"},
		{Method: "POST", Path: "/api/v1/admin/skills/proposals/{id}/review", Auth: "staff", Group: "staff_skills", Summary: "Review skill proposal", SummaryFA: "بررسی پیشنهاد مهارت"},

		{Method: "GET", Path: "/api/v1/admin/skill-catalog/", Auth: "staff", Group: "staff_skills", Summary: "Skill catalog (legacy path)", SummaryFA: "کاتالوگ مهارت (مسیر قدیمی)", Alias: true},
		{Method: "POST", Path: "/api/v1/admin/skill-catalog/groups", Auth: "staff", Group: "staff_skills", Summary: "Create skill group (legacy)", SummaryFA: "افزودن گروه (مسیر قدیمی)", Alias: true},
		{Method: "PUT", Path: "/api/v1/admin/skill-catalog/groups/{id}", Auth: "staff", Group: "staff_skills", Summary: "Update skill group (legacy)", SummaryFA: "ویرایش گروه (مسیر قدیمی)", Alias: true},
		{Method: "DELETE", Path: "/api/v1/admin/skill-catalog/groups/{id}", Auth: "staff", Group: "staff_skills", Summary: "Delete skill group (legacy)", SummaryFA: "حذف گروه (مسیر قدیمی)", Alias: true},
		{Method: "POST", Path: "/api/v1/admin/skill-catalog/skills", Auth: "staff", Group: "staff_skills", Summary: "Create skill (legacy)", SummaryFA: "افزودن مهارت (مسیر قدیمی)", Alias: true},
		{Method: "PUT", Path: "/api/v1/admin/skill-catalog/skills/{id}", Auth: "staff", Group: "staff_skills", Summary: "Update skill (legacy)", SummaryFA: "ویرایش مهارت (مسیر قدیمی)", Alias: true},
		{Method: "DELETE", Path: "/api/v1/admin/skill-catalog/skills/{id}", Auth: "staff", Group: "staff_skills", Summary: "Delete skill (legacy)", SummaryFA: "حذف مهارت (مسیر قدیمی)", Alias: true},
		{Method: "GET", Path: "/api/v1/admin/skill-proposals", Auth: "staff", Group: "staff_skills", Summary: "Skill proposals (legacy)", SummaryFA: "پیشنهاد مهارت (مسیر قدیمی)", Alias: true},
		{Method: "POST", Path: "/api/v1/admin/skill-proposals/{id}/review", Auth: "staff", Group: "staff_skills", Summary: "Review proposal (legacy)", SummaryFA: "بررسی پیشنهاد (مسیر قدیمی)", Alias: true},

		{Method: "POST", Path: "/api/v1/webhooks/events", Auth: "internal", Group: "integrations", Summary: "Inbound mission award from upstream systems", SummaryFA: "وب‌هوک پیشرفت ماموریت"},
	}
}

func catalogJSON() map[string]any {
	groups := []string{
		"health", "auth", "session", "tickets", "volunteer_profile", "volunteer_work",
		"missions", "certificates", "staff_ops", "staff_volunteers", "staff_tasks",
		"staff_assignments", "staff_missions", "staff_tickets", "staff_certificates",
		"staff_reports", "staff_skills", "integrations",
	}
	by := map[string][]map[string]any{}
	for _, g := range groups {
		by[g] = []map[string]any{}
	}
	for _, r := range CatalogRoutes() {
		item := map[string]any{"method": r.Method, "path": r.Path, "auth": r.Auth, "summary": r.Summary, "summary_fa": r.SummaryFA}
		if r.Alias {
			item["alias"] = true
		}
		by[r.Group] = append(by[r.Group], item)
	}
	out := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		out = append(out, map[string]any{"name": g, "items": by[g]})
	}
	return map[string]any{
		"product":    "Mahak Volunteer Management Platform",
		"product_fa": "سامانه مدیریت داوطلبان محک",
		"service":    "mahak-volunteer-api",
		"version":    "v1",
		"roles": []map[string]string{
			{"id": "volunteer", "fa": "داوطلب"},
			{"id": "operator", "fa": "بهره‌بردار / واحد پشتیبانی"},
			{"id": "admin", "fa": "ادمین سامانه"},
		},
		"auth": map[string]string{
			"jwt":      "Authorization: Bearer <jwt>",
			"staff":    "JWT with role admin or operator",
			"internal": "X-Internal-Token: <INTERNAL_API_TOKEN>",
		},
		"docs": map[string]string{
			"openapi":  "/docs/openapi.yaml",
			"handbook": "/docs/api.md",
			"handover": "/docs/integration.md",
			"postman":  "/postman/Mahak-Volunteer-Management.postman_collection.json",
		},
		"groups": out,
	}
}
