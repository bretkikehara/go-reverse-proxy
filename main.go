package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/bretkikehara/go-reverse-proxy/host-provider/etchosts"
	"github.com/bretkikehara/go-reverse-proxy/proxy"
)

func main() {
	// initialize a reverse proxy and pass the actual backend server url here
	p, err := etchosts.New("asdfexample.com")
	if err != nil {
		panic(err)
	}
	pxy := proxy.New(p)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":80"
	}
	fmt.Println(">>> listening on ", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: pxy,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
