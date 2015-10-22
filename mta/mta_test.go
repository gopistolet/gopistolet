package mta

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gopistolet/gopistolet/smtp"
	c "github.com/smartystreets/goconvey/convey"
)

// Dummy mail handler
func dummyHandler(*State) {

}

type testProtocol struct {
	t *testing.T
	// Goconvey context so it works in a different goroutine
	ctx     c.C
	cmds    []smtp.Cmd
	answers []interface{}
}

func getMailWithoutError(a string) *smtp.MailAddress {
	addr, _ := smtp.ParseAddress(a)
	return &addr
}

func (p *testProtocol) Send(cmd smtp.Cmd) {
	p.ctx.So(len(p.answers), c.ShouldBeGreaterThan, 0)

	//c.Printf("RECEIVED: %#v\n", cmd)

	answer := p.answers[0]
	p.answers = p.answers[1:]

	if cmdA, ok := cmd.(smtp.Answer); ok {
		p.ctx.So(answer, c.ShouldHaveSameTypeAs, cmdA)
		cmdE, _ := answer.(smtp.Answer)
		p.ctx.So(cmdE.Status, c.ShouldEqual, cmdA.Status)
	} else if cmdA, ok := cmd.(smtp.MultiAnswer); ok {
		p.ctx.So(answer, c.ShouldHaveSameTypeAs, cmdA)
		cmdE, _ := answer.(smtp.MultiAnswer)
		p.ctx.So(cmdE.Status, c.ShouldEqual, cmdA.Status)
	} else {
		p.t.Fatalf("Answer should be Answer or MultiAnswer")
	}
}

func (p *testProtocol) GetCmd() (*smtp.Cmd, error) {
	p.ctx.So(len(p.cmds), c.ShouldBeGreaterThan, 0)

	cmd := p.cmds[0]
	p.cmds = p.cmds[1:]

	if cmd == nil {
		return nil, errors.New("No more commands")
	}

	//c.Printf("SENDING: %#v\n", cmd)
	return &cmd, nil
}

func (p *testProtocol) Close() {
	// Did not expect connection to be closed, got more commands
	p.ctx.So(len(p.cmds), c.ShouldBeLessThanOrEqualTo, 0)

	// Did not expect connection to be closed, need more answers
	p.ctx.So(len(p.answers), c.ShouldBeLessThanOrEqualTo, 0)
}

// Tests answers for HELO,EHLO and QUIT
func TestAnswersHeloQuit(t *testing.T) {
	cfg := Config{
		Hostname: "home.sweet.home",
	}

	mta := New(cfg, HandlerFunc(dummyHandler))
	if mta == nil {
		t.Fatal("Could not create mta server")
	}

	c.Convey("Testing answers for HELO and QUIT.", t, func(ctx c.C) {

		// Test connection with HELO and QUIT
		proto := &testProtocol{
			t:   t,
			ctx: ctx,
			cmds: []smtp.Cmd{
				smtp.HeloCmd{
					Domain: "some.sender",
				},
				smtp.QuitCmd{},
			},
			answers: []interface{}{
				smtp.Answer{
					Status:  smtp.Ready,
					Message: cfg.Hostname + " Service Ready",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: cfg.Hostname,
				},
				smtp.Answer{
					Status:  smtp.Closing,
					Message: "Bye!",
				},
			},
		}
		mta.HandleClient(proto)
	})

	c.Convey("Testing answers for HELO and close connection.", t, func(ctx c.C) {
		proto := &testProtocol{
			t:   t,
			ctx: ctx,
			cmds: []smtp.Cmd{
				smtp.HeloCmd{
					Domain: "some.sender",
				},
				nil,
			},
			answers: []interface{}{
				smtp.Answer{
					Status:  smtp.Ready,
					Message: cfg.Hostname + " Service Ready",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: cfg.Hostname,
				},
			},
		}
		mta.HandleClient(proto)

	})

	c.Convey("Testing answers for EHLO and QUIT.", t, func(ctx c.C) {

		// Test connection with EHLO and QUIT
		proto := &testProtocol{
			t:   t,
			ctx: ctx,
			cmds: []smtp.Cmd{
				smtp.EhloCmd{
					Domain: "some.sender",
				},
				smtp.QuitCmd{},
			},
			answers: []interface{}{
				smtp.Answer{
					Status:  smtp.Ready,
					Message: cfg.Hostname + " Service Ready",
				},
				smtp.MultiAnswer{
					Status: smtp.Ok,
				},
				smtp.Answer{
					Status:  smtp.Closing,
					Message: "Bye!",
				},
			},
		}
		mta.HandleClient(proto)
	})

	c.Convey("Testing answers for EHLO and close connection.", t, func(ctx c.C) {
		proto := &testProtocol{
			t:   t,
			ctx: ctx,
			cmds: []smtp.Cmd{
				smtp.EhloCmd{
					Domain: "some.sender",
				},
				nil,
			},
			answers: []interface{}{
				smtp.Answer{
					Status:  smtp.Ready,
					Message: cfg.Hostname + " Service Ready",
				},
				smtp.MultiAnswer{
					Status: smtp.Ok,
				},
			},
		}
		mta.HandleClient(proto)

	})
}

