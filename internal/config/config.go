package config

import (
	"log"
	"os"
	
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	DnsName    string
	PfxPassword string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (using system environment)")
	}
	
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dnsName := os.Getenv("DNS_NAME")
	pfxPwd := os.Getenv("PFX_PASSWORD")
	if dnsName == "" {
		log.Fatal("DNS_NAME not set")
	}

	// godotenv.Load(".env.local")
	

	return &Config{
		ServerPort: 	port,
		DnsName:      	dnsName,
		PfxPassword:	pfxPwd,
	}
}
