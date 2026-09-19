package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"seyahat/backend/internal/models"
)

type UserRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) *UserRepository { return &UserRepository{pool} }

// QueryRow, saglik kontrolleri gibi tekil raw SQL okumalarinda kullanilir.
func (r *UserRepository) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return r.pool.QueryRow(ctx, sql, args...)
}

// Create, yeni kullanici kaydi olusturur.
func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (full_name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		u.FullName, u.Email, u.Phone, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt)
}

// UpsertAdmin, seed icin admin kullanicisini yoksa olusturur / varsa gunceller.
func (r *UserRepository) UpsertAdmin(ctx context.Context, fullName, email, hash string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (full_name, email, password_hash, role)
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = 'admin'
		RETURNING id`, fullName, email, hash).Scan(&id)
	return id, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, full_name, email, COALESCE(phone,''), password_hash, role,
		       COALESCE(totp_secret,''), totp_enabled, created_at
		FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Phone, &u.PasswordHash, &u.Role,
			&u.TotpSecret, &u.TotpEnabled, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, full_name, email, COALESCE(phone,''), password_hash, role,
		       COALESCE(totp_secret,''), totp_enabled, created_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Phone, &u.PasswordHash, &u.Role,
			&u.TotpSecret, &u.TotpEnabled, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateProfile, musterinin ad/telefon bilgisini gunceller.
func (r *UserRepository) UpdateProfile(ctx context.Context, id, fullName, phone string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET full_name=$1, phone=$2, updated_at=now() WHERE id=$3`,
		fullName, phone, id)
	return err
}

// List, admin panel icin kullanici listesi dondurur.
func (r *UserRepository) List(ctx context.Context, search string) ([]models.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, full_name, email, COALESCE(phone,''), role, created_at
		FROM users
		WHERE ($1 = '' OR full_name ILIKE '%'||$1||'%' OR email ILIKE '%'||$1||'%')
		ORDER BY created_at DESC LIMIT 500`, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.User, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Phone, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// --- TOTP (2FA) ---

// SetTOTPSecret, kurulum asamasinda uretilen gizli anahtari saklar.
func (r *UserRepository) SetTOTPSecret(ctx context.Context, id, secret string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET totp_secret=$1 WHERE id=$2`, secret, id)
	return err
}

// GetTOTPSecret, kullanicinin TOTP gizli anahtarini dondurur.
func (r *UserRepository) GetTOTPSecret(ctx context.Context, id string) (string, error) {
	var secret string
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(totp_secret,'') FROM users WHERE id=$1`, id).Scan(&secret)
	return secret, err
}

// SetTOTPEnabled, 2FA'yi acar/kapatir (kapatirken anahtar da temizlenir).
func (r *UserRepository) SetTOTPEnabled(ctx context.Context, id string, enabled bool) error {
	if enabled {
		_, err := r.pool.Exec(ctx, `UPDATE users SET totp_enabled=true WHERE id=$1`, id)
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE users SET totp_enabled=false, totp_secret=NULL WHERE id=$1`, id)
	return err
}

// GetTOTPEnabled, kullanicinin 2FA durumunu dondurur.
func (r *UserRepository) GetTOTPEnabled(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `SELECT totp_enabled FROM users WHERE id=$1`, id).Scan(&ok)
	return ok, err
}

// --- Refresh token saklama (token'in kendisi degil, SHA-256 hash'i saklanir) ---

func (r *UserRepository) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt interface{}) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return err
}

func (r *UserRepository) RefreshTokenExists(ctx context.Context, userID, tokenHash string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE user_id=$1 AND token_hash=$2 AND expires_at > now() AND revoked = false)`,
		userID, tokenHash).Scan(&ok)
	return ok, err
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, userID, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked = true WHERE user_id=$1 AND token_hash=$2`, userID, tokenHash)
	return err
}
