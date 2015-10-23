package smtp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
)

type StatusCode uint32

// SMTP status codes
const (
	Ready             StatusCode = 220
	Closing           StatusCode = 221
	Ok                StatusCode = 250
	StartData         StatusCode = 354
	ShuttingDown      StatusCode = 421
	SyntaxError       StatusCode = 500
	SyntaxErrorParam  StatusCode = 501
	NotImplemented    StatusCode = 502
	BadSequence       StatusCode = 503
	AbortMail         StatusCode = 552
	NoValidRecipients StatusCode = 554
)

// ErrLtl Line too long error
var ErrLtl = errors.New("Line too long")

var ErrNoDelims = errors.New("Delimiters not found")

// ErrIncomplete Incomplete data error
var ErrIncomplete = errors.New("Incomplete data")

type UntillReader struct {
	Delims     []byte
	N          int
	R          *bufio.Reader
	delimsRead int
}

func (u *UntillReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	lr := io.LimitedReader{
		R: u.R,
		N: int64(u.N),
	}

	for u.N > 0 && u.delimsRead < len(u.Delims) && n < len(p) {
		nb, err := lr.Read(p[n : n+1])
		if nb == 1 {
			if u.Delims[u.delimsRead] == p[n] {
				u.delimsRead++
			} else {
				u.delimsRead = 0
			}

			u.N--
			n++

			if u.delimsRead == len(u.Delims) {
				break
			}
		}

		if err != nil {
			if err == io.EOF {
				if lr.N == 0 {
					return n, ErrLtl
				}
				// EOF but not all delims are read.
				return n, ErrNoDelims
			}
			return n, err
		}
	}

	if u.N == 0 {
		if u.delimsRead != len(u.Delims) {
			return n, ErrLtl
		}
	}

	if n == len(p) {
		if u.delimsRead != len(u.Delims) {
			return n, nil
		}
	}

	if u.delimsRead == len(u.Delims) {
		return n, io.EOF
	}

	return n, nil
}

