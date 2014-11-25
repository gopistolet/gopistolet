package smtp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

type Handler func()

type Server struct {
	Addr    string
	Handler Handler
}

func (srv *Server) ListenAndServe() error {
	if srv.Addr == "" {
		srv.Addr = ":25"
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	return srv.Serve(ln)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			// Just a temporary error
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("Accept error: %v", err)
				continue
			}

			return err
		}

		conn := srv.newConn(c)
		go conn.serve()
	}

	return nil
}

// Creat a wrapper around net.Conn
func (srv *Server) newConn(c net.Conn) *Conn {
	return &Conn{
		c: c,
	}
}

type Conn struct {
	c net.Conn

	from string
	to   []string
	msg  []byte
}

func (conn *Conn) serve() error {
	defer conn.c.Close()

	log.Printf("Received new connection")
	conn.write(220, "GoPistolet at your service")

	br := bufio.NewReader(conn.c)
	for {
		line, _ := br.ReadString('\n')

		if line == "" {
			continue
		}

		verb, args := parseLine(line)
		switch verb {

		case "HELO":
			{
				// Initiate SMTP conversation
				log.Printf("    > SMTP request from %s", args)
				conn.write(250, "GoPistolet")
			}

		case "EHLO":
			{
				// Initiate (extended) SMTP conversation
				log.Printf("    > Extended SMTP request from %s", args)
				conn.write(502, "Command not implemented")
			}

		case "MAIL":
			{
				// MAIL FROM:<sender@example.com>

				if conn.from != "" {
					log.Printf("    > MAIL FROM already specified: %s", conn.from)
					conn.write(503, "Sender already specified")
					break
				}

				conn.from = parseFROM(args)

				// Check if we can parse the params
				if conn.from == "" {
					log.Printf("    > Could not parse email")
					conn.write(501, "Invalid syntax")
				} else {
					log.Printf("    > From: %s", conn.from)
					conn.write(250, "OK")
				}

			}

		case "RCPT":
			{
				// RCPT TO:<sender@example.com>

				rcpt := parseTO(args)

				// Check if we can parse the params
				if rcpt == "" {
					log.Printf("    > Could not parse email")
					conn.write(501, "Invalid syntax")
				} else {
					conn.to = append(conn.to, rcpt)
					log.Printf("    > To: %s", rcpt)
					conn.write(250, "OK")
				}

				/*
					RFC 5321:

					The minimum total number of recipients that MUST be buffered is 100
					recipients.  Rejection of messages (for excessive recipients) with
					fewer than 100 RCPT commands is a violation of this specification.
					The general principle that relaying SMTP server MUST NOT, and
					delivery SMTP servers SHOULD NOT, perform validation tests on message
					header fields suggests that messages SHOULD NOT be rejected based on
					the total number of recipients shown in header fields.  A server that
					imposes a limit on the number of recipients MUST behave in an orderly
					fashion, such as rejecting additional addresses over its limit rather
					than silently discarding addresses previously accepted.  A client
					that needs to deliver a message containing over 100 RCPT commands
					SHOULD be prepared to transmit in 100-recipient "chunks" if the
					server declines to accept more than 100 recipients in a single
					message.

						452 Too many recipients
				*/

				// TODO check if  email exists on our server
				/*
					RFC 821

					If the recipient is unknown the
					receiver-SMTP returns a 550 Failure reply.

					There are some cases where the destination information in the
					<forward-path> is incorrect, but the receiver-SMTP knows the
					correct destination.  In such cases, one of the following replies
					should be used to allow the sender to contact the correct
					destination.

					   251 User not local; will forward to <forward-path>

						  This reply indicates that the receiver-SMTP knows the user's
						  mailbox is on another host and indicates the correct
						  forward-path to use in the future.  Note that either the
						  host or user or both may be different.  The receiver takes
						  responsibility for delivering the message.

					   551 User not local; please try <forward-path>

						  This reply indicates that the receiver-SMTP knows the user's
						  mailbox is on another host and indicates the correct
						  forward-path to use.  Note that either the host or user or
						  both may be different.  The receiver refuses to accept mail
						  for this user, and the sender must either redirect the mail
						  according to the information provided or return an error
						  response to the originating user.
				*/

			}

		case "DATA":
			{
				// Read data until ending '.' line.
				conn.write(354, "Accepting mail input")

				for {

					data, _ := br.ReadString('\n')

					log.Printf("    > (%d) %q", len(data), data)

					if data == ".\r\n" || data == ".\r" || data == ".\n" {
						log.Printf("    > END")
						break
					} else {
						conn.msg = append(conn.msg, []byte(data)...)
						continue
					}

					// TODO break when there is no more content
					// TODO check for content too long
					/*
						RFC 5321:

						The maximum total length of a message content (including any message
						header section as well as the message body) MUST BE at least 64K
						octets.  Since the introduction of Internet Standards for multimedia
						mail (RFC 2045 [21]), message lengths on the Internet have grown
						dramatically, and message size restrictions should be avoided if at
						all possible.  SMTP server systems that must impose restrictions
						SHOULD implement the "SIZE" service extension of RFC 1870 [10], and
						SMTP client systems that will send large messages SHOULD utilize it
						when possible.

						552 Too much mail data
					*/

					// TODO check for time out while waiting (this might also be needed for the whole connection)

				}

				log.Printf("    > Data: %s", conn.msg)
				// TODO: Handle email

				// Reset our state so a new MAIL command can be executed
				conn.reset()
				conn.write(250, "OK")
			}

		case "RSET":
			{
				conn.reset()
				conn.write(250, "OK")
			}

		case "VRFY", "EXPN":
			{
				// Additional commands
				conn.write(502, "Command not implemented")
				/*
					RFC 821

					SMTP provides as additional features, commands to verify a user
					name or expand a mailing list.  This is done with the VRFY and
					EXPN commands
				*/

			}

		case "NOOP":
			{
				// Tell the client that the server isn't dead :)
				conn.write(250, "OK")
			}

		case "QUIT":
			{
				// Close connection
				log.Printf("    > Closing connection")
				conn.write(221, "Bye!")
				return nil
			}

		default:
			{
				log.Printf("    > Command unrecognized: '%s'", verb)
				conn.write(500, "Command unrecognized")
			}

			/*
				RFC 5321

				The maximum total length of a reply line including the reply code and
				the <CRLF> is 512 octets.  More information may be conveyed through
				multiple-line replies.
			*/

		}

	}

	return nil
}

