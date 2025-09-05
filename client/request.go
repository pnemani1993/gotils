package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"io"

	"bytes"
	"encoding/json"

	"golang.org/x/crypto/pkcs12"
)

type HttpX struct {
	HttpRequest     *http.Request
	HttpResponse    *http.Response
	TransportConfig *http.Transport
	retryEnabled    bool
	retryCount      int
	retryInterval   time.Duration
	logger          *log.Logger
}

type HttpXBuilder struct {
	urlPaths      Path
	Url           *url.URL
	headers       map[string][]string
	httpRequest   *http.Request
	tlsConfig     *tls.Config
	proxyUrl      string
	errorsList    *ClientError
	logger        *log.Logger
	isPeekEnabled bool
	retryEnabled  bool
	retryCount    int
	retryInterval time.Duration
}

func NewHttpXBuilder() *HttpXBuilder {
	requestContext, err := http.NewRequestWithContext(context.TODO(), "", "", nil)
	if err != nil {
		log.Default()
		fmt.Println("Error in the context")
		return nil
	}

	return &HttpXBuilder{
		urlPaths:    NewPath(),
		Url:         &url.URL{},
		headers:     make(map[string][]string),
		httpRequest: requestContext,
		errorsList:  &ClientError{errorsList: make([]error, 0, 5)},
	}
}

// Enter the host url along with the port number in the format:
// BaseUrl("https://hosturl.domain:port")
func (builder *HttpXBuilder) BaseUrl(url string) *HttpXBuilder {
	url, _ = strings.CutSuffix(url, "/")

	url = strings.Replace(url, "//localhost", "//127.0.0.1", 1)

	if strings.HasPrefix(url, "https://") {
		url, _ = strings.CutPrefix(url, "https://")
		builder.Url.Scheme = "https"
	} else if strings.HasPrefix(url, "http://") {
		url, _ = strings.CutPrefix(url, "http://")
		builder.Url.Scheme = "http"
	} else {
		builder.Url.Scheme = "https"
	}

	if index := strings.Index(url, "/"); index != -1 {
		builder.urlPaths.Add(url[index:])
		url = url[:index]
	}

	builder.Url.Host = url

	return builder
}

// Enter the path to the URL
func (builder *HttpXBuilder) SetPath(path string) *HttpXBuilder {
	err := builder.urlPaths.Add(path)
	if err != nil {
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, err)
	}
	return builder
}

// Enter path with the required substitutions
func (builder *HttpXBuilder) SetPathf(path string, values ...string) *HttpXBuilder {
	err := builder.urlPaths.Addf(path, values)
	if err != nil {
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, err)
	}
	return builder
}

// Enter the query parameters as key, value
func (builder *HttpXBuilder) SetQueryParameters(key string, value string) *HttpXBuilder {
	if len(builder.Url.RawQuery) == 0 {
		builder.Url.RawQuery = key + "=" + value
		return builder
	}
	builder.Url.RawQuery = builder.Url.RawQuery + "&" + key + "=" + value
	return builder
}

func (builder *HttpXBuilder) SetHeader(key string, value string) *HttpXBuilder {
	builder.headers[key] = []string{value}
	return builder
}

func (builder *HttpXBuilder) Get() *HttpXBuilder {
	builder.httpRequest.Method = "GET"
	return builder
}

func (builder *HttpXBuilder) Post() *HttpXBuilder {
	builder.httpRequest.Method = "POST"
	builder.httpRequest.ContentLength = -1
	return builder
}

func (builder *HttpXBuilder) WithStringAsBody(body string) *HttpXBuilder {
	requestBody := io.NopCloser(strings.NewReader(body))
	builder.httpRequest.Body = requestBody
	return builder
}

func (builder *HttpXBuilder) WithStructAsBody(body any) *HttpXBuilder {
	jsonData, err := json.Marshal(body)
	if err != nil {
		jsonError := &InvalidInput{1003, err.Error(), fmt.Sprint(body)}
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, jsonError)
		return builder
	}
	requestBody := io.NopCloser(bytes.NewReader(jsonData))
	builder.httpRequest.Body = requestBody
	return builder
}

func (builder *HttpXBuilder) Delete() *HttpXBuilder {
	builder.httpRequest.Method = "DELETE"
	return builder
}

func (builder *HttpXBuilder) Put() *HttpXBuilder {
	builder.httpRequest.Method = "PUT"
	builder.httpRequest.ContentLength = -1
	return builder
}

func (builder *HttpXBuilder) Patch() *HttpXBuilder {
	builder.httpRequest.Method = "PATCH"
	return builder
}

