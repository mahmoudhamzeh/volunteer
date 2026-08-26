package taskuc

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/scoring"
)

type Service struct {
	tasks      domain.TaskRepository
	volunteers domain.VolunteerRepository
	certs      domain.CertificateRepository
	locker     domain.Locker
	notify     domain.Notifier
	clock      domain.Clock
}

func New(tasks domain.TaskRepository, volunteers domain.VolunteerRepository, certs domain.CertificateRepository, locker domain.Locker, notify domain.Notifier, clock domain.Clock) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	return &Service{tasks: tasks, volunteers: volunteers, certs: certs, locker: locker, notify: notify, clock: clock}
}

type TaskInput struct {
	Title             string
	Description       string
	Location          string
	StartsAt          time.Time
	EndsAt            time.Time
	Capacity          int
	HourWeight        float64
	RequiredSkills    []string
	RequiredSkillIDs  []string
	MinScore          float64
	RequiredEducation string
	WorkMode          string
	DeliveryHint      string
	RequiresTraining  bool
	TrainingKind      string
	TrainingLocation  string
	TrainingAt        *time.Time
	Status            domain.TaskStatus
	Kind              string
	Slots             []domain.TaskSlot
}

func (s *Service) Create(ctx context.Context, actor uuid.UUID, in TaskInput) (*domain.Task, error) {
	if err := validateTask(in); err != nil {
		return nil, err
	}
	now := s.clock.Now()
	kind := strings.TrimSpace(in.Kind)
	if kind == "" {
		kind = domain.TaskOneOff
	}
	t := &domain.Task{
		ID:                uuid.New(),
		Title:             strings.TrimSpace(in.Title),
		Description:       strings.TrimSpace(in.Description),
		Location:          strings.TrimSpace(in.Location),
		StartsAt:          in.StartsAt.UTC(),
		EndsAt:            in.EndsAt.UTC(),
		Capacity:          in.Capacity,
		HourWeight:        in.HourWeight,
		RequiredSkills:    domain.ParseSkillCategories(in.RequiredSkills),
		RequiredSkillIDs:  parseUUIDs(in.RequiredSkillIDs),
		MinScore:          in.MinScore,
		RequiredEducation: strings.TrimSpace(in.RequiredEducation),
		WorkMode:          domain.ParseWorkMode(in.WorkMode),
		DeliveryHint:      strings.TrimSpace(in.DeliveryHint),
		Kind:              kind,
		Slots:             in.Slots,
		Status:            domain.TaskOpen,
		CreatedBy:         actor,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	applyTraining(t, in)
	if in.Status != "" {
		t.Status = in.Status
	}
	if kind == domain.TaskRecurring {
		occs, err := expandOccurrences(in)
		if err != nil {
			return nil, err
		}
		t.SeriesID = t.ID
		sum := 0
		for _, oc := range occs {
			sum += oc.Capacity
		}
		t.Capacity = sum
		if err := s.tasks.Create(ctx, t); err != nil {
			return nil, err
		}
		for _, oc := range occs {
			child := *t
			child.ID = uuid.New()
			child.Kind = domain.TaskOccurrence
			child.SeriesID = t.ID
			child.StartsAt = oc.Starts
			child.EndsAt = oc.Ends
			child.Capacity = oc.Capacity
			child.ReservedCount = 0
			child.Weekday = oc.Weekday
			child.Slots = nil
			if err := s.tasks.Create(ctx, &child); err != nil {
				return nil, err
			}
		}
		return t, nil
	}
	return t, s.tasks.Create(ctx, t)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in TaskInput) (*domain.Task, error) {
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := validateTask(in); err != nil {
		return nil, err
	}
	t.Title = strings.TrimSpace(in.Title)
	t.Description = strings.TrimSpace(in.Description)
	t.Location = strings.TrimSpace(in.Location)
	t.StartsAt = in.StartsAt.UTC()
	t.EndsAt = in.EndsAt.UTC()
	t.Capacity = in.Capacity
	t.HourWeight = in.HourWeight
	t.RequiredSkills = domain.ParseSkillCategories(in.RequiredSkills)
	t.RequiredSkillIDs = parseUUIDs(in.RequiredSkillIDs)
	t.MinScore = in.MinScore
	t.RequiredEducation = strings.TrimSpace(in.RequiredEducation)
	t.WorkMode = domain.ParseWorkMode(in.WorkMode)
	t.DeliveryHint = strings.TrimSpace(in.DeliveryHint)
	applyTraining(t, in)
	if in.Kind != "" {
		t.Kind = in.Kind
	}
	if t.Kind == domain.TaskRecurring {
		t.Slots = in.Slots
	} else if len(in.Slots) > 0 {
		t.Slots = in.Slots
	}
	if in.Status != "" {
		t.Status = in.Status
	}
	t.UpdatedAt = s.clock.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	if t.Kind == domain.TaskRecurring {
		if err := s.syncSeriesOccurrences(ctx, t, in); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (s *Service) SetStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) (*domain.Task, error) {
	if !domain.ValidTaskStatus(status) {
		return nil, domain.Invalid("وضعیت فعالیت نامعتبر است")
	}
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Status = status
	t.UpdatedAt = s.clock.Now()
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	if t.Kind == domain.TaskRecurring {
		children, _, err := s.tasks.List(ctx, domain.TaskFilter{SeriesID: t.ID, Kind: domain.TaskOccurrence, Limit: 500})
		if err == nil {
			for i := range children {
				children[i].Status = status
				children[i].UpdatedAt = t.UpdatedAt
				_ = s.tasks.Update(ctx, &children[i])
			}
		}
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.tasks.Delete(ctx, id)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return s.tasks.GetByID(ctx, id)
}

func (s *Service) GetAssignment(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	return s.tasks.GetAssignment(ctx, id)
}

func (s *Service) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 500 {
		f.Limit = 500
	}
	_ = s.CloseExpired(ctx)
	return s.tasks.List(ctx, f)
}

func (s *Service) ListEligible(ctx context.Context, userID uuid.UUID, f domain.TaskFilter) ([]domain.Task, int, error) {
	if f.Limit <= 0 {
		f.Limit = 200
	}
	if f.Limit > 500 {
		f.Limit = 500
	}
	_ = s.CloseExpired(ctx)
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if v.Status == domain.StatusSuspended {
		return []domain.Task{}, 0, nil
	}
	if !v.Status.CanViewTasks() {
		return nil, 0, domain.ErrNotApproved
	}
	f.Status = domain.TaskOpen
	f.Upcoming = true
	f.ExcludeVolunteerID = v.ID
	f.ExcludeKind = domain.TaskRecurring
	if skills, err := s.volunteers.ListVolunteerSkills(ctx, v.ID); err == nil {
		v.Skills = skills
	}
	tasks, total, err := s.tasks.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.Task, 0, len(tasks))
	for _, t := range tasks {
		if scoring.EligibleForTask(*v, t) == nil {
			out = append(out, t)
		}
	}
	return out, total, nil
}

func (s *Service) Accept(ctx context.Context, userID, taskID uuid.UUID) (*domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if v.Status == domain.StatusSuspended {
		return nil, domain.Invalid("در حال حاضر امکان درخواست این فعالیت وجود ندارد")
	}
	if skills, err := s.volunteers.ListVolunteerSkills(ctx, v.ID); err == nil {
		v.Skills = skills
	}
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !t.IsOpen() {
		return nil, domain.ErrNotEligible
	}
	if err := scoring.EligibleForTask(*v, *t); err != nil {
		return nil, err
	}
	unlock := func() {}
	if s.locker != nil {
		unlockFn, err := s.locker.Lock(ctx, "task:"+taskID.String(), 8*time.Second)
		if err != nil {
			return nil, err
		}
		unlock = unlockFn
	}
	defer unlock()
	asg, err := s.tasks.ApplySeat(ctx, taskID, v.ID)
	if err != nil {
		return nil, err
	}
	asg.Task = t
	asg.Volunteer = v
	if s.notify != nil {
		body := "درخواست شما برای «" + t.Title + "» ثبت شد و پس از تایید واحد پشتیبانی نهایی می‌شود."
		if t.RequiresTraining {
			body += " این فعالیت نیاز به آموزش دارد. پس از تایید درخواست، ابتدا باید در آموزش شرکت کنید تا واحد پشتیبانی حضور شما در آموزش را تایید کند."
		}
		_ = s.notify.Notify(ctx, v.UserID, "درخواست فعالیت ثبت شد", body)
		if sn, ok := s.notify.(interface {
			NotifyStaff(ctx context.Context, title, body string) error
		}); ok {
			staff := v.FullName + " برای «" + t.Title + "» درخواست داده است."
			if t.RequiresTraining {
				staff += " این فعالیت نیاز به آموزش دارد."
			}
			_ = sn.NotifyStaff(ctx, "درخواست فعالیت جدید", staff)
		}
	}
	return asg, nil
}

func (s *Service) Approve(ctx context.Context, assignmentID uuid.UUID) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AssignmentRequested {
		return nil, domain.ErrInvalidTransition
	}
	unlock := func() {}
	if s.locker != nil {
		unlockFn, err := s.locker.Lock(ctx, "task:"+a.TaskID.String(), 8*time.Second)
		if err != nil {
			return nil, err
		}
		unlock = unlockFn
	}
	defer unlock()
	a, err = s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	return s.promoteToReserved(ctx, a, false)
}

