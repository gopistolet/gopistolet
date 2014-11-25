package main

import "gopistolet/smtp"

func main() {
	s := smtp.Server{Addr: ":1024"}
	s.ListenAndServe()
}
