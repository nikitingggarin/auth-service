package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type EmailService struct {
	config     Config
	workerPool *WorkerPool
}

func NewEmailService() *EmailService {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", ""))

	return &EmailService{
		config: Config{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     port,
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", ""),
			FromName:     getEnv("FROM_NAME", ""),
		},
		workerPool: NewWorkerPool(5), // 5 одновременных отправок
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SendWelcomeEmail отправляет реальное welcome email
func (s *EmailService) SendWelcomeEmail(email, userName string) error {
	subject := "Добро пожаловать в наше приложение!"

	// Простой текстовый email
	body := fmt.Sprintf(`
Добро пожаловать, %s!

Мы рады приветствовать вас в нашем приложении.

Ваш аккаунт был успешно создан и готов к использованию.

Если у вас возникнут вопросы, не стесняйтесь обращаться в нашу поддержку.

С уважением,
Команда Auth Servise
`, userName)

	log.Printf("📧 Starting to send welcome email to: %s", email)

	if err := s.sendEmail(email, subject, body); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	log.Printf("✅ Welcome email sent successfully to: %s", email)
	return nil
}

// sendEmail отправляет email через SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	// Формируем сообщение
	message := s.buildMessage(to, subject, body)

	// Настройка аутентификации
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// Адрес SMTP сервера
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Отправка email
	return smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(message))
}

// buildMessage формирует email сообщение
func (s *EmailService) buildMessage(to, subject, body string) string {
	return fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.config.FromName,
		s.config.FromEmail,
		to,
		subject,
		body)
}

// SendWelcomeEmailAsync запускает отправку email в фоне
func (s *EmailService) SendWelcomeEmailAsync(email, userName string) {
    go func() {
        s.workerPool.Acquire() // Занимаем воркера
        defer s.workerPool.Release() // Освобождаем воркера
        
        if err := s.SendWelcomeEmail(email, userName); err != nil {
            log.Printf("❌ Failed to send welcome email to %s: %v", email, err)
        }
    }()
}
