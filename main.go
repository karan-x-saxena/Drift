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

	if *configYaml != "" {
		_, err := config.NewYamlConfig(*configYaml)
		if err != nil {
			slog.Error(err.Error())
			return
		}
	}

	s := &fasthttp.Server{
		Handler: ProxyHandler,

		// Every response will contain 'Server: My super server' header.
		Name: "My super server",

		// Other Server settings may be set here.
	}
	addr := fmt.Sprintf(":%s", *port)
	if err := s.ListenAndServe(addr); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
