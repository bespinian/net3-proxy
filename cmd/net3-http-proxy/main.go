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
	return func(res http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading request body: %w", err))
		}

		log.Println("Request headers:")
		for name, values := range req.Header {
			for _, v := range values {
				log.Printf("%s: %s", name, v)
			}
		}

		log.Println("Request body:")
		log.Println(string(body))

		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		targetUrl, err := url.Parse(fmt.Sprintf("http://%s:%v", targetHost, targetPort))
		if err != nil {
			log.Print(fmt.Errorf("Invalid proxy target url: %w", err))
		}
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		proxy.ServeHTTP(res, req)
	}
}
