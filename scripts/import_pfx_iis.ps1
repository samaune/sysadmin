<#
.SYNOPSIS
    Automation script SSL renewal with arguments and -DomainName -Force switch.

.DESCRIPTION
    Import a PFX certificate, checks if it exists in LocalMachine\My, and does not expired today.
#>

param(
  [Parameter(Mandatory=$true, Position=0)]
  [string]$DomainName,
  
  # [Parameter(Mandatory=$false)]
  # [string]$Password = "",

  [Parameter(Mandatory=$false, HelpMessage="Force import even if certificate exists or not expired")]
  [switch]$Force
)

if(!$DomainName) {
  Write-Host "Automation script SSL renewal with arguments and -DomainName -Force switch"
  exit 0
}

$port=443;
$pfxPath = "uploads\$DomainName.pfx"

# $tcp=New-Object System.Net.Sockets.TcpClient; 
# $tcp.Connect($DomainName, $port);
# $ssl=New-Object System.Net.Security.SslStream($tcp.GetStream(), $false, ({$true}));
# $ssl.AuthenticateAsClient($DomainName);
# $resCert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2 $ssl.RemoteCertificate;
# $tcp.Close(); $ssl.Close();
# $remainingDay = ([math]::Ceiling(($resCert.NotAfter - (Get-Date)).TotalDays));
$remainingDay = 1

if (($remainingDay -eq 0) -or ($Force)) {
  ## Download new PFX
  # try {
  #   if (Test-Path $pfxPath) {
  #     $pfxNewPath = $pfxPath -replace "\.pfx$", "_$(Get-Date -Format "yyyyMMdd").pfx";
  #     Move-Item -Path $pfxPath -Destination  $pfxNewPath -Force
  #     Write-Host "Moved $pfxPath"
  #   }
  #   $PfxUrl="https://developer.singha.app/downloads/certs/pfx/$domainName.pfx"; 
  #   Invoke-WebRequest -Uri $PfxUrl -OutFile $pfxPath -UseBasicParsing
  #   Write-Host "Downloaded PFX to $pfxPath"    
  # } catch {
  #   Write-Error "Failed to download PFX: $_"
  #   exit 1
  # }

  $Password = ""
  # Read-Host "Enter password" -AsSecureString | ConvertFrom-SecureString
  $pfxPwd = ConvertTo-SecureString $Password -AsPlainText -Force
  # if (!$Password) {
  #   $pfxPwd = Read-Host -Prompt "Enter PFX password" -AsSecureString
  # }

  # Write-Host "Certificate expires today!";
  $cert = Get-PfxCertificate -FilePath $pfxPath -Password $pfxPwd
  $existingCert = Get-ChildItem Cert:\LocalMachine\My ` | Where-Object { $_.Thumbprint -eq ($cert.Thumbprint) } `
  
  if (!$existingCert) {
    $pfx= Import-PfxCertificate -FilePath $pfxPath -Password $pfxPwd -CertStoreLocation Cert:\LocalMachine\My
    Write-Host "Certificate imported with thumbprint:" $pfx.Thumbprint
    $thumbprint = $pfx.Thumbprint
  }else{
    Write-Host "Certificate existing with thumbprint:" $existingCert.Thumbprint
    $thumbprint = $existingCert.Thumbprint
  }
  
  # Binding - IIS Binding update cert to IIS (Requied enabled DNI)
  $binding = "${DomainName}:${port}"
  Write-Host "Updating binding $binding..."
  netsh http delete sslcert hostnameport=$binding
  netsh http add sslcert hostnameport=$binding certhash=$thumbprint certstorename=My appid='{98525227-7F70-4B89-908D-BE5F94026C65}'

  # Cleanup - Remove expired certificate
  Get-ChildItem Cert:\LocalMachine\My `
  | Where-Object { ($_.NotAfter -lt (Get-Date)) } `
  | Remove-Item
} else {
  Write-Host "Certificate does not expire today. Use -Force to re-import."
  exit 0
}
