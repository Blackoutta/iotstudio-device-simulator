package simulator

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"net/url"
)

func GenerateSasToken(a AuthMessage, pid, dname, dk string) (string, error) {
	res := fmt.Sprintf("products/%v/devices/%v", pid, dname)
	// make raw auth message
	raw := fmt.Sprintf("%v\n%v\n%v\n%v", a.Et, a.Method, res, a.Version)
	sign, err := GetSign(raw, a.Method, dk)
	if err != nil {
		return "", err
	}

	st := fmt.Sprintf("version=%v&res=%v&sign=%v&et=%v&method=%v",
		a.Version, res, sign, a.Et, a.Method)
	return st, err

}

func AuthJson(a AuthMessage, pid, dname, dk string) ([]byte, error) {
	st, err := GenerateSasToken(a, pid, dname, dk)
	if err != nil {
		return nil, err
	}

	j := struct {
		Lt int64  `json:"lt,omitempty"`
		St string `json:"st"`
	}{
		Lt: a.Lt,
		St: st,
	}

	buffer := &bytes.Buffer{}
	je := json.NewEncoder(buffer)
	je.SetEscapeHTML(false)
	je.Encode(j)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func GetSign(rawMsg string, method string, dk string) (string, error) {
	// Decode device key
	decodedDeviceKey, err := base64.StdEncoding.DecodeString(dk)
	if err != nil {
		return "", err
	}
	// encrypt auth message with device key
	var h hash.Hash
	switch method {
	case AUTH_MD5:
		h = hmac.New(md5.New, []byte(decodedDeviceKey))
	case AUTH_SHA1:
		h = hmac.New(sha1.New, []byte(decodedDeviceKey))
	case AUTH_SHA256:
		h = hmac.New(sha256.New, []byte(decodedDeviceKey))
	}
	_, err = h.Write([]byte(rawMsg))
	if err != nil {
		return "", err
	}
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	sign = url.QueryEscape(sign)
	return sign, nil
}

type AuthMessage struct {
	Lt int64
	SasToken
}

type SasToken struct {
	Version string
	Et      int64
	Method  string
}
