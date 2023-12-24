package main

import (
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	Host   string
	Port   string
	Name   string
	InUsed bool
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
			Host:   "http://localhost",
			Port:   "9000",
			Name:   "Host 1",
			InUsed: false,
		},
		{
			Host:   "http://localhost",
			Port:   "9001",
			Name:   "Host 2",
			InUsed: false,
		},
		{
			Host:   "http://localhost",
			Port:   "9002",
			Name:   "Host 3",
			InUsed: false,
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpClient := &http.Client{}
		req := extractDownstreamRequest(r)
		target := getTarget(targets)
		fmt.Printf("[INFO] starting hit the target of %s:%s\n", target.Host, target.Port)
		url := fmt.Sprintf("%s:%s", target.Host, target.Port)
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

	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}
}

func getTarget(servers []Server) Server {
	if len(servers) == 1 {
		return servers[0]
	}

	// a, b, c
	// if all in clear state, mark a as inUsed, return b

	// a (used), b, c
	// if a in used, go to next
	// if b is unused, mark b as inUsed, return b

	// a (used), b (used), c
	// if a is used, go to next
	// if b is used, go to next
	// if c is unused, mark c as inUsed, check if c is the end of registered server, cleanup the previous servers

	var usedIndex int
	for index, server := range servers {
		// If reaches the end of registered servers
		if server.InUsed && index == len(servers)-1 {
			// clean up previous servers state
			cleanupServers(servers)

			// Mark initial registered server to inUsed
			servers[0].InUsed = true
			return servers[0]
		}

		// Move to the next registered server if accessed indexed server is in use
		if server.InUsed {
			continue
		}

		usedIndex = index
		break
	}

	fmt.Println(usedIndex)

	servers[usedIndex].InUsed = true
	return servers[usedIndex]
}

func cleanupServers(servers []Server) {
	for index, server := range servers {
		servers[index].InUsed = false
		servers[index].Host = server.Host
	}
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
