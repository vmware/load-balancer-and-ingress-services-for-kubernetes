// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package session

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vmware/alb-sdk/go/logger"

	"github.com/golang/glog"
)

type AviResult struct {
	// Code should match the HTTP status code.
	Code int `json:"code"`

	// Message should contain a short description of the result of the requested
	// operation.
	Message *string `json:"message"`
}

// AviError represents an error resulting from a request to the Avi Controller
type AviError struct {
	// aviresult holds the standard header (code and message) that is included in
	// responses from Avi.
	AviResult

	// verb is the HTTP verb (GET, POST, PUT, PATCH, or DELETE) that was
	// used in the request that resulted in the error.
	Verb string

	// url is the URL that was used in the request that resulted in the error.
	Url string

	// HttpStatusCode is the HTTP response status code (e.g., 200, 404, etc.).
	HttpStatusCode int

	// err contains a descriptive error object for error cases other than HTTP
	// errors (i.e., non-2xx responses), such as socket errors or malformed JSON.
	err error
}

// HttpClient allows callers to inject their own implementations for the SDK to use.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PostMultipartRequest performs a POST API call and uploads multipart data to API fileobject/upload
func (avisess *AviSession) PostMultipartWafAppSignatureObjectRequest(fileLocPtr *os.File, uri string, tenant string, fileParams map[string]string) error {
	url := avisess.prefix + "/api/wafapplicationsignatureprovider/" + uri
	return avisess.restMultipartFileObjectUploadRequest("POST", fileLocPtr, url, nil, 0, tenant, fileParams)
}

// PostMultipartRequest performs a POST API call and uploads multipart data to API fileobject/upload
func (avisess *AviSession) PostMultipartFileObjectRequest(fileLocPtr *os.File, tenant string, fileParams map[string]string) error {

	url := avisess.prefix + "/api/fileobject/upload"
	return avisess.restMultipartFileObjectUploadRequest("POST", fileLocPtr, url, nil, 0, tenant, fileParams)
}

// restMultipartFileObjectUploadRequest makes a REST request to the Avi Controller's fileobject/upload REST API using
// POST to upload a file.
// Return status of multipart upload.
func (avisess *AviSession) restMultipartFileObjectUploadRequest(verb string, filePathPtr *os.File, url string,
	lastErr error, retryNum int, tenant string, fileParams map[string]string) error {

	if errorResult := avisess.checkRetryForSleep(retryNum, verb, url, lastErr); errorResult != nil {
		return errorResult
	}
	if avisess.lazyAuthentication && avisess.sessionid == "" {
		avisess.initiateSession()
	}

	errorResult := AviError{Verb: verb, Url: url}
	//Prepare a file that you will submit to an URL.
	values := map[string]io.Reader{
		"file": filePathPtr,
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		var err error
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				glog.Errorf("restMultipartFileObjectUploadRequest Error in adding file: %v ", err)
				return err
			}
		}
		if _, err := io.Copy(fw, r); err != nil {
			glog.Errorf("restMultipartFileObjectUploadRequest Error io.Copy %v ", err)
			return err
		}

	}

	var err error
	for fieldName, fieldValue := range fileParams {
		err = w.WriteField(fieldName, fieldValue)
		if err != nil {
			errorResult.err = fmt.Errorf("restMultipartFileObjectUploadRequest Adding URI field %v failed: %v", fieldName, err)
			return errorResult
		}
	}

	// Closing the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	req, errorResult := avisess.newAviRequest(context.Background(), verb, url, &b, tenant)
	if errorResult.err != nil {
		return errorResult
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := avisess.client.Do(req)
	if err != nil {
		glog.Errorf("restMultipartFileObjectUploadRequest Error during client request: %v ", err)
		dump, err := httputil.DumpRequestOut(req, true)
		debug(dump, err)
		return err
	}

	defer resp.Body.Close()

	errorResult.HttpStatusCode = resp.StatusCode
	avisess.collectCookiesFromResp(resp)
	glog.Infof("Response code: %v", resp.StatusCode)

	retryReq := false
	if resp.StatusCode == 401 && len(avisess.sessionid) != 0 {
		resp.Body.Close()
		err := avisess.initiateSession()
		if err != nil {
			return err
		}
		retryReq = true
	} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
		resp.Body.Close()
		retryReq = true
		glog.Infof("Retrying %d due to Status Code %d", retryNum, resp.StatusCode)
	}

	if retryReq {
		check, _, err := avisess.CheckControllerStatus()
		if check == false {
			glog.Errorf("restMultipartFileObjectUploadRequest Error during checking controller state")
			return err
		}
		// Doing this so that a new request is made to the
		return avisess.restMultipartFileObjectUploadRequest("POST", filePathPtr, url, err, retryNum+1, tenant, fileParams)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		glog.Errorf("Error: %v", resp)
		bres, berr := ioutil.ReadAll(resp.Body)
		if berr == nil {
			mres, _ := convertAviResponseToMapInterface(bres)
			glog.Infof("Error resp: %v", mres)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		}
		return errorResult
	}

	if resp.StatusCode == 201 {
		// File Created and upload to server
		fmt.Printf("restMultipartFileObjectUploadRequest Response: %v", resp.Status)
		return nil
	}
	return err
}

// Error implements the error interface.
func (err AviError) Error() string {
	var msg string

	if err.err != nil {
		msg = fmt.Sprintf("error: %v", err.err)
	} else if err.Message != nil {
		msg = fmt.Sprintf("HTTP code: %d; error from Controller: %s",
			err.HttpStatusCode, *err.Message)
	} else {
		msg = fmt.Sprintf("HTTP code: %d.", err.HttpStatusCode)
	}

	return fmt.Sprintf("Encountered an error on %s request to URL %s: %s",
		err.Verb, err.Url, msg)
}

