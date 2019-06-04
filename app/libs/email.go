package libs

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

var emailService *EmailService

type EmailService struct {
	Dialer *gomail.Dialer
	From string
}

func  InitEmail(c EmailConfig) {


	emailService = &EmailService{
		Dialer: gomail.NewDialer(c.Host, c.Port, c.User, c.Password),
		From: c.From,
	}
}

func (s *EmailService) SendEmail(to, data string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	m.Attach("/home/Alex/lolcat.jpg")

	return s.Dialer.DialAndSend(m)
}

func GetEmail() *EmailService {
	if emailService == nil {
		log.Panic().Msg("Email service is not init")
	}
	return emailService
}