func (conn *Conn) write(code int, str string) {
	fmt.Fprintf(conn.c, "%d %s\r\n", code, str)
}

func (conn *Conn) reset() {
	conn.from = ""
	conn.to = make([]string, 0)
	conn.msg = make([]byte, 0)
}

func parseLine(line string) (verb string, args string) {
	i := strings.Index(line, " ")
	if i == -1 {
		verb = strings.ToUpper(strings.TrimSpace(line))
		return
	}

	verb = strings.ToUpper(line[:i])
	args = strings.TrimSpace(line[i+1 : len(line)])
	return

	/*
		RFC 5321

		The maximum total length of a text line including the <CRLF> is 1000
		octets (not counting the leading dot duplicated for transparency).
		This number may be increased by the use of SMTP Service Extensions.

		--

		The maximum total length of a command line including the command word
		and the <CRLF> is 512 octets.  SMTP extensions may be used to
		increase this limit.

			500 Line too long
	*/
}

// some regexes we don't want to compile for each request
var (
	fromRegex = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:[\ ]?<(.+@.+)>`)
	toRegex   = regexp.MustCompile(`[Tt][Oo]:<(.+@.+)>.*`)
)

func parseFROM(line string) string {

	matches := fromRegex.FindStringSubmatch(line)

	if len(matches) == 2 {
		return matches[1]
	} else {
		return ""
	}

}

func parseTO(line string) string {

	matches := toRegex.FindStringSubmatch(line)

	if len(matches) == 2 {
		return matches[1]
	} else {
		return ""
	}

}

/*
	RFC 5321

	The maximum total length of a domain name or number is 255 octets.
*/
