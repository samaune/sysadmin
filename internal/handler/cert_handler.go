package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"ndk/internal/kube"
	"ndk/internal/service"
	"net/http"
)

func GetCerts(w http.ResponseWriter, r *http.Request) {
	// --- Serve file as download ---
	pfxData := kube.GetCert("example.com")
	w.Header().Set("Content-Type", "application/x-pkcs12") // MIME type for PFX
	w.Header().Set("Content-Disposition", `attachment; filename="tls.pfx"`)
	w.Header().Set("Content-Length", fmt.Sprint(len(pfxData)))

	var err error
	_, err = w.Write(pfxData)
	if err != nil {
		log.Println("Failed to write response:", err)
	}
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users := service.GetAllUsers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
