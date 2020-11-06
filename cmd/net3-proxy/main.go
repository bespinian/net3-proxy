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
	log.Println("Starting proxy")

	portStr := getEnv("NET3_HTTP_PROXY_PORT", "81")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid value for port: %w", err))
	}
	targetHost := getEnv("NET3_HTTP_PROXY_TARGET_HOST", "localhost")
	targetPortStr := getEnv("NET3_HTTP_PROXY_TARGET_PORT", "80")
	targetPort, err := strconv.Atoi(targetPortStr)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid value for target port: %w", err))
	}

	handleFunc, err := makeProxyHandleFunc(targetHost, targetPort)
	if err != nil {
		log.Fatal(fmt.Errorf("error making proxy handle func: %w", err))
	}
	http.HandleFunc("/", handleFunc)

	log.Printf("Listening on localhost:%v", port)
	log.Printf("Forwarding to %s:%v", targetHost, targetPort)

	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatal(fmt.Errorf("could not start proxy server: %w", err))
	}
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Environment variable %q not set. Falling back to default %q", key, fallback)
		return fallback
	}
	return value
}

func makeProxyHandleFunc(targetHost string, targetPort int) (func(http.ResponseWriter, *http.Request), error) {
	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%v", targetHost, targetPort))
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy target URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(res http.ResponseWriter, req *http.Request) {
		logLines := make([]string, 0)

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Print(fmt.Errorf("error reading request body: %w", err))
		}

		logLines = append(logLines, "")
		logLines = append(logLines, "-------------------")

		logLines = append(logLines, "")
		logLines = append(logLines, fmt.Sprintf("%s %s", req.Method, req.RequestURI))

		logLines = append(logLines, "")
		logLines = append(logLines, "Request Headers")
		for name, values := range req.Header {
			for _, v := range values {
				logLines = append(logLines, fmt.Sprintf("%s: %s", name, v))
			}
		}

		logLines = append(logLines, "")
		logLines = append(logLines, "Request Body")
		logLines = append(logLines, string(body))

		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		proxy.ModifyResponse = makeLogResponseFunc(logLines)
		proxy.ServeHTTP(res, req)
	}, nil
}

func makeLogResponseFunc(logLines []string) func(*http.Response) error {
	return func(resp *http.Response) error {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		logLines = append(logLines, "")
		logLines = append(logLines, "Response Headers")
		for name, values := range resp.Header {
			for _, v := range values {
				logLines = append(logLines, fmt.Sprintf("%s: %s", name, v))
			}
		}

		logLines = append(logLines, "")
		logLines = append(logLines, "Response Body")
		logLines = append(logLines, string(body))

		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))

		log.Println(strings.Join(logLines, "\n"))

		return nil
	}
}
