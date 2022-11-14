package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/bretkikehara/go-reverse-proxy/host-provider/etchosts"
	"github.com/bretkikehara/go-reverse-proxy/proxy"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERR: panic", r)
			debug.PrintStack()
		}
	}()
	// initialize a reverse proxy and pass the actual backend server url here
	host := "asdfexample.com"
	p, err := etchosts.New(host)
	if err != nil {
		panic(err)
	}
	defer func() {
		p.RemoveTLD(host)
	}()

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":80"
	}
	fmt.Println(">>> listening on ", addr)
	srv := &http.Server{
		Addr: addr,
		Handler: proxy.New(p, []proxy.TargetConfig{
			{
				Subdomain: "app",
				Port:      8000,
			},
		}),
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
