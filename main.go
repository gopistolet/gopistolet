package main

import "gopistolet/smtp"

func main() {
	config := smtp.Config{Port: 1234, Hostname: "", Key: "cert/server.key", Cert: "cert/server.crt"}
	s := smtp.NewMSAServer(config)
	s.ListenAndServe()
}
