package authentication

import (
	"fmt"
	"strings"

	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gospf"
	"github.com/gopistolet/gospf/dns"
	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"
)

func New(c *mta.Config) *Authentication {
	return &Authentication{
		config: c,
	}
}

type Authentication struct {
	config    *mta.Config
	spfResult string
}

func (handler *Authentication) Handle(state *smtp.State) {
	// create SPF instance
	spf, err := gospf.New(state.From.GetDomain(), &dns.GoSPFDNS{})
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Infof("Could not create spf: %v", err)
		return
	}

	// check the given IP on that instance
	check, err := spf.CheckIP(state.Ip.String())
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Errorf("Error while checking ip in spf: %v", err)
		return
	}

	log.WithFields(log.Fields{
		"Ip":     state.Ip.String(),
		"Domain": state.From.GetDomain(),
	}).Info("SPF returned " + check)

	handler.spfResult = check

	handler.authenticationResultsHeader(state)
	handler.receivedSpfHeader(state)
}

func (handler *Authentication) authenticationResultsHeader(state *smtp.State) {
	/*
		header field is defined in RFC 5451 section 2.2 and RFC 7601
		Authentication-Results: receiver.example.org; spf=pass smtp.mailfrom=example.com;
	*/
	headerField := fmt.Sprintf("Authentication-Results: %s; spf=%s smtp.mailfrom=%s;\r\n", handler.config.Hostname, strings.ToLower(handler.spfResult), state.From.GetDomain())
	state.Data = append([]byte(headerField), state.Data...)
}

func (handler *Authentication) receivedSpfHeader(state *smtp.State) {
	/*
		RFC 4408

		header-field     = "Received-SPF:" [CFWS] result FWS [comment FWS]
						   [ key-value-list ] CRLF

		result           = "Pass" / "Fail" / "SoftFail" / "Neutral" /
						   "None" / "TempError" / "PermError"

		key-value-list   = key-value-pair *( ";" [CFWS] key-value-pair )
						   [";"]

		key-value-pair   = key [CFWS] "=" ( dot-atom / quoted-string )

		key              = "client-ip" / "envelope-from" / "helo" /
						   "problem" / "receiver" / "identity" /
							mechanism / "x-" name / name

		identity         = "mailfrom"   ; for the "MAIL FROM" identity
						   / "helo"     ; for the "HELO" identity
						   / name       ; other identities

		dot-atom         = <unquoted word as per [RFC2822]>
		quoted-string    = <quoted string as per [RFC2822]>
		comment          = <comment string as per [RFC2822]>
		CFWS             = <comment or folding white space as per [RFC2822]>
		FWS              = <folding white space as per [RFC2822]>
		CRLF             = <standard end-of-line token as per [RFC2822]>
	*/
	headerField := fmt.Sprintf("Received-SPF: %s client-ip=%s; receiver=%s;\r\n", handler.spfResult, state.Ip, handler.config.Hostname)
	state.Data = append([]byte(headerField), state.Data...)
}
