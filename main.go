package main

import "GoPistolet/smtp"

func main() {
	s := smtp.Server{Addr: ":1024"}
	s.ListenAndServe()
}
