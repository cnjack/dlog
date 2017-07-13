package dlog

import (
	"testing"
)

func TestSignature(t *testing.T) {
	headers := map[string]string{
		DATE:                 "Mon, 09 Nov 2015 06:03:03 GMT",
		HOST:                 "test-project.cn-hangzhou-devcommon-intranet.sls.aliyuncs.com",
		LOG_VERSION:          "0.6.0",
		CONTENT_LENTH:        "52",
		LOG_BODYRAWSIZE:      "50",
		CONTENT_TYPE:         "application/x-protobuf",
		"x-log-compresstype": "lz4",
		LOG_SIGNATUREMETHOD:  "hmac-sha1",
		CONTENT_MD5:          "1DD45FA4A70A9300CC9FE7305AF2C494",
	}

	AccessKeySecret := "4fdO2fTDDnZPU/L7CHNdemB2Nsk="
	method := "POST"
	resource := "/logstores/test-logstore"
	expect_signature := "XWLGYHGg2F2hcfxWxMLiNkGki6g="
	signature, err := NewAliLogCredential(AccessKeySecret).Signature(
		method,
		headers,
		resource)
	if expect_signature == signature && nil == err {
		t.Log("Success\n  hope:" + expect_signature + "\n ouput:" + signature)
	} else {
		t.Error("Fail\n  output:" + signature + "\n" + "error:" + err.Error())
	}
}
