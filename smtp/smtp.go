package smtp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"regexp"
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
	to []string
	msg []byte
}

func (conn *Conn) serve() error {
	defer conn.c.Close()

	log.Printf("Received new connection")
	fmt.Fprintf(conn.c, "220 SMTP ready\n")

	br := bufio.NewReader(conn.c)
	for {
		line, _ := br.ReadString('\n')
		
		if line == "" {
			continue
		}

		verb, args := parseLine(line)
		log.Printf("%s -------- %s", verb, args)
		
		switch verb {
			
			case "HELO": {
				// Initiate SMTP conversation
				log.Printf("    > SMTP request from %s", args)
				fmt.Fprintf(conn.c, "250 OK\n")
			}
				
			case "EHLO": {
				// Initiate (extended) SMTP conversation
				log.Printf("    > Extended SMTP request from %s", args)
				fmt.Fprintf(conn.c, "502 Command not implemented\n")
			}
	
			case "MAIL": {
				// MAIL FROM:<sender@example.com>
				conn.from = parseFROM(args)
				log.Printf("    > From: %s", conn.from)
				fmt.Fprintf(conn.c, "250 OK\n")
			}
			
			case "RCPT": {
				// RCPT TO:<sender@example.com>
				conn.to = append(conn.to, parseTO(args))
				log.Printf("    > To: %s", conn.to)
				fmt.Fprintf(conn.c, "250 OK\n")
			}
			
			case "DATA": {
				// Read data until ending '.' line.
				
				for {
					
					data, _ := br.ReadString('\n')
					
					log.Printf("    > (%d) %q", len(data), data)
					
					if data == ".\r\n" {
						log.Printf("    > END")
						break
					} else {
						conn.msg = append(conn.msg, []byte(data)...)
						continue
					}
					
				}
				
				log.Printf("    > Data: %s", conn.msg)
				
			}
			
			case "RSET": {
				// reset all sent information
				// don't close connection!
				conn.from = "";
				conn.to = make([]string, 0)
				conn.msg = make([]byte, 0)
			}
			
			case "VRFY": {
				// placeholder
				fmt.Fprintf(conn.c, "502 Command not implemented\n")	
			}
			
			case "NOOP": {
				// Tell the client that the server isn't dead :)
				fmt.Fprintf(conn.c, "250 OK\n")
			}
			
			case "QUIT": {
				// Close connection
				log.Printf("    > Closing connection")
				fmt.Fprintf(conn.c, "221 QUIT\n")
				return nil
			}
	
			default: {
				fmt.Fprintf(conn.c, "500 ERROR\n")
				log.Printf("    > Unsupported command: '%s'", verb)
			}
	
		}
		
	}

	return nil
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
}


// some regexes we don't want to compile for each request
var (
	fromRegex = regexp.MustCompile(`FROM:<(.+@.+)>`)
	toRegex = regexp.MustCompile(`TO:<(.+@.+)>`)
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
