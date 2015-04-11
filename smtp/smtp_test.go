package smtp

import (
	_ "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	Convey("FROM", t, func() {

		{
			line := "MAIL FROM: <example@example.com>"
			verb, args := parseLine(line)

			So(verb, ShouldEqual, "MAIL")

			So(strings.Join(args, " "), ShouldEqual, "FROM: <example@example.com>")
		}

	})

}

func TestParseFrom(t *testing.T) {
	Convey("FROM", t, func() {

		{ // Most simple test for email FROM
			line := "MAIL FROM:<example.email@example.com>"
			_, args := parseLine(line)

			email, err := parseFROM(args)

			So(err, ShouldEqual, nil)
			So(email.Local, ShouldEqual, "example.email")
			So(email.Domain, ShouldEqual, "example.com")
		}

		{ // With space between FROM: and email
			line := "MAIL FROM: <example.email@example.com>"
			_, args := parseLine(line)

			email, err := parseFROM(args)

			So(err, ShouldEqual, nil)
			So(email.Local, ShouldEqual, "example.email")
			So(email.Domain, ShouldEqual, "example.com")
		}

		{ // Quoted string
			line := `MAIL FROM: <" example@email"@example.com>`
			_, args := parseLine(line)

			email, err := parseFROM(args)

			So(err, ShouldEqual, nil)
			So(email.Local, ShouldEqual, " example@email")
			So(email.Domain, ShouldEqual, "example.com")
		}

		{ // With name
			line := `MAIL FROM: "Bob Example" <bob@example.com>`
			_, args := parseLine(line)

			email, err := parseFROM(args)

			So(err, ShouldEqual, nil)
			So(email.Local, ShouldEqual, "bob")
			So(email.Domain, ShouldEqual, "example.com")
			So(email.Name, ShouldEqual, "Bob Example")
		}

	})
}

func TestParseTo(t *testing.T) {
	Convey("TO", t, func() {

		{ // Most simple test for email FROM
			line := "RCPT TO:<example.email@example.com>"
			_, args := parseLine(line)

			email, err := parseTO(args)

			So(err, ShouldEqual, nil)
			So(email.Local, ShouldEqual, "example.email")
			So(email.Domain, ShouldEqual, "example.com")
		}

		{ // With space between FROM: and email
			line := "RCPT TO: <example.email@example.com>"
			_, args := parseLine(line)

			email, err := parseTO(args)

			So(err, ShouldEqual, nil)
			So(email.Local, ShouldEqual, "example.email")
			So(email.Domain, ShouldEqual, "example.com")
		}

	})
}
