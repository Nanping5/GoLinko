package sms

import (
	"GoLinko/internal/config"
	myredis "GoLinko/internal/service/redis"
	constants "GoLinko/pkg/const"
	"GoLinko/pkg/utils"
	"GoLinko/pkg/zlog"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
	"go.uber.org/zap"
)

func VerifyEmailCode(email2 string) (msg string, ret int) {
	key := "email_code:" + email2

	code, err := myredis.GetKey(key)
	if err != nil {
		zlog.GetLogger().Error("", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}
	if code != "" {
		msg = "验证码已发送，请稍后再试"
		zlog.GetLogger().Info(msg, zap.String("email", email2))
		return msg, -2
	}
	//生成验证码并存储到 Redis，设置过期时间为 3 分钟
	code = strconv.Itoa(utils.GenerateRandomCode(6))
	err = myredis.SetKeyEx(key, code, 180)
	if err != nil {
		zlog.GetLogger().Error("", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}

	targetMailBox := email2
	smtpServer := config.GetConfig().Smtp.SmtpServer
	emailAddr := config.GetConfig().Smtp.EmailAddr
	smtpKey := config.GetConfig().Smtp.SmtpKey

	// 发送邮件
	em := email.NewEmail()
	em.From = fmt.Sprintf("GoLinko <%s>", emailAddr)
	em.To = []string{targetMailBox}
	em.Subject = "GoLinko 验证码"
	em.Text = []byte(fmt.Sprintf("您的验证码是: %s, 有效期为3分钟", code))

	// QQ 邮箱使用 465 端口需要 TLS 连接
	err = em.SendWithTLS(
		smtpServer+":465",
		smtp.PlainAuth("", emailAddr, smtpKey, smtpServer),
		&tls.Config{ServerName: smtpServer},
	)
	if err != nil {
		zlog.GetLogger().Error("发送验证码邮件失败", zap.Error(err))
		return constants.SYSTEM_ERROR, -1
	}
	msg = "验证码已发送，请注意查收"
	zlog.GetLogger().Info(msg, zap.String("email", email2))
	return msg, 0
}
