package handler

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"seyahat/backend/internal/models"
	"seyahat/backend/internal/service"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Register, yeni musteri kaydi olusturur ve token cifti dondurur.
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := DecodeBody(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz istek govdesi")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.FullName == "" || !emailRe.MatchString(req.Email) || len(req.Password) < 8 {
		ErrorJSON(w, http.StatusBadRequest, "ad, gecerli e-posta ve en az 8 karakterli sifre zorunludur")
		return
	}
	if exists, _ := a.Users.FindByEmail(r.Context(), req.Email); exists != nil {
		ErrorJSON(w, http.StatusConflict, "bu e-posta ile kayitli bir hesap var")
		return
	}
	hash, err := service.HashPassword(req.Password)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "sifre islenemedi")
		return
	}
	u := &models.User{FullName: req.FullName, Email: req.Email, Phone: req.Phone, PasswordHash: hash, Role: "customer"}
	if err := a.Users.Create(r.Context(), u); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kayit olusturulamadi")
		return
	}
	a.issueTokens(w, r, u)
}

// Login, e-posta + sifre ile giris yapar.
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := DecodeBody(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz istek govdesi")
		return
	}
	u, err := a.Users.FindByEmail(r.Context(), strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil || u == nil || !service.CheckPassword(u.PasswordHash, req.Password) {
		ErrorJSON(w, http.StatusUnauthorized, "e-posta veya sifre hatali")
		return
	}
	// 2FA (TOTP) aciksa tek kullanimlik kod zorunlu
	if u.TotpEnabled {
		if !service.VerifyTOTP(u.TotpSecret, req.TotpCode) {
			ErrorJSON(w, http.StatusUnauthorized, "iki asamali dogrulama kodu gerekli veya hatali")
			return
		}
	}
	a.issueTokens(w, r, u)
}

// issueTokens, token cifti uretir; refresh token hash'i veritabanina yazilir.
func (a *API) issueTokens(w http.ResponseWriter, r *http.Request, u *models.User) {
	pair, hash, err := a.Tokens.GeneratePair(u)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "token uretilemedi")
		return
	}
	expires := time.Now().Add(7 * 24 * time.Hour)
	if err := a.Users.SaveRefreshToken(r.Context(), u.ID, hash, expires); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "oturum kaydedilemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]interface{}{
		"tokens": pair,
		"user":   u,
	})
}

// Refresh, refresh token ile yeni access token uretir (rotation uygulanir).
func (a *API) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
		UserID       string `json:"user_id"`
	}
	if err := DecodeBody(r, &req); err != nil || req.RefreshToken == "" || req.UserID == "" {
		ErrorJSON(w, http.StatusBadRequest, "refresh token ve user_id zorunlu")
		return
	}
	hash := service.HashRefreshToken(req.RefreshToken)
	ok, err := a.Users.RefreshTokenExists(r.Context(), req.UserID, hash)
	if err != nil || !ok {
		ErrorJSON(w, http.StatusUnauthorized, "oturum suresi dolmus, lutfen tekrar giris yapin")
		return
	}
	u, err := a.Users.FindByID(r.Context(), req.UserID)
	if err != nil || u == nil {
		ErrorJSON(w, http.StatusUnauthorized, "kullanici bulunamadi")
		return
	}
	// Rotation: eski refresh token iptal edilir, yeni cift uretilir.
	_ = a.Users.RevokeRefreshToken(r.Context(), req.UserID, hash)
	a.issueTokens(w, r, u)
}

// Logout, refresh token'i iptal eder.
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
		UserID       string `json:"user_id"`
	}
	if err := DecodeBody(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz istek")
		return
	}
	_ = a.Users.RevokeRefreshToken(r.Context(), req.UserID, service.HashRefreshToken(req.RefreshToken))
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me, giris yapmis kullanicinin profilini dondurur.
func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	u, err := a.Users.FindByID(r.Context(), claims.UserID)
	if err != nil || u == nil {
		ErrorJSON(w, http.StatusNotFound, "kullanici bulunamadi")
		return
	}
	JSON(w, http.StatusOK, u)
}

// UpdateProfile, musterinin ad/telefon bilgisini gunceller.
func (a *API) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}
	if err := DecodeBody(r, &req); err != nil || req.FullName == "" {
		ErrorJSON(w, http.StatusBadRequest, "ad zorunlu")
		return
	}
	if err := a.Users.UpdateProfile(r.Context(), claims.UserID, req.FullName, req.Phone); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "profil guncellenemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Sosyal giris (Google / Apple) ----------

