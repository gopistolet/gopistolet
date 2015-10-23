package smtp

import (
	"bufio"
	_ "fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParser(t *testing.T) {

	Convey("Testing parser", t, func() {
		commands :=
			`HELO relay.example.org
EHLO other.example.org
MAIL FROM:<bob@example.org>
RCPT TO:<alice@example.com>
RCPT TO:<theboss@example.com>
SEND
SOML
SAML
RSET
VRFY jones
EXPN staff
NOOP
QUIT
`

		br := bufio.NewReader(strings.NewReader(commands))

		p := parser{}

		expectedCommands := []Cmd{
			HeloCmd{Domain: "relay.example.org"},
			EhloCmd{Domain: "other.example.org"},
			MailCmd{From: &MailAddress{Address: "bob@example.org"}},
			RcptCmd{To: &MailAddress{Address: "alice@example.com"}},
			RcptCmd{To: &MailAddress{Address: "theboss@example.com"}},
			SendCmd{},
			SomlCmd{},
			SamlCmd{},
			RsetCmd{},
			VrfyCmd{Param: "jones"},
			ExpnCmd{ListName: "staff"},
			NoopCmd{},
			QuitCmd{},
		}

		for _, expectedCommand := range expectedCommands {
			command, err := p.ParseCommand(br)
			So(err, ShouldEqual, nil)
			So(command, ShouldResemble, expectedCommand)
		}

	})

	Convey("Testing parser DATA cmd", t, func() {
		commands :=
			`DATA
Some usefull data.
.
quit
`

		br := bufio.NewReader(strings.NewReader(commands))
		p := parser{}

		command, err := p.ParseCommand(br)
		So(err, ShouldEqual, nil)
		So(command, ShouldHaveSameTypeAs, DataCmd{})

		dcmd, _ := command.(DataCmd)
		br2 := bufio.NewReader(&dcmd.R)
		br2.ReadString('\n')

		command, err = p.ParseCommand(br)
		So(err, ShouldEqual, nil)
		So(command, ShouldHaveSameTypeAs, QuitCmd{})

	})

	Convey("Testing parser with invalid commands", t, func() {

		commands := `RCPT
helo
ehlo

  
RCPT TO:some invalid email
rcpt :valid@mail.be
RCPT :valid@mail.be
RCPT TA:valid@mail.be
MAIL
MAIL from:some invalid email
MAIL :valid@mail.be
MAIL FROA:valid@mail.be
MAIL To some@invalid
UNKN some unknown command
`

		br := bufio.NewReader(strings.NewReader(commands))

		p := parser{}

		expectedCommands := []Cmd{
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			UnknownCmd{},
			UnknownCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			InvalidCmd{},
			UnknownCmd{},
		}

		for _, expectedCommand := range expectedCommands {
			command, err := p.ParseCommand(br)
			So(err, ShouldEqual, nil)
			So(command, ShouldHaveSameTypeAs, expectedCommand)
		}

	})

	Convey("Testing parseLine()", t, func() {

		tests := []struct {
			line string
			verb string
			args []string
		}{
			{
				line: "HELO\r\n",
				verb: "HELO",
			},
			{
				line: "HELO relay.example.org\r\n",
				verb: "HELO",
				args: []string{"relay.example.org"},
			},
			{
				line: "MAIL FROM:<bob@example.org>\r\n",
				verb: "MAIL",
				args: []string{"FROM:<bob@example.org>"},
			},
		}

		for _, test := range tests {
			verb, args, err := parseLine(bufio.NewReader(strings.NewReader(test.line)))
			So(err, ShouldEqual, nil)
			So(verb, ShouldEqual, test.verb)
			So(args, ShouldResemble, test.args)
		}

	})

	Convey("Testing parseTo()", t, func() {

		tests := []struct {
			line          string
			addressString string
		}{
			{
				line:          "RCPT TO:<alice@example.com>\r\n",
				addressString: "alice@example.com",
			},
		}

		for _, test := range tests {
			_, args, err := parseLine(bufio.NewReader(strings.NewReader(test.line)))
			So(err, ShouldEqual, nil)

			addr, err := parseTO(args)
			So(err, ShouldEqual, nil)
			So(addr.GetAddress(), ShouldEqual, test.addressString)
		}

	})

}
