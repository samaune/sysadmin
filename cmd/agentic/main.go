package main

import (
	// "crypto/tls"
	// "crypto/x509"
	// "io/ioutil"
	// "net/http"

	// "software.sslmate.com/src/go-pkcs12"
	"ndk/internal/config"
	"ndk/internal/exec"
	"bytes"
	"fmt"
	"strings"
)

func main() {
	cfg := config.Load()
	hostname := cfg.DnsName
	pfxPath := fmt.Sprintf("uploads\\%s.pfx", hostname)
	pfxPwd := cfg.PfxPassword
	port := 443

	// Import the PFX certificate into the "My" store under LocalMachine
	thumb := ImportPfxCertificate(pfxPath, pfxPwd)

	// Bind the cert to IIS using netsh
	bindCmd := fmt.Sprintf(`netsh http add sslcert hostnameport=%s:%d certhash=%s appid="{00112233-4455-6677-8899-AABBCCDDEEFF}" certstorename=MY`, hostname, port, thumb)
	out, err := exec.RunCmd("cmd", "/C", bindCmd)
	if err != nil {
		panic(fmt.Errorf("failed to bind cert to IIS: %v\nOutput: %s", err, out))
	}
	fmt.Println("Certificate bound to IIS successfully.")
}

func ImportPfxCertificate(pfxPath string, pfxPwd string) string {
	thumb := GetThumbprint(pfxPath, pfxPwd)
	if IsExistsPfxCertificate(thumb) {
		fmt.Println("Certificate %s already exists.", thumb)
		psCmd := fmt.Sprintf(`Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Thumbprint -eq "%s" } | Remove-Item;`, thumb)
		out, err := exec.RunPwsh(psCmd)
		if err != nil {
			panic(fmt.Errorf("failed to delete PFX: %v\nOutput: %s", err, out))
		}
	} else {
		// Import the PFX certificate into the "My" store under LocalMachine
		psCmd := fmt.Sprintf(`
		$pfxPwd = ConvertTo-SecureString "%s" -AsPlainText -Force;
		$pfx = Import-PfxCertificate -FilePath "%s" -CertStoreLocation Cert:\LocalMachine\My -Password $pfxPwd;
		$pfx.Thumbprint
		`, pfxPwd, pfxPath)

		out, err := exec.RunPwsh(psCmd)
		if err != nil {
			panic(fmt.Errorf("failed to import PFX: %v\nOutput: %s", err, out))
		}
		fmt.Println("Certificate imported successfully.")
		out = bytes.TrimSpace(out)
		thumb = string(out)
	}
	fmt.Println("%s", thumb)
	return thumb
}

func IsExistsPfxCertificate(thumb string) bool {
	psCmd := fmt.Sprintf(`
	$cert = Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Thumbprint -eq "%s" };
	if ($cert) { Write-Output "FOUND" } else { Write-Output "NOTFOUND" }
	`, strings.ToUpper(thumb))
	out, err := exec.RunPwsh(psCmd)
	if err != nil {
		panic(fmt.Errorf("failed to get thumbprint: %v\nOutput: %s", err, out))
	}
	result := strings.TrimSpace(string(out))
	return strings.EqualFold(result, "FOUND")
}

func GetThumbprint(pfxPath string, pfxPwd string) string {
	getThumbCmd := fmt.Sprintf(`
	$pfxPwd = ConvertTo-SecureString "%s" -AsPlainText -Force;
	(Get-PfxCertificate "%s" -Password $pfxPwd).Thumbprint;
	`, pfxPwd, pfxPath)
	thumb, err := exec.RunPwsh(getThumbCmd)
	if err != nil {
		panic(fmt.Errorf("failed to get thumbprint: %v\nOutput: %s", err, thumb))
	}
	thumb = bytes.TrimSpace(thumb)
	return string(thumb)
}