func (s *Service) AssignVolunteer(ctx context.Context, taskID, volunteerID uuid.UUID) (*domain.Assignment, error) {
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return nil, err
	}
	if !v.Status.CanViewTasks() {
		return nil, domain.Invalid("فقط داوطلب تاییدشده را می‌توان به فعالیت تخصیص داد")
	}
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t.Status != domain.TaskOpen {
		return nil, domain.Invalid("فقط فعالیت باز را می‌توان تخصیص داد")
	}
	unlock := func() {}
	if s.locker != nil {
		unlockFn, err := s.locker.Lock(ctx, "task:"+taskID.String(), 8*time.Second)
		if err != nil {
			return nil, err
		}
		unlock = unlockFn
	}
	defer unlock()
	existing, err := s.tasks.GetAssignmentByTaskVolunteer(ctx, taskID, volunteerID)
	if err == nil {
		if existing.Status == domain.AssignmentRequested {
			a, err := s.promoteToReserved(ctx, existing, true)
			if err != nil {
				return nil, err
			}
			a.Volunteer = v
			return a, nil
		}
		if existing.Status.BlocksReapply() {
			return nil, domain.ErrAlreadyAssigned
		}
	} else if err != domain.ErrNotFound {
		return nil, err
	}
	asg, err := s.tasks.ApplySeat(ctx, taskID, volunteerID)
	if err != nil {
		return nil, err
	}
	a, err := s.promoteToReserved(ctx, asg, true)
	if err != nil {
		return nil, err
	}
	a.Volunteer = v
	return a, nil
}

