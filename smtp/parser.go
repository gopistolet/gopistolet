package smtp

import "bufio"

import "strings"
import "errors"

type parser struct {
}

func (p *parser) ParseCommand(br *bufio.Reader) (command Cmd, err error) {
	/*
		RFC 5321 2.3.8

		Lines consist of zero or more data characters terminated by the
		sequence ASCII character "CR" (hex value 0D) followed immediately by
		ASCII character "LF" (hex value 0A).  This termination sequence is
		denoted as <CRLF> in this document.  Conforming implementations MUST
		NOT recognize or generate any other character or character sequence
		as a line terminator.  Limits MAY be imposed on line lengths by
		servers (see Section 4).


		RFC 2821
		The maximum total length of a command line including the command
		word and the <CRLF> is 512 characters.  SMTP extensions may be
		used to increase this limit.

	*/

	var address *MailAddress
	verb, args, err := parseLine(br)
	if err != nil {
		return nil, err
	}
	//conn.write(500, err.Error())
	//conn.c.Close()

	switch verb {

	case "HELO":
		{
			// TODO: < 1 or == 1?
			if len(args) < 1 {
				command = InvalidCmd{Cmd: "HELO", Info: "HELO requires valid address"}
				break
			}
			command = HeloCmd{Domain: args[0]}
		}

	case "EHLO":
		{
			// TODO: < 1 or == 1?
			if len(args) < 1 {
				command = InvalidCmd{Cmd: "EHLO", Info: "EHLO requires valid address"}
				break
			}
			command = EhloCmd{Domain: args[0]}
		}

	case "MAIL":
		{
			address, err = parseFROM(args)
			if err != nil {
				command = InvalidCmd{Cmd: verb, Info: err.Error()}
				err = nil
			} else {
				command = MailCmd{From: address}
			}
		}

	case "RCPT":
		{
			address, err = parseTO(args)
			if err != nil {
				command = InvalidCmd{Cmd: verb, Info: err.Error()}
				err = nil
			} else {
				command = RcptCmd{To: address}
			}
		}

	case "DATA":
		{
			// TODO: write tests for this
			command = DataCmd{
				R: *NewDataReader(br),
			}
		}

	case "RSET":
		{
			command = RsetCmd{}
		}

	case "SEND":
		{
			command = SendCmd{}
		}

	case "SOML":
		{
			command = SomlCmd{}
		}

	case "SAML":
		{
			command = SamlCmd{}
		}

	case "VRFY":
		{
			//conn.write(502, "Command not implemented")
			/*
					RFC 821
					SMTP provides as additional features, commands to verify a user
					name or expand a mailing list.  This is done with the VRFY and
					EXPN commands
					RFC 5321
					As discussed in Section 3.5, individual sites may want to disable
					either or both of VRFY or EXPN for security reasons (see below).  As
					a corollary to the above, implementations that permit this MUST NOT
					appear to have verified addresses that are not, in fact, verified.
					If a site disables these commands for security reasons, the SMTP
					server MUST return a 252 response, rather than a code that could be
					confused with successful or unsuccessful verification.
					Returning a 250 reply code with the address listed in the VRFY
					command after having checked it only for syntax violates this rule.
					Of course, an implementation that "supports" VRFY by always returning
					550 whether or not the address is valid is equally not in
					conformance.
				From what I have read, 502 is better than 252...
			*/
			command = VrfyCmd{Param: args[0]}
		}

	case "EXPN":
		{
			command = ExpnCmd{ListName: args[0]}
		}

	case "NOOP":
		{
			command = NoopCmd{}
		}

	case "QUIT":
		{
			command = QuitCmd{}
		}

	default:
		{
			command = UnknownCmd{Cmd: verb, Line: strings.TrimSuffix(verb, "\n")}
		}

	}

	return
}

// parseLine returns the verb of the line and a list of all comma separated arguments
func parseLine(r *bufio.Reader) (verb string, args []string, err error) {

	/*
		RFC 5321
		4.5.3.1.4.  Command Line

		The maximum total length of a command line including the command word
		and the <CRLF> is 512 octets.  SMTP extensions may be used to
		increase this limit.
	*/
	line, err := r.ReadString('\n')
	if err != nil {
		return line, []string{}, err
	}

	// Strip \n and \r
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	i := strings.Index(line, " ")
	if i == -1 {
		verb = strings.ToUpper(strings.TrimSpace(line))
		return
	}

	verb = strings.ToUpper(line[:i])
	args = strings.Split(strings.TrimSpace(line[i+1:len(line)]), " ")
	return
}

func parseFROM(args []string) (*MailAddress, error) {
	if len(args) < 1 {
		return nil, errors.New("No FROM given")
	}

	joined_args := strings.Join(args, " ")
	index := strings.Index(joined_args, ":")
	if index == -1 {
		return nil, errors.New("No FROM given (didn't find ':')")
	}
	if strings.ToLower(joined_args[0:index]) != "from" {
		return nil, errors.New("No FROM given")
	}

	address_str := joined_args[index+1 : len(joined_args)]

	address, err := ParseAddress(address_str)
	if err != nil {
		return nil, err
	} else {
		return &address, nil
	}

}

func parseTO(args []string) (*MailAddress, error) {
	if len(args) < 1 {
		return nil, errors.New("No TO given")
	}

	joined_args := strings.Join(args, " ")
	index := strings.Index(joined_args, ":")
	if index == -1 {
		return nil, errors.New("No TO given (didn't find ':')")
	}
	if strings.ToLower(joined_args[0:index]) != "to" {
		return nil, errors.New("No TO given")
	}

	address_str := joined_args[index+1 : len(joined_args)]

	address, err := ParseAddress(address_str)
	if err != nil {
		return nil, err
	} else {
		return &address, nil
	}
}
