package authuc

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users      domain.UserRepository
	volunteers domain.VolunteerRepository
	secret     []byte
	ttl        time.Duration
}

func New(users domain.UserRepository, volunteers domain.VolunteerRepository, secret string, ttl time.Duration) *Service {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &Service{users: users, volunteers: volunteers, secret: []byte(secret), ttl: ttl}
}

type Claims struct {
	UserID uuid.UUID    `json:"uid"`
	Role   domain.Role  `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) Register(ctx context.Context, email, password, fullName string, role domain.Role) (*domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(password) < 8 || strings.TrimSpace(fullName) == "" {
		return nil, "", domain.ErrInvalidInput
	}
	if role == "" {
		role = domain.RoleVolunteer
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return nil, "", domain.ErrConflict
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	now := time.Now().UTC()
	u := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err
	}
	if role == domain.RoleVolunteer {
		v := &domain.Volunteer{
			ID:              uuid.New(),
			UserID:          u.ID,
			FullName:        strings.TrimSpace(fullName),
			Status:          domain.StatusDraft,
			SkillCategories: []domain.SkillCategory{},
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.volunteers.Create(ctx, v); err != nil {
			return nil, "", err
		}
	}
	token, err := s.issue(u)
	return u, token, err
}

func (s *Service) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", domain.ErrUnauthorized
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, "", domain.ErrUnauthorized
	}
	token, err := s.issue(u)
	return u, token, err
}

func (s *Service) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, domain.ErrUnauthorized
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

// FromExternalToken maps a trusted upstream Auth token subject to a local user.
// For the MAHAK ecosystem the gateway validates the token; this service accepts
// a pre-extracted user id (sub) and role claim.
func (s *Service) UpsertFromExternal(ctx context.Context, externalID, email, fullName string, role domain.Role) (*domain.User, string, error) {
	if externalID == "" {
		return nil, "", domain.ErrInvalidInput
	}
	u, err := s.users.GetByExternalID(ctx, externalID)
	now := time.Now().UTC()
	if err == domain.ErrNotFound {
		if role == "" {
			role = domain.RoleVolunteer
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
		u = &domain.User{
			ID:             uuid.New(),
			Email:          strings.ToLower(email),
			PasswordHash:   string(hash),
			Role:           role,
			ExternalUserID: externalID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.users.Create(ctx, u); err != nil {
			return nil, "", err
		}
		if role == domain.RoleVolunteer {
			_ = s.volunteers.Create(ctx, &domain.Volunteer{
				ID: uuid.New(), UserID: u.ID, FullName: fullName,
				Status: domain.StatusDraft, CreatedAt: now, UpdatedAt: now,
			})
		}
	} else if err != nil {
		return nil, "", err
	}
	token, err := s.issue(u)
	return u, token, err
}

func (s *Service) issue(u *domain.User) (string, error) {
	claims := Claims{
		UserID: u.ID,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}
