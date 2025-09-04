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
	"sync"

	"io"

	"bytes"
	"encoding/json"

	"golang.org/x/crypto/pkcs12"
)

type HttpX struct {
	HttpRequest     *http.Request
	HttpResponse    *http.Response
	TransportConfig *http.Transport
	wg              sync.WaitGroup
}

type HttpXBuilder struct {
	UrlPaths      Path
	Url           *url.URL
	Headers       map[string][]string
	HttpRequest   *http.Request
	TlsConfig     *tls.Config
	ProxyUrl      string
	wg            sync.WaitGroup
	Error         *ClientError
	logger        *log.Logger
	isPeekEnabled bool
}

func NewHttpXBuilder() *HttpXBuilder {
	requestContext, err := http.NewRequestWithContext(context.TODO(), "", "", nil)
	if err != nil {
		log.Default()
		fmt.Println("Error in the context")
		return nil
	}

	return &HttpXBuilder{
		UrlPaths:    NewPath(),
		Url:         &url.URL{},
		Headers:     make(map[string][]string),
		HttpRequest: requestContext,
		Error:       &ClientError{errorsList: make([]error, 0, 5)},
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
		builder.UrlPaths.Add(url[index:])
		url = url[:index]
	}

	builder.Url.Host = url

	return builder
}

// Enter the path to the URL
func (builder *HttpXBuilder) SetPath(path string) *HttpXBuilder {
	err := builder.UrlPaths.Add(path)
	if err != nil {
		builder.Error.errorsList = append(builder.Error.errorsList, err)
	}
	return builder
}