// AviSession maintains a session to the specified Avi Controller
type AviSession struct {
	// host specifies the hostname or IP address of the Avi Controller
	host string

	// username specifies the username with which we should authenticate with the
	// Avi Controller.
	username string

	// password specifies the password with which we should authenticate with the
	// Avi Controller.
	password string

	// auth token generated by Django, for use in token mode
	authToken string

	// optional callback function passed in by the client which generates django auth token
	refreshAuthToken func() string

	// optional callback function V2 passed in by the client which generates django auth token with error handling
	refreshAuthTokenV2 func() (string, error)

	// insecure specifies whether we should perform strict certificate validation
	// for connections to the Avi Controller.
	insecure bool

	// timeout specifies time limit for API request. Default value set to 60 seconds
	timeout time.Duration

	// optional tenant string to use for API request
	tenant string

	// optional version string to use for API request
	version string

	// internal: session id for this session
	sessionid string

	// internal: csrfToken for this session
	csrfToken string

	// internal: referer field string to use in requests
	prefix string

	// internal: re-usable transport to enable connection reuse
	transport *http.Transport

	// internal: reusable client
	client HttpClient

	// optional lazy authentication flag. This will trigger login when the first API call is made.
	// The authentication is not performed when the Session object is created.
	lazyAuthentication bool

	// optional maximum api retry count
	max_api_retries int

	// optional api retry interval in milliseconds
	api_retry_interval int

	// Number of retries the SDK should attempt to check controller status.
	ctrlStatusCheckRetryCount int
	// Time interval in seconds within each retry to check controller status.
	ctrlStatusCheckRetryInterval int

	// this flag disables the checkcontrollerstatus method, instead client do their own retries
	disableControllerStatusCheck bool

	// Lock to synchronise the cookies collection from API response
	cookiesCollectLock sync.Mutex

	// Update the request header with custom headers
	user_headers map[string]string

	// CSP_HOST specifies the CSP Host name
	CSP_HOST string

	// CSP_TOKEN specifies the API token of the csp host with which we can generate the access token
	CSP_TOKEN string

	// internal: variable to store generated csp access token
	CSP_ACCESS_TOKEN string
}

const DEFAULT_AVI_VERSION = "18.2.6"
const DEFAULT_API_TIMEOUT = time.Duration(60 * time.Second)
const DEFAULT_API_TENANT = "admin"
const DEFAULT_MAX_API_RETRIES = 3
const DEFAULT_API_RETRY_INTERVAL = 500
const DEFAULT_CSP_HOST = "console.cloud.vmware.com"

// NewAviSession initiates a session to AviController and returns it
func NewAviSession(host string, username string, options ...func(*AviSession) error) (*AviSession, error) {
	if flag.Parsed() == false {
		flag.Parse()
	}
	avisess := &AviSession{
		host:     host,
		username: username,
	}
	avisess.sessionid = ""
	avisess.csrfToken = ""

	avisess.prefix = "https://" + avisess.host + "/"

	ip := GetIPVersion(avisess.host)
	if ip != nil && ip.To4() == nil {
		avisess.prefix = fmt.Sprintf("https://[%s]/", avisess.host)
	}

	avisess.tenant = ""
	avisess.insecure = false
	// The default behaviour was for 10 iterations, if client does not init session with specific retry
	// count option the controller status will be checked 10 times.
	avisess.ctrlStatusCheckRetryCount = 10
	for _, option := range options {
		err := option(avisess)
		if err != nil {
			return avisess, err
		}
	}

	if avisess.tenant == "" {
		avisess.tenant = DEFAULT_API_TENANT
	}
	if avisess.version == "" {
		avisess.version = DEFAULT_AVI_VERSION
	}

	if avisess.max_api_retries == 0 {
		avisess.max_api_retries = DEFAULT_MAX_API_RETRIES
	}

	if avisess.api_retry_interval == 0 {
		avisess.api_retry_interval = DEFAULT_API_RETRY_INTERVAL
	}

	if avisess.CSP_HOST == "" {
		avisess.CSP_HOST = DEFAULT_CSP_HOST
	}

	// set default timeout
	if avisess.timeout == 0 {
		avisess.timeout = DEFAULT_API_TIMEOUT
	}

	if avisess.client == nil {
		// create default transport object
		if avisess.transport == nil {
			avisess.transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		// attach transport object to client
		avisess.client = &http.Client{
			Transport: avisess.transport,
			Timeout:   avisess.timeout,
		}
	}

	if avisess.CSP_TOKEN != "" {
		err := avisess.getCSPAccessToken()
		return avisess, err
	}

	if !avisess.lazyAuthentication {
		err := avisess.initiateSession()
		return avisess, err
	}
	return avisess, nil
}

func requestForAccessToken(retries int, url string, payload *strings.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		glog.Errorf("Request error: %v ", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Response error: %v ", err)
	}
	return resp, err
}

func (avisess *AviSession) getCSPAccessToken() error {
	url := "https://" + avisess.CSP_HOST + "/csp/gateway/am/api/auth/api-tokens/authorize"
	payload := strings.NewReader("api_token=" + avisess.CSP_TOKEN)
	var (
		retries  int = 0
		resp     *http.Response
		err      error
		response map[string]interface{}
	)
	for retries < DEFAULT_MAX_API_RETRIES {
		resp, err = requestForAccessToken(retries, url, payload)
		if err != nil {
			glog.Errorf("Request error: %v ", err)
		}
		if resp.StatusCode == 200 {
			break
		} else {
			glog.Errorf("Unable to get the access token, retrying : %v", retries)
			time.Sleep(10 * time.Second)
			retries += 1
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	json.Unmarshal(body, &response)
	if retries == DEFAULT_MAX_API_RETRIES && resp.StatusCode != 200 {
		glog.Errorf("Unable to get the access token due to: %v", response["message"].(string))
		// Exiting from here because not getting any error in err var, From CSP Always getting an error in resp var.
		os.Exit(0)
	} else {
		access_token := response["access_token"].(string)
		avisess.CSP_ACCESS_TOKEN = access_token
	}
	return nil
}

func (avisess *AviSession) initiateSession() error {
	if avisess.insecure == true {
		glog.Warning("Strict certificate verification is *DISABLED*")
	}

	// If refresh auth token is provided, use callback function provided
	if avisess.isTokenAuth() {
		switch {
		case avisess.refreshAuthToken != nil:
			avisess.setAuthToken(avisess.refreshAuthToken())
		case avisess.refreshAuthTokenV2 != nil:
			if token, err := avisess.refreshAuthTokenV2(); err != nil {
				return err
			} else {
				avisess.setAuthToken(token)
			}
		}
	}

	// initiate http session here
	// first set the csrf token
	var res interface{}
	//rerror := avisess.Get("", res)

	// now login to get session_id, csrfToken
	cred := make(map[string]string)
	cred["username"] = avisess.username

	if avisess.isTokenAuth() {
		cred["token"] = avisess.authToken
	} else {
		cred["password"] = avisess.password
	}

	rerror := avisess.Post("login", cred, res)
	if rerror != nil {
		glog.Errorf("response error: %v ", rerror)
		return rerror
	}

	glog.Infof("response: %v", res)
	if res != nil && reflect.TypeOf(res).Kind() != reflect.String {
		glog.Infof("results: %v error %v", res.(map[string]interface{}), rerror)
	}

	return nil
}

// SetPassword - Use this for NewAviSession option argument for setting password
func SetPassword(password string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setPassword(password)
	}
}

func (avisess *AviSession) setPassword(password string) error {
	avisess.password = password
	return nil
}

// SetVersion - Use this for NewAviSession option argument for setting version
func SetVersion(version string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setVersion(version)
	}
}

func (avisess *AviSession) setVersion(version string) error {
	avisess.version = version
	return nil
}

// SetAuthToken - Use this for NewAviSession option argument for setting authToken
func SetAuthToken(authToken string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setAuthToken(authToken)
	}
}

