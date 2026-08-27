package taskuc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

func tehranLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return time.FixedZone("IRST", 3*3600+30*60)
	}
	return loc
}

var jalaliMonths = []string{"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}

func formatJalaliDateTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	local := t.In(tehranLoc())
	jy, jm, jd := gregorianToJalali(local.Year(), int(local.Month()), local.Day())
	month := ""
	if jm >= 1 && jm <= 12 {
		month = jalaliMonths[jm-1]
	}
	return fmt.Sprintf("%d %s %d، ساعت %02d:%02d", jd, month, jy, local.Hour(), local.Minute())
}

func gregorianToJalali(gy, gm, gd int) (int, int, int) {
	gDays := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	gy2 := gy - 1600
	gm2 := gm - 1
	gDayNo := 365*gy2 + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400
	for i := 0; i < gm2; i++ {
		gDayNo += gDays[i]
	}
	if gm2 > 1 && ((gy%4 == 0 && gy%100 != 0) || gy%400 == 0) {
		gDayNo++
	}
	gDayNo += gd - 1
	jDayNo := gDayNo - 79
	jNp := jDayNo / 12053
	jDayNo %= 12053
	jy := 979 + 33*jNp + 4*(jDayNo/1461)
	jDayNo %= 1461
	if jDayNo >= 366 {
		jy += (jDayNo - 1) / 365
		jDayNo = (jDayNo - 1) % 365
	}
	jMonths := [12]int{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29}
	for i := 0; i < 11; i++ {
		if jDayNo < jMonths[i] {
			return jy, i + 1, jDayNo + 1
		}
		jDayNo -= jMonths[i]
	}
	return jy, 12, jDayNo + 1
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

func dayKey(t time.Time) string {
	d := t.In(tehranLoc())
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

func copySeriesMeta(dst *domain.Task, parent *domain.Task) {
	dst.Title = parent.Title
	dst.Description = parent.Description
	dst.Location = parent.Location
	dst.HourWeight = parent.HourWeight
	dst.RequiredSkills = parent.RequiredSkills
	dst.RequiredSkillIDs = parent.RequiredSkillIDs
	dst.MinScore = parent.MinScore
	dst.RequiredEducation = parent.RequiredEducation
	dst.WorkMode = parent.WorkMode
	dst.DeliveryHint = parent.DeliveryHint
	dst.RequiresTraining = parent.RequiresTraining
	dst.TrainingCourseID = parent.TrainingCourseID
	dst.TrainingCourse = parent.TrainingCourse
	dst.TrainingKind = parent.TrainingKind
	dst.TrainingLocation = parent.TrainingLocation
	dst.TrainingAt = parent.TrainingAt
	dst.UpdatedAt = parent.UpdatedAt
}

func (s *Service) syncSeriesOccurrences(ctx context.Context, parent *domain.Task, in TaskInput) error {
	wanted, err := expandOccurrences(in)
	if err != nil {
		return err
	}
	children, _, err := s.tasks.List(ctx, domain.TaskFilter{SeriesID: parent.ID, Kind: domain.TaskOccurrence, Limit: 500})
	if err != nil {
		return err
	}
	byDay := map[string]*domain.Task{}
	for i := range children {
		cp := children[i]
		byDay[dayKey(cp.StartsAt)] = &children[i]
	}
	seen := map[string]struct{}{}
	sum := 0
	for _, oc := range wanted {
		sum += oc.Capacity
		k := dayKey(oc.Starts)
		seen[k] = struct{}{}
		if existing, ok := byDay[k]; ok {
			copySeriesMeta(existing, parent)
			existing.StartsAt = oc.Starts
			existing.EndsAt = oc.Ends
			existing.Weekday = oc.Weekday
			existing.Capacity = oc.Capacity
			if existing.ReservedCount > existing.Capacity {
				existing.Capacity = existing.ReservedCount
			}
			if existing.Status == domain.TaskClosed && parent.Status == domain.TaskOpen {
				existing.Status = domain.TaskOpen
			}
			if err := s.tasks.Update(ctx, existing); err != nil {
				return err
			}
			continue
		}
		child := *parent
		child.ID = uuid.New()
		child.Kind = domain.TaskOccurrence
		child.SeriesID = parent.ID
		child.StartsAt = oc.Starts
		child.EndsAt = oc.Ends
		child.Capacity = oc.Capacity
		child.ReservedCount = 0
		child.Weekday = oc.Weekday
		child.Slots = nil
		if err := s.tasks.Create(ctx, &child); err != nil {
			return err
		}
	}
	for k, leftover := range byDay {
		if _, ok := seen[k]; ok {
			continue
		}
		asgs, _, err := s.tasks.ListAssignments(ctx, domain.AssignmentFilter{TaskID: leftover.ID, Limit: 50})
		if err != nil {
			return err
		}
		keep := false
		for _, a := range asgs {
			if a.Status != domain.AssignmentCancelled && a.Status != domain.AssignmentRejected {
				keep = true
				break
			}
		}
		if keep {
			leftover.Status = domain.TaskClosed
			leftover.UpdatedAt = parent.UpdatedAt
			if err := s.tasks.Update(ctx, leftover); err != nil {
				return err
			}
			continue
		}
		if err := s.tasks.Delete(ctx, leftover.ID); err != nil {
			leftover.Status = domain.TaskClosed
			leftover.UpdatedAt = parent.UpdatedAt
			_ = s.tasks.Update(ctx, leftover)
		}
	}
	parent.Capacity = sum
	parent.Slots = in.Slots
	return s.tasks.Update(ctx, parent)
}
