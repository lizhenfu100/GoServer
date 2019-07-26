package cdn

import (
	"common"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const ( //SecretKey|SecretId from https://console.qcloud.com/capi
	kRequesturl = "cdn.api.qcloud.com/v2/index.php"
	kSecretKey  = "56B9wbK2azDPQ3b7Xqy770fDHF3wlS3U"
	kSecretId   = "AKIDz4ciT8GgOyUErV6kG0KVIz2HNYvB6V6a"
)

func RefreshUrl(urls []string) string {
	args := make(map[string]interface{})
	args["Action"] = "RefreshCdnUrl"
	for i, v := range urls {
		args["urls."+toStr(i)] = v
	}
	req := Signature(kSecretKey, kSecretId, args, "POST", kRequesturl)
	return SendRequest(kRequesturl, req, "POST")
}

func Signature(secretKey, secretId string, args map[string]interface{}, method string, requesturl string) map[string]interface{} {
	timenow := time.Now()
	args["SecretId"] = secretId
	args["Timestamp"] = timenow.Unix()
	args["Nonce"] = rand.New(rand.NewSource(timenow.UnixNano())).Intn(1000)
	/**sort all the args to make signPlainText**/
	sigUrl := method + requesturl + "?"
	var keys []string
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	isfirst := true
	for _, key := range keys {
		if !isfirst {
			sigUrl = sigUrl + "&"
		}
		isfirst = false
		if strings.Contains(key, "_") {
			strings.Replace(key, ".", "_", -1)
		}
		value := toStr(args[key])
		sigUrl = sigUrl + key + "=" + value
	}
	sig, _ := sign(sigUrl, secretKey)
	args["Signature"] = sig
	return args
}
func SendRequest(requesturl string, args map[string]interface{}, method string) string {
	requesturl = "https://" + requesturl
	if method == "GET" {
		return httpGet(requesturl + "?" + toStr2(args))
	} else if method == "POST" {
		return httpPost(requesturl, args)
	} else {
		return "unsuppported http method"
	}
}

func toStr2(args map[string]interface{}) string {
	isfirst := true
	requesturl := ""
	for k, v := range args {
		if !isfirst {
			requesturl += "&"
		}
		isfirst = false
		if strings.Contains(k, "_") {
			strings.Replace(k, ".", "_", -1)
		}
		requesturl += k + "=" + url.QueryEscape(toStr(v))
	}
	return requesturl
}
func toStr(t interface{}) string {
	switch v := t.(type) {
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case int64:
		return strconv.Itoa(int(v))
	default:
		return ""
	}
}
func sign(signPlainText string, secretKey string) (string, string) {
	hash := hmac.New(sha1.New, common.S2B(secretKey))
	hash.Write(common.S2B(signPlainText))
	sig := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	encd_sig := url.QueryEscape(sig)
	return sig, encd_sig
}

func httpGet(url string) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(3) * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err.Error()
	}
	body, e := ioutil.ReadAll(resp.Body)
	if resp.Body.Close(); e != nil {
		return e.Error()
	}
	return common.B2S(body)
}
func httpPost(requesturl string, args map[string]interface{}) string {
	req, err := http.NewRequest("POST", requesturl, strings.NewReader(toStr2(args)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(3) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	body, e := ioutil.ReadAll(resp.Body)
	if resp.Body.Close(); e != nil {
		return e.Error()
	}
	return common.B2S(body)
}