func (avisess *AviSession) setAuthToken(authToken string) error {
	avisess.authToken = authToken
	return nil
}

// SetAuthToken - Use this for NewAviSession option argument for setting authToken
func SetRefreshAuthTokenCallback(f func() string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setRefreshAuthTokenCallback(f)
	}
}

func (avisess *AviSession) setRefreshAuthTokenCallback(f func() string) error {
	avisess.refreshAuthToken = f
	return nil
}

// SetAuthToken V2 - Use this for NewAviSession option argument for setting authToken with option to return error found
// during token generation
func SetRefreshAuthTokenCallbackV2(f func() (string, error)) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setRefreshAuthTokenCallbackV2(f)
	}
}

func (avisess *AviSession) setRefreshAuthTokenCallbackV2(f func() (string, error)) error {
	avisess.refreshAuthTokenV2 = f
	return nil
}

// SetTenant - Use this for NewAviSession option argument for setting tenant
func SetTenant(tenant string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setTenant(tenant)
	}
}

func (avisess *AviSession) setTenant(tenant string) error {
	avisess.tenant = tenant
	return nil
}

// SetInsecure - Use this for NewAviSession option argument for allowing insecure connection to AviController
func SetInsecure(avisess *AviSession) error {
	avisess.insecure = true
	return nil
}

// SetControllerStatusCheckLimits allows client to limit the number of tries the SDK should
// attempt to reach the controller at the time gap of specified time intervals.
func SetControllerStatusCheckLimits(numRetries, retryInterval int) func(*AviSession) error {
	return func(avisess *AviSession) error {
		if numRetries <= 0 || retryInterval <= 0 {
			return errors.New("Retry count and retry interval should be greater than zero")
		}
		avisess.ctrlStatusCheckRetryCount = numRetries
		avisess.ctrlStatusCheckRetryInterval = retryInterval
		return nil
	}
}

func DisableControllerStatusCheckOnFailure(controllerStatusCheck bool) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.disableControllerStatusCheckOnFailure(controllerStatusCheck)
	}
}

func (avisess *AviSession) disableControllerStatusCheckOnFailure(controllerStatusCheck bool) error {
	avisess.disableControllerStatusCheck = controllerStatusCheck
	return nil
}

// SetTransport - Use this for NewAviSession option argument for configuring http transport to enable connection
func SetTransport(transport *http.Transport) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setTransport(transport)
	}
}

func (avisess *AviSession) setTransport(transport *http.Transport) error {
	if avisess.client != nil {
		return errors.New("Cannot set custom Transport for external clients")
	}
	avisess.transport = transport
	return nil
}

// SetClient allows callers to inject their own HTTP client.
func SetClient(client HttpClient) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setClient(client)
	}
}

func (avisess *AviSession) setClient(client HttpClient) error {
	if avisess.transport != nil {
		return errors.New("Cannot set custom client when transport is already set to http.Transport")
	}
	avisess.client = client
	return nil
}

// SetTimeout -
func SetTimeout(timeout time.Duration) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setTimeout(timeout)
	}
}

func (avisess *AviSession) setTimeout(timeout time.Duration) error {
	avisess.timeout = timeout
	return nil
}

// SetUserHeader -
func SetUserHeader(user_headers map[string]string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setUserHeader(user_headers)
	}
}

func (avisess *AviSession) setUserHeader(user_headers map[string]string) error {
	avisess.user_headers = user_headers
	return nil
}

// SetCSPToken
func SetCSPToken(csptoken string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setCSPToken(csptoken)
	}
}

func (avisess *AviSession) setCSPToken(csptoken string) error {
	avisess.CSP_TOKEN = csptoken
	return nil
}

func SetCSPHost(csphost string) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setCSPHost(csphost)
	}
}

func (avisess *AviSession) setCSPHost(csphost string) error {
	avisess.CSP_HOST = csphost
	return nil
}

func (avisess *AviSession) isTokenAuth() bool {
	return avisess.authToken != "" || avisess.refreshAuthToken != nil || avisess.refreshAuthTokenV2 != nil
}

// SetTimeout -
func SetLazyAuthentication(lazyAuthentication bool) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setLazyAuthentication(lazyAuthentication)
	}
}

func (avisess *AviSession) setLazyAuthentication(lazyAuthentication bool) error {
	avisess.lazyAuthentication = lazyAuthentication
	return nil
}

