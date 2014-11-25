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
				
					/*
					RFC 5321
					
					An EHLO command MAY be issued by a client later in the session.  If
					it is issued after the session begins and the EHLO command is
					acceptable to the SMTP server, the SMTP server MUST clear all buffers
					and reset the state exactly as if a RSET command had been issued.  In
					other words, the sequence of RSET followed immediately by EHLO is
					redundant, but not harmful other than in the performance cost of
					executing unnecessary commands.
					*/
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

				if conn.from == "" {
					conn.write(503, "Need MAIL before RCPT")
					break
				}

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
						  
					
					RFC 5321
						  
					When an SMTP server receives a message for delivery or further
					processing, it MUST insert trace ("time stamp" or "Received")
					information at the beginning of the message content, as discussed in
					Section 4.1.1.4.
					
					This line MUST be structured as follows:
					
					o  The FROM clause, which MUST be supplied in an SMTP environment,
					   SHOULD contain both (1) the name of the source host as presented
					   in the EHLO command and (2) an address literal containing the IP
					   address of the source, determined from the TCP connection.
					
					o  The ID clause MAY contain an "@" as suggested in RFC 822, but this
					   is not required.
					
					o  If the FOR clause appears, it MUST contain exactly one <path>
					   entry, even when multiple RCPT commands have been given.  Multiple
					   <path>s raise some security issues and have been deprecated, see
					   Section 7.2.
					   
					---
					
					Any system that includes an SMTP server supporting mail relaying or
					delivery MUST support the reserved mailbox "postmaster" as a case-
					insensitive local name.  This postmaster address is not strictly
					necessary if the server always returns 554 on connection opening (as
					described in Section 3.1).  The requirement to accept mail for
					postmaster implies that RCPT commands that specify a mailbox for
					postmaster at any of the domains for which the SMTP server provides
					mail service, as well as the special case of "RCPT TO:<Postmaster>"
					(with no domain specification), MUST be supported.
				*/

			}

		case "DATA":
			{
				if conn.from == "" {
					conn.write(503, "Need MAIL before DATA")
					break
				}

				if len(conn.to) < 1 {
					conn.write(503, "Need RCPT before DATA")
				}

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
						
						---
						
						Without some provision for data transparency, the character sequence
						"<CRLF>.<CRLF>" ends the mail text and cannot be sent by the user.
						In general, users are not aware of such "forbidden" sequences.  To
						allow all user composed text to be transmitted transparently, the
						following procedures are used:
						
						o  Before sending a line of mail text, the SMTP client checks the
						   first character of the line.  If it is a period, one additional
						   period is inserted at the beginning of the line.
						
						o  When a line of mail text is received by the SMTP server, it checks
						   the line.  If the line is composed of a single period, it is
						   treated as the end of mail indicator.  If the first character is a
						   period and there are other characters on the line, the first
						   character is deleted.
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
			
		case "SEND", "SOML", "SAML": {
				// Obsolete
				conn.write(502, "Command not implemented")
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
