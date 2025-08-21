package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"realtimechatserver/internal/cluster"
	"realtimechatserver/internal/server"
	"realtimechatserver/internal/transport"
)

func main() {
	h := server.NewHub()

	// Cluster mode if REDIS_URL present
	if os.Getenv("REDIS_URL") != "" {
		rb, err := cluster.NewRedisBroker()
		if err != nil {
			log.Fatal(err)
		}
		if err := h.EnableCluster(rb); err != nil {
			log.Fatal(err)
		}
		log.Println("Cluster Mode On")
	} else {
		log.Println("Standalone Mode On")
	}

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
