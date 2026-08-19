package volunteeruc

import (
	"strings"
	"unicode"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func splitName(first, last, full string) (string, string) {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if first != "" || last != "" {
		return first, last
	}
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func validatePersianName(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return domain.Invalid("نام را با حروف فارسی وارد کنید")
	}
	for _, r := range s {
		if unicode.IsSpace(r) || r == '\u200c' {
			continue
		}
		if r >= 0x0600 && r <= 0x06FF {
			continue
		}
		return domain.Invalid("نام و نام خانوادگی فقط با حروف فارسی مجاز است")
	}
	return nil
}

func normalizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '۰' && r <= '۹':
			b.WriteRune('0' + (r - '۰'))
		case r >= '٠' && r <= '٩':
			b.WriteRune('0' + (r - '٠'))
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		}
	}
	return b.String()
}

func validateNationalID(s string) error {
	if len(s) != 10 {
		return domain.Invalid("کد ملی باید دقیقاً ۱۰ رقم باشد")
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return domain.Invalid("کد ملی فقط عدد است")
		}
	}
	return nil
}
