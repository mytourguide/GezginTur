package service

// Google ve Apple "ID token" (JWT, RS256) imzasini issuer'un JWKS'ine karsi dogrular.
// Dis bagimlilik yok: sadece net/http + crypto/rsa ile uygulanir.

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type oauthClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Sub           string `json:"sub"`
	Iss           string `json:"iss"`
	Aud           string `json:"aud"`
	Exp           int64  `json:"exp"`
}

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Use string `json:"use"`
}

type jwksCache struct {
	mu    sync.Mutex
	keys  []jwksKey
	fetch time.Time
	url   string
}

var caches = map[string]*jwksCache{}

func loadJWKS(url string) ([]jwksKey, error) {
	cachesMu.Lock()
	c, ok := caches[url]
	if !ok {
		c = &jwksCache{url: url}
		caches[url] = c
	}
	cachesMu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.fetch) < time.Hour && len(c.keys) > 0 {
		return c.keys, nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var doc struct {
		Keys []jwksKey `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	if len(doc.Keys) == 0 {
		return nil, errors.New("bos jwks")
	}
	c.keys, c.fetch = doc.Keys, time.Now()
	return c.keys, nil
}

var cachesMu sync.Mutex

func b64urlBig(s string) (*big.Int, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(b), nil
}

// verifyJWT: RS256 imzali JWT'yi dogrular; kid eslesen anahtarla.
func verifyJWT(jwksURL, rawToken string) (*oauthClaims, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("gecersiz token bicimi")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil || header.Alg != "RS256" {
		return nil, errors.New("desteklenmeyen token algoritmasi")
	}

	keys, err := loadJWKS(jwksURL)
	if err != nil {
		return nil, err
	}
	var selected *jwksKey
	for i := range keys {
		if keys[i].Kid == header.Kid {
			selected = &keys[i]
			break
		}
	}
	if selected == nil {
		// Anahtar donmus olabilir: zorla tazele
		clearCache(jwksURL)
		keys, err = loadJWKS(jwksURL)
		if err != nil {
			return nil, err
		}
		for i := range keys {
			if keys[i].Kid == header.Kid {
				selected = &keys[i]
				break
			}
		}
		if selected == nil {
			return nil, errors.New("eslesen anahtar bulunamadi")
		}
	}

	n, err := b64urlBig(selected.N)
	if err != nil {
		return nil, err
	}
	eb, err := b64urlBig(selected.E)
	if err != nil {
		return nil, err
	}
	pub := &rsa.PublicKey{N: n, E: int(eb.Int64())}

	signed := []byte(parts[0] + "." + parts[1])
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	// RS256 = PKCS#1 v1.5 with SHA-256
	sum := sha256.Sum256(signed)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
		return nil, fmt.Errorf("imza dogrulanamadi: %w", err)
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims oauthClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

func clearCache(url string) {
	cachesMu.Lock()
	if c, ok := caches[url]; ok {
		c.mu.Lock()
		c.keys = nil
		c.mu.Unlock()
	}
	cachesMu.Unlock()
}

// --- Google ---

type GoogleIdentity struct {
	Email, Name, Sub string
}

// VerifyGoogleIDToken, Google Identity Services'in dondurdugu id_token'i dogrular.
func VerifyGoogleIDToken(idToken, clientID string) (*GoogleIdentity, error) {
	if clientID == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID tanimli degil")
	}
	claims, err := verifyJWT("https://www.googleapis.com/oauth2/v3/certs", idToken)
	if err != nil {
		return nil, err
	}
	if claims.Aud != clientID {
		return nil, errors.New("google token kitle uyusmadi")
	}
	if claims.Iss != "https://accounts.google.com" && claims.Iss != "accounts.google.com" {
		return nil, errors.New("gecersiz saglayici (iss)")
	}
	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("token suresi dolmus")
	}
	if claims.Email == "" {
		return nil, errors.New("token icinde e-posta yok")
	}
	return &GoogleIdentity{Email: strings.ToLower(claims.Email), Name: claims.Name, Sub: claims.Sub}, nil
}

// --- Apple ---

type AppleIdentity struct {
	Email, Sub string
}

// VerifyAppleIDToken, Sign in with Apple'in dondurdugu id_token'i dogrular.
func VerifyAppleIDToken(idToken, serviceID string) (*AppleIdentity, error) {
	if serviceID == "" {
		return nil, errors.New("APPLE_SERVICE_ID tanimli degil")
	}
	claims, err := verifyJWT("https://appleid.apple.com/auth/keys", idToken)
	if err != nil {
		return nil, err
	}
	if claims.Aud != serviceID {
		return nil, errors.New("apple token kitle uyusmadi")
	}
	if claims.Iss != "https://appleid.apple.com" {
		return nil, errors.New("gecersiz saglayici (iss)")
	}
	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("token suresi dolmus")
	}
	if claims.Email == "" {
		return nil, errors.New("apple token icinde e-posta yok (ilk giris disinda donmeyebilir)")
	}
	return &AppleIdentity{Email: strings.ToLower(claims.Email), Sub: claims.Sub}, nil
}