func SetMaxApiRetries(max_api_retries int) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setMaxApiRetries(max_api_retries)
	}
}

func (avisess *AviSession) setMaxApiRetries(max_api_retries int) error {
	avisess.max_api_retries = max_api_retries
	return nil
}

func SetApiRetryInterval(api_retry_interval int) func(*AviSession) error {
	return func(sess *AviSession) error {
		return sess.setApiRetryInterval(api_retry_interval)
	}
}

func (avisess *AviSession) setApiRetryInterval(api_retry_interval int) error {
	avisess.api_retry_interval = api_retry_interval
	return nil
}

func (avisess *AviSession) checkRetryForSleep(retry int, verb string, url string, lastErr error) error {
	if retry == 0 {
		return nil
	} else if retry < avisess.max_api_retries {
		time.Sleep(time.Duration(avisess.api_retry_interval) * time.Millisecond)
	} else {
		if lastErr != nil {
			glog.Errorf("Aborting after %v times. Last error %v", retry, lastErr)
			return lastErr
		}
		errorResult := AviError{Verb: verb, Url: url}
		errorResult.err = fmt.Errorf("tried %v times and failed", retry)
		return errorResult
	}
	return nil
}

func (avisess *AviSession) newAviRequest(ctx context.Context, verb string, url string, payload io.Reader, tenant string) (*http.Request, AviError) {
	req, err := http.NewRequest(verb, url, payload)
	errorResult := AviError{Verb: verb, Url: url}
	if err != nil {
		errorResult.err = fmt.Errorf("http.NewRequest failed: %v", err)
		return nil, errorResult
	}
	if avisess.CSP_ACCESS_TOKEN != "" {
		req.Header.Set("Authorization", "Bearer "+string(avisess.CSP_ACCESS_TOKEN))
	}
	req.Header.Set("Content-Type", "application/json")

	if avisess.user_headers != nil {
		for k, v := range avisess.user_headers {
			req.Header.Set(k, v)
		}
	}
	traceID := logger.GetTraceID(ctx)
	if traceID != "" {
		req.Header.Set("X-Request-ID", traceID)
	}
	//req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Avi-Version", avisess.version)
	if tenant == "" {
		tenant = avisess.tenant
	}
	if !strings.HasSuffix(url, "login") && avisess.csrfToken != "" {
		req.Header["X-CSRFToken"] = []string{avisess.csrfToken}
		req.AddCookie(&http.Cookie{Name: "csrftoken", Value: avisess.csrfToken})
	}
	if avisess.prefix != "" {
		req.Header.Set("Referer", avisess.prefix)
	}
	if tenant != "" {
		req.Header.Set("X-Avi-Tenant", tenant)
	}

	if !strings.HasSuffix(url, "login") && avisess.sessionid != "" {
		req.AddCookie(&http.Cookie{Name: "sessionid", Value: avisess.sessionid})
		req.AddCookie(&http.Cookie{Name: "avi-sessionid", Value: avisess.sessionid})
	}
	return req, errorResult
}

//
// Helper routines for REST calls.
//

func (avisess *AviSession) collectCookiesFromResp(resp *http.Response) {
	// collect cookies from the resp
	avisess.cookiesCollectLock.Lock()
	defer avisess.cookiesCollectLock.Unlock()

	var csrfToken string
	var sessionID string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			csrfToken = cookie.Value
		}
		if cookie.Name == "sessionid" || cookie.Name == "avi-sessionid" {
			sessionID = cookie.Value
		}
	}
	if csrfToken != "" && sessionID != "" {
		avisess.csrfToken = csrfToken
		avisess.sessionid = sessionID
	}
}

// RestRequest exports restRequest from the SDK
// Returns http.Response for accessing the whole http Response struct including headers and response body
func (avisess *AviSession) RestRequest(verb string, uri string, payload interface{}, tenant string, lastError error,
	retryNum ...int) (*http.Response, error) {
	return avisess.restRequest(context.Background(), verb, uri, payload, tenant, nil)
}

// restRequest makes a REST request to the Avi Controller's REST API.
// Returns http.Response if successful
// Note: The caller of the function is responsible for doing resp.Body.Close()
func (avisess *AviSession) restRequest(ctx context.Context, verb string, uri string, payload interface{}, tenant string, lastError error, retryNum ...int) (*http.Response, error) {
	url := avisess.prefix + uri
	// If optional retryNum arg is provided, then count which retry number this is
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
	}
	if errorResult := avisess.checkRetryForSleep(retry, verb, url, lastError); errorResult != nil {
		return nil, errorResult
	}

	if avisess.lazyAuthentication && avisess.sessionid == "" && !(uri == "" || uri == "login") {
		avisess.initiateSession()
	}

	var payloadIO io.Reader
	if payload != nil {
		jsonStr, err := json.Marshal(payload)
		if err != nil {
			return nil, AviError{Verb: verb, Url: url, err: err}
		}
		payloadIO = bytes.NewBuffer(jsonStr)
	}

	req, errorResult := avisess.newAviRequest(ctx, verb, url, payloadIO, tenant)
	if errorResult.err != nil {
		return nil, errorResult
	}
	retryReq := false
	resp, err := avisess.client.Do(req)
	if err != nil {
		// retry until controller status check limits.
		glog.Errorf("Client error for URI: %+v. Error: %+v", uri, err.Error())
		dump, dumpErr := httputil.DumpRequestOut(req, true)
		if dumpErr != nil {
			glog.Error("Error while dumping request. Still retrying.")
		}
		debug(dump, dumpErr)
		retryReq = true
	}
	if resp != nil && resp.StatusCode == 500 {
		if _, err = avisess.fetchBody(verb, uri, resp); err != nil {
			glog.Errorf("Client error for URI: %+v. Error: %+v", uri, err.Error())
		}
		if err != nil {
			return nil, err
		} else {
			retryReq = true
		}
	}
	if !retryReq {
		glog.Infof("Req for %s uri %v tenant %s RespCode %v", verb, url, tenant, resp.StatusCode)
		errorResult.HttpStatusCode = resp.StatusCode

		if uri == "login" {
			avisess.collectCookiesFromResp(resp)
		}
		if resp.StatusCode == 401 && uri != "login" {
			resp.Body.Close()
			glog.Infof("Retrying url %s; retry %d due to Status Code %d", url, retry, resp.StatusCode)
			err := avisess.initiateSession()
			if err != nil {
				return nil, err
			}
			retryReq = true
		} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
			resp.Body.Close()
			retryReq = true
			glog.Infof("Retrying url: %s; retry: %d due to Status Code %d", url, retry, resp.StatusCode)
		}
	}
	if retryReq {
		if !avisess.disableControllerStatusCheck {
			check, httpResp, err := avisess.CheckControllerStatus()
			if check == false {
				if resp != nil && resp.Body != nil {
					glog.Infof("Body is not nil, close it.")
					resp.Body.Close()
				}
				glog.Errorf("restRequest Error during checking controller state. Error: %s", err)
				return httpResp, err
			}
			if err := avisess.initiateSession(); err != nil {
				if resp != nil && resp.Body != nil {
					glog.Infof("Body is not nil, close it.")
					resp.Body.Close()
				}
				return nil, err
			}
			return avisess.restRequest(context.Background(), verb, uri, payload, tenant, errorResult, retry+1)
		} else {
			glog.Error("CheckControllerStatus is disabled for this session, not going to retry.")
			if err != nil {
				glog.Errorf("Failed to invoke API. Error: %s", err.Error())
			}
			return nil, fmt.Errorf("Rest request error, returning to caller: %s", err.Error())

		}
	}
	return resp, nil
}

