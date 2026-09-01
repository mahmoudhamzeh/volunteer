package volunteeruc

import (
	"strings"
	"time"
	"unicode"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

const minVolunteerAge = 18
const birthDateLayout = "2006-01-02"

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

const occupationOther = "other"
const occupationOtherMax = 80

var genders = map[string]struct{}{
	"male":   {},
	"female": {},
}

var occupations = map[string]struct{}{
	"student":       {},
	"employee":      {},
	"teacher":       {},
	"medical":       {},
	"engineer":      {},
	"worker":        {},
	"self_employed": {},
	"homemaker":     {},
	"retired":       {},
	"unemployed":    {},
	"soldier":       {},
	"seminary":      {},
	"other":         {},
}

func validGender(s string) bool {
	_, ok := genders[strings.TrimSpace(s)]
	return ok
}

func validOccupation(s string) bool {
	_, ok := occupations[strings.TrimSpace(s)]
	return ok
}

func applyGenderOccupation(v *domain.Volunteer, in ProfileInput) error {
	if g := strings.TrimSpace(in.Gender); g != "" {
		if !validGender(g) {
			return domain.Invalid("جنسیت را از فهرست انتخاب کنید")
		}
		v.Gender = g
	}
	if occ := strings.TrimSpace(in.Occupation); occ != "" {
		if !validOccupation(occ) {
			return domain.Invalid("شغل را از فهرست انتخاب کنید")
		}
		v.Occupation = occ
		if occ == occupationOther {
			other := strings.TrimSpace(in.OccupationOther)
			if countRunes(other) > occupationOtherMax {
				return domain.Invalid("شرح شغل نباید بیشتر از ۸۰ حرف باشد")
			}
			v.OccupationOther = other
		} else {
			v.OccupationOther = ""
		}
	}
	return nil
}

func countRunes(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func parseBirthDate(s string) (time.Time, error) {
	bd, err := time.Parse(birthDateLayout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, domain.Invalid("تاریخ تولد نامعتبر است")
	}
	return bd, nil
}

func ageYears(birth, now time.Time) int {
	by, bm, bd := birth.Date()
	ny, nm, nd := now.Date()
	age := ny - by
	if nm < bm || (nm == bm && nd < bd) {
		age--
	}
	return age
}

func validateBirthDate(s string, now time.Time) error {
	birth, err := parseBirthDate(s)
	if err != nil {
		return err
	}
	todayY, todayM, todayD := now.Date()
	if birth.After(time.Date(todayY, todayM, todayD, 0, 0, 0, 0, time.UTC)) {
		return domain.Invalid("تاریخ تولد نمی‌تواند در آینده باشد")
	}
	if ageYears(birth, now) < minVolunteerAge {
		return domain.Invalid("حداقل سن داوطلبی ۱۸ سال تمام است")
	}
	return nil
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
