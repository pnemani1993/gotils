package client

import (
	"log"
	"net/http"
	"time"

	"io"
)

type HttpX struct {
	HttpRequest     *http.Request
	HttpResponse    *http.Response
	TransportConfig *http.Transport
	retryEnabled    bool
	retryCount      int
	retryInterval   time.Duration
	logger          *log.Logger
	timeout         time.Duration
	responseTime    time.Duration
}

func (httpX *HttpX) Send() (*http.Response, error) {

	httpX.logger = log.Default()

	httpClient := &http.Client{
		Transport: httpX.TransportConfig,
		Timeout:   httpX.timeout,
	}
	var response *http.Response
	var err error
	count := 0
	var startTime time.Time
	var responseTime time.Duration

	if httpX.retryEnabled {
		for count = range httpX.retryCount {
			startTime = time.Now()
			response, err = httpClient.Do(httpX.HttpRequest)
			responseTime = time.Since(startTime)
			if err != nil {
				time.Sleep(httpX.retryInterval)
				httpX.logger.Printf("Retrying after %d attempt", count+1)
				continue
			}

			if response.StatusCode > 299 {
				time.Sleep(httpX.retryInterval)
				httpX.logger.Printf("Retrying after %d attempt", count+1)
				continue
			}
			break
		}
	} else {
		startTime = time.Now()
		response, err = httpClient.Do(httpX.HttpRequest)
		responseTime = time.Since(startTime)
	}

	if err != nil {
		errorResponse := &InvalidOperation{1005, "Error sending HTTP/HTTPS request: " + err.Error()}
		if httpX.retryEnabled {
			httpX.logger.Printf("Request failed after %d retries", count+1)
		}
		httpX.HttpResponse = nil
		return nil, errorResponse
	}

	httpX.HttpResponse = response
	httpX.responseTime = responseTime
	if response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(response.Body)
		errorResponse := &HttpError{response.StatusCode, string(responseMessage)}
		if httpX.retryEnabled {
			httpX.logger.Printf("Request failed after %d retries", count+1)
		}
		return response, errorResponse
	}

	httpX.logger.Printf("Http Request successful: %s", httpX.HttpRequest.URL.Host)
	return response, nil
}

func (httpX *HttpX) SendAsync() *HttpFuture {
	httpX.logger = log.Default()
	httpFuture := &HttpFuture{responseChannel: make(chan int), isDone: false}
	go httpX.sendingAsync(httpFuture)
	return httpFuture
}

func (httpX *HttpX) sendingAsync(httpFuture *HttpFuture) {

	httpClient := &http.Client{
		Transport: httpX.TransportConfig,
		Timeout:   httpX.timeout,
	}

	var response *http.Response
	var err error
	count := 0
	var startTime time.Time
	var responseTime time.Duration

	if httpX.retryEnabled {
		for count = range httpX.retryCount {
			startTime = time.Now()
			response, err = httpClient.Do(httpX.HttpRequest)
			responseTime = time.Since(startTime)
			if err != nil {
				time.Sleep(httpX.retryInterval)
				httpX.logger.Printf("Retrying after %d attempt", count+1)
				continue
			}

			if response.StatusCode > 299 {
				time.Sleep(httpX.retryInterval)
				httpX.logger.Printf("Retrying after %d attempt", count+1)
				continue
			}
			break
		}
	} else {
		startTime = time.Now()
		response, err = httpClient.Do(httpX.HttpRequest)
		responseTime = time.Since(startTime)
	}

	if err != nil {
		errorResponse := &InvalidOperation{1005, "Error sending HTTP/HTTPS request: " + err.Error()}
		if httpX.retryEnabled {
			httpX.logger.Printf("Request failed after %d retries", count+1)
		}
		httpFuture.response = nil
		httpFuture.err = errorResponse
		httpFuture.isDone = true
		httpFuture.responseChannel <- 1
		return
	}

	if response.StatusCode > 299 {
		responseMessage, _ := io.ReadAll(response.Body)
		errorResponse := &HttpError{response.StatusCode, string(responseMessage)}
		if httpX.retryEnabled {
			httpX.logger.Printf("Request failed after %d retries", count+1)
		}
		httpFuture.response = response
		httpFuture.err = errorResponse
		httpFuture.isDone = true
		httpFuture.responseChannel <- 2
		return
	}
	httpX.HttpResponse = response
	httpX.responseTime = responseTime
	httpFuture.response = response
	httpFuture.err = nil
	httpFuture.isDone = true
	httpFuture.responseChannel <- 0
	httpX.logger.Printf("Http Request successful: %s", httpX.HttpRequest.URL.Host)
}

func (httpX *HttpX) GetResponseTime() time.Duration {
	return httpX.responseTime
}
