package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"software.sslmate.com/src/go-pkcs12"
)

func ConvertToPfx(crtData []byte, keyData []byte) []byte {
	certBlock, _ := pem.Decode([]byte(crtData))
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		log.Fatal("failed to decode PEM certificate")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	keyBlock, _ := pem.Decode([]byte(keyData))
	if keyBlock == nil {
		log.Fatal("failed to decode PEM key")
	}

	var privKey interface{}
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		log.Fatalf("unsupported key type %q", keyBlock.Type)
	}
	if err != nil {
		log.Fatal(err)
	}

	pfxData, err := pkcs12.Modern.Encode(privKey, cert, nil, "mypassword")

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PFX file created: tls.pfx")

	return pfxData
}
