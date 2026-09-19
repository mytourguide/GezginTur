package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// SignUpload, istemci tarafindan S3'e dogrudan yukleme icin imzali PUT URL uretir.
// Donen public_url, tur gorseli olarak admin formuna kaydedilir.
func (a *API) SignUpload(w http.ResponseWriter, r *http.Request) {
	if !a.Uploads.Enabled() {
		ErrorJSON(w, http.StatusServiceUnavailable, "dosya yukleme yapilandirilmamis (S3 ayarlari eksik)")
		return
	}
	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Filename == "" {
		ErrorJSON(w, http.StatusBadRequest, "filename zorunlu")
		return
	}
	if req.ContentType == "" {
		req.ContentType = "application/octet-stream"
	}
	// Guvenlik: yalnizca gorsel MIME turlerine izin ver
	allowed := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/webp": true, "image/gif": true,
	}
	if !allowed[strings.ToLower(req.ContentType)] {
		ErrorJSON(w, http.StatusBadRequest, "yalnizca JPEG/PNG/WebP/GIF yuklenebilir")
		return
	}

	// Anahtar: uploads/YYYYMM/rastgele-ozelik
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "anahtar uretilemedi")
		return
	}
	ext := strings.ToLower(filepath.Ext(req.Filename))
	key := fmt.Sprintf("uploads/%s/%s%s", time.Now().Format("200601"), hex.EncodeToString(raw), ext)

	url, err := a.Uploads.SignPut(r.Context(), key, req.ContentType)
	if err != nil {
		ErrorJSON(w, http.StatusBadGateway, "imzali URL uretilemedi: "+err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{
		"upload_url": url,
		"public_url": a.Uploads.PublicURLFor(key),
		"key":        key,
	})
}
