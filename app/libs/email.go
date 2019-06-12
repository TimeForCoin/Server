package libs

import (
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/gomail.v2"
)

var emailService *EmailService

// EmailService 邮件服务
type EmailService struct {
	Dialer *gomail.Dialer
	From   string
}

// InitEmail 初始化邮件服务
func InitEmail(c EmailConfig) {
	emailService = &EmailService{
		Dialer: gomail.NewDialer(c.Host, c.Port, c.User, c.Password),
		From:   c.From,
	}
}

// SendAuthEmail 发送认证邮件
func (s *EmailService) SendAuthEmail(id primitive.ObjectID, to, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "[闲得一币] 身份认证")
	m.SetBody("text/html",
		"<p>请点击下面的链接完成邮箱身份认证，有效期为30分钟</p><p><a>https://coin.zhenly.cn/api/certification/auth?code="+
			code+"&user="+id.Hex()+
			"</a></p><p>如果不是本人操作请勿点击</p>")
	return s.Dialer.DialAndSend(m)
}

// GetEmail 获取全局邮件服务实例
func GetEmail() *EmailService {
	if emailService == nil {
		log.Panic().Msg("Email service is not init")
	}
	return emailService
}