// fetchBody fetches the response body from the http.Response returned from restRequest
func (avisess *AviSession) fetchBody(verb, uri string, resp *http.Response) (result []byte, err error) {
	url := avisess.prefix + uri
	errorResult := AviError{HttpStatusCode: resp.StatusCode, Verb: verb, Url: url}

	if resp.StatusCode == 204 {
		// no content in the response
		return result, nil
	}
	// It cannot be assumed that the error will always be from server side in response.
	// Error could be from HTTP client side which will not have body in response.
	// Need to change our API resp handling design if we want to handle client side errors separately.

	// Below block will take care for errors without body.
	if resp.Body == nil {
		glog.Errorf("Encountered client side error: %+v", resp)
		errorResult.Message = &resp.Status
		return result, errorResult
	}

	defer resp.Body.Close()
	result, err = ioutil.ReadAll(resp.Body)
	if err == nil {
		if resp.StatusCode < 200 || resp.StatusCode > 299 || resp.StatusCode == 500 {
			mres, merr := convertAviResponseToMapInterface(result)
			glog.Infof("Error code %v parsed resp: %v err %v",
				resp.StatusCode, mres, merr)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		} else {
			return result, nil
		}
	} else {
		errmsg := fmt.Sprintf("Response body read failed: %v", err)
		errorResult.Message = &errmsg
		glog.Errorf("Error in reading uri %v %v", uri, err)
	}
	return result, errorResult
}

// restMultipartUploadRequest makes a REST request to the Avi Controller's REST API using POST to upload a file.
// Return status of multipart upload.
func (avisess *AviSession) restMultipartUploadRequest(verb string, uri string, file_path_ptr *os.File, tenant string, lastErr error,
	retryNum ...int) error {
	url := avisess.prefix + "/api/fileservice/" + uri

	// If optional retryNum arg is provided, then count which retry number this is
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
	}

	if errorResult := avisess.checkRetryForSleep(retry, verb, url, lastErr); errorResult != nil {
		return errorResult
	}

	if avisess.lazyAuthentication && avisess.sessionid == "" && !(uri == "" || uri == "login") {
		avisess.initiateSession()
	}

	errorResult := AviError{Verb: verb, Url: url}
	//Prepare a file that you will submit to an URL.
	values := map[string]io.Reader{
		"file": file_path_ptr,
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		var err error
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				if err != nil {
					glog.Errorf("restMultipartUploadRequest Error in adding file: %v ", err)
					return err
				}
			}
		}
		if _, err := io.Copy(fw, r); err != nil {
			if err != nil {
				glog.Errorf("restMultipartUploadRequest Error io.Copy %v ", err)
				return err
			}
		}

	}
	// Closing the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()
	uri_temp := "controller://" + strings.Split(uri, "?")[0]
	err := w.WriteField("uri", uri_temp)
	if err != nil {
		errorResult.err = fmt.Errorf("restMultipartUploadRequest Adding URI field failed: %v", err)
		return errorResult
	}
	req, errorResult := avisess.newAviRequest(context.Background(), verb, url, &b, tenant)
	if errorResult.err != nil {
		return errorResult
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := avisess.client.Do(req)
	if err != nil {
		glog.Errorf("restMultipartUploadRequest Error during client request: %v ", err)
		dump, err := httputil.DumpRequestOut(req, true)
		debug(dump, err)
		return err
	}

	defer resp.Body.Close()

	errorResult.HttpStatusCode = resp.StatusCode
	avisess.collectCookiesFromResp(resp)
	glog.Infof("Response code: %v", resp.StatusCode)

	retryReq := false
	if resp.StatusCode == 401 && len(avisess.sessionid) != 0 && uri != "login" {
		resp.Body.Close()
		err := avisess.initiateSession()
		if err != nil {
			return err
		}
		retryReq = true
	} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
		resp.Body.Close()
		retryReq = true
		glog.Infof("Retrying %d due to Status Code %d", retry, resp.StatusCode)
	}

	if retryReq {
		check, _, err := avisess.CheckControllerStatus()
		if check == false {
			glog.Errorf("restMultipartUploadRequest Error during checking controller state")
			return err
		}
		// Doing this so that a new request is made to the
		return avisess.restMultipartUploadRequest(verb, uri, file_path_ptr, tenant, errorResult, retry+1)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		glog.Errorf("Error: %v", resp)
		bres, berr := ioutil.ReadAll(resp.Body)
		if berr == nil {
			mres, _ := convertAviResponseToMapInterface(bres)
			glog.Infof("Error resp: %v", mres)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		}
		return errorResult
	}

	if resp.StatusCode == 201 {
		// File Created and upload to server
		fmt.Printf("restMultipartUploadRequest Response: %v", resp.Status)
		return nil
	}

	return err
}

