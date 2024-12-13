package pkg

import (
	"log"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type HttpClient struct {
	url      string
	client   fasthttp.Client
	pool     sync.Pool
	IsOnline bool
}

func (h *HttpClient) HeathCheck(url string, timer uint) {
	req := h.pool.Get().(*fasthttp.Request)
	defer func() {
		req.Reset()
		h.pool.Put(req) // Return the request object to the pool
	}()

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Set request details
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	// for key, value := range headers {
	// 	req.Header.Set(key, value)
	// }
	// if len(body) > 0 {
	// 	req.SetBody(body)
	// }

	for {
		var statusCode int

		err := h.client.Do(req, resp)
		if err != nil {
			statusCode = 400
		}
		statusCode = resp.StatusCode()

		if statusCode == 200 {
			h.IsOnline = true
		} else {
			h.IsOnline = false
		}

		time.Sleep(time.Second * time.Duration(timer))
	}

}

func (h *HttpClient) ProxyHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		req := h.pool.Get().(*fasthttp.Request)
		defer func() {
			req.Reset()
			h.pool.Put(req)
		}()

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		ctx.Request.CopyTo(req)
		req.SetRequestURIBytes(append([]byte(h.url), ctx.URI().PathOriginal()...))

		if err := fasthttp.Do(req, resp); err != nil {
			log.Printf("Proxy error: %v", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBodyString("Internal Server Error")
			return
		}

		resp.CopyTo(&ctx.Response)
	}
}

func NewHttpClient(url string, maxConn uint, maxConnTimeout uint, maxConnDuration uint, maxIdemponentCallAttempts uint, maxIdleConnDuration uint) *HttpClient {
	httpClient := &HttpClient{
		url: url,
		client: fasthttp.Client{
			MaxConnsPerHost:           int(maxConn),
			MaxConnWaitTimeout:        time.Duration(time.Duration(maxConnTimeout).Seconds()),
			MaxConnDuration:           time.Duration(maxConnDuration),
			MaxIdemponentCallAttempts: int(maxIdemponentCallAttempts),
			MaxIdleConnDuration:       time.Duration(maxIdleConnDuration),
		},
		pool: sync.Pool{
			New: func() interface{} {
				return fasthttp.AcquireRequest()
			},
		},
	}

	return httpClient
}
