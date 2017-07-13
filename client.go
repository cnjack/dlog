package dlog

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mreiferson/go-httpclient"
)

const (
	DefaultTimeout int64 = 35
)
const (
	version         = "0.6.0"
	signaturemethod = "hmac-sha1"
)

type LogClient interface {
	Send(method string, headers map[string]string, logdata interface{}, resource string) (resp *http.Response, err error)
}

type aliLogClient struct {
	Timeout      int64
	url          string
	accessKey    string
	credential   Credential
	client       *http.Client
	clientLocker sync.Mutex
}

func NewAliLogClient(url, accessKey, accessKeySecret string) (LogClient, error) {
	if url == "" {
		return nil, errors.New("dlog: log url is empty")
	}

	credential := NewAliLogCredential(accessKeySecret)

	cli := new(aliLogClient)
	cli.credential = credential
	cli.accessKey = accessKey
	cli.url = url

	if 5 != len(strings.Split(url, ".")) {
		return nil, errors.New("dlog: log url is invalid")
	}

	cli.initClient()
	return cli, nil

}

func (a *aliLogClient) initClient() {
	a.clientLocker.Lock()
	defer a.clientLocker.Unlock()
	timeoutInt := DefaultTimeout

	if a.Timeout > 0 {
		timeoutInt = a.Timeout
	}

	timeout := time.Second * time.Duration(timeoutInt)
	transport := &httpclient.Transport{
		ConnectTimeout:        time.Second * 3,
		RequestTimeout:        timeout,
		ResponseHeaderTimeout: timeout + time.Second,
	}
	a.client = &http.Client{Transport: transport}
}

func (a *aliLogClient) authorization(method string, headers map[string]string, resource string) (authHeader string, err error) {
	if signature, e := a.credential.Signature(method, headers, resource); e != nil {
		return "", e
	} else {
		authHeader = fmt.Sprintf("LOG %s:%s", a.accessKey, signature)
	}
	return
}

func (a *aliLogClient) Send(method string, headers map[string]string, logdata interface{}, resource string) (resp *http.Response, err error) {
	var logContent []byte
	if nil == logdata {
		logContent = []byte{}
	} else {
		switch m := logdata.(type) {
		case []byte:
			{
				logContent = m
			}
		default:
			err = errors.New("[ali_log][Send] logdata type Invalid")
			return

		}
	}

	logMD5 := md5.Sum(logContent)
	strMd5 := strings.ToUpper(fmt.Sprintf("%x", logMD5))
	if 0 == len(method) {
		method = "POST"
	}
	if nil == headers {
		headers = make(map[string]string)
	}

	headers[LOG_VERSION] = version
	headers[CONTENT_TYPE] = "application/x-protobuf"
	headers[CONTENT_MD5] = strMd5
	headers[LOG_SIGNATUREMETHOD] = signaturemethod
	headers[CONTENT_LENTH] = fmt.Sprintf("%v", len(logContent))
	headers[LOG_BODYRAWSIZE] = "0"
	headers[HOST] = a.url

	headers[DATE] = time.Now().UTC().Format(http.TimeFormat)

	if authHeader, e := a.authorization(method, headers, fmt.Sprintf("/%s", resource)); e != nil {
		err = errors.New("[ali_log][Send][authorization] " + e.Error())
		return
	} else {
		headers[AUTHORIZATION] = authHeader
	}

	url := a.url + "/" + resource
	if !strings.HasPrefix(a.url, "http://") || strings.HasPrefix(a.url, "https://") {
		url = "http://" + a.url + "/" + resource
	}
	postBodyReader := bytes.NewBuffer(logContent)

	var req *http.Request
	if req, err = http.NewRequest(method, url, postBodyReader); err != nil {
		err = errors.New("[ali_log][Send][NewRequest] " + err.Error())
		return
	}
	for header, value := range headers {
		req.Header.Add(header, value)
	}

	if resp, err = a.client.Do(req); err != nil {
		err = errors.New("[ali_log][Send][Do] " + err.Error())
		return
	}
	return
}
