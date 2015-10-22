package smtp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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

// ErrIncomplete Incomplete data error
var ErrIncomplete = errors.New("Incomplete data")

// ReadUntill reads a string that ends with delims. It returns an error if more than maxBytes are read or no delims were found.
func ReadUntill(delims []byte, maxBytes int, r io.Reader) (string, error) {
	delimsRead := 0
	bytesRead := 0
	result := ""

	buffer := make([]byte, 1)
	for {
		if bytesRead >= maxBytes {
			break
		}

		if delimsRead >= len(delims) {
			break
		}

		n, err := r.Read(buffer)
		if n > 0 {
			//fmt.Printf("Read: %v\n", buffer)
			bytesRead += n
			if bytesRead > maxBytes {
				panic("Can't happen...")
			}

			for _, b := range buffer {
				if b == delims[delimsRead] {
					delimsRead++
				} else {
					delimsRead = 0
				}
			}

			result += string(buffer[0:n])
		}

		if err != nil {
			return result, err
		}
	}

	if bytesRead > maxBytes {
		return result, ErrLtl
	}

	// Only way we get here is:
	// bytesRead == maxBytes     => if not all delims are read now, then line will be too long.
	// bytesRead < maxBytes      => only happens if delimsRead >= len(delims), which is ok.
	if delimsRead != len(delims) {
		return result, ErrLtl
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
	r      io.Reader
	buffer []byte
}

func NewDataReader(r io.Reader) *DataReader {
	dr := &DataReader{
		r:      r,
		buffer: make([]byte, 0, MAX_LINE),
	}

	return dr
}

func (r *DataReader) Read(p []byte) (int, error) {
	var n int = 0

	if len(r.buffer) > 0 {
		n = copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		return n, nil
	}

	limited := &LimitedReader{
		R:     r.r,
		N:     MAX_LINE + 1,
		Delim: '\n',
	}

	br := bufio.NewReader(limited)

	line, err := br.ReadBytes('\n')
	lineLen := len(line)
	if lineLen > 0 && line[len(line)-1] != '\n' {
		buf := make([]byte, 1)

		for n, err := r.r.Read(buf); ; {
			if n > 0 {
				if buf[0] == '\n' {
					break
				}
			}

			if err != nil {
				break
			}

			n, err = r.r.Read(buf)
		}
	}
	fmt.Printf("Read %d bytes\n", lineLen)

	if bytes.Compare(line, []byte(".\r\n")) == 0 ||
		bytes.Compare(line, []byte(".\r")) == 0 ||
		bytes.Compare(line, []byte(".\n")) == 0 {

		return 0, io.EOF
	} else if lineLen > 2 && line[0] == '.' {
		line = line[1:]
		lineLen--
	}

	if lineLen > MAX_LINE {
		return 0, ErrLtl
	}

	n = copy(p, line)
	r.buffer = r.buffer[0 : lineLen-n]
	copy(r.buffer, line[n:])

	if err == io.EOF {
		return 0, ErrIncomplete
	}

	return n, nil
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
	br     *bufio.Reader
	parser parser
}

// NewMtaProtocol Creates a protocol that works over a socket.
// the net.Conn parameter will be closed when done.
func NewMtaProtocol(c net.Conn) *MtaProtocol {
	proto := &MtaProtocol{
		c:      c,
		br:     bufio.NewReader(c),
		parser: parser{},
	}

	return proto
}

func (p *MtaProtocol) Send(c Cmd) {
	fmt.Fprintf(p.c, "%s\r\n", c)
}

// GetCmd returns the next command.
func (p *MtaProtocol) GetCmd() (*Cmd, error) {
	cmd, err := p.parser.ParseCommand(p.br)
	if err != nil {
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
