// Package auth xmlsec.go is taken from:
// https://github.com/RobotsAndPencils/go-saml/blob/master/xmlsec.go
package auth

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	xmlResponseID = "urn:oasis:names:tc:SAML:2.0:protocol:Response"
)

func verifyResponseSignature(xml []byte, publicCertPath string) error {
	samlXmlsecInput, err := ioutil.TempFile(os.TempDir(), "tmpgs")
	if err != nil {
		return err
	}
	samlXmlsecInput.Write(xml)
	samlXmlsecInput.Close()
	defer os.Remove(samlXmlsecInput.Name())

	_, err = exec.Command(
		"xmlsec1", "--verify",
		"--pubkey-cert-pem", publicCertPath,
		"--id-attr:ID", xmlResponseID,
		samlXmlsecInput.Name(),
	).CombinedOutput()
	if err != nil {
		return errors.New("error verifing signature: " + err.Error())
	}
	return nil
}
