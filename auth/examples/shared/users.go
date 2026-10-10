package shared

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/pgutil"
)

var (
	ErrUserNotFound = errors.NotFoundError("user not found")
	ErrEmailTaken   = errors.NewResultErrorWithStatus("email already registered", "CONFLICT", http.StatusConflict)
)

type User struct {
	Id           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Active       bool      `json:"active"`
	Role         string    `json:"role"`
}

type Users struct {
	pool *pgxpool.Pool
}

func NewUsers(pool *pgxpool.Pool) Users {
	return Users{pool: pool}
}

const userColumns = `id, email, password_hash, active, role`

func (u Users) Create(ctx context.Context, email, passwordHash string, active bool) (User, error) {
	user, err := scanUser(u.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, active) VALUES ($1, $2, $3) RETURNING `+userColumns,
		email, passwordHash, active))
	if pgutil.IsUniqueViolation(err) {
		return User{}, ErrEmailTaken
	}

	return user, err
}

// passwordless logins create the account on first use
func (u Users) FindOrCreate(ctx context.Context, email string) (User, error) {
	return scanUser(u.pool.QueryRow(ctx,
		`INSERT INTO users (email) VALUES ($1)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING `+userColumns, email))
}

func (u Users) ByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(u.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
}

func (u Users) ById(ctx context.Context, id uuid.UUID) (User, error) {
	return scanUser(u.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (u Users) SetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := u.pool.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, id, passwordHash)
	return err
}

func (u Users) Activate(ctx context.Context, email string) error {
	_, err := u.pool.Exec(ctx, `UPDATE users SET active = true WHERE email = $1`, email)
	return err
}

func scanUser(row pgx.Row) (User, error) {
	var user User
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}

	return user, err
}
