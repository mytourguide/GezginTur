package service

import "golang.org/x/crypto/bcrypt"

// HashPassword, bcrypt ile sifre hashler.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword, duz metin sifreyi hash ile karsilastirir.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