func (builder *HttpXBuilder) Head() *HttpXBuilder {
	builder.httpRequest.Method = "HEAD"
	return builder
}

func (builder *HttpXBuilder) RetryEnabled(count int) *HttpXBuilder {
	builder.retryEnabled = true
	builder.retryCount = count
	builder.retryInterval = time.Millisecond * 1000
	return builder
}

func (builder *HttpXBuilder) RetryInterval(interval int) *HttpXBuilder {
	builder.retryInterval = time.Millisecond * time.Duration(interval)
	return builder
}

func (builder *HttpXBuilder) InsecureSkipVerify() *HttpXBuilder {
	builder.tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return builder
}

func (builder *HttpXBuilder) TLSConfigPEM(path string) *HttpXBuilder {
	builder.tlsConfig = builder.tlsConfigPEM(path)
	return builder
}

func (builder *HttpXBuilder) TLSConfigPKCS12(path string, password string) *HttpXBuilder {
	builder.tlsConfig = builder.tlsConfigPKCS12(path, password)
	return builder
}

func (builder *HttpXBuilder) SetProxy(proxyUrl string) *HttpXBuilder {
	builder.proxyUrl = proxyUrl
	return builder
}

func (builder *HttpXBuilder) Peek() *HttpXBuilder {
	builder.isPeekEnabled = true
	builder.logger = log.Default()
	return builder
}

func (builder *HttpXBuilder) Build() (*HttpX, error) {
	builder.Url.Path = builder.urlPaths.GetPath()
	builder.httpRequest.URL = builder.Url
	builder.httpRequest.Header = builder.headers

	if (builder.httpRequest.Method == "GET" || builder.httpRequest.Method == "DELETE") && builder.httpRequest.Body != nil {
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, &InvalidOperation{1002, "request body should not be provided for the given request method"})
	} else if (builder.httpRequest.Method == "POST" || builder.httpRequest.Method == "PUT" || builder.httpRequest.Method == "PATCH") && builder.httpRequest.Body == nil {
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, &InvalidOperation{1003, "request body should be provided for the given request method"})
	}

	transport := &http.Transport{
		TLSClientConfig: builder.tlsConfig,
	}

	proxy, exists := builder.getProxy()
	if exists {
		transport.Proxy = http.ProxyURL(proxy)
	}

	if len(builder.errorsList.errorsList) != 0 {
		return nil, builder.errorsList
	}

	if builder.isPeekEnabled {
		builder.logger.Printf("%s %s%s\n", builder.httpRequest.Method, builder.Url.Path, builder.Url.RawQuery)
		builder.logger.Printf("Host: %s\n", builder.Url.Host)
		for key, value := range builder.headers {
			if key == "Authorization" {
				builder.logger.Printf("%s: %s", key, "<Token>")
			}
			builder.logger.Printf("%s: %s\n", key, value)
		}
		if builder.httpRequest.Method == "POST" || builder.httpRequest.Method == "PUT" || builder.httpRequest.Method == "PATCH" {
			body, _ := io.ReadAll(builder.httpRequest.Body)
			builder.logger.Printf("\n%s", string(body))
		}
	}

	return &HttpX{
		HttpRequest:     builder.httpRequest,
		TransportConfig: transport,
		retryEnabled:    builder.retryEnabled,
		retryCount:      builder.retryCount,
		retryInterval:   builder.retryInterval,
	}, nil
}

func (httpX *HttpX) Send() (*http.Response, error) {

	httpX.logger = log.Default()

	httpClient := &http.Client{
		Transport: httpX.TransportConfig,
	}
	var response *http.Response
	var err error
	count := 0

	if httpX.retryEnabled {
		for count = range httpX.retryCount {
			response, err = httpClient.Do(httpX.HttpRequest)
			if err != nil {
				time.Sleep(httpX.retryInterval)
				httpX.logger.Printf("Retrying after %d attempt", count)
				continue
			}

			if response.StatusCode > 299 {
				time.Sleep(httpX.retryInterval)
				httpX.logger.Printf("Retrying after %d attempt", count)
				continue
			}
			break
		}
	} else {
		response, err = httpClient.Do(httpX.HttpRequest)
	}

	if err != nil {
		errorResponse := &InvalidOperation{1005, "Error sending HTTP/HTTPS request: " + err.Error()}
		if count > 0 {
			httpX.logger.Printf("Request failed after &d retries", count)
		}
		httpX.HttpResponse = nil
		return nil, errorResponse
	}

	httpX.HttpResponse = response

	if response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(response.Body)
		errorResponse := &HttpError{response.StatusCode, string(responseMessage)}
		if count > 0 {
			httpX.logger.Printf("Request failed after &d retries", count)
		}
		return response, errorResponse
	}

	httpX.logger.Printf("Http Request successful: %s", httpX.HttpRequest.URL.RawPath)
	return response, nil
}

