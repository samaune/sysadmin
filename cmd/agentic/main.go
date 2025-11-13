package main

import (
	"bytes"
	"fmt"
	"ndk/internal/config"
	"ndk/internal/exec"
	"regexp"
	"strings"
)

func main() {
	cfg := config.Load()

	port := 443
	hostname := cfg.DnsName
	binding := fmt.Sprintf("%s:%d", hostname, port)

	// Import the PFX certificate into the "My" store under LocalMachine
	// pfxPath := fmt.Sprintf("%s\\uploads\\%s.pfx", cfg.PfxPath, hostname)
	// pfxPwd := cfg.PfxPwd
	// thumb := ImportPfxCertificate(pfxPath, pfxPwd)

	// thumb := "A8DAFB74120C2A9E790B0E8366A89854A1ADDB84"
	// BindingIIS(binding, thumb)
	out, _ := GetIISBindingAppId(binding)
	fmt.Println(out)
}

func GetIISBindingAppId(binding string) (string, bool) {
	psCmd := fmt.Sprintf(` netsh http show sslcert | Select-String "%s" -Context 0,8`, binding)
	out, err := exec.RunPwsh(psCmd)
	if err != nil {
		panic(fmt.Errorf("failed to get binding: %v\nOutput: %s", err, out))
	}
	result := strings.TrimSpace(string(out))
	//re := regexp.MustCompile(`Application ID\s*:\s*({[0-9a-fA-F\-]+})`)
	re := regexp.MustCompile(`Application ID\s*:\s*{([0-9a-fA-F\-]+)}`)
	match := re.FindStringSubmatch(result)

	if len(match) > 1 {
		//fmt.Println("Application ID:", match[1])
		return match[1], true
	} else {
		//fmt.Println("Application ID not found")
		return "", false
	}
}

func BindingIIS(binding string, thumb string) {
	appId := "4dc3e181-e14b-4a21-b022-59fc669b0914"

	//psCmd := fmt.Sprintf(`netsh http add sslcert hostnameport=%s certhash=%s appid={%s} certstorename=MY`, binding, thumb, appId)
	psCmd := fmt.Sprintf(`netsh http update sslcert hostnameport=%s certhash=%s appid={%s} certstorename=MY`, binding, thumb, appId)
	// if IsExistsPfxCertificate(appId) {
	// }
	fmt.Println(psCmd)
	out, err := exec.RunCmd(psCmd)
	if err != nil {
		panic(fmt.Errorf("failed to bind cert to IIS: %v\nOutput: %s", err, out))
	}
	out = bytes.TrimSpace(out)
	fmt.Println("Certificate bound to IIS successfully.")
	result := strings.TrimSpace(string(out))
	fmt.Printf("Certificate IIS: %v", result)
}

func ImportPfxCertificate(pfxPath string, pfxPwd string) string {
	thumb := GetThumbprint(pfxPath, pfxPwd)
	if IsExistsPfxCertificate(thumb) {
		fmt.Printf("Certificate %s already exists.", thumb)
		psCmd := fmt.Sprintf(`Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Thumbprint -eq "%s" } | Remove-Item;`, thumb)
		out, err := exec.RunPwsh(psCmd)
		if err != nil {
			panic(fmt.Errorf("failed to delete PFX: %v\nOutput: %s", err, out))
		}
	}

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

	fmt.Printf("%s", thumb)
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
