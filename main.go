package main

import (
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	Host string
	Port string
	Name string
}

type DownstreamRequest struct {
	Method      string
	Accept      string
	UserAgent   string
	ContentType string
	Proto       string
	Host        string
	Body        []byte
}

func main() {
	targets := []Server{
		{
			Host: "http://localhost",
			Port: "9000",
			Name: "Host 1",
		},
		{
			Host: "http://localhost",
			Port: "9001",
			Name: "Host 2",
		},
	}

	httpClient := &http.Client{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Just printing a dummy targets
		for _, target := range targets {
			fmt.Println(target.Host + ":" + target.Port)
		}

		req := extractDownstreamRequest(r)
		// TODO: use algorithm to pickup the downstream target.
		url := fmt.Sprintf("%s:%s", targets[0].Host, targets[0].Port)
		httpRequest, err := http.NewRequest(req.Method, url, nil)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		response, err := httpClient.Do(httpRequest)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer response.Body.Close()

		byteResponse, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		w.Write(byteResponse)
	})

	http.ListenAndServe(":3000", nil)
}

func extractDownstreamRequest(r *http.Request) DownstreamRequest {
	method := r.Method
	accept := r.Header.Get("Accept")
	userAgent := r.Header.Get("User-Agent")
	contentType := r.Header.Get("Content-Type")
	proto := r.Proto
	host := r.Host

	fmt.Printf("Method: %s\n", method)
	fmt.Printf("Accept: %s\n", accept)
	fmt.Printf("User-Agent: %s\n", userAgent)
	fmt.Printf("Content-Type: %s\n", contentType)
	fmt.Printf("Proto: %s\n", proto)
	fmt.Printf("Host: %s\n", host)

	return DownstreamRequest{
		Method:      method,
		Accept:      accept,
		UserAgent:   userAgent,
		ContentType: contentType,
		Proto:       proto,
		Host:        host,
	}
}
