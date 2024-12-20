package services

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"os"
	"strconv"
)

type EmailService struct {
	dialer *gomail.Dialer
}

func NewEmailService() *EmailService {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		port = 587 // default port
	}

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return &EmailService{dialer: d}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	fmt.Printf("Sending email to %s with subject: %s\n", to, subject)
	if err := s.dialer.DialAndSend(m); err != nil {
		fmt.Printf("Error sending email: %v\n", err)
		return err
	}
	fmt.Println("Email sent successfully")
	return nil
}
