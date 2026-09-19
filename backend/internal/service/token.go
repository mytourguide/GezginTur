package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"seyahat/backend/internal/models"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// TokenService, JWT access + refresh token uretimi ve dogrulamasindan sorumludur.
type TokenService struct{ secret []byte }

func NewTokenService(secret string) *TokenService { return &TokenService{secret: []byte(secret)} }

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GeneratePair, kullanici icin access + refresh token cifti uretir.
func (s *TokenService) GeneratePair(u *models.User) (*models.TokenPair, string, error) {
	now := time.Now()

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: u.ID, Role: u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   u.ID,
		},
	})
	accessStr, err := access.SignedString(s.secret)
	if err != nil { return nil, "", err }

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil { return nil, "", err }
	refreshStr := base64.URLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(refreshStr))

	return &models.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int64(AccessTokenTTL.Seconds()),
	}, hex.EncodeToString(sum[:]), nil
}

// ParseAccessToken, access token'i dogrular ve claim'leri dondurur.
func (s *TokenService) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("beklenmeyen imza yontemi")
		}
		return s.secret, nil
	})
	if err != nil { return nil, err }
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid { return nil, errors.New("gecersiz token") }
	return claims, nil
}

// HashRefreshToken, saklama/dogrulama icin refresh token'in SHA-256 hash'ini uretir.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
