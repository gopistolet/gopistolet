package smtp

import (
	"errors"
	_ "fmt"
	"log"
	"net"
	"net/mail"
	"strings"
)

type MailAddress struct {
	Name   string
	Local  string
	Domain string
}

func (m *MailAddress) String() string {
	a := mail.Address{Name: m.Name, Address: m.Local + "@" + m.Domain}
	return a.String()
}

// Validate the email adress
/*
   RFC 5321

   address-literal  = "[" ( IPv4-address-literal /
                    IPv6-address-literal /
                    General-address-literal ) "]"
                    ; See Section 4.1.3

   Mailbox        = Local-part "@" ( Domain / address-literal )

   Local-part     = Dot-string / Quoted-string
                  ; MAY be case-sensitive


   Dot-string     = Atom *("."  Atom)

   Atom           = 1*atext

   Quoted-string  = DQUOTE *QcontentSMTP DQUOTE

   QcontentSMTP   = qtextSMTP / quoted-pairSMTP

   quoted-pairSMTP  = %d92 %d32-126
                    ; i.e., backslash followed by any ASCII
                    ; graphic (including itself) or SPace

   qtextSMTP      = %d32-33 / %d35-91 / %d93-126
                  ; i.e., within a quoted string, any
                  ; ASCII graphic or space is permitted
                  ; without blackslash-quoting except
                  ; double-quote and the backslash itself.

   String         = Atom / Quoted-string




   RFC 5322 (since RFC 5321 never mentions atext rule)

   atext           =   ALPHA / DIGIT /    ; Printable US-ASCII
                       "!" / "#" /        ;  characters not including
                       "$" / "%" /        ;  specials.  Used for atoms.
                       "&" / "'" /
                       "*" / "+" /
                       "-" / "/" /
                       "=" / "?" /
                       "^" / "_" /
                       "`" / "{" /
                       "|" / "}" /
                       "~"
*/

func ParseAddress(address_str string) (*MailAddress, error) {
	address, err := mail.ParseAddress(address_str)
	if err != nil {
		return nil, err
	}

	index := strings.LastIndex(address.Address, "@")
	local := address.Address[0:index]
	domain := address.Address[index+1 : len(address.Address)]

	m := MailAddress{Name: address.Name, Local: local, Domain: domain}

	if valid, msg := m.Validate(); !valid {
		return nil, errors.New(msg)
	}

	return &m, nil

}

func (m *MailAddress) Validate() (bool, string) {
	/*
		 RFC 5321

			4.5.3.1.1.  Local-part

			   The maximum total length of a user name or other local-part is 64
			   octets.

			4.5.3.1.2.  Domain

			   The maximum total length of a domain name or number is 255 octets.

			4.5.3.1.3.  Path

			   The maximum total length of a reverse-path or forward-path is 256
			   octets (including the punctuation and element separators).
	*/
	if len(m.Local) > 64 {
		return false, "Local too long"
	}
	if len(m.Domain) > 253 {
		return false, "Domain too long"
	}
	if len(m.Domain)+len(m.Local) > 254 {
		return false, "MailAddress too long"
	}
	return true, ""
}

/*
   RFC 5321

   The maximum total length of a domain name or number is 255 octets.
*/


// ValidateDomainAddress will check if the sender's IP is authorized to send from the domain
func (m *MailAddress) ValidateDomainAddress(conn *conn) (bool, error) {
	
	// TODO
	// check for IP address
	ip := net.ParseIP(m.Domain)
	connAddr, ok := (conn.c.RemoteAddr()).(*net.TCPAddr)
	if !ok {
		return false, errors.New("Connection " + conn.c.RemoteAddr().String() + " isn't a tcp connection")
	}
	
	if ip != nil {
		// it's an IP
		if !ip.Equal(connAddr.IP) {
			return false, errors.New("IP in from(" + ip.String() + ") doesn't match real IP(" + connAddr.IP.String() + ")")
		}
	
	} else {
		// try to interpret is as a domain
		// Lookup A and AAAA records
		addresses, err := net.LookupIP(m.Domain)
		if err != nil {
			return false, err
		}
		for _,address := range addresses {
			if address.Equal(connAddr.IP) {
				return true, nil
			}
		}
	
		// Lookup SPF reocrds
		// TODO
	}
	
	return false, errors.New("End of non-void function")

	
}


// Check if m.Domain reverses to conn.
func (m *MailAddress) HasReverseDns(conn *conn) bool {
	// TODO
	// check for IP address
	ip := net.ParseIP(m.Domain)
	connAddr, ok := (conn.c.RemoteAddr()).(*net.TCPAddr)
	if !ok {
		log.Printf("    > Connection %s isn't a tcp connection", conn.c.RemoteAddr())
		return false
	}

	if ip != nil {
		// it's an IP
		if !ip.Equal(connAddr.IP) {
			log.Printf("    > IP in from(%s) doesn't match real IP(%s)", ip, connAddr.IP)
			return false
		}

	} else {
		// try to interpret is as a domain
		// check for rDNS of client IP
		domains, err := net.LookupAddr(connAddr.IP.String())
		if err != nil {
			log.Printf("    > rDNS lookup failed: %s", err)
			return false
		}

		if !stringInSlice(m.Domain, domains) {
			log.Printf("    > rDNS(%s) didn't match Domain(%s)", domains, m.Domain)
			return false
		}

		// if no rDNS match found, check for the SPF record
		// TODO
	}

	return true
}

// Check if we are m.Domain.
func (m *MailAddress) IsLocal(conn *conn) bool {
	// TODO: Check the domain for real :p
	return m.Domain == "gopistolet.be"
}

func stringInSlice(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