func (s *Service) promoteToReserved(ctx context.Context, a *domain.Assignment, byAdmin bool) (*domain.Assignment, error) {
	if a.Status != domain.AssignmentRequested {
		return nil, domain.ErrInvalidTransition
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if t.ReservedCount >= t.Capacity {
		return nil, domain.ErrCapacityFull
	}
	now := s.clock.Now()
	t.ReservedCount++
	t.UpdatedAt = now
	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	needTrain := t.RequiresTraining
	trained := false
	if needTrain {
		trained, err = s.tasks.HasCompletedTraining(ctx, a.VolunteerID, t)
		if err != nil {
			return nil, err
		}
	}
	if needTrain && !trained {
		a.Status = domain.AssignmentTrainingPending
	} else {
		a.Status = domain.AssignmentReserved
	}
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	title, body := s.approvalNotice(t, byAdmin, needTrain, trained)
	s.notifyVolunteer(ctx, a.VolunteerID, title, body)
	if needTrain && !trained && t.TrainingAt != nil {
		s.notifyVolunteerReminder(ctx, a.VolunteerID, "یادآوری آموزش",
			"آموزش فعالیت «"+t.Title+"» — "+trainingDetail(t), *t.TrainingAt)
	}
	a.Task = t
	return a, nil
}

func (s *Service) approvalNotice(t *domain.Task, byAdmin, needTrain, trained bool) (string, string) {
	if byAdmin {
		title, body := "به فعالیت تخصیص داده شدید", "واحد پشتیبانی شما را به فعالیت «"+t.Title+"» تخصیص داد."
		if needTrain && !trained {
			return title, body + " ابتدا در آموزش این فعالیت شرکت کنید. پس از برگزاری، واحد پشتیبانی حضور شما در آموزش را تایید می‌کند. " + trainingDetail(t)
		}
		if needTrain && trained {
			return title, body + " آموزش این فعالیت قبلاً در پرونده شما ثبت شده است و نیاز به حضور مجدد در آموزش نیست. برای انجام فعالیت متناسب با زمان‌بندی در محل حضور داشته باشید."
		}
		return title, body
	}
	title, body := "فعالیت تایید شد", "درخواست شما برای «"+t.Title+"» توسط واحد پشتیبانی تایید شد."
	if needTrain && !trained {
		return title, body + " ابتدا در آموزش این فعالیت شرکت کنید. پس از برگزاری، واحد پشتیبانی حضور شما در آموزش را تایید می‌کند. " + trainingDetail(t)
	}
	if needTrain && trained {
		return title, body + " آموزش این فعالیت قبلاً در پرونده شما ثبت شده است و نیاز به حضور مجدد در آموزش نیست. برای انجام فعالیت متناسب با زمان‌بندی در محل حضور داشته باشید."
	}
	return title, body
}

func (s *Service) ConfirmTraining(ctx context.Context, assignmentID, confirmedBy uuid.UUID) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AssignmentTrainingPending {
		return nil, domain.ErrInvalidTransition
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if !t.RequiresTraining {
		return nil, domain.Invalid("این فعالیت نیاز به آموزش ندارد")
	}
	already, err := s.tasks.HasCompletedTraining(ctx, a.VolunteerID, t)
	if err != nil {
		return nil, err
	}
	if !already {
		vt := &domain.VolunteerTraining{
			ID:               uuid.New(),
			VolunteerID:      a.VolunteerID,
			SeriesID:         t.TrainingSeriesID(),
			TrainingKind:     t.TrainingKind,
			TrainingLocation: t.TrainingLocation,
			TrainingAt:       t.TrainingAt,
			SourceTaskID:     t.ID,
			SourceTaskTitle:  t.Title,
			AssignmentID:     a.ID,
			ConfirmedBy:      confirmedBy,
			ConfirmedAt:      s.clock.Now(),
		}
		if err := s.tasks.CreateVolunteerTraining(ctx, vt); err != nil {
			return nil, err
		}
	}
	pending, _, err := s.tasks.ListAssignments(ctx, domain.AssignmentFilter{
		VolunteerID: a.VolunteerID,
		Status:      domain.AssignmentTrainingPending,
		Limit:       200,
	})
	if err != nil {
		return nil, err
	}
	var current *domain.Assignment
	for i := range pending {
		p := &pending[i]
		task := p.Task
		if task == nil {
			got, err := s.tasks.GetByID(ctx, p.TaskID)
			if err != nil {
				continue
			}
			task = got
		}
		if p.ID != a.ID && !courseMatches(t, task) {
			continue
		}
		p.Status = domain.AssignmentReserved
		if err := s.tasks.UpdateAssignment(ctx, p); err != nil {
			return nil, err
		}
		if p.ID == a.ID {
			cp := *p
			current = &cp
		}
	}
	if current == nil {
		a.Status = domain.AssignmentReserved
		if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
			return nil, err
		}
		current = a
	}
	s.notifyVolunteer(ctx, a.VolunteerID, "آموزش تایید شد",
		"حضور شما در آموزش «"+t.Title+"» تایید شد و این دوره به فهرست آموزش‌های شما اضافه شد. برای انجام فعالیت متناسب با زمان‌بندی فعالیت در محل حضور داشته باشید.")
	current.Task = t
	return current, nil
}