// restMultipartDownloadRequest makes a REST request to the Avi Controller's REST API.
// Returns multipart download and write data to file
func (avisess *AviSession) restMultipartDownloadRequest(verb string, uri string, file_path_ptr *os.File, tenant string, lastErr error, retryNum ...int) error {
	url := avisess.prefix + "/api/fileservice/" + uri

	// If optional retryNum arg is provided, then count which retry number this is
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
	}

	if errorResult := avisess.checkRetryForSleep(retry, verb, url, lastErr); errorResult != nil {
		return errorResult
	}

	req, errorResult := avisess.newAviRequest(context.Background(), verb, url, nil, tenant)
	if errorResult.err != nil {
		return errorResult
	}
	req.Header.Set("Accept", "application/json")
	resp, err := avisess.client.Do(req)
	if err != nil {
		errorResult.err = fmt.Errorf("restMultipartDownloadRequest Error for during client request: %v", err)
		dump, err := httputil.DumpRequestOut(req, true)
		debug(dump, err)
		return errorResult
	}

	errorResult.HttpStatusCode = resp.StatusCode
	avisess.collectCookiesFromResp(resp)
	glog.Infof("Response code: %v", resp.StatusCode)

	retryReq := false
	if resp.StatusCode == 401 && len(avisess.sessionid) != 0 && uri != "login" {
		resp.Body.Close()
		err := avisess.initiateSession()
		if err != nil {
			return err
		}
		retryReq = true
	} else if resp.StatusCode == 419 || (resp.StatusCode >= 500 && resp.StatusCode < 599) {
		resp.Body.Close()
		retryReq = true
		glog.Infof("Retrying %d due to Status Code %d", retry, resp.StatusCode)
	}

	if retryReq {
		check, _, err := avisess.CheckControllerStatus()
		if check == false {
			glog.Errorf("restMultipartDownloadRequest Error during checking controller state")
			return err
		}
		return avisess.restMultipartDownloadRequest(verb, uri, file_path_ptr, tenant, errorResult)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		// no content in the response
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		glog.Errorf("Error: %v", resp)
		bres, berr := ioutil.ReadAll(resp.Body)
		if berr == nil {
			mres, _ := convertAviResponseToMapInterface(bres)
			glog.Infof("Error resp: %v", mres)
			emsg := fmt.Sprintf("%v", mres)
			errorResult.Message = &emsg
		}
		return errorResult
	}

	_, err = io.Copy(file_path_ptr, resp.Body)
	defer file_path_ptr.Close()

	if err != nil {
		glog.Errorf("Error while downloading %v", err)
	}
	return err
}

func convertAviResponseToMapInterface(resbytes []byte) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(resbytes, &result)
	return result, err
}

// AviCollectionResult for representing the collection type results from Avi
type AviCollectionResult struct {
	Count   int
	Results json.RawMessage
	Next    string
}

func removeSensitiveFields(data []byte) []byte {
	dataString := string(data)
	re := regexp.MustCompile(`"password":"([^\s]+?)","username":"([^\s]+?)"`)
	updatedDataString := re.ReplaceAllString(dataString, "")
	return []byte(updatedDataString)
}

func debug(data []byte, err error) {
	if err == nil {
		data = removeSensitiveFields(data)
		glog.Infof("%s\n\n", data)
	} else {
		glog.Errorf("%s\n\n", err)
	}
}

// Checking for controller up state.
// Flexible to wait on controller status infinitely or for fixed time span.
func (avisess *AviSession) CheckControllerStatus() (bool, *http.Response, error) {
	url := avisess.prefix + "/api/cluster/status"
	var isControllerUp bool
	for round := 0; round < avisess.ctrlStatusCheckRetryCount; round++ {
		checkReq, err := http.NewRequest("GET", url, nil)
		if err != nil {
			glog.Errorf("CheckControllerStatus Error %v while generating http request.", err)
			return false, nil, err
		}
		//Getting response from controller's API
		if stateResp, err := avisess.client.Do(checkReq); err == nil {
			defer stateResp.Body.Close()
			//Checking controller response
			if stateResp.StatusCode != 503 && stateResp.StatusCode != 502 && stateResp.StatusCode != 500 {
				isControllerUp = true
				break
			} else {
				glog.Infof("CheckControllerStatus Error while generating http request %d %v",
					stateResp.StatusCode, err)
			}
		} else {
			glog.Errorf("CheckControllerStatus Error while generating http request %v %v", url, err)
		}
		// if controller status check interval is not set during client init, use the default SDK
		// behaviour.
		if avisess.ctrlStatusCheckRetryInterval == 0 {
			time.Sleep(getMinTimeDuration((time.Duration(math.Exp(float64(round))*3) * time.Second), (time.Duration(30) * time.Second)))
		} else {
			// controller status will be polled at intervals specified during client init.
			time.Sleep(time.Duration(avisess.ctrlStatusCheckRetryInterval) * time.Second)
		}
		glog.Errorf("CheckControllerStatus Controller %v Retrying. round %v..!", url, round)
	}
	return isControllerUp, &http.Response{Status: "408 Request Timeout", StatusCode: 408}, nil
}

// getMinTimeDuration returns the minimum time duration between two time values.
func getMinTimeDuration(durationFirst, durationSecond time.Duration) time.Duration {
	if durationFirst <= durationSecond {
		return durationFirst
	}
	return durationSecond
}

func (avisess *AviSession) restRequestInterfaceResponse(verb string, url string,
	payload interface{}, response interface{}, options ...ApiOptionsParams) error {
	opts, err := getOptions(options)
	if err != nil {
		return err
	}
	if len(opts.params) != 0 {
		url = updateUri(url, opts)
	}
	httpResponse, rerror := avisess.restRequest(opts.ctx, verb, url, payload, opts.tenant, nil)
	if rerror != nil {
		return rerror
	}
	var res []byte
	if res, err = avisess.fetchBody(verb, url, httpResponse); err != nil {
		return err
	}

	if len(res) == 0 {
		return nil
	} else {
		return json.Unmarshal(res, &response)
	}
}

