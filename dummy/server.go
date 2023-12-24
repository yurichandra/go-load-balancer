package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("empty port arguments, use --port=${PORT}")
		return
	}

	portArgs := args[1]
	if !strings.Contains(portArgs, "--port") || !strings.Contains(portArgs, "=") {
		fmt.Println("wrong port arguments, use --port=${PORT}")
		return
	}

	port := strings.Split(portArgs, "=")
	if len(port) != 2 {
		fmt.Println("wrong port arguments, use --port=${PORT}")
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from downstream server"))
	})

	fmt.Printf("start to listen on port %s\n", port[1])
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port[1]), nil); err != nil {
		fmt.Println(err.Error())
	}
}
