package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

// ErrLinkNotFound — предопределённая ошибка, которую возвращаем,
// когда ссылка не найдена в БД.
var ErrLinkNotFound = errors.New("link not found")

// Link — основная бизнес-сущность "короткая ссылка".
type Link struct {
	// Уникальный идентификатор в БД (первичный ключ)
	ID int64
	// Короткий код
	ShortCode string
	// Оригинальная длинная ссылка
	LongURL string
	// Счётчик переходов по короткой ссылке
	Clicks int
	// Когда ссылка была создана
	CreatedAt time.Time
}

// GenerateShortCode генерирует криптографически случайный короткий код заданной длины.
func GenerateShortCode(length int) (string, error) {
	// Создаём слайс байт нужной длины
	bytes := make([]byte, length)
	// rand.Read заполняет слайс криптографически случайными байтами
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Кодируем байты в URL-safe base64 (без символов + и /)
	// Обрезаем до нужной длины
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
