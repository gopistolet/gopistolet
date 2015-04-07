package smtp

import (
	_ "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

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
