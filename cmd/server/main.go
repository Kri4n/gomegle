package main

import (
	"fmt"
	"log"
	"net/http"
	"realtimechatserver/internal/server"
	"realtimechatserver/internal/transport"
)

func main() {
	h := server.NewHub()
	go h.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		transport.ServeWS(h, w, r)
	})

	// Tiny health check
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/client.html")
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		s := h.Snapshot()
		out := fmt.Sprintf(
			"connected %d\nwaiting %d\npaired %d\n",
			s.Connected, s.Waiting, s.Paired,
		)
		w.Write([]byte(out))
	})

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":5040"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}