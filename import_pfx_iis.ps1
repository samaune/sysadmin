<#
.SYNOPSIS
    Automation script SSL renewal with arguments and -DomainName -Force switch.

.DESCRIPTION
    Import a PFX certificate, checks if it exists in LocalMachine\My, and does not expired today.
#>

param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$DomainName,
  
  [Parameter(Mandatory=$false)]
    [string]$Password,

    [Parameter(Mandatory=$false, HelpMessage="Force import even if certificate exists or not expired")]
    [switch]$Force
)

if(!$DomainName) {
  Write-Host "Automation script SSL renewal with arguments and -DomainName -Force switch"
  exit 0
}

$port=443;
$pfxPath = "uploads\$DomainName.pfx"
$PfxUrl="https://developer.singha.app/downloads/certs/pfx/$domainName.pfx"; 
# Read-Host "Enter password" -AsSecureString | ConvertFrom-SecureString
# "secret" | ConvertTo-SecureString
$pfxPwd = ConvertTo-SecureString $Password -AsPlainText -Force
if (!$Password) {
  $pfxPwd = Read-Host -Prompt "Enter PFX password" -AsSecureString
}

$tcp=New-Object System.Net.Sockets.TcpClient; 
$tcp.Connect($DomainName,$port);
$ssl=New-Object System.Net.Security.SslStream($tcp.GetStream(),$false,({$true}));
$ssl.AuthenticateAsClient($DomainName);
$resCert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2 $ssl.RemoteCertificate;
$tcp.Close(); $ssl.Close();

$certExpiryDate = $resCert.NotAfter; 
$remainingDay = ([math]::Ceiling(($resCert.NotAfter - (Get-Date)).TotalDays));

if (($remainingDay -eq 0) -or ($Force)) {
  # Download new PFX
  try {
    if (Test-Path $pfxPath) {
      #Remove-Item $pfxPath -Force
      $DateStr = Get-Date -Format "yyyyMMdd"
      $pfxNewPath = $pfxPath -replace "\.pfx$", "_$(Get-Date -Format "yyyyMMdd").pfx";
      Move-Item -Path $pfxPath -Destination  $pfxNewPath -Force
      Write-Host "Moved $pfxPath"
    }
    Invoke-WebRequest -Uri $PfxUrl -OutFile $pfxPath -UseBasicParsing
    Write-Host "Downloaded PFX to $pfxPath"    
  } catch {
    Write-Error "Failed to download PFX: $_"
    exit 1
  }

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
  netsh http add sslcert hostnameport=$binding certhash=$thumbprint certstorename=My appid='{00112233-4455-6677-8899-AABBCCDDEEFF}'

  # Cleanup - Remove expired certificate
  Get-ChildItem Cert:\LocalMachine\My `
  | Where-Object { ($_.NotAfter -lt (Get-Date)) } `
  | Remove-Item
} else {
  Write-Host "Certificate does not expire today. Use -Force to re-import."
  exit 0
}