func courseMatches(src, other *domain.Task) bool {
	if src == nil || other == nil {
		return false
	}
	vt := domain.VolunteerTraining{
		SeriesID:         src.TrainingSeriesID(),
		TrainingKind:     src.TrainingKind,
		TrainingLocation: src.TrainingLocation,
	}
	return vt.CoversTask(*other)
}

func (s *Service) MessageApplicant(ctx context.Context, assignmentID uuid.UUID, body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return domain.Invalid("متن پیام را وارد کنید")
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return err
	}
	title := "پیام واحد پشتیبانی"
	if a.Task != nil && a.Task.Title != "" {
		title = "پیام واحد پشتیبانی درباره «" + a.Task.Title + "»"
	}
	s.notifyVolunteer(ctx, a.VolunteerID, title, body)
	return nil
}

func (s *Service) notifyVolunteer(ctx context.Context, volunteerID uuid.UUID, title, body string) {
	if s.notify == nil {
		return
	}
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return
	}
	_ = s.notify.Notify(ctx, v.UserID, title, body)
}

func (s *Service) notifyVolunteerReminder(ctx context.Context, volunteerID uuid.UUID, title, body string, remindAt time.Time) {
	if s.notify == nil {
		return
	}
	v, err := s.volunteers.GetByID(ctx, volunteerID)
	if err != nil {
		return
	}
	if rn, ok := s.notify.(interface {
		NotifyReminder(ctx context.Context, userID uuid.UUID, title, body string, remindAt time.Time) error
	}); ok {
		_ = rn.NotifyReminder(ctx, v.UserID, title, body, remindAt)
		return
	}
	_ = s.notify.Notify(ctx, v.UserID, title, body)
}