func (httpX *HttpX) SendAsync() *HttpFuture {
	httpFuture := &HttpFuture{responseChannel: make(chan *http.Response)}
	go httpX.sendingAsync(httpFuture)
	return httpFuture
}

func (httpX *HttpX) sendingAsync(httpFuture *HttpFuture) {

	httpClient := &http.Client{
		Transport: httpX.TransportConfig,
	}
	response, err := httpClient.Do(httpX.HttpRequest)
	if err != nil {
		errorResponse := &InvalidOperation{1005, "Error sending HTTP/HTTPS request: " + err.Error()}
		httpFuture.responseChannel <- nil
		httpFuture.err = errorResponse
		return
	}

	if response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(response.Body)
		errorResponse := &HttpError{response.StatusCode, string(responseMessage)}
		httpFuture.responseChannel <- response
		httpFuture.err = errorResponse
		return
	}
	httpX.HttpResponse = response
	httpFuture.responseChannel <- response
	httpFuture.err = nil
}

func (builder *HttpXBuilder) tlsConfigPEM(path string) *tls.Config {
	caCert, err := os.ReadFile(path)
	if err != nil {
		errorResponse := &InvalidInput{1004, "Error loading CA certificate: " + err.Error(), path}
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
		return nil
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	return tlsConfig
}

func (builder *HttpXBuilder) tlsConfigPKCS12(path string, password string) *tls.Config {
	pfxData, err := os.ReadFile(path)
	if err != nil {
		errorResponse := &InvalidInput{1004, "Error reading PKCS12 file: " + err.Error(), path}
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
	}

	blocks, err := pkcs12.ToPEM(pfxData, password)
	if err != nil {
		errorResponse := &InvalidInput{1004, "Error decoding PKCS12/Invalid password: " + err.Error(), path}
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
	}

	var pemCert []byte
	var pemKey []byte

	for _, b := range blocks {
		if b.Type == "CERTIFICATE" {
			pemCert = append(pemCert, pem.EncodeToMemory(b)...)
		} else if b.Type == "PRIVATE KEY" {
			pemKey = append(pemKey, pem.EncodeToMemory(b)...)
		}
	}

	cert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		errorResponse := &InvalidInput{1004, "Error creating X509 key pair: " + err.Error(), path}
		builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig
}

func (builder *HttpXBuilder) getProxy() (*url.URL, bool) {
	if len(builder.proxyUrl) == 0 {

		if no_proxy, exists := os.LookupEnv("NO_PROXY"); exists {
			no_proxy_urls := strings.Split(no_proxy, ",")
			for _, value := range no_proxy_urls {
				if strings.HasPrefix(builder.Url.Host, value) {
					return nil, false
				}
			}
		}

		if builder.Url.Scheme == "https" {
			if proxy, exists := os.LookupEnv("HTTPS_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading HTTPS_PROXY: " + err.Error()}
					builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			} else if proxy, exists := os.LookupEnv("ALL_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading ALL_PROXY: " + err.Error()}
					builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			}
		} else if builder.Url.Scheme == "http" {
			if proxy, exists := os.LookupEnv("HTTP_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading HTTP_PROXY: " + err.Error()}
					builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			} else if proxy, exists := os.LookupEnv("ALL_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading ALL_PROXY: " + err.Error()}
					builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			}
		}

	} else {
		proxyUrl, err := url.Parse(builder.proxyUrl)
		if err != nil {
			errorResponse := &InvalidInput{1005, "Error reading proxy input: " + err.Error(), builder.proxyUrl}
			builder.errorsList.errorsList = append(builder.errorsList.errorsList, errorResponse)
			return nil, false
		}
		return proxyUrl, true
	}
	return nil, false
}

type ClientError struct {
	errorsList []error
}

func (err *ClientError) Error() string {
	errorString := ""
	for _, errs := range err.errorsList {
		errorString = errorString + errs.Error() + "\n"
	}
	return errorString
}

type HttpFuture struct {
	responseChannel chan *http.Response
	err             error
}

func (future *HttpFuture) Get() (*http.Response, error) {
	response := <-future.responseChannel

	if future.err != nil {
		return response, future.err
	}

	if response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(response.Body)
		errorResponse := &HttpError{response.StatusCode, string(responseMessage)}
		return response, errorResponse
	}
	return response, nil
}
