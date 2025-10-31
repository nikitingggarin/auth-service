package email

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func TestSendWelcomeEmail(t *testing.T) {
	// Загружаем .env из корня проекта
	envPath := filepath.Join("..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Logf("❌ Failed to load .env from %s: %v", envPath, err)
		// Пробуем загрузить из текущей директории
		if err := godotenv.Load(); err != nil {
			t.Skip("Skipping test - .env file not found")
		}
	}

	// Проверяем что SMTP настройки есть
	if os.Getenv("SMTP_USERNAME") == "" {
		t.Errorf("❌ SMTP_USERNAME not set")
		return
	}
	if os.Getenv("SMTP_PASSWORD") == "" {
		t.Errorf("❌ SMTP_PASSWORD not set")
		return
	}

	emailService := NewEmailService()

	// Логируем настройки (без пароля)
	t.Logf("🔧 SMTP settings: %s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))
	t.Logf("📧 From: %s", os.Getenv("FROM_EMAIL"))

	// Используем почту из .env (нашу Яндекс почту)
	testEmail := os.Getenv("FROM_EMAIL") // Отправляем самому себе
	testName := "Test User"

	t.Logf("🧪 Testing email sending to ourselves: %s", testEmail)

	// Пробуем отправить email
	if err := emailService.SendWelcomeEmail(testEmail, testName); err != nil {
		t.Errorf("❌ Failed to send welcome email: %v", err)
	} else {
		t.Logf("✅ Email sent successfully! Check inbox: %s", testEmail)
	}
}