func (s *Service) StartWork(ctx context.Context, userID, assignmentID uuid.UUID) (*domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.VolunteerID != v.ID {
		return nil, domain.ErrForbidden
	}
	if a.Status != domain.AssignmentReserved {
		return nil, domain.ErrInvalidTransition
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if !t.IsRemote() {
		return nil, domain.Invalid("شروع فعالیت فقط برای کارهای دورکار است. حضور در فعالیت حضوری را واحد پشتیبانی ثبت می‌کند")
	}
	a.Status = domain.AssignmentInProgress
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	s.notifyVolunteer(ctx, a.VolunteerID, "فعالیت شروع شد", "فعالیت «"+t.Title+"» شروع شد. پس از انجام کار، نتیجه را ارسال کنید.")
	a.Task = t
	a.Volunteer = v
	return a, nil
}

func (s *Service) ConfirmAttendance(ctx context.Context, assignmentID uuid.UUID, in AttendanceInput) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if t.IsRemote() {
		return nil, domain.Invalid("این فعالیت دورکار است و نیاز به حضور حضوری ندارد")
	}
	already := a.Status == domain.AssignmentAttended
	if !already && a.Status != domain.AssignmentReserved && a.Status != domain.AssignmentInProgress && a.Status != domain.AssignmentSubmitted {
		return nil, domain.ErrInvalidTransition
	}
	checkIn := s.clock.Now()
	if in.CheckInAt != nil {
		checkIn = in.CheckInAt.UTC()
	} else if a.CheckInAt != nil {
		checkIn = *a.CheckInAt
	}
	var checkOut *time.Time
	if in.CheckOutAt != nil {
		co := in.CheckOutAt.UTC()
		checkOut = &co
	} else if already {
		checkOut = a.CheckOutAt
	}
	if checkOut != nil && checkOut.Before(checkIn) {
		return nil, domain.Invalid("ساعت خروج نمی‌تواند قبل از ساعت ورود باشد")
	}
	a.Status = domain.AssignmentAttended
	a.CheckInAt = &checkIn
	a.CheckOutAt = checkOut
	a.AttendedAt = &checkIn
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	a.Task = t
	return a, nil
}

type AttendanceInput struct {
	CheckInAt  *time.Time
	CheckOutAt *time.Time
}

func (s *Service) MarkAbsent(ctx context.Context, assignmentID uuid.UUID) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AssignmentReserved && a.Status != domain.AssignmentInProgress && a.Status != domain.AssignmentAttended && a.Status != domain.AssignmentSubmitted {
		return nil, domain.ErrInvalidTransition
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	wasOccupied := a.Status.OccupiesSeat()
	a.Status = domain.AssignmentAbsent
	if wasOccupied && t.ReservedCount > 0 {
		t.ReservedCount--
		t.UpdatedAt = s.clock.Now()
		if err := s.tasks.Update(ctx, t); err != nil {
			return nil, err
		}
	}
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	s.notifyVolunteer(ctx, a.VolunteerID, "عدم حضور ثبت شد", "حضور شما در فعالیت «"+t.Title+"» ثبت نشد.")
	a.Task = t
	return a, nil
}

type DeliveryInput struct {
	Note      string
	FileName  string
	ObjectKey string
	Mime      string
}

