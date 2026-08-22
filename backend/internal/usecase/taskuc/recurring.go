package taskuc

import (
	"fmt"
	"strings"
	"time"

	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func tehranLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return time.FixedZone("IRST", 3*3600+30*60)
	}
	return loc
}

func parseHM(s, fallback string) (int, int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		s = fallback
	}
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, domain.Invalid("ساعت نوبت نامعتبر است؛ قالب درست مثلا ۰۹:۰۰ است")
	}
	return h, m, nil
}

type occurrence struct {
	Starts   time.Time
	Ends     time.Time
	Weekday  int
	Capacity int
}

func expandOccurrences(in TaskInput) ([]occurrence, error) {
	if len(in.Slots) == 0 {
		return nil, domain.Invalid("برای فعالیت جاری حداقل یک روز هفته با ظرفیت مشخص کنید")
	}
	loc := tehranLoc()
	start := in.StartsAt.In(loc)
	end := in.EndsAt.In(loc)
	day0 := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	day1 := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)
	if day1.Before(day0) {
		return nil, domain.Invalid("تاریخ پایان بازه باید بعد از تاریخ شروع باشد")
	}
	byWD := map[int]domain.TaskSlot{}
	for _, sl := range in.Slots {
		if sl.Weekday < 0 || sl.Weekday > 6 {
			return nil, domain.Invalid("روز هفته نامعتبر است")
		}
		if sl.Capacity < 1 {
			return nil, domain.Invalid("ظرفیت هر روز هفته باید حداقل ۱ نفر باشد")
		}
		byWD[sl.Weekday] = sl
	}
	var out []occurrence
	for d := day0; !d.After(day1); d = d.AddDate(0, 0, 1) {
		sl, ok := byWD[int(d.Weekday())]
		if !ok {
			continue
		}
		sh, sm, err := parseHM(sl.StartTime, "09:00")
		if err != nil {
			return nil, err
		}
		eh, em, err := parseHM(sl.EndTime, "13:00")
		if err != nil {
			return nil, err
		}
		st := time.Date(d.Year(), d.Month(), d.Day(), sh, sm, 0, 0, loc)
		en := time.Date(d.Year(), d.Month(), d.Day(), eh, em, 0, 0, loc)
		if !en.After(st) {
			return nil, domain.Invalid("ساعت پایان نوبت باید بعد از ساعت شروع همان روز باشد")
		}
		out = append(out, occurrence{Starts: st.UTC(), Ends: en.UTC(), Weekday: sl.Weekday, Capacity: sl.Capacity})
		if len(out) > 180 {
			return nil, domain.Invalid("تعداد نوبت‌ها بیش از حد است؛ بازه یا روزهای هفته را کمتر کنید")
		}
	}
	if len(out) == 0 {
		return nil, domain.Invalid("در این بازه هیچ نوبتی برای روزهای انتخاب‌شده وجود ندارد")
	}
	return out, nil
}
