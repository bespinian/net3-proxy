package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("Starting proxy")

	portStr := getEnv("NET3_HTTP_PROXY_PORT", "81")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(fmt.Errorf("Invalid value for port: %w", err))
	}
	targetProtocol := getEnv("NET3_HTTP_PROXY_TARGET_PROTOCOL", "http")
	targetHost := getEnv("NET3_HTTP_PROXY_TARGET_HOST", "localhost")
	targetPortStr := getEnv("NET3_HTTP_PROXY_TARGET_PORT", "80")
	targetPort, err := strconv.Atoi(targetPortStr)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(fmt.Errorf("Invalid value for target port: %w", err))
	}

	log.Printf("Listening on port %v", port)
	log.Printf("Forwarding to host %q on port %v", targetHost, targetPort)

	http.HandleFunc("/", makeProxyHandleFunc(targetProtocol, targetHost, targetPort))
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(fmt.Errorf("Could not start proxy server: %w", err))
	}
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Variable %q not set. Falling back to default %q", key, fallback)
		return fallback
	}
	return value
}

func makeProxyHandleFunc(targetProtocol, targetHost string, targetPort int) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		bodyCopy, err := req.GetBody()
		if err != nil {
			log.SetOutput(os.Stderr)
			log.Print(fmt.Errorf("error getting request body: %w", err))
		}

		body, err := ioutil.ReadAll(bodyCopy)
		if err != nil {
			log.SetOutput(os.Stderr)
			log.Print(fmt.Errorf("error reading request body: %w", err))
		}

		log.SetOutput(os.Stdout)

		log.Println("Request headers:")
		for name, values := range req.Header {
			for _, v := range values {
				log.Printf("%s: %s", name, v)
			}
		}

		log.Println("Request body:")
		log.Println(body)

		targetUrl, err := url.Parse(fmt.Sprintf("%s://%s:%v", targetProtocol, targetHost, targetPort))
		if err != nil {
			log.SetOutput(os.Stderr)
			log.Print(fmt.Errorf("Invalid proxy target url: %w", err))
		}
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		proxy.ServeHTTP(res, req)
	}
}