// socialLogin, dogrulanmis dis kimlikle (email) kullanici bulur veya olusturur, token verir.
func (a *API) socialLogin(w http.ResponseWriter, r *http.Request, email, name string) {
	u, err := a.Users.FindByEmail(r.Context(), email)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kullanici sorgulanamadi")
		return
	}
	if u == nil {
		// Otomatik hesap olustur: parola kullanilmaz (rasgele hash)
		rnd := make([]byte, 16)
		if _, err := rand.Read(rnd); err != nil {
			ErrorJSON(w, http.StatusInternalServerError, "hesap olusturulamadi")
			return
		}
		hash, _ := service.HashPassword(fmt.Sprintf("social-%x", rnd))
		if name == "" {
			name = strings.Split(email, "@")[0]
		}
		u = &models.User{FullName: name, Email: email, PasswordHash: hash, Role: "customer"}
		if err := a.Users.Create(r.Context(), u); err != nil {
			ErrorJSON(w, http.StatusInternalServerError, "hesap olusturulamadi")
			return
		}
	}
	a.issueTokens(w, r, u)
}

// OAuthGoogle, Google Identity Services'in verdigi id_token'i dogrular ve giris yapar.
func (a *API) OAuthGoogle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"id_token"`
	}
	if err := DecodeBody(r, &req); err != nil || req.IDToken == "" {
		ErrorJSON(w, http.StatusBadRequest, "id_token zorunlu")
		return
	}
	id, err := service.VerifyGoogleIDToken(req.IDToken, a.Cfg.GoogleClientID)
	if err != nil {
		ErrorJSON(w, http.StatusUnauthorized, "google dogrulamasi basarisiz: "+err.Error())
		return
	}
	a.socialLogin(w, r, id.Email, id.Name)
}

// OAuthApple, Sign in with Apple'in verdigi id_token'i dogrular ve giris yapar.
func (a *API) OAuthApple(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"id_token"`
		Name    string `json:"name"` // Apple ad yalnizca ilk giris islev cevabinda doner
	}
	if err := DecodeBody(r, &req); err != nil || req.IDToken == "" {
		ErrorJSON(w, http.StatusBadRequest, "id_token zorunlu")
		return
	}
	id, err := service.VerifyAppleIDToken(req.IDToken, a.Cfg.AppleServiceID)
	if err != nil {
		ErrorJSON(w, http.StatusUnauthorized, "apple dogrulamasi basarisiz: "+err.Error())
		return
	}
	a.socialLogin(w, r, id.Email, req.Name)
}

// ---------- Iki asamali dogrulama (TOTP) ----------

// Setup2FA, yeni TOTP gizli anahtari uretir ve otpauth:// URL'i dondurur.
// Kullanici bu URL'i Google Authenticator'a taratip Enable2FA ile dogrular.
func (a *API) Setup2FA(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	secret, err := service.GenerateTOTPSecret()
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "gizli anahtar uretilemedi")
		return
	}
	if err := a.Users.SetTOTPSecret(r.Context(), claims.UserID, secret); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "gizli anahtar kaydedilemedi")
		return
	}
	u, _ := a.Users.FindByID(r.Context(), claims.UserID)
	account := claims.UserID
	if u != nil {
		account = u.Email
	}
	JSON(w, http.StatusOK, map[string]string{
		"secret":  secret,
		"otp_url": service.TOTPUrlFor("GezginTur", account, secret),
	})
}

// Enable2FA, Authenticator'daki 6 haneli kodu dogrulayip 2FA'yi aktif eder.
func (a *API) Enable2FA(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	var req struct {
		Code string `json:"code"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Code == "" {
		ErrorJSON(w, http.StatusBadRequest, "dogrulama kodu zorunlu")
		return
	}
	secret, err := a.Users.GetTOTPSecret(r.Context(), claims.UserID)
	if err != nil || secret == "" {
		ErrorJSON(w, http.StatusBadRequest, "once 2FA kurulumu yapin (/auth/2fa/setup)")
		return
	}
	if !service.VerifyTOTP(secret, req.Code) {
		ErrorJSON(w, http.StatusBadRequest, "kod hatali, lutfen Authenticator'daki guncel kodu girin")
		return
	}
	if err := a.Users.SetTOTPEnabled(r.Context(), claims.UserID, true); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "2FA aktif edilemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

// Disable2FA, 2FA'yi kapatir ve gizli anahtari siler.
func (a *API) Disable2FA(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	if err := a.Users.SetTOTPEnabled(r.Context(), claims.UserID, false); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "2FA kapatilamadi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

// TOTPStatus, 2FA'nin acik olup olmadigini dondurur.
func (a *API) TOTPStatus(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	enabled, err := a.Users.GetTOTPEnabled(r.Context(), claims.UserID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "durum okunamadi")
		return
	}
	JSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}
