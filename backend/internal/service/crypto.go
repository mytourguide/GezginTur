package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// CryptoService, KVKK geregi kimlik/pasaport numarasi gibi hassas verileri
// AES-256-GCM ile sifreler. Anahtar 32 byte olmalidir (base64 ile verilir).
type CryptoService struct {
	gcm cipher.AEAD
}

func NewCryptoService(keyB64 string) *CryptoService {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil || len(key) != 32 {
		// Gelistirme kolayligi icin bos anahtarda sabit gelistirme anahtari kullanilir.
		// URETIM ortaminda DATA_ENCRYPTION_KEY mutlaka 32 byte'lik rastgele base64 olmalidir.
		key = make([]byte, 32)
		copy(key, []byte("dev-encryption-key-degistir!"))
	}
	block, err := aes.NewCipher(key)
	if err != nil { panic(err) }
	gcm, err := cipher.NewGCM(block)
	if err != nil { panic(err) }
	return &CryptoService{gcm: gcm}
}

// Encrypt, duz metni base64(nonce||ciphertext) formatinda sifreler.
func (c *CryptoService) Encrypt(plain string) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil { return "", err }
	sealed := c.gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt, Encrypt ciktisini cozer.
func (c *CryptoService) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil { return "", err }
	if len(data) < c.gcm.NonceSize() { return "", errors.New("sifreli veri bozuk") }
	nonce, ct := data[:c.gcm.NonceSize()], data[c.gcm.NonceSize():]
	plain, err := c.gcm.Open(nil, nonce, ct, nil)
	if err != nil { return "", err }
	return string(plain), nil
}
