package main

// "encoding/json"
// "log"
// "net/http"
// "os"
// "os/exec"
// "path/filepath"

// "fmt"
// "ndk/internal/certs"
//import "ndk/internal/certs"
import (
	//"ndk/internal/kube"
	"log"
	"ndk/internal/config"
	"ndk/internal/server"
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
