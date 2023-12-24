package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"net/http"
	"syscall"
	"time"
)

type Server struct {
	Host     string
	Port     string
	Name     string
	HitCount int
	InUsed   bool
	Active   bool
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

var ErrNoActiveServer = errors.New("no active server available")

func main() {
	targets := []Server{
		{
			Host:     "http://localhost",
			Port:     "9000",
			Name:     "Host 1",
			InUsed:   false,
			Active:   true,
			HitCount: 0,
		},
		{
			Host:     "http://localhost",
			Port:     "9001",
			Name:     "Host 2",
			InUsed:   false,
			Active:   true,
			HitCount: 0,
		},
		{
			Host:     "http://localhost",
			Port:     "9002",
			Name:     "Host 3",
			InUsed:   false,
			Active:   true,
			HitCount: 0,
		},
	}

	go func() {
		for range time.Tick(10 * time.Second) {
			handleHealthcheck(targets)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpClient := &http.Client{}
		req := extractDownstreamRequest(r)
		target, err := getTarget(targets)
		if err != nil {
			if errors.Is(err, ErrNoActiveServer) {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("no service available"))
				return
			}

			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Info("starting hit the target", "host", target.Host, "port", target.Port)
		url := fmt.Sprintf("%s:%s", target.Host, target.Port)
		httpRequest, err := http.NewRequest(req.Method, url, nil)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response, err := httpClient.Do(httpRequest)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		byteResponse, err := io.ReadAll(response.Body)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		slog.Info("hit the target completed", "host", target.Host, "port", target.Port)
		w.Write(byteResponse)
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}
}

func handleHealthcheck(servers []Server) {
	for index, server := range servers {
		httpClient := &http.Client{
			Timeout: 3 * time.Second,
		}

		slog.Info(
			"starting hit the target for healtcheck",
			"host",
			servers[index].Host,
			"port",
			servers[index].Port,
		)
		url := fmt.Sprintf("%s:%s/healtcheck", server.Host, server.Port)
		httpRequest, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			slog.Error("error occured", "error", err.Error())
			return
		}

		response, err := httpClient.Do(httpRequest)
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				servers[index].Active = false

				slog.Info(
					"healtcheck request completed",
					"host",
					servers[index].Host,
					"port",
					servers[index].Port,
					"active",
					servers[index].Active,
				)

				continue
			}

			servers[index].Active = false
			slog.Error("error occured", "error", err.Error())
			return
		}
		defer response.Body.Close()

		_, err = io.ReadAll(response.Body)
		if err != nil {
			slog.Error("error occured", "error", err.Error())
			return
		}

		if response.StatusCode != http.StatusOK {
			servers[index].Active = false
		} else {
			servers[index].Active = true
		}

		slog.Info(
			"healtcheck request completed",
			"host",
			servers[index].Host,
			"port",
			servers[index].Port,
			"active",
			servers[index].Active,
		)
	}
}

func getTarget(servers []Server) (Server, error) {
	if len(servers) == 1 && servers[0].Active {
		return servers[0], nil
	} else if len(servers) == 1 && !servers[0].Active {
		return Server{}, ErrNoActiveServer
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
		// If last registered server is not active, try to fallback to the first server.
		if !server.Active && index == len(servers)-1 {
			resetInUsedFlag(servers)
			// Mark initial registered server to inUsed
			servers[0].InUsed = true
			servers[0].HitCount++
			return servers[0], nil
		}

		// If reaches the end of registered servers
		if server.InUsed && index == len(servers)-1 {
			// clean up previous servers state
			resetInUsedFlag(servers)

			// Mark initial registered server to inUsed
			servers[0].InUsed = true
			servers[0].HitCount++
			return servers[0], nil
		}

		// Move to the next registered server if accessed indexed server is in use
		if server.InUsed {
			continue
		}

		usedIndex = index
		break
	}

	servers[usedIndex].HitCount++
	servers[usedIndex].InUsed = true
	return servers[usedIndex], nil
}

func resetInUsedFlag(servers []Server) {
	for index, server := range servers {
		servers[index].InUsed = false

		// Just for the sake of accessing server
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