// Enter path with the required substitutions
func (builder *HttpXBuilder) SetPathf(path string, values ...string) *HttpXBuilder {
	err := builder.UrlPaths.Addf(path, values)
	if err != nil {
		builder.Error.errorsList = append(builder.Error.errorsList, err)
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
	builder.Headers[key] = []string{value}
	return builder
}

func (builder *HttpXBuilder) Get() *HttpXBuilder {
	builder.HttpRequest.Method = "GET"
	return builder
}

func (builder *HttpXBuilder) Post() *HttpXBuilder {
	builder.HttpRequest.Method = "POST"
	builder.HttpRequest.ContentLength = -1
	return builder
}

func (builder *HttpXBuilder) WithStringAsBody(body string) *HttpXBuilder {
	requestBody := io.NopCloser(strings.NewReader(body))
	builder.HttpRequest.Body = requestBody
	return builder
}

func (builder *HttpXBuilder) WithStructAsBody(body any) *HttpXBuilder {
	jsonData, err := json.Marshal(body)
	if err != nil {
		jsonError := &InvalidInput{1003, err.Error(), fmt.Sprint(body)}
		builder.Error.errorsList = append(builder.Error.errorsList, jsonError)
		return builder
	}
	requestBody := io.NopCloser(bytes.NewReader(jsonData))
	builder.HttpRequest.Body = requestBody
	return builder
}

func (builder *HttpXBuilder) Delete() *HttpXBuilder {
	builder.HttpRequest.Method = "DELETE"
	return builder
}

func (builder *HttpXBuilder) Put() *HttpXBuilder {
	builder.HttpRequest.Method = "PUT"
	builder.HttpRequest.ContentLength = -1
	return builder
}

func (builder *HttpXBuilder) Patch() *HttpXBuilder {
	builder.HttpRequest.Method = "PATCH"
	return builder
}

func (builder *HttpXBuilder) Head() *HttpXBuilder {
	builder.HttpRequest.Method = "HEAD"
	return builder
}

func (builder *HttpXBuilder) InsecureSkipVerify() *HttpXBuilder {
	builder.TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	return builder
}

func (builder *HttpXBuilder) TLSConfigPEM(path string) *HttpXBuilder {
	builder.TlsConfig = builder.tlsConfigPEM(path)
	return builder
}

func (builder *HttpXBuilder) TLSConfigPKCS12(path string, password string) *HttpXBuilder {
	builder.TlsConfig = builder.tlsConfigPKCS12(path, password)
	return builder
}

func (builder *HttpXBuilder) SetProxy(proxyUrl string) *HttpXBuilder {
	builder.ProxyUrl = proxyUrl
	return builder
}

func (builder *HttpXBuilder) Peek() *HttpXBuilder {
	builder.isPeekEnabled = true
	builder.logger = log.Default()
	return builder
}

func (builder *HttpXBuilder) Build() (*HttpX, error) {
	builder.Url.Path = builder.UrlPaths.GetPath()
	builder.HttpRequest.URL = builder.Url
	builder.HttpRequest.Header = builder.Headers

	if (builder.HttpRequest.Method == "GET" || builder.HttpRequest.Method == "DELETE") && builder.HttpRequest.Body != nil {
		builder.Error.errorsList = append(builder.Error.errorsList, &InvalidOperation{1002, "request body should not be provided for the given request method"})
	} else if (builder.HttpRequest.Method == "POST" || builder.HttpRequest.Method == "PUT" || builder.HttpRequest.Method == "PATCH") && builder.HttpRequest.Body == nil {
		builder.Error.errorsList = append(builder.Error.errorsList, &InvalidOperation{1003, "request body should be provided for the given request method"})
	}

	transport := &http.Transport{
		TLSClientConfig: builder.TlsConfig,
	}

	proxy, exists := builder.getProxy()
	if exists {
		transport.Proxy = http.ProxyURL(proxy)
	}

	if len(builder.Error.errorsList) != 0 {
		return nil, builder.Error
	}

	if builder.isPeekEnabled {
		builder.logger.Printf("%s %s%s\n", builder.HttpRequest.Method, builder.Url.Path, builder.Url.RawQuery)
		builder.logger.Printf("Host: %s\n", builder.Url.Host)
		for key, value := range builder.Headers {
			if key == "Authorization" {
				builder.logger.Printf("%s: %s", key, "<Token>")
			}
			builder.logger.Printf("%s: %s\n", key, value)
		}
		if builder.HttpRequest.Method == "POST" || builder.HttpRequest.Method == "PUT" || builder.HttpRequest.Method == "PATCH" {
			body, _ := io.ReadAll(builder.HttpRequest.Body)
			builder.logger.Printf("\n%s", string(body))
		}
	}

	return &HttpX{
		HttpRequest:     builder.HttpRequest,
		TransportConfig: transport,
	}, nil
}

func (httpX *HttpX) Send() (*http.Response, error) {

	httpClient := &http.Client{
		Transport: httpX.TransportConfig,
	}
	response, err := httpClient.Do(httpX.HttpRequest)
	if err != nil {
		errorResponse := &InvalidOperation{1005, "Error sending HTTP/HTTPS request: " + err.Error()}
		return nil, errorResponse
	}

	if response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(response.Body)
		errorResponse := &HttpError{response.StatusCode, string(responseMessage)}
		return response, errorResponse
	}
	httpX.HttpResponse = response
	return response, nil
}

func (builder *HttpXBuilder) tlsConfigPEM(path string) *tls.Config {
	caCert, err := os.ReadFile(path)
	if err != nil {
		errorResponse := &InvalidInput{1004, "Error loading CA certificate: " + err.Error(), path}
		builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
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
		builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
	}

	blocks, err := pkcs12.ToPEM(pfxData, password)
	if err != nil {
		errorResponse := &InvalidInput{1004, "Error decoding PKCS12/Invalid password: " + err.Error(), path}
		builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
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
		builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return tlsConfig
}

func (builder *HttpXBuilder) getProxy() (*url.URL, bool) {
	if len(builder.ProxyUrl) == 0 {

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
					builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			} else if proxy, exists := os.LookupEnv("ALL_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading ALL_PROXY: " + err.Error()}
					builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			}
		} else if builder.Url.Scheme == "http" {
			if proxy, exists := os.LookupEnv("HTTP_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading HTTP_PROXY: " + err.Error()}
					builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			} else if proxy, exists := os.LookupEnv("ALL_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					errorResponse := &InvalidOperation{1004, "Error reading ALL_PROXY: " + err.Error()}
					builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
					return nil, false
				}
				return proxy, true
			}
		}

	} else {
		proxyUrl, err := url.Parse(builder.ProxyUrl)
		if err != nil {
			errorResponse := &InvalidInput{1005, "Error reading proxy input: " + err.Error(), builder.ProxyUrl}
			builder.Error.errorsList = append(builder.Error.errorsList, errorResponse)
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
