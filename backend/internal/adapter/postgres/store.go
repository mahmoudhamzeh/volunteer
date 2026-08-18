package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mahmoudhamzeh/volunteer/backend/internal/domain"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *DB { return &DB{Pool: pool} }

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}

type UserRepo struct{ db *DB }

func (d *DB) Users() *UserRepo { return &UserRepo{d} }

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO users (id, email, phone, password_hash, role, external_user_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8)`,
		u.ID, u.Email, u.Phone, u.PasswordHash, u.Role, u.ExternalUserID, u.CreatedAt, u.UpdatedAt)
	return mapErr(err)
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, userCols+` WHERE id=$1`, id))
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, userCols+` WHERE email=$1`, email))
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, userCols+` WHERE phone=$1 AND phone <> ''`, phone))
}

func (r *UserRepo) GetByExternalID(ctx context.Context, externalID string) (*domain.User, error) {
	return scanUser(r.db.Pool.QueryRow(ctx, userCols+` WHERE external_user_id=$1`, externalID))
}

const userCols = `SELECT id,email,COALESCE(phone,''),password_hash,role,COALESCE(external_user_id,''),created_at,updated_at FROM users`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.ExternalUserID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}