func (s *Service) SubmitDelivery(ctx context.Context, userID, assignmentID uuid.UUID, in DeliveryInput) (*domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.VolunteerID != v.ID {
		return nil, domain.ErrForbidden
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if !t.IsRemote() {
		return nil, domain.Invalid("ارسال نتیجه فقط برای کارهای دورکار است. حضور در فعالیت حضوری را واحد پشتیبانی ثبت می‌کند")
	}
	if a.Status != domain.AssignmentReserved && a.Status != domain.AssignmentInProgress && a.Status != domain.AssignmentSubmitted && a.Status != domain.AssignmentRevisionRequested {
		return nil, domain.ErrInvalidTransition
	}
	note := strings.TrimSpace(in.Note)
	if note == "" && strings.TrimSpace(in.ObjectKey) == "" && a.DeliveryObjectKey == "" {
		return nil, domain.Invalid("شرح نتیجه یا فایل را ارسال کنید")
	}
	now := s.clock.Now()
	if note != "" {
		a.DeliveryNote = note
	}
	if strings.TrimSpace(in.ObjectKey) != "" {
		a.DeliveryFileName = strings.TrimSpace(in.FileName)
		a.DeliveryObjectKey = strings.TrimSpace(in.ObjectKey)
		a.DeliveryMime = strings.TrimSpace(in.Mime)
	}
	a.DeliveredAt = &now
	a.Status = domain.AssignmentSubmitted
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	s.notifyVolunteer(ctx, a.VolunteerID, "نتیجه فعالیت ثبت شد", "نتیجه «"+t.Title+"» برای بررسی واحد پشتیبانی ارسال شد.")
	if sn, ok := s.notify.(interface {
		NotifyStaff(ctx context.Context, title, body string) error
	}); ok {
		_ = sn.NotifyStaff(ctx, "نتیجه فعالیت ارسال شد", v.FullName+" نتیجه «"+t.Title+"» را ارسال کرد و منتظر بررسی است.")
	}
	a.Task = t
	a.Volunteer = v
	return a, nil
}

func (s *Service) RequestRevision(ctx context.Context, assignmentID uuid.UUID, comment string) (*domain.Assignment, error) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return nil, domain.Invalid("توضیح اصلاح یا تکمیل را بنویسید")
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if !t.IsRemote() {
		return nil, domain.Invalid("درخواست اصلاح فقط برای نتیجه فعالیت دورکار است")
	}
	if a.Status != domain.AssignmentSubmitted {
		return nil, domain.ErrInvalidTransition
	}
	a.Status = domain.AssignmentRevisionRequested
	a.AdminComment = comment
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	s.notifyVolunteer(ctx, a.VolunteerID, "نیاز به اصلاح نتیجه", "واحد پشتیبانی برای فعالیت «"+t.Title+"» درخواست اصلاح یا تکمیل کرده است. "+comment)
	a.Task = t
	return a, nil
}

func (s *Service) Complete(ctx context.Context, assignmentID uuid.UUID, discipline, expertise, ethics int, comment string) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if t.IsRemote() {
		if a.Status != domain.AssignmentSubmitted {
			return nil, domain.Invalid("ابتدا داوطلب باید نتیجه را ارسال کند")
		}
	} else if a.Status != domain.AssignmentAttended {
		return nil, domain.Invalid("ابتدا حضور داوطلب را ثبت کنید")
	}
	score, err := scoring.CompositeScore(discipline, expertise, ethics)
	if err != nil {
		return nil, err
	}
	v, err := s.volunteers.GetByID(ctx, a.VolunteerID)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	a.AdminDiscipline = &discipline
	a.AdminExpertise = &expertise
	a.AdminEthics = &ethics
	a.AdminComment = comment
	a.CompositeScore = &score
	a.HoursAwarded = t.HourWeight
	a.Status = domain.AssignmentCompleted
	a.CompletedAt = &now
	if a.AttendedAt == nil {
		a.AttendedAt = &now
	}
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	scoring.UpdateVolunteerTotals(v, score, t.HourWeight)
	v.UpdatedAt = now
	if err := s.volunteers.Update(ctx, v); err != nil {
		return nil, err
	}
	if s.notify != nil {
		_ = s.notify.Notify(ctx, v.UserID, "فعالیت تکمیل شد", "امتیاز شما ثبت شد. پس از تکمیل می‌توانید تقدیرنامه این فعالیت را درخواست کنید.")
	}
	a.Task = t
	a.Volunteer = v
	return a, nil
}

