package auth

import (
	"encoding/xml"

	saml "github.com/RobotsAndPencils/go-saml"
)

// structs have been taken from https://github.com/RobotsAndPencils/go-saml
// and modified to fit our needs
// most notable making Attribute.AttributeValue an array for group attributes
// example -
//  <saml2:Attribute Name="groups" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified">
//    <saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Airware Data Analysts</saml2:AttributeValue>
//    <saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Airware Administrators</saml2:AttributeValue>
//  </saml2:Attribute>

// Response is the struct for the SAML response
type Response struct {
	XMLName      xml.Name
	SAMLP        string `xml:"xmlns:samlp,attr"`
	SAML         string `xml:"xmlns:saml,attr"`
	SAMLSIG      string `xml:"xmlns:samlsig,attr"`
	Destination  string `xml:"Destination,attr"`
	ID           string `xml:"ID,attr"`
	Version      string `xml:"Version,attr"`
	IssueInstant string `xml:"IssueInstant,attr"`
	InResponseTo string `xml:"InResponseTo,attr"`

	Assertion Assertion      `xml:"Assertion"`
	Signature saml.Signature `xml:"Signature"`
	Issuer    saml.Issuer    `xml:"Issuer"`
	Status    saml.Status    `xml:"Status"`

	originalString string
}

// Assertion is the struct for the assertion in the SAML response
type Assertion struct {
	XMLName            xml.Name
	ID                 string      `xml:"ID,attr"`
	Version            string      `xml:"Version,attr"`
	XS                 string      `xml:"xmlns:xs,attr"`
	XSI                string      `xml:"xmlns:xsi,attr"`
	SAML               string      `xml:"saml,attr"`
	IssueInstant       string      `xml:"IssueInstant,attr"`
	Issuer             saml.Issuer `xml:"Issuer"`
	Subject            saml.Subject
	Conditions         saml.Conditions
	AttributeStatement AttributeStatement
}

// Attribute is the struct for all the values in the SAML assertion
type Attribute struct {
	XMLName        xml.Name
	Name           string `xml:",attr"`
	FriendlyName   string `xml:",attr"`
	NameFormat     string `xml:",attr"`
	AttributeValue []AttributeValue
}

// AttributeValue is the struct for the values in an Attribute
type AttributeValue struct {
	XMLName xml.Name
	Type    string `xml:"xsi:type,attr"`
	Value   string `xml:",innerxml"`
}

// AttributeStatement is the struct for all the attributes that come back in the SAML Assertion
type AttributeStatement struct {
	XMLName    xml.Name
	Attributes []Attribute `xml:"Attribute"`
}

// GetAttribute by Name or by FriendlyName. Return blank string if not found
func (r *Response) GetAttribute(name string) string {
	for _, attr := range r.Assertion.AttributeStatement.Attributes {
		if attr.Name == name || attr.FriendlyName == name {
			return attr.AttributeValue[0].Value
		}
	}
	return ""
}

// GetGroupAttribute by Name. Return Attribute if found
func (r *Response) GetGroupAttribute(groupName string) *Attribute {
	for _, attr := range r.Assertion.AttributeStatement.Attributes {
		if attr.Name == groupName {
			return &attr
		}
	}
	return nil
}
