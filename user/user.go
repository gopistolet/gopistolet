package user

import "github.com/gopistolet/gopistolet/smtp"

// Implementation of User for our SMTP service
type User struct {
	Name     string
	Email    smtp.MailAddress
	Password string
}

func (u *User) CheckPassword(password string) bool {
	return password == u.Password
}
