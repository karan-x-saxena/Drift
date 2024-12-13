package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

	"drift/config"
	"drift/pkg"

	"github.com/valyala/fasthttp"
)

func OProxyHandler(ctx *fasthttp.RequestCtx) {
	const targetURL = "https://www.google.com" // Use HTTPS directly to avoid redirection overhead

	// Acquire request and response objects from pools
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	// Defer their release to avoid memory leaks
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// Efficiently copy the incoming request to the target request
	ctx.Request.CopyTo(req)

	// Set the target URL while preserving the incoming URI path
	req.SetRequestURIBytes(append([]byte(targetURL), ctx.URI().PathOriginal()...))

	// Send the request to the target and check for errors
	if err := fasthttp.Do(req, resp); err != nil {
		log.Printf("Proxy error: %v", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString("Internal Server Error")
		return
	}

	resp.CopyTo(&ctx.Response)

	// // Optimize response copying
	// ctx.Response.SetStatusCode(resp.StatusCode())
	// ctx.Response.Header.SetContentTypeBytes(resp.Header.Peek("Content-Type"))
	// ctx.Response.Header.SetBytesV("Content-Length", resp.Header.Peek("Content-Length"))
	// ctx.Response.SetBodyRaw(resp.Body())
}

func ProxyHandler(ctx *fasthttp.RequestCtx) {
	targetURL := "http://google.com"

	// Create a request for the target
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Copy the incoming request to the target request
	ctx.Request.CopyTo(req)

	// Modify the target host and URL
	req.SetRequestURI(targetURL + string(ctx.Request.URI().Path()))

	// Create a response for the target
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Forward the request to the target
	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Printf("Failed to proxy request: %v", err)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}

	// Copy the target's response back to the client
	resp.CopyTo(&ctx.Response)
}

func main() {
	port := flag.String("p", "8080", "Port to run the HTTP server")
	logFileName := flag.String("l", "", "Directory for log files")
	generateDefaultConfig := flag.String("g", "", "Generate Default COnfiguration for Drift")
	configYaml := flag.String("y", "", "Yaml config for Drift")

	flag.Parse()

	var logger pkg.Logger
	if *logFileName != "" {
		logger = pkg.Logger{LogFileName: *logFileName}
	} else {
		logger = pkg.Logger{}
	}

	logger.InitLogger()

	if *generateDefaultConfig != "" {
		err := config.BaseYamlFile()
		if err != nil {
			panic(err)
		}
		return
	}

	slog.Info("drift starts here!")
	var c config.Config

	if *configYaml != "" {
		var err error
		c, err = config.NewYamlConfig(*configYaml)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}

	var handler *pkg.HttpClient

	for _, server := range c.Servers {
		handler = pkg.NewHttpClient(server.Url, server.MaxConnection, server.MaxConnectionTimeout, server.MaxConnectionDuration, server.MaxIdemponentCallAttempts, server.MaxIdleConnectionDuration)
		go handler.HeathCheck(server.HealthCheckPath, c.HeathCheckTimer)

	}
	s := &fasthttp.Server{
		Handler: handler.ProxyHandler(),

		// Every response will contain 'Server: My super server' header.
		Name: "My super server",

		// Other Server settings may be set here.
	}
	addr := fmt.Sprintf(":%s", *port)
	if err := s.ListenAndServe(addr); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
