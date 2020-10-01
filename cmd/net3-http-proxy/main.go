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
)

func main() {

	log.SetOutput(os.Stdout)
	log.Printf("Starting proxy.")

	net3HttpProxyPort := getEnv("NET3_HTTP_PROXY_PORT", "81")
	net3HttpProxyTargetProtocol := getEnv("NET3_HTTP_PROXY_TARGET_PROTOCOL", "http")
	net3HttpProxyTargetHost := getEnv("NET3_HTTP_PROXY_TARGET_HOST", "127.0.0.1")
	net3HttpProxyTargetPort := getEnv("NET3_HTTP_PROXY_TARGET_PORT", "80")

	log.Printf("Listening on port %q", net3HttpProxyPort)
	log.Printf("Forwarding to host %q on port %q", net3HttpProxyTargetHost, net3HttpProxyTargetPort)

	http.HandleFunc("/", getRequestAndForwardHandler(net3HttpProxyTargetProtocol, net3HttpProxyTargetHost, net3HttpProxyTargetPort))
	if err := http.ListenAndServe(":"+net3HttpProxyPort, nil); err != nil {
		log.SetOutput(os.Stderr)
		log.Print(fmt.Errorf("Could not start proxy server: %w", err))
		panic(err)
	}

}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Variable %q not set. Falling back to default %q", key, fallback)
	return fallback
}

func getRequestAndForwardHandler(targetProtocol, targetHost, targetPort string) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.SetOutput(os.Stderr)
			log.Print(fmt.Errorf("error reading body: %w", err))
			http.Error(res, "can't read body", http.StatusBadRequest)
			return
		}

		log.SetOutput(os.Stdout)

		log.Println("Request headers:")
		for name, values := range req.Header {
			for _, value := range values {
				log.Printf("%q:%q", name, value)
			}
		}

		log.Printf("Request body: %q", body)

		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		targetUrl, err := url.Parse(targetProtocol + "://" + targetHost + ":" + targetPort)
		if err != nil {
			log.SetOutput(os.Stderr)
			log.Print(fmt.Errorf("invalid proxy target url: %w", err))
			http.Error(res, "invalid target url", http.StatusBadRequest)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		proxy.ServeHTTP(res, req)
	}
}