func (s *Service) RateByVolunteer(ctx context.Context, userID, assignmentID uuid.UUID, rating int, comment string) (*domain.Assignment, error) {
	if rating < 1 || rating > 5 {
		return nil, domain.ErrInvalidInput
	}
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.VolunteerID != v.ID {
		return nil, domain.ErrForbidden
	}
	if a.Status != domain.AssignmentCompleted && a.Status != domain.AssignmentAttended {
		return nil, domain.ErrInvalidTransition
	}
	a.VolunteerRating = &rating
	a.VolunteerComment = strings.TrimSpace(comment)
	return a, s.tasks.UpdateAssignment(ctx, a)
}

func (s *Service) CancelByVolunteer(ctx context.Context, userID, assignmentID uuid.UUID) (*domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if a.VolunteerID != v.ID {
		return nil, domain.ErrForbidden
	}
	return s.Cancel(ctx, assignmentID, false)
}

func (s *Service) Reject(ctx context.Context, assignmentID uuid.UUID, comment string) (*domain.Assignment, error) {
	return s.cancel(ctx, assignmentID, true, comment)
}

func (s *Service) Cancel(ctx context.Context, assignmentID uuid.UUID, byAdmin bool) (*domain.Assignment, error) {
	return s.cancel(ctx, assignmentID, byAdmin, "")
}

func (s *Service) cancel(ctx context.Context, assignmentID uuid.UUID, byAdmin bool, comment string) (*domain.Assignment, error) {
	a, err := s.tasks.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if !a.Status.Cancellable() {
		return nil, domain.ErrInvalidTransition
	}
	wasOccupied := a.Status.OccupiesSeat()
	if byAdmin {
		a.Status = domain.AssignmentRejected
		if c := strings.TrimSpace(comment); c != "" {
			a.AdminComment = c
		}
	} else {
		a.Status = domain.AssignmentCancelled
	}
	t, err := s.tasks.GetByID(ctx, a.TaskID)
	if err != nil {
		return nil, err
	}
	if wasOccupied && t.ReservedCount > 0 {
		t.ReservedCount--
		t.UpdatedAt = s.clock.Now()
		_ = s.tasks.Update(ctx, t)
	}
	if err := s.tasks.UpdateAssignment(ctx, a); err != nil {
		return nil, err
	}
	title, body := "انصراف از فعالیت", "درخواست شما برای «"+t.Title+"» لغو شد."
	if byAdmin {
		title, body = "درخواست فعالیت رد شد", "درخواست شما برای «"+t.Title+"» توسط واحد پشتیبانی رد شد."
		if a.AdminComment != "" {
			body += " دلیل: " + a.AdminComment
		}
	}
	s.notifyVolunteer(ctx, a.VolunteerID, title, body)
	a.Task = t
	return a, nil
}

func (s *Service) ListAssignments(ctx context.Context, f domain.AssignmentFilter) ([]domain.Assignment, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 200 {
		f.Limit = 200
	}
	return s.tasks.ListAssignments(ctx, f)
}

func (s *Service) MyAssignments(ctx context.Context, userID uuid.UUID) ([]domain.Assignment, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items, _, err := s.tasks.ListAssignments(ctx, domain.AssignmentFilter{VolunteerID: v.ID, Limit: 100})
	return items, err
}

func (s *Service) MyTrainings(ctx context.Context, userID uuid.UUID) ([]domain.VolunteerTraining, error) {
	v, err := s.volunteers.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.tasks.ListVolunteerTrainings(ctx, v.ID)
}

func (s *Service) ListVolunteerTrainings(ctx context.Context, volunteerID uuid.UUID) ([]domain.VolunteerTraining, error) {
	return s.tasks.ListVolunteerTrainings(ctx, volunteerID)
}

func (s *Service) CloseExpired(ctx context.Context) error {
	if s.notify != nil {
		if f, ok := s.notify.(interface {
			FireDueReminders(ctx context.Context, now time.Time) error
		}); ok {
			_ = f.FireDueReminders(ctx, s.clock.Now())
		}
	}
	if s.tasks == nil {
		return nil
	}
	_, err := s.tasks.CloseExpired(ctx, s.clock.Now())
	return err
}

