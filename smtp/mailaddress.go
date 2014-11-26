package smtp

import (
    "fmt"
)

type MailAddress struct {
    Local	string
    Domain	string
}

func (m *MailAddress) String() string {
    return fmt.Sprintf("%s@%s", m.Local, m.Domain)
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
    */
func (m *MailAddress) Validate() bool {
    return true
}

/*
    RFC 5321

    The maximum total length of a domain name or number is 255 octets.
*/



// ValidateFrom will check if the email address is valid
// and if the email domain/address matches the clients remote address
func (m *MailAddress) ValidateFrom(conn *Conn) bool {
    // TODO
    return true
}


// ValidateTo will check if the recepient email address is valid
func (m *MailAddress) ValidateTo(conn *Conn) bool {
    // TODO
    return true
}