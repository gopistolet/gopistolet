package smtp

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseAddress(t *testing.T) {

	Convey("Testing ParseAddress()", t, func() {

		mails := []struct {
			str    string
			parsed MailAddress
		}{
			{
				str: `"Bob" <bob@example.com>`,
				parsed: MailAddress{
					Name:   `Bob`,
					Local:  `bob`,
					Domain: `example.com`,
				},
			},
			{
				str: `   <bob@example.com> `,
				parsed: MailAddress{
					Name:   ``,
					Local:  `bob`,
					Domain: `example.com`,
				},
			},
		}

		for _, mail := range mails {
			address, err := ParseAddress(mail.str)
			fmt.Println(address)
			So(err, ShouldEqual, nil)
			So(address.String(), ShouldEqual, mail.parsed.String())
		}

	})

}

func TestValidate(t *testing.T) {
	Convey("Testing Validate()", t, func() {

		valid_locals := []string{
			"mathias",
			"foo,!#",
			"!def!xyz%abc",
			"$A12345",
			//"Fred Bloggs",
			"customer/department=shipping",
		}

		for _, m := range valid_locals {
			m := MailAddress{Local: m, Domain: "example.com"}
			valid, _ := m.Validate()
			So(valid, ShouldEqual, true)
		}

	})

}
