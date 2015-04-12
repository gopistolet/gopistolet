package main

import "github.com/gopistolet/gopistolet/smtp"

func main() {
	config := smtp.Config{Port: 1234, Hostname: ""}
	s := smtp.NewMSAServer(config)
	s.ListenAndServe()
}
