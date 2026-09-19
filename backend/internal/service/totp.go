package service

// TOTP (RFC 6238) zaman bazli tek kullanimlik sifre uretimi/dogrulamasi.
// Harici kutuphane gerektirmez: HMAC-SHA1 + dinamik kesme.
// Google Authenticator, Authy vb. tum standart uygulamalarla uyumludur.

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const totpPeriod = 30 // saniye

// GenerateTOTPSecret, 20 byte'lik rastgele base32 (kucuk harf, pad'siz) gizli anahtar uretir.
func GenerateTOTPSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)), nil
}

// TOTPUrlFor, Google Authenticator'a taratilacak otpauth:// URL'i uretir.
func TOTPUrlFor(issuer, account, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&period=%d&digits=6",
		issuer, account, strings.ToUpper(secret), issuer, totpPeriod)
}

// hotpValue, RFC 4226 dinamik kesme ile 6 haneli kod uretir.
func hotpValue(secret string, counter uint64) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", err
	}
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(msg)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff) % 1000000
	return fmt.Sprintf("%06d", code), nil
}

// VerifyTOTP, verilen 6 haneli kodu dogrular; ±1 zaman penceresi toleransli (saat kaymasi).
func VerifyTOTP(secret, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	counter := uint64(time.Now().Unix() / totpPeriod)
	for _, c := range []uint64{counter - 1, counter, counter + 1} {
		expected, err := hotpValue(secret, c)
		if err == nil && hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}
