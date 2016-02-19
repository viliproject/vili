package auth

import (
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	saml "github.com/RobotsAndPencils/go-saml"
	"github.com/labstack/echo"

	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
)

// OktaConfig is the configuration for the OktaAuthService
type OktaConfig struct {
	Entrypoint string
	Issuer     string
	Cert       string
	Domain     string
}

// OktaAuthService is the auth service that uses Okta to authenticate users
type OktaAuthService struct {
	config   *OktaConfig
	certPath string
}

// InitOktaAuthService creates a new instance of OktaAuthService from the given
// config and sets it as the default auth service
func InitOktaAuthService(config *OktaConfig) error {
	certFile, err := ioutil.TempFile(os.TempDir(), "tmpgs")
	if err != nil {
		return err
	}
	certFile.WriteString(config.Cert)
	certFile.Close()
	log.Debugf("Wrote Okta cert to %s", certFile.Name())
	block, _ := pem.Decode([]byte(config.Cert))
	if block == nil {
		return fmt.Errorf("Failed to parse Okta cert")
	}
	service = &OktaAuthService{
		config:   config,
		certPath: certFile.Name(),
	}
	return nil
}

// AddHandlers implements the Service interface
func (s *OktaAuthService) AddHandlers(srv *server.Server) {
	srv.Echo().Get("/login", s.loginHandler)
	srv.Echo().Post("/login/callback", s.loginCallbackHandler)
	srv.Echo().Get("/login/failed", s.loginFailedHandler)
}

// Cleanup implements the Service interface
func (s *OktaAuthService) Cleanup() {
	os.Remove(s.certPath)
}

func (s *OktaAuthService) loginHandler(c *echo.Context) error {
	return c.Redirect(http.StatusFound, s.config.Entrypoint)
}

func (s *OktaAuthService) loginCallbackHandler(c *echo.Context) error {
	r := c.Request()
	err := r.ParseForm()
	if err != nil {
		return err
	}
	encodedXML := r.FormValue("SAMLResponse")

	if encodedXML == "" {
		c.String(http.StatusBadRequest, "SAMLResponse form value missing")
		return nil
	}
	bytesXML, err := base64.StdEncoding.DecodeString(encodedXML)
	if err != nil {
		c.String(http.StatusBadRequest, "SAMLResponse parse: "+err.Error())
		return nil
	}

	response := &saml.Response{}
	err = xml.Unmarshal(bytesXML, response)
	if err != nil {
		c.String(http.StatusBadRequest, "SAMLResponse parse: "+err.Error())
		return nil
	}

	err = s.Validate(response, bytesXML)
	if err != nil {
		c.String(http.StatusBadRequest, "SAMLResponse validation: "+err.Error())
		return nil
	}

	email := response.GetAttribute("email")
	if email == "" {
		c.String(http.StatusBadRequest, "SAML attribute identifier email missing")
		return nil
	}
	splitEmail := strings.Split(email, "@")
	if len(splitEmail) != 2 {
		c.String(http.StatusBadRequest, "SAML attribute identifier email is invalid")
		return nil
	}
	if splitEmail[1] != s.config.Domain {
		c.String(http.StatusBadRequest, "SAML attribute identifier email is not in the correct domain")
		return nil
	}

	err = session.Login(r, c.Response(), &session.User{
		Email:     email,
		Username:  splitEmail[0],
		FirstName: response.GetAttribute("firstName"),
		LastName:  response.GetAttribute("lastName"),
	})
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func (s *OktaAuthService) loginFailedHandler(c *echo.Context) error {
	return c.String(http.StatusOK, "Login Failed")
}

// Validate validates the SAML response
// taken from https://github.com/RobotsAndPencils/go-saml/blob/master/authnresponse.go#L49
func (s *OktaAuthService) Validate(r *saml.Response, originalBytes []byte) error {
	if r.Version != "2.0" {
		return errors.New("unsupported SAML Version")
	}

	if len(r.ID) == 0 {
		return errors.New("missing ID attribute on SAML Response")
	}

	if len(r.Assertion.ID) == 0 {
		return errors.New("no Assertions")
	}

	if len(r.Signature.SignatureValue.Value) == 0 {
		return errors.New("no signature")
	}

	if r.Assertion.Subject.SubjectConfirmation.Method != "urn:oasis:names:tc:SAML:2.0:cm:bearer" {
		return errors.New("assertion method exception")
	}

	err := verifyResponseSignature(originalBytes, s.certPath)
	if err != nil {
		return err
	}

	// CHECK TIMES
	expires := r.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.NotOnOrAfter
	notOnOrAfter, e := time.Parse(time.RFC3339, expires)
	if e != nil {
		return e
	}
	if notOnOrAfter.Before(time.Now()) {
		return errors.New("assertion has expired on: " + expires)
	}

	return nil
}
