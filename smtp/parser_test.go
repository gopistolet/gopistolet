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
		commands := ""
		commands += "HELO relay.example.org\r\n"
		commands += "EHLO other.example.org\r\n"
		commands += "MAIL FROM:<bob@example.org>\r\n"
		commands += "RCPT TO:<alice@example.com>\r\n"
		commands += "RCPT TO:<theboss@example.com>\r\n"
		commands += "SEND\r\n"
		commands += "SOML\r\n"
		commands += "SAML\r\n"
		commands += "RSET\r\n"
		commands += "VRFY jones\r\n"
		commands += "EXPN staff\r\n"
		commands += "NOOP\r\n"
		commands += "QUIT\r\n"

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
			Print(expectedCommand)
			command, err := p.ParseCommand(br)
			So(err, ShouldEqual, nil)
			So(command, ShouldResemble, expectedCommand)
		}

	})

	Convey("Testing parser DATA cmd", t, func() {
		commands := ""
		commands += "DATA\r\n"
		commands += "Some usefull data.\r\n"
		commands += ".\r\n"
		commands += "quit\r\n"

		br := bufio.NewReader(strings.NewReader(commands))
		p := parser{}

		command, err := p.ParseCommand(br)
		So(err, ShouldEqual, nil)
		So(command, ShouldHaveSameTypeAs, DataCmd{})
		dataCommand, ok := command.(DataCmd)
		So(ok, ShouldEqual, true)
		br2 := bufio.NewReader(dataCommand.R.br)
		line, _ := br2.ReadString('\n')
		So(line, ShouldEqual, "Some usefull data.\r\n")
		line, _ = br2.ReadString('\n')
		So(line, ShouldEqual, ".\r\n")

		command, err = p.ParseCommand(br)
		So(err, ShouldEqual, nil)
		So(command, ShouldHaveSameTypeAs, QuitCmd{})

	})

	Convey("Testing parser with invalid commands", t, func() {

		commands := ""
		commands += "RCPT\r\n"
		commands += "helo\r\n"
		commands += "ehlo\r\n"
		commands += "\r\n"
		commands += "  \r\n"
		commands += "RCPT TO:some invalid email\r\n"
		commands += "rcpt :valid@mail.be\r\n"
		commands += "RCPT :valid@mail.be\r\n"
		commands += "RCPT TA:valid@mail.be\r\n"
		commands += "MAIL\r\n"
		commands += "MAIL from:some invalid email\r\n"
		commands += "MAIL :valid@mail.be\r\n"
		commands += "MAIL FROA:valid@mail.be\r\n"
		commands += "MAIL To some@invalid\r\n"
		commands += "UNKN some unknown command\r\n"

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
			{
				line: "HELO some_ctrl_char\r\n",
				verb: "HELO",
				args: []string{"some_ctrl_char"},
			},
			{
				line: "HELO some_ctrl_char\n",
				verb: "HELO",
				args: []string{"some_ctrl_char"},
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
