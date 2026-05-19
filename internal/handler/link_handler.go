package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Siargio/shortener/internal/service"
)

// LinkHandler — структура, содержащая HTTP-обработчики для работы со ссылками.
// Внедряет зависимость от сервиса (dependency injection).
type LinkHandler struct {
	service *service.LinkService
}

// NewLinkHandler — конструктор обработчика.
// Принимает сервис и возвращает готовый обработчик.
func NewLinkHandler(service *service.LinkService) *LinkHandler {
	return &LinkHandler{service: service}
}

// ShortenRequest — DTO для запроса на создание короткой ссылки.
// Тег `json:"url"` указывает, как поле называется в JSON-запросе.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse — DTO для ответа на создание короткой ссылки.
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// StatsResponse — DTO для ответа со статистикой по ссылке.
type StatsResponse struct {
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url"`
	Clicks    int    `json:"clicks"`
}

// Shorten — HTTP-обработчик для POST /shorten
// Создаёт короткую ссылку из длинной.
func (h *LinkHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	// 1. Декодируем JSON-тело запроса
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// 2. Валидация: ссылка не должна быть пустой
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}
	// 3. Вызываем бизнес-логику (сервис)
	shortURL, err := h.service.Shorten(r.Context(), req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to shorten url")
		return
	}
	// 4. Отдаём успешный ответ с кодом 201 Created
	writeJSON(w, http.StatusCreated, ShortenResponse{ShortURL: shortURL})
}

// Redirect — HTTP-обработчик для GET /{code}
// Перенаправляет с короткой ссылки на длинную.
func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Извлекаем короткий код из пути
	shortCode := r.PathValue("code")
	if shortCode == "" {
		writeError(w, http.StatusBadRequest, "short code is required")
		return
	}
	// Получаем длинную ссылку через сервис
	longURL, err := h.service.GetLongURL(r.Context(), shortCode)
	if err != nil {
		// Если ссылка не найдена — 404
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	// HTTP 301 Moved Permanently — постоянное перенаправление
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

// Stats — HTTP-обработчик для GET /stats/{code}
// Возвращает статистику по короткой ссылке.
func (h *LinkHandler) Stats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("code")
	if shortCode == "" {
		writeError(w, http.StatusBadRequest, "short code is required")
		return
	}

	link, err := h.service.GetStats(r.Context(), shortCode)
	if err != nil {
		// Если ссылка не найдена — 404
		writeError(w, http.StatusNotFound, "link not found")
		return
	}

	writeJSON(w, http.StatusOK, StatsResponse{
		ShortCode: link.ShortCode,
		LongURL:   link.LongURL,
		Clicks:    link.Clicks,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	// указываем, что ответ — JSON
	w.Header().Set("Content-Type", "application/json")
	// устанавливаем HTTP-статус
	w.WriteHeader(status)
	// кодируем и отправляем
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
