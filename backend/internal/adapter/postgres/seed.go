package postgres

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/authuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/missionuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/taskuc"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/usecase/volunteeruc"
	"golang.org/x/crypto/bcrypt"
)

func Demo(ctx context.Context, users domain.UserRepository, volunteers domain.VolunteerRepository, tasks *taskuc.Service, missions *missionuc.Service, vol *volunteeruc.Service, auth *authuc.Service) {
	if _, err := users.GetByEmail(ctx, "admin@mahak.ir"); err == nil {
		return
	}
	log.Println("seeding demo accounts and sample data")

	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	now := time.Now().UTC()
	admin := &domain.User{
		ID: uuid.New(), Email: "admin@mahak.ir", PasswordHash: string(hash),
		Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now,
	}
	if err := users.Create(ctx, admin); err != nil {
		log.Println("seed admin:", err)
		return
	}

	_, _, err := auth.Register(ctx, "volunteer@mahak.ir", "Volunteer@123", "سارا محمدی", domain.RoleVolunteer)
	if err != nil {
		log.Println("seed volunteer:", err)
		return
	}
	u, _ := users.GetByEmail(ctx, "volunteer@mahak.ir")
	v, _ := volunteers.GetByUserID(ctx, u.ID)
	v.NationalID = "0012345678"
	v.Phone = "09121234567"
	v.City = "تهران"
	v.Bio = "طراح گرافیک و علاقه‌مند به فعالیت‌های حمایتی کودکان"
	v.SkillCategories = []domain.SkillCategory{domain.SkillArtistic, domain.SkillAdministrative}
	v.EducationField = "گرافیک"
	v.Status = domain.StatusApproved
	v.UpdatedAt = now
	_ = volunteers.Update(ctx, v)

	pendingHash, _ := bcrypt.GenerateFromPassword([]byte("Volunteer@123"), bcrypt.DefaultCost)
	pendingUser := &domain.User{ID: uuid.New(), Email: "pending@mahak.ir", PasswordHash: string(pendingHash), Role: domain.RoleVolunteer, CreatedAt: now, UpdatedAt: now}
	_ = users.Create(ctx, pendingUser)
	_ = volunteers.Create(ctx, &domain.Volunteer{
		ID: uuid.New(), UserID: pendingUser.ID, FullName: "علی رضایی", NationalID: "0023456789",
		Phone: "09351234567", City: "اصفهان", SkillCategories: []domain.SkillCategory{domain.SkillMedical},
		EducationField: "پزشکی", MedicalLicense: "12345", Status: domain.StatusPending, CreatedAt: now, UpdatedAt: now,
	})

	_, _ = tasks.Create(ctx, admin.ID, taskuc.TaskInput{
		Title: "طراحی پوستر هفته حمایت از کودک", Description: "طراحی پوستر دیجیتال برای کمپین جذب کمک‌های مردمی محک.",
		Location: "دورکاری / دفتر مرکزی محک", StartsAt: now.Add(24 * time.Hour), EndsAt: now.Add(72 * time.Hour),
		Capacity: 4, HourWeight: 6, RequiredSkills: []string{"artistic"}, MinScore: 0, RequiredEducation: "گرافیک",
	})
	_, _ = tasks.Create(ctx, admin.ID, taskuc.TaskInput{
		Title: "همراهی کودکان در بخش بستری", Description: "حضور در بیمارستان و همراهی بازی و قصه برای کودکان بستری.",
		Location: "بیمارستان محک، تجریش", StartsAt: now.Add(48 * time.Hour), EndsAt: now.Add(52 * time.Hour),
		Capacity: 6, HourWeight: 4, RequiredSkills: []string{"psychological", "education"},
	})
	_, _ = tasks.Create(ctx, admin.ID, taskuc.TaskInput{
		Title: "پشتیبانی اداری بایگانی مدارک", Description: "مرتب‌سازی پرونده‌های داوطلبی و ورود داده.",
		Location: "دفتر مرکزی", StartsAt: now.Add(12 * time.Hour), EndsAt: now.Add(20 * time.Hour),
		Capacity: 3, HourWeight: 3, RequiredSkills: []string{"administrative"},
	})

	h72 := 72
	_, _ = missions.Create(ctx, missionuc.MissionInput{
		Title: "تکمیل پروفایل داوطلبی", Description: "پروفایل، مهارت‌ها و تقویم زمانی را کامل کنید.",
		Kind: domain.MissionCompleteProfile, HourWeight: 1, TargetCount: 1,
	})
	_, _ = missions.Create(ctx, missionuc.MissionInput{
		Title: "دعوت از ۵ کاربر جدید", Description: "۵ نفر را به اپلیکیشن محک دعوت کنید.",
		Kind: domain.MissionInviteUsers, HourWeight: 2, DeadlineHours: &h72, TargetCount: 5, WebhookEvent: "user.invited",
	})
	_ = vol
}
