package mail
import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
)


type Sender interface {
	SendVerificationCode(toEmail, code string) error
}

func NewFromEnv() Sender {
	return &smtpSender{
		host:     os.Getenv("SMTP_HOST"),
		port:     envPort("SMTP_PORT", 587),
		user:     os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

type smtpSender struct {
	host, user, password, from string
	port                       int
}

func envPort(key string, def int) int {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func (s *smtpSender) SendVerificationCode(toEmail, code string) error {
	subject := "Your verification code"
	body := fmt.Sprintf("Your 4-digit verification code is: %s\n\nIt expires in 15 minutes.", code)

	if s.host == "" {
		log.Printf("[mail:dev] verification code for %s: %s", toEmail, code)
		return nil
	}

	if s.from == "" {
		return fmt.Errorf("SMTP_FROM is not set")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", s.from, toEmail, subject, body))

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}
	return smtp.SendMail(addr, auth, s.from, []string{toEmail}, msg)
}
