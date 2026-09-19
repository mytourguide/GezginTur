package handler

import (
	"net/http"
	"strconv"

	"seyahat/backend/internal/repository"
)

// ListTours, filtrelerle tur listesi dondurur (musteri tarafi).
func (a *API) ListTours(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.TourFilters{
		CategorySlug: q.Get("category"),
		Search:       q.Get("search"),
		Sort:         q.Get("sort"),
		Featured:     q.Get("featured") == "true",
	}
	if v := q.Get("min_price"); v != "" { if n, err := strconv.ParseFloat(v, 64); err == nil { f.MinPrice = &n } }
	if v := q.Get("max_price"); v != "" { if n, err := strconv.ParseFloat(v, 64); err == nil { f.MaxPrice = &n } }
	if v := q.Get("duration_max"); v != "" { if n, err := strconv.Atoi(v); err == nil { f.DurationMax = &n } }
	f.StartDate = q.Get("start_date")
	f.EndDate = q.Get("end_date")
	f.Page, _ = strconv.Atoi(q.Get("page"))
	if f.Page <= 0 { f.Page = 1 }
	f.Limit, _ = strconv.Atoi(q.Get("limit"))
	if f.Limit <= 0 { f.Limit = 12 }

	tours, total, err := a.Tours.List(r.Context(), f)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "turlar listelenemedi")
		return
	}
	if q.Get("lang") == "en" { // dil destegi: en cevirileri ana alanlara uygula
		for i := range tours {
			tours[i].Localize()
		}
	}
	JSON(w, http.StatusOK, map[string]interface{}{
		"tours": tours, "total": total, "page": f.Page, "limit": f.Limit,
	})
}

// GetTour, slug ile tur detayini dondurur (SEO icin SSR/SSG sayfa besler).
func (a *API) GetTour(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tour, err := a.Tours.GetBySlug(r.Context(), slug)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "tur yuklenemedi")
		return
	}
	if tour == nil {
		ErrorJSON(w, http.StatusNotFound, "tur bulunamadi")
		return
	}
	if r.URL.Query().Get("lang") == "en" {
		tour.Localize()
	}
	JSON(w, http.StatusOK, tour)
}

// ListCategories, kategori listesi dondurur.
func (a *API) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := a.Tours.Categories(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kategoriler yuklenemedi")
		return
	}
	if r.URL.Query().Get("lang") == "en" {
		for i := range cats {
			cats[i].Localize()
		}
	}
	JSON(w, http.StatusOK, cats)
}