// ReadUntill reads a string that ends with delims. It returns an error if more than maxBytes are read or no delims were found.
func ReadUntill(delims []byte, maxBytes int, r *bufio.Reader) (string, error) {
	ru := UntillReader{
		R:      r,
		N:      maxBytes,
		Delims: delims,
	}

	result := ""
	buffer := make([]byte, 64)
	for {
		n, err := ru.Read(buffer)
		result += string(buffer[0:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return result, err
		}
	}

	return result, nil
}

type LimitedReader struct {
	R     io.Reader // underlying reader
	N     int       // max bytes remaining
	Delim byte
}

func (l *LimitedReader) Read(p []byte) (int, error) {
	if l.N <= 0 {
		return 0, io.EOF
	}

	if len(p) > l.N {
		p = p[0:l.N]
	}

	bytesRead := 0
	buf := make([]byte, 1)
	for l.N > 0 && bytesRead < len(p) {
		n, err := l.R.Read(buf)

		if n > 0 {
			p[bytesRead] = buf[0]
			l.N -= n
			bytesRead += n
			if buf[0] == l.Delim {
				break
			}

		}

		if err != nil {
			return bytesRead, err
		}
	}

	return bytesRead, nil
}

const (
	MAX_LINE = 1000
)

// DataReader implements the reader that will read the data from a MAIL cmd
type DataReader struct {
	br *bufio.Reader
}

func NewDataReader(br *bufio.Reader) *DataReader {
	dr := &DataReader{
		br: br,
	}

	return dr
}

func (r *DataReader) Read(p []byte) (int, error) {
	dr := textproto.NewReader(r.br).DotReader()
	return dr.Read(p)
}

// Cmd All SMTP answers/commands should implement this interface.
type Cmd interface {
	fmt.Stringer
}

// Answer A raw SMTP answer. Used to send a status code + message.
type Answer struct {
	Status  StatusCode
	Message string
}

func (c Answer) String() string {
	return fmt.Sprintf("%d %s", c.Status, c.Message)
}

// MultiAnswer A multiline answer.
type MultiAnswer struct {
	Status   StatusCode
	Messages []string
}

func (c MultiAnswer) String() string {
	if len(c.Messages) == 0 {
		return fmt.Sprintf("%d", c.Status)
	}

	result := ""
	for i := 0; i < len(c.Messages)-1; i++ {
		result += fmt.Sprintf("%d-%s", c.Status, c.Messages[i])
		result += "\r\n"
	}

	result += fmt.Sprintf("%d %s", c.Status, c.Messages[len(c.Messages)-1])

	return result
}

// InvalidCmd is a known command with invalid arguments or syntax
type InvalidCmd struct {
	// The command
	Cmd  string
	Info string
}

func (c InvalidCmd) String() string {
	return fmt.Sprintf("%s %s", c.Cmd, c.Info)
}

// UnknownCmd is a command that is none of the other commands. i.e. not implemented
type UnknownCmd struct {
	// The command
	Cmd  string
	Line string
}

func (c UnknownCmd) String() string {
	return fmt.Sprintf("%s", c.Cmd)
}

type HeloCmd struct {
	Domain string
}

func (c HeloCmd) String() string {
	return ""
}

type EhloCmd struct {
	Domain string
}

func (c EhloCmd) String() string {
	return ""
}

type QuitCmd struct {
}

func (c QuitCmd) String() string {
	return ""
}

type MailCmd struct {
	From *MailAddress
}

func (c MailCmd) String() string {
	return ""
}

type RcptCmd struct {
	To *MailAddress
}

func (c RcptCmd) String() string {
	return ""
}

type DataCmd struct {
	Data []byte
	R    DataReader
}

func (c DataCmd) String() string {
	return ""
}

type RsetCmd struct {
}

func (c RsetCmd) String() string {
	return ""
}

type NoopCmd struct{}

func (c NoopCmd) String() string {
	return ""
}

// Not implemented because of security concerns
type VrfyCmd struct {
	Param string
}

func (c VrfyCmd) String() string {
	return ""
}

type ExpnCmd struct {
	ListName string
}

func (c ExpnCmd) String() string {
	return ""
}

type SendCmd struct{}

func (c SendCmd) String() string {
	return ""
}

type SomlCmd struct{}

func (c SomlCmd) String() string {
	return ""
}

type SamlCmd struct{}

func (c SamlCmd) String() string {
	return ""
}

// Protocol Used as communication layer so we can easily switch between a real socket
// and a test implementation.
type Protocol interface {
	// Send a SMTP command.
	Send(Cmd)
	// Receive a command(will block while waiting for it).
	// Returns false if there are no more commands left. Otherwise a command will be returned.
	// We need the bool because if we just return nil, the nil will also implement the empty interface...
	GetCmd() (*Cmd, error)
	// Close the connection.
	Close()
}

type MtaProtocol struct {
	c      net.Conn
	lr     *io.LimitedReader
	br     *bufio.Reader
	parser parser
}

// NewMtaProtocol Creates a protocol that works over a socket.
// the net.Conn parameter will be closed when done.
func NewMtaProtocol(c net.Conn) *MtaProtocol {
	proto := &MtaProtocol{
		c:      c,
		lr:     &io.LimitedReader{R: c, N: MAX_LINE},
		parser: parser{},
	}
	proto.br = bufio.NewReader(proto.lr)

	return proto
}

func (p *MtaProtocol) Send(c Cmd) {
	fmt.Fprintf(p.c, "%s\r\n", c)
}

func (p *MtaProtocol) SkipTillNewline() error {
	LIMIT := 1024
	for {

		p.lr.N = int64(LIMIT)
		_, err := p.br.ReadBytes('\n')
		if err == io.EOF {
			// EOF didn't come from the limitedreader.
			if p.lr.N > 0 {
				return err
			}

			// Could be from limitedreader but also from underlying reader.
			// If EOF came from limitedreader, it will be reset in the next iteration.
			// Otherwise we will get in the if above in the next iteration.
			continue
		}

		// Can't handle this error...
		if err != nil {
			return err
		}

		// No error, so we've read untill a newline.
		return nil
	}
}

// GetCmd returns the next command.
func (p *MtaProtocol) GetCmd() (*Cmd, error) {
	p.lr.N = int64(512)
	cmd, err := p.parser.ParseCommand(p.br)
	if err != nil {
		// Line too long.
		if err == io.EOF && p.lr.N == 0 {
			log.Printf("Line is too long")
			skipErr := p.SkipTillNewline()
			if skipErr != nil {
				return nil, skipErr
			}

			return nil, ErrLtl
		}
		log.Printf("Could not parse command: %v", err)
		return nil, err
	}

	return &cmd, nil
}

func (p *MtaProtocol) Close() {
	err := p.c.Close()
	if err != nil {
		log.Printf("Error while closing protocol: %v", err)
	}
}
