package api

import (
	"context"
	"net/http"
	"time"

	"github.com/pedramkousari/abshar-toolbox-new/scripts/update"
)

func HandleFunc(server *Server) {
	server.HandleFunc("/ping", pingHandle)
	server.HandleFunc("/patch", patchHandle)
	server.HandleFunc("/state", stateHandle)

	server.HandleFunc("/stop", stopHandle(server))
}

func pingHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func patchHandle(w http.ResponseWriter, r *http.Request) {
	us := update.NewUpdateService()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	go func() {
		<-time.After(time.Second * 10)
		cancel()
	}()

	resChan := make(chan bool)
	go us.Handle(ctx, resChan)

	if res := <-resChan; res {
		w.Write([]byte("OK"))
	} else {
		w.Write([]byte("NO"))
	}

}
func stateHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func stopHandle(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		server.Stop()
	}
}