// Get issues a GET request against the avisess REST API.
func (avisess *AviSession) Get(uri string, response interface{}, options ...ApiOptionsParams) error {
	return avisess.restRequestInterfaceResponse("GET", uri, nil, response, options...)
}

// Post issues a POST request against the avisess REST API.
func (avisess *AviSession) Post(uri string, payload interface{}, response interface{}, options ...ApiOptionsParams) error {
	return avisess.restRequestInterfaceResponse("POST", uri, payload, response, options...)
}

// Put issues a PUT request against the avisess REST API.
func (avisess *AviSession) Put(uri string, payload interface{}, response interface{}, options ...ApiOptionsParams) error {
	return avisess.restRequestInterfaceResponse("PUT", uri, payload, response, options...)
}

// Post issues a PATCH request against the avisess REST API.
// allowed patchOp - add, replace, remove
func (avisess *AviSession) Patch(uri string, payload interface{}, patchOp string, response interface{}, options ...ApiOptionsParams) error {
	var patchPayload = make(map[string]interface{})
	patchPayload[patchOp] = payload
	glog.Infof(" PATCH OP %v data %v", patchOp, payload)
	return avisess.restRequestInterfaceResponse("PATCH", uri, patchPayload, response, options...)
}

// Delete issues a DELETE request against the avisess REST API.
func (avisess *AviSession) Delete(uri string, params ...interface{}) error {
	var payload, response interface{}
	if len(params) > 0 {
		payload = params[0]
		if len(params) == 2 {
			response = params[1]
		}
	}
	return avisess.restRequestInterfaceResponse("DELETE", uri, payload, response)
}

// GetCollectionRaw issues a GET request and returns a AviCollectionResult with unmarshaled (raw) results section.
func (avisess *AviSession) GetCollectionRaw(uri string, options ...ApiOptionsParams) (AviCollectionResult, error) {
	var result AviCollectionResult
	opts, err := getOptions(options)
	if err != nil {
		return result, err
	}
	if len(opts.params) != 0 {
		uri = updateUri(uri, opts)
	}
	httpResponse, rerror := avisess.restRequest(context.Background(), "GET", uri, nil, opts.tenant, nil)
	if rerror != nil || httpResponse == nil {
		return result, rerror
	}

	var res []byte
	if res, err = avisess.fetchBody("GET", uri, httpResponse); err != nil {
		return result, err
	}

	if strings.Contains(uri, "cluster?") {
		result.Results = res
		result.Count = 1
	}
	err = json.Unmarshal(res, &result)
	return result, err
}

// GetCollection performs a collection API call and unmarshals the results into objList, which should be an array type
func (avisess *AviSession) GetCollection(uri string, objList interface{}, options ...ApiOptionsParams) error {
	result, err := avisess.GetCollectionRaw(uri, options...)
	if err != nil {
		return err
	}
	if result.Count == 0 {
		return nil
	}
	return json.Unmarshal(result.Results, &objList)
}

// GetRaw performs a GET API call and returns raw data
func (avisess *AviSession) GetRaw(uri string, options ...ApiOptionsParams) ([]byte, error) {
	opts, err := getOptions(options)
	if err != nil {
		return nil, err
	}
	if len(opts.params) != 0 {
		uri = updateUri(uri, opts)
	}
	resp, rerror := avisess.restRequest(context.Background(), "GET", uri, nil, opts.tenant, nil)
	if rerror != nil || resp == nil {
		return nil, rerror
	}

	return avisess.fetchBody("GET", uri, resp)
}

// PostRaw performs a POST API call and returns raw data
func (avisess *AviSession) PostRaw(uri string, payload interface{}, options ...ApiOptionsParams) ([]byte, error) {
	opts, err := getOptions(options)
	if err != nil {
		return nil, err
	}
	if len(opts.params) != 0 {
		uri = updateUri(uri, opts)
	}
	resp, rerror := avisess.restRequest(context.Background(), "POST", uri, payload, opts.tenant, nil)
	if rerror != nil || resp == nil {
		return nil, rerror
	}
	return avisess.fetchBody("POST", uri, resp)
}

// PutRaw performs a POST API call and returns raw data
func (avisess *AviSession) PutRaw(uri string, payload interface{}, options ...ApiOptionsParams) ([]byte, error) {
	opts, err := getOptions(options)
	if err != nil {
		return nil, err
	}
	if len(opts.params) != 0 {
		uri = updateUri(uri, opts)
	}
	resp, rerror := avisess.restRequest(context.Background(), "PUT", uri, payload, opts.tenant, nil)
	if rerror != nil || resp == nil {
		return nil, rerror
	}
	return avisess.fetchBody("PUT", uri, resp)
}

// GetMultipartRaw performs a GET API call and returns multipart raw data (File Download)
// The verb input is ignored and kept only for backwards compatibility
func (avisess *AviSession) GetMultipartRaw(verb string, uri string, file_loc_ptr *os.File, options ...ApiOptionsParams) error {
	opts, err := getOptions(options)
	if err != nil {
		return err
	}
	return avisess.restMultipartDownloadRequest("GET", uri, file_loc_ptr, opts.tenant, nil)
}

// PostMultipartRequest performs a POST API call and uploads multipart data
// The verb input is ignored and kept only for backwards compatibility
func (avisess *AviSession) PostMultipartRequest(verb string, uri string, file_loc_ptr *os.File, options ...ApiOptionsParams) error {
	opts, err := getOptions(options)
	if err != nil {
		return err
	}
	return avisess.restMultipartUploadRequest("POST", uri, file_loc_ptr, opts.tenant, nil)
}

type ApiOptions struct {
	name        string
	cloud       string
	cloudUUID   string
	tenant      string
	skipDefault bool
	includeName bool
	payload     interface{}
	result      interface{}
	params      map[string]string
	ctx         context.Context
}

