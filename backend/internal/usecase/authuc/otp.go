package authuc

import (
	"context"
	"crypto/rand"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type otpChallenge struct {
	hash      string
	expiresAt time.Time
	sentAt    time.Time
	attempts  int
}

type OTPSendResult struct {
	Phone       string `json:"phone"`
	TTLSeconds  int    `json:"ttl_seconds"`
	ResendAfter int    `json:"resend_after"`
	IsNew       bool   `json:"is_new"`
	DevCode     string `json:"dev_code,omitempty"`
}

func (s *Service) SendOTP(ctx context.Context, rawPhone string) (*OTPSendResult, error) {
	phone, err := NormalizeMobile(rawPhone)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	s.otpMu.Lock()
	if cur, ok := s.otp[phone]; ok && now.Sub(cur.sentAt) < s.otpCooldown {
		wait := int((s.otpCooldown - now.Sub(cur.sentAt)).Seconds()) + 1
		s.otpMu.Unlock()
		return nil, domain.Invalid("برای ارسال مجدد " + strconv.Itoa(wait) + " ثانیه صبر کنید")
	}
	s.otpMu.Unlock()

	code, err := randomDigits(5)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	s.otpMu.Lock()
	s.otp[phone] = otpChallenge{
		hash:      string(hash),
		expiresAt: now.Add(s.otpTTL),
		sentAt:    now,
	}
	s.otpMu.Unlock()

	log.Printf("otp sent to %s", phone)
	isNew := true
	if _, err := s.findVolunteerUser(ctx, phone); err == nil {
		isNew = false
	}
	out := &OTPSendResult{
		Phone:       phone,
		TTLSeconds:  int(s.otpTTL.Seconds()),
		ResendAfter: int(s.otpCooldown.Seconds()),
		IsNew:       isNew,
	}
	if s.revealOTP {
		out.DevCode = code
	}
	return out, nil
}

func (s *Service) VerifyOTP(ctx context.Context, rawPhone, code, fullName string) (*domain.User, string, bool, error) {
	phone, err := NormalizeMobile(rawPhone)
	if err != nil {
		return nil, "", false, err
	}
	code = normalizeDigits(strings.TrimSpace(code))
	if len(code) != 5 {
		return nil, "", false, domain.Invalid("کد ۵ رقمی پیامک را وارد کنید")
	}
	now := time.Now().UTC()
	s.otpMu.Lock()
	ch, ok := s.otp[phone]
	if !ok || now.After(ch.expiresAt) {
		s.otpMu.Unlock()
		return nil, "", false, domain.Invalid("کد منقضی شده است؛ دوباره پیامک بفرستید")
	}
	if ch.attempts >= 5 {
		delete(s.otp, phone)
		s.otpMu.Unlock()
		return nil, "", false, domain.Invalid("تعداد تلاش بیش از حد است؛ دوباره پیامک بفرستید")
	}
	if bcrypt.CompareHashAndPassword([]byte(ch.hash), []byte(code)) != nil {
		ch.attempts++
		s.otp[phone] = ch
		s.otpMu.Unlock()
		return nil, "", false, domain.Invalid("کد واردشده نادرست است")
	}
	delete(s.otp, phone)
	s.otpMu.Unlock()

	if u, err := s.findVolunteerUser(ctx, phone); err == nil {
		token, err := s.issue(u)
		return u, token, false, err
	}

	name := strings.TrimSpace(fullName)
	if name == "" {
		name = "داوطلب"
	}
	u, token, err := s.registerByPhone(ctx, phone, name)
	return u, token, true, err
}

func (s *Service) findVolunteerUser(ctx context.Context, phone string) (*domain.User, error) {
	if u, err := s.users.GetByPhone(ctx, phone); err == nil {
		return u, nil
	}
	v, err := s.volunteers.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	return s.users.GetByID(ctx, v.UserID)
}

func (s *Service) registerByPhone(ctx context.Context, phone, fullName string) (*domain.User, string, error) {
	now := time.Now().UTC()
	hash, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	u := &domain.User{
		ID:           uuid.New(),
		Email:        phone + "@otp.mahak.local",
		Phone:        phone,
		PasswordHash: string(hash),
		Role:         domain.RoleVolunteer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err
	}
	if err := s.volunteers.Create(ctx, &domain.Volunteer{
		ID:              uuid.New(),
		UserID:          u.ID,
		FullName:        fullName,
		Phone:           phone,
		Status:          domain.StatusDraft,
		SkillCategories: []domain.SkillCategory{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		return nil, "", err
	}
	token, err := s.issue(u)
	return u, token, err
}

func NormalizeMobile(raw string) (string, error) {
	d := normalizeDigits(raw)
	var b strings.Builder
	for _, r := range d {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	n := b.String()
	switch {
	case strings.HasPrefix(n, "0098") && len(n) == 14:
		n = "0" + n[4:]
	case strings.HasPrefix(n, "98") && len(n) == 12:
		n = "0" + n[2:]
	case strings.HasPrefix(n, "9") && len(n) == 10:
		n = "0" + n
	}
	if len(n) != 11 || !strings.HasPrefix(n, "09") {
		return "", domain.Invalid("شماره موبایل را به‌صورت ۰۹۱۲۱۲۳۴۵۶۷ وارد کنید")
	}
	return n, nil
}

func normalizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '۰' && r <= '۹':
			b.WriteRune('0' + (r - '۰'))
		case r >= '٠' && r <= '٩':
			b.WriteRune('0' + (r - '٠'))
		case unicode.IsSpace(r) || r == '-' || r == '_':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func randomDigits(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, v := range buf {
		out[i] = '0' + (v % 10)
	}
	if out[0] == '0' {
		out[0] = '1'
	}
	return string(out), nil
}
