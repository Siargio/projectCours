package domain

import (
	"testing"
)

// Тест генерации короткого кода
func TestGenerateShortCode(t *testing.T) {
	// Test case 1: проверка длины
	t.Run("returns correct length", func(t *testing.T) {
		code, err := GenerateShortCode(8)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 8 {
			t.Errorf("expected length 8, got %d", len(code))
		}
	})

	// Test case 2: проверка уникальности (генерация двух разных кодов)
	t.Run("generates unique codes", func(t *testing.T) {
		code1, _ := GenerateShortCode(8)
		code2, _ := GenerateShortCode(8)

		if code1 == code2 {
			t.Errorf("expected different codes, got same: %s", code1)
		}
	})

	// Test case 3: проверка разных длин
	t.Run("handles different lengths", func(t *testing.T) {
		testCases := []struct {
			length int
		}{
			{4}, {6}, {8}, {10}, {12},
		}

		for _, tc := range testCases {
			code, err := GenerateShortCode(tc.length)
			if err != nil {
				t.Fatalf("length %d: unexpected error: %v", tc.length, err)
			}
			if len(code) != tc.length {
				t.Errorf("length %d: expected %d, got %d", tc.length, tc.length, len(code))
			}
		}
	})
}
