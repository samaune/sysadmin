package main

import (
	"log"
	"ndk/internal/config"
	"ndk/internal/server"
	//"ndk/internal/kube"
)

func main() {
	cfg := config.Load()
	srv := server.New(cfg)

	log.Println("🚀 Starting server on", cfg.ServerPort)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	//kube.GetCert("dagu.nagaworld.com")
	//handlers.Watcher()
}
