package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	log.Printf("Starting proxy")

	portStr := getEnv("NET3_HTTP_PROXY_PORT", "81")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(fmt.Errorf("Invalid value for port: %w", err))
	}
	targetHost := getEnv("NET3_HTTP_PROXY_TARGET_HOST", "localhost")
	targetPortStr := getEnv("NET3_HTTP_PROXY_TARGET_PORT", "80")
	targetPort, err := strconv.Atoi(targetPortStr)
	if err != nil {
		log.Fatal(fmt.Errorf("Invalid value for target port: %w", err))
	}

	log.Printf("Listening on port %v", port)
	log.Printf("Forwarding to host %q on port %v", targetHost, targetPort)

	http.HandleFunc("/", makeProxyHandleFunc(targetHost, targetPort))
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
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

func makeProxyHandleFunc(targetHost string, targetPort int) func(res http.ResponseWriter, req *http.Request) {
	targetUrl, err := url.Parse(fmt.Sprintf("http://%s:%v", targetHost, targetPort))
	if err != nil {
		log.Print(fmt.Errorf("Invalid proxy target url: %w", err))
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	return func(res http.ResponseWriter, req *http.Request) {
		logLines := make([]string, 0)

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading request body: %w", err))
		}

		logLines = append(logLines, "")
		logLines = append(logLines, "---------------")

		logLines = append(logLines, "")
		logLines = append(logLines, fmt.Sprintf("%s %s", req.Method, req.RequestURI))

		logLines = append(logLines, "")
		logLines = append(logLines, "Request headers")
		for name, values := range req.Header {
			for _, v := range values {
				logLines = append(logLines, fmt.Sprintf("%s: %s", name, v))
			}
		}

		if len(body) > 0 {
			logLines = append(logLines, "")
			logLines = append(logLines, "Request body")
			logLines = append(logLines, string(body))
		}

		log.Println(strings.Join(logLines, "\n"))

		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		proxy.ServeHTTP(res, req)
	}
}
