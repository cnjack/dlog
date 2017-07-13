package dlog

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	AUTHORIZATION       = "Authorization"
	CONTENT_TYPE        = "Content-Type"
	CONTENT_LENTH       = "Content-Length"
	CONTENT_MD5         = "Content-MD5"
	DATE                = "Date"
	HOST                = "Host"
	LOG_VERSION         = "x-log-apiversion"
	LOG_SIGNATUREMETHOD = "x-log-signaturemethod"
	LOG_BODYRAWSIZE     = "x-log-bodyrawsize"
)

type Credential interface {
	Signature(method string, headers map[string]string, resource string) (signature string, err error)
}

type AliLogCredential struct {
	accessKeySecret string
}

func NewAliLogCredential(accessKeySecret string) *AliLogCredential {
	aliLogCredential := new(AliLogCredential)
	aliLogCredential.accessKeySecret = accessKeySecret
	return aliLogCredential
}

func (a *AliLogCredential) Signature(method string, headers map[string]string, resource string) (signature string, err error) {

	signItems := []string{}
	signItems = append(signItems, method)

	var contentMD5, contentType string
	date := time.Now().UTC().Format(http.TimeFormat)

	if v, exist := headers[CONTENT_MD5]; exist {
		contentMD5 = v
	}
	if v, exist := headers[CONTENT_TYPE]; exist {
		contentType = v
	}
	if v, exist := headers[DATE]; exist {
		date = v
	}

	logHeaders := []string{}
	for k, v := range headers {
		if strings.HasPrefix(k, "x-log") || strings.HasPrefix(k, "x-acs") {
			logHeaders = append(logHeaders, k+":"+strings.TrimSpace(v))
		}
	}

	sort.Sort(sort.StringSlice(logHeaders))

	stringToSign := method + "\n" +
		contentMD5 + "\n" +
		contentType + "\n" +
		date + "\n" +
		strings.Join(logHeaders, "\n") + "\n" +
		resource

	sha1Hash := hmac.New(sha1.New, []byte(a.accessKeySecret))
	if _, e := sha1Hash.Write([]byte(stringToSign)); e != nil {
		err = e
		return
	}
	signature = base64.StdEncoding.EncodeToString(sha1Hash.Sum(nil))
	return
}