func SetContext(ctx context.Context) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setContext(ctx)
	}
}

func (opts *ApiOptions) setContext(ctx context.Context) error {
	opts.ctx = ctx
	return nil
}

func SetOptTenant(tenant string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setOptTenant(tenant)
	}
}

func (opts *ApiOptions) setOptTenant(tenant string) error {
	opts.tenant = tenant
	return nil
}

func SetName(name string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setName(name)
	}
}

func (opts *ApiOptions) setName(name string) error {
	opts.name = name
	return nil
}

func SetCloud(cloud string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setCloud(cloud)
	}
}

func (opts *ApiOptions) setCloud(cloud string) error {
	opts.cloud = cloud
	return nil
}

func SetCloudUUID(cloudUUID string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setCloudUUID(cloudUUID)
	}
}

func (opts *ApiOptions) setCloudUUID(cloudUUID string) error {
	opts.cloudUUID = cloudUUID
	return nil
}

func SetSkipDefault(skipDefault bool) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setSkipDefault(skipDefault)
	}
}

func (opts *ApiOptions) setSkipDefault(skipDefault bool) error {
	opts.skipDefault = skipDefault
	return nil
}

func SetIncludeName(includeName bool) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setIncludeName(includeName)
	}
}

func (opts *ApiOptions) setIncludeName(includeName bool) error {
	opts.includeName = includeName
	return nil
}

func SetResult(result interface{}) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setResult(result)
	}
}

func (opts *ApiOptions) setResult(result interface{}) error {
	opts.result = result
	return nil
}

func SetParams(params map[string]string) func(*ApiOptions) error {
	return func(opts *ApiOptions) error {
		return opts.setParams(params)
	}
}

func (opts *ApiOptions) setParams(params map[string]string) error {
	opts.params = params
	return nil
}

type ApiOptionsParams func(*ApiOptions) error

// GetUri returns the URI
func (avisess *AviSession) GetUri(obj string, options ...ApiOptionsParams) (string, error) {
	opts, err := getOptions(options)
	if err != nil {
		return "", err
	}
	if opts.result == nil {
		return "", errors.New("reference to result provided")
	}

	if opts.name == "" {
		return "", errors.New("Name not specified")
	}

	uri := "api/" + obj + "?name=" + url.QueryEscape(opts.name)
	if opts.cloud != "" {
		uri = uri + "&cloud=" + url.QueryEscape(opts.cloud)
	} else if opts.cloudUUID != "" {
		uri = uri + "&cloud_ref.uuid=" + url.QueryEscape(opts.cloudUUID)
	}
	if opts.skipDefault {
		uri = uri + "&skip_default=true"
	}
	if opts.includeName {
		uri = uri + "&include_name=true"
	}
	return uri, nil
}

// DeleteObject performs DELETE Operation and delete the data
func (avisess *AviSession) DeleteObject(uri string, options ...ApiOptionsParams) error {
	opts, err := getOptions(options)
	if err != nil {
		return err
	}
	return avisess.restRequestInterfaceResponse("DELETE", uri, opts.payload, opts.result, options...)
}

func getOptions(options []ApiOptionsParams) (*ApiOptions, error) {
	opts := &ApiOptions{}
	for _, opt := range options {
		err := opt(opts)
		if err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// GetObject performs GET and return object data
func (avisess *AviSession) GetObject(obj string, options ...ApiOptionsParams) error {
	opts, err := getOptions(options)
	uri, err := avisess.GetUri(obj, options...)
	if err != nil {
		return err
	}
	res, err := avisess.GetCollectionRaw(uri, options...)
	if err != nil {
		return err
	}
	if strings.Contains(uri, "cluster?") {
		return json.Unmarshal(res.Results, &opts.result)
	}
	if res.Count == 0 {
		return errors.New("No object of type " + obj + " with name " + opts.name + " is found")
	} else if res.Count > 1 {
		return errors.New("More than one object of type " + obj + " with name " + opts.name + " is found")
	}
	elems := make([]json.RawMessage, 1)
	err = json.Unmarshal(res.Results, &elems)
	if err != nil {
		return err
	}
	return json.Unmarshal(elems[0], &opts.result)

}

// GetObjectByName performs GET with name filter
func (avisess *AviSession) GetObjectByName(obj string, name string, result interface{}, options ...ApiOptionsParams) error {
	opts, err := getOptions(options)
	if err != nil {
		return err
	}
	return avisess.GetObject(obj, SetName(name), SetResult(result), SetOptTenant(opts.tenant))
}

// Utility functions

// GetControllerVersion gets the version number from the Avi Controller
func (avisess *AviSession) GetControllerVersion() (string, error) {
	var resp interface{}
	err := avisess.Get("/api/initial-data", &resp)
	if err != nil {
		return "", err
	}
	version := resp.(map[string]interface{})["version"].(map[string]interface{})["Version"].(string)
	return version, nil
}

// Logout performs log out operation of the Avi Controller
func (avisess *AviSession) Logout() error {
	url := avisess.prefix + "logout"
	req, _ := avisess.newAviRequest(context.Background(), "POST", url, nil, avisess.tenant)
	_, err := avisess.client.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func (avisess *AviSession) ResetPassword(password string) error {
	avisess.Logout()
	avisess.password = password
	return nil
}

func updateUri(uri string, opts *ApiOptions) string {
	if strings.Contains(uri, "?") {
		uri += "&"
	} else {
		uri += "?"
	}
	for k, v := range opts.params {
		if (k == "name" && opts.name != "") || (opts.cloud != "" && k == "cloud") ||
			(opts.includeName && k == "include_name") || (opts.skipDefault && k == "skip_default") ||
			(opts.cloudUUID != "" && k == "cloud_ref.uuid") {
			continue
		} else {
			uri += k + "=" + v + "&"
		}
	}
	return uri
}

func GetIPVersion(ipAddr string) net.IP {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		glog.Errorf("Controller Host is not valid.")
		return nil
	}
	return ip
}
