package missionuc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func (s *Service) verifiedCount(ctx context.Context, v *domain.Volunteer, m *domain.Mission) (int, string, error) {
	switch m.VerifyMode {
	case domain.VerifyInternal:
		return s.internalCount(ctx, v, m)
	case domain.VerifyOutbound:
		return s.outboundCount(ctx, v, m)
	case domain.VerifyInbound:
		if strings.TrimSpace(m.VerifyURL) != "" {
			return s.outboundCount(ctx, v, m)
		}
		return 0, "این ماموریت فقط وقتی جلو می‌رود که سرویس خارجی با توکن ماموریت، وب‌هوک را صدا بزند", nil
	default:
		return 0, "روش تأیید این ماموریت نامعتبر است", nil
	}
}

func (s *Service) internalCount(ctx context.Context, v *domain.Volunteer, m *domain.Mission) (int, string, error) {
	if m.Kind == domain.MissionCompleteProfile {
		if reason := profileIncompleteReason(ctx, s.volunteers, v); reason != "" {
			return 0, reason, nil
		}
		return m.TargetCount, "", nil
	}
	return 0, "برای این نوع ماموریت بررسی داخلی تعریف نشده است؛ وب‌سرویس یا وب‌هوک را در پنل ادمین تنظیم کنید", nil
}

type outboundPayload struct {
	Event       string `json:"event"`
	MissionID   string `json:"mission_id"`
	MissionKind string `json:"mission_kind"`
	VolunteerID string `json:"volunteer_id"`
	Phone       string `json:"phone,omitempty"`
	NationalID  string `json:"national_id,omitempty"`
	TargetCount int    `json:"target_count"`
}

type outboundReply struct {
	OK        *bool  `json:"ok"`
	Completed *bool  `json:"completed"`
	Progress  *int   `json:"progress"`
	Count     *int   `json:"count"`
	Message   string `json:"message"`
}

func (s *Service) outboundCount(ctx context.Context, v *domain.Volunteer, m *domain.Mission) (int, string, error) {
	body, _ := json.Marshal(outboundPayload{
		Event:       "mission.verify",
		MissionID:   m.ID.String(),
		MissionKind: string(m.Kind),
		VolunteerID: v.ID.String(),
		Phone:       v.Phone,
		NationalID:  v.NationalID,
		TargetCount: m.TargetCount,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.VerifyURL, bytes.NewReader(body))
	if err != nil {
		return 0, "", domain.Invalid("آدرس وب‌سرویس تأیید نامعتبر است")
	}
	req.Header.Set("Content-Type", "application/json")
	if m.VerifyToken != "" {
		req.Header.Set("Authorization", "Bearer "+m.VerifyToken)
	}
	res, err := s.http.Do(req)
	if err != nil {
		return 0, "سرویس تأیید در دسترس نیست", nil
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return 0, "وب‌سرویس تأیید، توکن را نپذیرفت", nil
	}
	if res.StatusCode >= 400 {
		msg := "وب‌سرویس تأیید این ماموریت را تأیید نکرد"
		var reply outboundReply
		if json.Unmarshal(raw, &reply) == nil && strings.TrimSpace(reply.Message) != "" {
			msg = reply.Message
		}
		return 0, msg, nil
	}
	var reply outboundReply
	if err := json.Unmarshal(raw, &reply); err != nil {
		return 0, "پاسخ وب‌سرویس تأیید قابل خواندن نیست", nil
	}
	if reply.Message == "" {
		reply.Message = "سرویس مربوطه انجام ماموریت را تأیید نکرد"
	}
	if reply.Progress != nil {
		return *reply.Progress, reply.Message, nil
	}
	if reply.Count != nil {
		return *reply.Count, reply.Message, nil
	}
	if reply.Completed != nil && *reply.Completed {
		return m.TargetCount, "", nil
	}
	if reply.OK != nil && *reply.OK {
		return m.TargetCount, "", nil
	}
	return 0, reply.Message, nil
}

func profileIncompleteReason(ctx context.Context, volunteers domain.VolunteerRepository, v *domain.Volunteer) string {
	if v.Status != domain.StatusPending && v.Status != domain.StatusApproved && v.Status != domain.StatusSuspended {
		return "پروفایل هنوز ارسال نشده است. فرم ثبت‌نام را کامل کنید و برای بررسی بفرستید"
	}
	if strings.TrimSpace(v.FirstName) == "" && strings.TrimSpace(v.FullName) == "" {
		return "نام در پروفایل ثبت نشده است"
	}
	if strings.TrimSpace(v.LastName) == "" && !strings.Contains(strings.TrimSpace(v.FullName), " ") {
		return "نام خانوادگی در پروفایل ثبت نشده است"
	}
	if !isTenDigitID(v.NationalID) {
		return "کد ملی معتبر در پروفایل ثبت نشده است"
	}
	if strings.TrimSpace(v.Phone) == "" {
		return "شماره موبایل در پروفایل ثبت نشده است"
	}
	if strings.TrimSpace(v.BirthDate) == "" {
		return "تاریخ تولد در پروفایل ثبت نشده است"
	}
	if strings.TrimSpace(v.Province) == "" || strings.TrimSpace(v.City) == "" {
		return "استان و شهر در پروفایل ثبت نشده است"
	}
	if strings.TrimSpace(v.EducationLevel) == "" {
		return "میزان تحصیلات در پروفایل ثبت نشده است"
	}
	docs, err := volunteers.ListDocuments(ctx, v.ID)
	if err != nil {
		return "مدارک پروفایل قابل بررسی نیست"
	}
	hasNational := false
	for _, d := range docs {
		if d.Kind == domain.DocNationalID {
			hasNational = true
			break
		}
	}
	if !hasNational {
		return "تصویر کارت ملی در پروفایل بارگذاری نشده است"
	}
	skills, err := volunteers.ListVolunteerSkills(ctx, v.ID)
	if err != nil {
		return "مهارت‌های پروفایل قابل بررسی نیست"
	}
	if len(skills) == 0 && len(v.SkillCategories) == 0 {
		return "حداقل یک مهارت در پروفایل انتخاب نشده است"
	}
	return ""
}

func isTenDigitID(s string) bool {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= '۰' && r <= '۹':
			b.WriteRune('0' + (r - '۰'))
		case unicode.IsSpace(r):
		default:
			return false
		}
	}
	return b.Len() == 10
}
