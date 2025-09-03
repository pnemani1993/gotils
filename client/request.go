package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
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
	HttpResponse    http.Response
	TransportConfig *http.Transport
	wg              sync.WaitGroup
	Error           []error
}

type HttpXBuilder struct {
	UrlPaths     Path
	Url          *url.URL
	Headers      map[string][]string
	HttpRequest  *http.Request
	HttpResponse http.Response
	TlsConfig    *tls.Config
	ProxyUrl     string
	wg           sync.WaitGroup
	Error        []error
}

func NewHttpXBuilder() *HttpXBuilder {
	requestContext, err := http.NewRequestWithContext(context.TODO(), "", "", nil)
	if err != nil {
		fmt.Println("Error in the context")
		return nil
	}

	return &HttpXBuilder{
		UrlPaths:    NewPath(),
		Url:         &url.URL{},
		Headers:     make(map[string][]string),
		HttpRequest: requestContext,
		Error:       make([]error, 0, 5),
	}
}

// Enter the host url along with the port number in the format:
// BaseUrl("https://hosturl.domain:port")
func (builder *HttpXBuilder) BaseUrl(url string) *HttpXBuilder {
	url, _ = strings.CutSuffix(url, "/")

	if strings.Contains(url, "//localhost") {
		url = strings.Replace(url, "//localhost", "//127.0.0.1", 1)
	}

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
		builder.Error = append(builder.Error, err)
	}
	return builder
}

// Enter path with the required substitutions
func (builder *HttpXBuilder) SetPathf(path string, values ...string) *HttpXBuilder {
	err := builder.UrlPaths.Addf(path, values)
	if err != nil {
		builder.Error = append(builder.Error, err)
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
		builder.Error = append(builder.Error, err)
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
	builder.TlsConfig = tlsConfigPEM(path)
	return builder
}

func (builder *HttpXBuilder) TLSConfigPKCS12(path string, password string) *HttpXBuilder {
	builder.TlsConfig = tlsConfigPKCS12(path, password)
	return builder
}

func (builder *HttpXBuilder) SetProxy(proxyUrl string) *HttpXBuilder {
	builder.ProxyUrl = proxyUrl
	return builder
}

func (builder *HttpXBuilder) Build() (*HttpX, error) {
	builder.Url.Path = builder.UrlPaths.GetPath()
	builder.HttpRequest.URL = builder.Url
	builder.HttpRequest.Header = builder.Headers

	if (builder.HttpRequest.Method == "GET" || builder.HttpRequest.Method == "DELETE") && builder.HttpRequest.Body != nil {
		builder.Error = append(builder.Error, errors.New("request body should not be provided for the given request method"))
	} else if (builder.HttpRequest.Method == "POST" || builder.HttpRequest.Method == "PUT" || builder.HttpRequest.Method == "PATCH") && builder.HttpRequest.Body == nil {
		builder.Error = append(builder.Error, errors.New("request body should be provided for the given request method"))
	}

	transport := &http.Transport{
		TLSClientConfig: builder.TlsConfig,
	}

	proxy, exists := builder.getProxy()
	if exists {
		transport.Proxy = http.ProxyURL(proxy)
	}

	if len(builder.Error) != 0 {
		var errorString string
		for _, erro := range builder.Error {
			errorString = errorString + "\n" + erro.Error()
		}
		return nil, errors.New(errorString)
	}

	return &HttpX{
		HttpRequest:     builder.HttpRequest,
		Error:           builder.Error,
		TransportConfig: transport,
	}, nil
}

func (httpX *HttpX) Send() (*http.Response, error) {

	httpClient := &http.Client{
		Transport: httpX.TransportConfig,
	}
	response, err := httpClient.Do(httpX.HttpRequest)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func tlsConfigPEM(path string) *tls.Config {
	caCert, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error loading CA certificate:", err)
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

func tlsConfigPKCS12(path string, password string) *tls.Config {
	pfxData, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Errorf("Error reading PKCS12 file: %v", err)
	}

	// Replace "your_password" with the actual password for your PKCS12 file
	blocks, err := pkcs12.ToPEM(pfxData, password)
	if err != nil {
		fmt.Errorf("Error decoding PKCS12: %v", err)
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
		fmt.Errorf("Error creating X509 key pair: %v", err)
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
					builder.Error = append(builder.Error, errors.New("proxy error"))
					return nil, false
				}
				return proxy, true
			} else if proxy, exists := os.LookupEnv("ALL_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					builder.Error = append(builder.Error, errors.New("proxy error"))
					return nil, false
				}
				return proxy, true
			}
		} else if builder.Url.Scheme == "http" {
			if proxy, exists := os.LookupEnv("HTTP_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					builder.Error = append(builder.Error, errors.New("proxy error"))
					return nil, false
				}
				return proxy, true
			} else if proxy, exists := os.LookupEnv("ALL_PROXY"); exists {
				proxy, err := url.Parse(proxy) // Replace with your proxy details
				if err != nil {
					builder.Error = append(builder.Error, errors.New("proxy error"))
					return nil, false
				}
				return proxy, true
			}
		}

	} else {
		proxyUrl, err := url.Parse(builder.ProxyUrl)
		if err != nil {
			builder.Error = append(builder.Error, errors.New("proxy error"))
			return nil, false
		}
		return proxyUrl, true
	}
	return nil, false
}
