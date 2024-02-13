package captcha

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/panshiqu/dysms"
	"html/template"
	"math/rand"
	"net/smtp"
	"strings"
	"time"
)

var (
	EmailHost     = "" // 己方邮箱服务，例如smtp.exmail.qq.com:465
	EmailUser     = "" // 己方邮箱账号 例如xxx@xxx.com
	EmailPassword = "" // 己方邮箱密码
	mimeTypeHtml  = "text/html"
	mimeTypeText  = "text/plain"

	SmsKey    = "" // 阿里云平台短信服务key
	SmsSecret = "" // 阿里云平台短信服务secret
	SmsName   = "" // 阿里云平台短信服务短信头名称，例如 elozo、羽苏生物
	SmsTmpl   = "" // 阿里云平台短信服务模板编码
)

type Subject struct {
	Email string
	Code  string
}

type Sms struct {
	Key    string
	Secret string
	Name   string
	Tmpl   string
}

var sms = Sms{
	Key:    SmsKey,
	Secret: SmsSecret,
	Name:   SmsName,
	Tmpl:   SmsTmpl,
}

func init() {
	rand.New(rand.NewSource(time.Now().Unix()))
}

// SendToMail  //
func SendToMail(user, password, host, to, subject, body, mimeType string) error {
	auth := smtp.PlainAuth("", user, password, strings.Split(host, ":")[0])
	e := email.NewEmail()
	e.From = user
	e.To = []string{to}
	e.Subject = subject
	switch mimeType {
	case mimeTypeHtml:
		e.HTML = []byte(body)
	case mimeTypeText:
		e.Text = []byte(body)
	default:
		return errors.New("invalid mimeType:" + mimeType)
	}
	err := e.SendWithTLS(host, auth, &tls.Config{ServerName: strings.Split(host, ":")[0]})
	return err
}

func SendVerifyMail(to, code, subject, originalTmpl string) error {

	tmpl, err := template.New("verify-code.gohtml").Parse(originalTmpl)
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err := tmpl.ExecuteTemplate(&tpl, "verify-code.gohtml", Subject{Email: to, Code: code}); err != nil {
		return err
	}

	err = SendToMail(EmailUser, EmailPassword, EmailHost, to, subject, tpl.String(), "text/html")
	return err
}

// SendMobileCaptcha 发送验证码
func SendMobileCaptcha(mobile, code, name string) error {
	params := struct {
		Code string `json:"code"`
	}{
		code,
	}

	buf, err := json.Marshal(&params)
	if err != nil {
		return err
	}
	sms.Name = name
	if err := dysms.SendSms(sms.Key, sms.Secret, mobile, sms.Name, string(buf), sms.Tmpl); err != nil {
		return err
	}

	return nil
}

func GetCaptchaCode() string {
	return fmt.Sprintf("%06v", rand.Int31n(1000000))
}