// Test answers if we are given a sequence of MAIL,RCPT,DATA commands.
func TestMailAnswersCorrectSequence(t *testing.T) {
	cfg := Config{
		Hostname: "home.sweet.home",
	}

	mta := New(cfg, HandlerFunc(dummyHandler))
	if mta == nil {
		t.Fatal("Could not create mta server")
	}

	c.Convey("Testing correct sequence of MAIL,RCPT,DATA commands.", t, func(ctx c.C) {

		proto := &testProtocol{
			t:   t,
			ctx: ctx,
			cmds: []smtp.Cmd{
				smtp.HeloCmd{
					Domain: "some.sender",
				},
				smtp.MailCmd{
					From: getMailWithoutError("someone@somewhere.test"),
				},
				smtp.RcptCmd{
					To: getMailWithoutError("guy1@somewhere.test"),
				},
				smtp.RcptCmd{
					To: getMailWithoutError("guy2@somewhere.test"),
				},
				smtp.DataCmd{
					R: *smtp.NewDataReader(bytes.NewReader([]byte("Some test email\n.\n"))),
				},
				smtp.QuitCmd{},
			},
			answers: []interface{}{
				smtp.Answer{
					Status:  smtp.Ready,
					Message: cfg.Hostname + " Service Ready",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: cfg.Hostname,
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: "OK",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: "OK",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: "OK",
				},
				smtp.Answer{
					Status:  smtp.StartData,
					Message: "OK",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: "OK",
				},
				smtp.Answer{
					Status:  smtp.Closing,
					Message: "Bye!",
				},
			},
		}
		mta.HandleClient(proto)
	})

	c.Convey("Testing wrong sequence of MAIL,RCPT,DATA commands.", t, func(ctx c.C) {
		c.Convey("RCPT before MAIL", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.HeloCmd{
						Domain: "some.sender",
					},
					smtp.RcptCmd{
						To: getMailWithoutError("guy1@somewhere.test"),
					},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: cfg.Hostname,
					},
					smtp.Answer{
						Status:  smtp.BadSequence,
						Message: "Need mail before RCPT",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

		c.Convey("DATA before MAIL", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.HeloCmd{
						Domain: "some.sender",
					},
					smtp.DataCmd{},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: cfg.Hostname,
					},
					smtp.Answer{
						Status:  smtp.BadSequence,
						Message: "Need mail before DATA",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

		c.Convey("DATA before RCPT", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.HeloCmd{
						Domain: "some.sender",
					},
					smtp.MailCmd{
						From: getMailWithoutError("guy@somewhere.test"),
					},
					smtp.DataCmd{},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: cfg.Hostname,
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.BadSequence,
						Message: "Need RCPT before DATA",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

		c.Convey("Multiple MAIL commands.", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.HeloCmd{
						Domain: "some.sender",
					},
					smtp.MailCmd{
						From: getMailWithoutError("guy@somewhere.test"),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("someone@somewhere.test"),
					},
					smtp.MailCmd{
						From: getMailWithoutError("someguy@somewhere.test"),
					},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: cfg.Hostname,
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.BadSequence,
						Message: "Sender already specified",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

	})
}

// Tests if our state gets reset correctly.
func TestReset(t *testing.T) {
	cfg := Config{
		Hostname: "home.sweet.home",
	}

	mta := New(cfg, HandlerFunc(dummyHandler))
	if mta == nil {
		t.Fatal("Could not create mta server")
	}

	c.Convey("Testing reset", t, func(ctx c.C) {

		c.Convey("Test reset after sending mail.", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.HeloCmd{
						Domain: "some.sender",
					},
					smtp.MailCmd{
						From: getMailWithoutError("someone@somewhere.test"),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("guy1@somewhere.test"),
					},
					smtp.DataCmd{
						R: *smtp.NewDataReader(bytes.NewReader([]byte("Some email content\n.\n"))),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("someguy@somewhere.test"),
					},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: cfg.Hostname,
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.StartData,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.BadSequence,
						Message: "Need mail before RCPT",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

		c.Convey("Manually reset", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.HeloCmd{
						Domain: "some.sender",
					},
					smtp.MailCmd{
						From: getMailWithoutError("someone@somewhere.test"),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("guy1@somewhere.test"),
					},
					smtp.RsetCmd{},
					smtp.MailCmd{
						From: getMailWithoutError("someone@somewhere.test"),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("guy1@somewhere.test"),
					},
					smtp.DataCmd{
						R: *smtp.NewDataReader(bytes.NewReader([]byte("Some email\n.\n"))),
					},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: cfg.Hostname,
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.StartData,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

		// EHLO should reset state.
		c.Convey("Reset with EHLO", func() {
			proto := &testProtocol{
				t:   t,
				ctx: ctx,
				cmds: []smtp.Cmd{
					smtp.EhloCmd{
						Domain: "some.sender",
					},
					smtp.MailCmd{
						From: getMailWithoutError("someone@somewhere.test"),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("guy1@somewhere.test"),
					},
					smtp.EhloCmd{
						Domain: "some.sender",
					},
					smtp.MailCmd{
						From: getMailWithoutError("someone@somewhere.test"),
					},
					smtp.RcptCmd{
						To: getMailWithoutError("guy1@somewhere.test"),
					},
					smtp.DataCmd{
						R: *smtp.NewDataReader(bytes.NewReader([]byte("Some email\n.\n"))),
					},
					smtp.QuitCmd{},
				},
				answers: []interface{}{
					smtp.Answer{
						Status:  smtp.Ready,
						Message: cfg.Hostname + " Service Ready",
					},
					smtp.MultiAnswer{
						Status: smtp.Ok,
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.MultiAnswer{
						Status: smtp.Ok,
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.StartData,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Ok,
						Message: "OK",
					},
					smtp.Answer{
						Status:  smtp.Closing,
						Message: "Bye!",
					},
				},
			}
			mta.HandleClient(proto)
		})

	})
}

// Tests answers if we send an unknown command.
func TestAnswersUnknownCmd(t *testing.T) {
	cfg := Config{
		Hostname: "home.sweet.home",
	}

	mta := New(cfg, HandlerFunc(dummyHandler))
	if mta == nil {
		t.Fatal("Could not create mta server")
	}

	c.Convey("Testing answers for unknown cmds.", t, func(ctx c.C) {
		proto := &testProtocol{
			t:   t,
			ctx: ctx,
			cmds: []smtp.Cmd{
				smtp.HeloCmd{
					Domain: "some.sender",
				},
				smtp.UnknownCmd{
					Cmd: "someinvalidcmd",
				},
				smtp.QuitCmd{},
			},
			answers: []interface{}{
				smtp.Answer{
					Status:  smtp.Ready,
					Message: cfg.Hostname + " Service Ready",
				},
				smtp.Answer{
					Status:  smtp.Ok,
					Message: "OK",
				},
				smtp.Answer{
					Status:  smtp.SyntaxError,
					Message: cfg.Hostname,
				},
				smtp.Answer{
					Status:  smtp.Closing,
					Message: "Bye!",
				},
			},
		}
		mta.HandleClient(proto)
	})
}
