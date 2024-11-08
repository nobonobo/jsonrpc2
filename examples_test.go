package jsonrpc2

import (
	"context"
	"net/http"

	"github.com/nobonobo/jsonrpc2"
)

type Version struct {
	Tag string
}

func ExampleServer() {
	server := jsonrpc2.NewServer()
	err := server.HandleFunc("version", func(ctx context.Context) (Version, error) {
		return Version{"1.0.0"}, nil
	})
	if err != nil {
		panic(err)
	}

	http.Handle("/api", server)
	http.ListenAndServe(":4545", nil)
}
