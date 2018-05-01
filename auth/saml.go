package auth

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/util"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// SAMLConfig is the configuration for the SAMLAuthService
type SAMLConfig struct {
	URL            string
	IDPMetadataURL string
	SPCert         string
	SPPrivateKey   string
}

// SAMLAuthService is the auth service that uses SAML to authenticate users
type SAMLAuthService struct {
	config         *SAMLConfig
	samlMiddleware *samlsp.Middleware
}

var jwtSigningMethod = jwt.SigningMethodHS256

// InitSAMLAuthService creates a new instance of SAMLAuthService from the given
// config and sets it as the default auth service
func InitSAMLAuthService(config *SAMLConfig) error {
	opts := samlsp.Options{
		ForceAuthn: false,
	}

	keyPair, err := tls.X509KeyPair(
		[]byte(config.SPCert),
		[]byte(config.SPPrivateKey),
	)
	if err != nil {
		return fmt.Errorf("Failed to parse SP SAML keypair")
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return fmt.Errorf("Failed to parse SP SAML keypair certificate")
	}
	opts.Key = keyPair.PrivateKey.(*rsa.PrivateKey)
	opts.Certificate = keyPair.Leaf

	idpMetadataURL, err := url.Parse(config.IDPMetadataURL)
	if err != nil {
		return fmt.Errorf("Failed to parse IDP metadata url")
	}
	opts.IDPMetadataURL = idpMetadataURL

	samlMiddleware, err := samlsp.New(opts)
	if err != nil {
		return fmt.Errorf("Failed to create SAML client")
	}
	metadataURL, err := url.Parse(config.URL)
	if err != nil {
		return fmt.Errorf("Failed to add metadata URL")
	}
	samlMiddleware.ServiceProvider.MetadataURL = *metadataURL
	acsURL, err := url.Parse(config.URL + "/login/callback")
	if err != nil {
		return fmt.Errorf("Failed to add ACS URL")
	}
	samlMiddleware.ServiceProvider.AcsURL = *acsURL

	service = &SAMLAuthService{
		config:         config,
		samlMiddleware: samlMiddleware,
	}
	return nil
}

// AddHandlers implements the Service interface
func (s *SAMLAuthService) AddHandlers(srv *server.Server) {
	srv.Echo().GET("/login", s.loginHandler)
	srv.Echo().POST("/login/callback", s.loginCallbackHandler)
	srv.Echo().GET("/login/failed", s.loginFailedHandler)
}

// Cleanup implements the Service interface
func (s *SAMLAuthService) Cleanup() {
}

func (s *SAMLAuthService) loginHandler(c echo.Context) error {
	binding := saml.HTTPRedirectBinding
	bindingLocation := s.samlMiddleware.ServiceProvider.GetSSOBindingLocation(binding)

	req, err := s.samlMiddleware.ServiceProvider.MakeAuthenticationRequest(bindingLocation)
	if err != nil {
		return err
	}

	relayState := util.RandString(80)
	signedState, err := s.getJWTToken(req.ID, c.Request().URL.Query().Get("redirect"))
	if err != nil {
		return err
	}

	s.samlMiddleware.ClientState.SetState(c.Response(), c.Request(), relayState, signedState)
	return c.Redirect(http.StatusFound, req.Redirect(relayState).String())
}

// getJWTToken gets a signed JWT token for the given id and redirect URI
func (s *SAMLAuthService) getJWTToken(id, redirectURI string) (string, error) {
	secretBlock := x509.MarshalPKCS1PrivateKey(s.samlMiddleware.ServiceProvider.Key)
	state := jwt.New(jwtSigningMethod)
	claims := state.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["uri"] = redirectURI
	return state.SignedString(secretBlock)
}

func (s *SAMLAuthService) loginCallbackHandler(c echo.Context) error {
	r := c.Request()
	if err := r.ParseForm(); err != nil {
		return err
	}

	// check assertion
	assertion, err := s.samlMiddleware.ServiceProvider.ParseResponse(r, s.getPossibleRequestIDs(c))
	if err != nil {
		switch e := err.(type) {
		case *saml.InvalidResponseError:
			log.
				WithError(e.PrivateErr).
				// WithField("response", e.Response).
				Info("invalid response error")
		}
		return s.loginFailedHandler(c)
	}
	a := Assertion(*assertion)

	secretBlock := x509.MarshalPKCS1PrivateKey(s.samlMiddleware.ServiceProvider.Key)

	var redirectURI string
	if relayState := r.FormValue("RelayState"); relayState != "" {
		stateValue := s.samlMiddleware.ClientState.GetState(r, relayState)
		if stateValue == "" {
			log.Infof("cannot find corresponding state: %s", relayState)
			return s.loginFailedHandler(c)
		}

		jwtParser := jwt.Parser{
			ValidMethods: []string{jwtSigningMethod.Name},
		}
		state, err := jwtParser.Parse(stateValue, func(t *jwt.Token) (interface{}, error) {
			return secretBlock, nil
		})
		if err != nil || !state.Valid {
			log.WithError(err).WithField("stateValue", stateValue).Infof("Cannot decode state JWT")
			return s.loginFailedHandler(c)
		}
		claims := state.Claims.(jwt.MapClaims)
		redirectURI = claims["uri"].(string)

		// delete the cookie
		s.samlMiddleware.ClientState.DeleteState(c.Response(), r, relayState)
	}
	if redirectURI == "" {
		redirectURI = "/"
	}

	email := a.GetAttributeValue("email")
	if email == "" {
		return c.String(http.StatusBadRequest, "SAML attribute identifier email missing")
	}
	splitEmail := strings.Split(email, "@")
	if len(splitEmail) != 2 {
		return c.String(http.StatusBadRequest, "SAML attribute identifier email is invalid")
	}

	user := &session.User{
		Email:     email,
		Username:  splitEmail[0],
		FirstName: a.GetAttributeValue("firstName"),
		LastName:  a.GetAttributeValue("lastName"),
		Groups:    a.GetAttributeValues("groups"),
	}

	err = session.Login(r, c.Response(), user)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, redirectURI)
}

func (s *SAMLAuthService) getPossibleRequestIDs(c echo.Context) []string {
	// allow IDP initiated requests, with an empty request id
	rv := []string{""}
	for _, value := range s.samlMiddleware.ClientState.GetStates(c.Request()) {
		jwtParser := jwt.Parser{
			ValidMethods: []string{jwtSigningMethod.Name},
		}
		token, err := jwtParser.Parse(value, func(t *jwt.Token) (interface{}, error) {
			secretBlock := x509.MarshalPKCS1PrivateKey(s.samlMiddleware.ServiceProvider.Key)
			return secretBlock, nil
		})
		if err != nil || !token.Valid {
			log.Warnf("invalid token %s", err)
			continue
		}
		claims := token.Claims.(jwt.MapClaims)
		rv = append(rv, claims["id"].(string))
	}

	return rv
}

func (s *SAMLAuthService) loginFailedHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Login Failed")
}

// Assertion is a wrapper around saml.Assertion
type Assertion saml.Assertion

// GetAttributeValues returns the values for the given key
func (a *Assertion) GetAttributeValues(key string) (values []string) {
	for _, statement := range a.AttributeStatements {
		for _, attribute := range statement.Attributes {
			if attribute.Name == key {
				for _, attributeValue := range attribute.Values {
					values = append(values, attributeValue.Value)
				}
			}
		}
	}
	return
}

// GetAttributeValue returns the first value for the given key
func (a *Assertion) GetAttributeValue(key string) string {
	values := a.GetAttributeValues(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
