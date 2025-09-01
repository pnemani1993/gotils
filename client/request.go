package client

import (
	"net/http"
	"net/url"
	"sync"
)

type HttpX struct {
	UrlPaths     Path
	Url          url.URL
	HttpRequest  http.Request
	HttpResponse http.Response
	wg           sync.WaitGroup
}
