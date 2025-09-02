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
	"strings"
	"sync"

	"golang.org/x/crypto/pkcs12"
)

type HttpX struct {
	HttpRequest  *http.Request
	HttpResponse http.Response
	TlsConfig    *tls.Config
	wg           sync.WaitGroup
	Error        []error
}

type HttpXBuilder struct {
	UrlPaths     Path
	Url          *url.URL
	Headers      map[string][]string
	HttpRequest  *http.Request
	HttpResponse http.Response
	TlsConfig    *tls.Config
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
	if strings.HasPrefix(url, "https://") {
		url, _ = strings.CutPrefix(url, "https://")
		builder.Url.Scheme = "https"
	} else if strings.HasPrefix(url, "http://") {
		url, _ = strings.CutPrefix(url, "http://")
		builder.Url.Scheme = "http"
	} else {
		builder.Url.Scheme = "https"
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
	return builder
}

func (builder *HttpXBuilder) Delete() *HttpXBuilder {
	builder.HttpRequest.Method = "DELETE"
	return builder
}

func (builder *HttpXBuilder) Put() *HttpXBuilder {
	builder.HttpRequest.Method = "PUT"
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

func (builder *HttpXBuilder) Build() (*HttpX, error) {
	builder.Url.Path = builder.UrlPaths.GetPath()
	builder.HttpRequest.URL = builder.Url
	builder.HttpRequest.Header = builder.Headers
	builder.HttpRequest.ContentLength = -1
	if len(builder.Error) != 0 {
		var errorString string
		for _, erro := range builder.Error {
			errorString = errorString + "\n" + erro.Error()
		}
		return nil, errors.New(errorString)
	}

	return &HttpX{
		HttpRequest: builder.HttpRequest,
		Error:       builder.Error,
		TlsConfig:   builder.TlsConfig,
	}, nil
}

func (httpX *HttpX) Send() (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: httpX.TlsConfig,
	}
	httpClient := &http.Client{
		Transport: transport,
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
