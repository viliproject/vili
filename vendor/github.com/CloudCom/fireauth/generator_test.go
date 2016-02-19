package fireauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	if gen := New("foo"); gen == nil {
		t.Fatal("generator should not be nil")
	}
}

func TestCreateTokenData(t *testing.T) {
	gen := New("foo")
	data := Data{"uid": "1"}

	token, err := gen.CreateToken(data, nil)
	if err != nil {
		t.Fatal(err)
	}

	tokenParts := strings.Split(token, TokenSep)
	if len(tokenParts) != 3 {
		t.Fatal("token is not composed correctly")
	}

	bytes, err := base64.URLEncoding.DecodeString(tokenParts[1] + "==")
	if err != nil {
		t.Fatal(err)
	}

	claim := struct {
		Version  int   `json:"v"`
		Data     Data  `json:"d"`
		IssuedAt int64 `json:"iat"`
	}{}
	if err := json.Unmarshal(bytes, &claim); err != nil {
		t.Fatal(err)
	}

	if claim.Version != Version {
		t.Fatalf("Expected: %d\nActual: %d", Version, claim.Version)
	}

	if !reflect.DeepEqual(data, claim.Data) {
		t.Fatalf("auth data is not the same.Expected: %s\nActual: %s", data, claim.Data)
	}
}

func TestCreateTokenFailure(t *testing.T) {
	if _, err := New("foo").CreateToken(nil, nil); err == nil {
		t.Fatal("CreateToken without data nor option should fail")
	}
	if _, err := New("foo").CreateToken(Data{}, nil); err == nil {
		t.Fatal("CreateToken with invalid data should fail")
	}
	ch := make(chan struct{})
	defer close(ch)
	if _, err := New("foo").CreateToken(Data{"uid": "1234", "invalid": ch}, &Option{}); err == nil {
		t.Fatal("Invalid data types should make the token creation fail")
	}
}

func TestCreateTokenAdminNoData(t *testing.T) {
	if _, err := New("foo").CreateToken(nil, &Option{Admin: true}); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTokenTooLong(t *testing.T) {
	if _, err := New("foo").CreateToken(Data{"uid": "1", "bigKey": randData(t, 1024)}, nil); err == nil {
		t.Fatal("Token too long should have failed")
	}
}

func randData(t *testing.T, size int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatal(err)
	}
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func TestValidate(t *testing.T) {
	if err := validate(nil, false); err != ErrNoUIDKey {
		t.Fatalf("Unexpected error. Expected: %s, Got: %v", ErrNoUIDKey, err)
	}
	if err := validate(Data{"uid": 42}, true); err != ErrUIDNotString {
		t.Fatalf("Unexpected error. Expected: %s, Got: %v", ErrUIDNotString, err)
	}
	if err := validate(Data{"uid": strings.Repeat(" ", MaxUIDLen+1)}, true); err != ErrUIDTooLong {
		t.Fatalf("Unexpected error. Expected: %s, Got: %v", ErrUIDTooLong, err)
	}
	// No uid in admin mode should not fail.
	if err := validate(Data{}, true); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateClaim(t *testing.T) {
	buf, err := generateClaim(Data{"uid": "42"}, &Option{Admin: true}, 0)
	if err != nil {
		t.Fatal(err)
	}

	if expect, got := `{"admin":true,"v":0,"d":{"uid":"42"},"iat":0}`, string(buf); expect != got {
		t.Fatalf("Unexpected claim\nExpect:\t%s\nGot:\t%s", expect, got)
	}
}

func TestSign(t *testing.T) {
	g := struct {
		key string
		in  string
		out string
	}{
		key: "you cannot see me",
		in:  "the winter is comming",
		out: "ytb5HiGUKtRhJg02DXS-serVBwbxud08FFNcx6dty78",
		// use the following code segmentation to generate the output
		// h := hmac.New(sha256.New, []byte(key))
		// h.Write([]byte(in))
		// out := encode(h.Sum(nil))
	}

	got := sign(g.in, g.key)
	if g.out != got {
		t.Fatalf("expect:%v, but got:%v", g.out, got)
	}
}