func validateTask(in TaskInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return domain.Invalid("عنوان فعالیت را وارد کنید")
	}
	if strings.TrimSpace(in.Description) == "" {
		return domain.Invalid("شرح فعالیت را وارد کنید")
	}
	if err := validateTraining(in); err != nil {
		return err
	}
	if in.StartsAt.IsZero() {
		return domain.Invalid("تاریخ شروع نامعتبر است؛ تاریخ و ساعت شروع را از تقویم انتخاب کنید")
	}
	if in.EndsAt.IsZero() {
		return domain.Invalid("تاریخ پایان نامعتبر است؛ تاریخ و ساعت پایان را از تقویم انتخاب کنید")
	}
	if in.Kind == domain.TaskRecurring {
		if len(in.Slots) == 0 {
			return domain.Invalid("برای فعالیت جاری حداقل یک روز هفته را با ظرفیت انتخاب کنید")
		}
		seen := map[int]struct{}{}
		for _, sl := range in.Slots {
			if sl.Weekday < 0 || sl.Weekday > 6 {
				return domain.Invalid("روز هفته نامعتبر است")
			}
			if _, ok := seen[sl.Weekday]; ok {
				return domain.Invalid("هر روز هفته فقط یک‌بار قابل انتخاب است")
			}
			seen[sl.Weekday] = struct{}{}
			if sl.Capacity < 1 {
				return domain.Invalid("ظرفیت هر روز هفته باید حداقل ۱ نفر باشد")
			}
		}
		if in.HourWeight <= 0 {
			return domain.Invalid("وزن ساعتی باید بزرگ‌تر از صفر باشد")
		}
		_, err := expandOccurrences(in)
		return err
	}
	if !in.EndsAt.After(in.StartsAt) {
		return domain.Invalid("تاریخ پایان باید بعد از تاریخ شروع باشد")
	}
	if in.Capacity < 1 {
		return domain.Invalid("ظرفیت باید حداقل ۱ نفر باشد")
	}
	if in.HourWeight <= 0 {
		return domain.Invalid("وزن ساعتی باید بزرگ‌تر از صفر باشد")
	}
	return nil
}

func validateTraining(in TaskInput) error {
	if !in.RequiresTraining {
		return nil
	}
	if !domain.ValidTrainingKind(in.TrainingKind) {
		return domain.Invalid("نوع آموزش را مشخص کنید")
	}
	if strings.TrimSpace(in.TrainingLocation) == "" {
		return domain.Invalid("محل آموزش را وارد کنید")
	}
	if in.TrainingAt == nil || in.TrainingAt.IsZero() {
		return domain.Invalid("زمان آموزش را مشخص کنید")
	}
	return nil
}

func applyTraining(t *domain.Task, in TaskInput) {
	t.RequiresTraining = in.RequiresTraining
	if !in.RequiresTraining {
		t.TrainingKind = ""
		t.TrainingLocation = ""
		t.TrainingAt = nil
		return
	}
	t.TrainingKind = in.TrainingKind
	t.TrainingLocation = strings.TrimSpace(in.TrainingLocation)
	if in.TrainingAt != nil && !in.TrainingAt.IsZero() {
		at := in.TrainingAt.UTC()
		t.TrainingAt = &at
		return
	}
	t.TrainingAt = nil
}

func trainingKindFa(kind string) string {
	switch kind {
	case domain.TrainingOnline:
		return "آنلاین"
	case domain.TrainingHybrid:
		return "ترکیبی"
	case domain.TrainingWorkshop:
		return "کارگاه"
	case domain.TrainingOther:
		return "سایر"
	default:
		return "حضوری"
	}
}

func trainingDetail(t *domain.Task) string {
	when := "—"
	if t.TrainingAt != nil && !t.TrainingAt.IsZero() {
		when = formatJalaliDateTime(*t.TrainingAt)
	}
	return "نوع آموزش: " + trainingKindFa(t.TrainingKind) + ". محل آموزش: " + t.TrainingLocation + ". زمان آموزش: " + when + "."
}

func parseUUIDs(in []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(in))
	for _, s := range in {
		id, err := uuid.Parse(strings.TrimSpace(s))
		if err != nil || id == uuid.Nil {
			continue
		}
		out = append(out, id)
	}
	return out
}
