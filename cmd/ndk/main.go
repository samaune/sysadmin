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
	//"ndk/internal/certs"
	"ndk/internal/handlers"
)

func main() {
	//certs.GetCert("commonName")
	handlers.Watcher()
}
