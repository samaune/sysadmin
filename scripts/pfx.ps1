param( 
  # [Parameter(Mandatory=$false)]
  # [string]$PfxPwd = "",

  [Parameter(Mandatory=$false, HelpMessage="Force import even if certificate exists or not expired")]
  [switch]$Force
)
$env = (Get-Content ".env" | Where-Object { $_ -match '=' }) -replace '"', '' -join "`n" | ConvertFrom-StringData
function Get-DayExpiredHttp($domain){
    $tcp = New-Object System.Net.Sockets.TcpClient; 
    $tcp.Connect($domain, 443);
    $ssl = New-Object System.Net.Security.SslStream($tcp.GetStream(), $false, ({$true}));
    $ssl.AuthenticateAsClient($domain);
    $resCert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2 $ssl.RemoteCertificate;
    $tcp.Close(); $ssl.Close();
    $remainingDay = ([math]::Ceiling(($resCert.NotAfter - (Get-Date)).TotalDays));
    Write-Host "[$domain] has expired in $remainingDay day(s)"
    return $remainingDay
}


function Get-DownloadPfx($pfxUrl, $pfxPath, $pfxPwd){
  try {
    $dir = Split-Path $pfxPath -Parent
    # Create directory if not exists
    if (-not (Test-Path $dir)) {
      New-Item -ItemType Directory -Path $dir | Out-Null
    }
    if (Test-Path $pfxPath) {
      $pfxNewPath = $pfxPath -replace "\.pfx$", "_$(Get-Date -Format "yyyyMMdd").pfx";
      Move-Item -Path $pfxPath -Destination  $pfxNewPath -Force
      Write-Host "Moved backup $pfxPath"
    }

    $encoded = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes($env.PFX_BASIC_AUTH))
    Invoke-WebRequest -Uri $pfxUrl `
    -Headers @{ Authorization = "Basic $encoded" } `
    -OutFile $pfxPath -UseBasicParsing -SkipCertificateCheck
    return Get-PfxCertificate -FilePath $pfxPath -Password (ConvertTo-SecureString $pfxPwd -AsPlainText -Force)
  } catch {
    Write-Host "Failed to download pfx: $_"
    return $null
    # exit 1
  }
}
# Return $False nothin
function Import-PfxCert($domain, $pfxPath, $pfxPwd) {
  try {
    $pfxUrl= $env.PFX_URL + "/$domain.pfx"; 
    $pfx = Get-DownloadPfx $pfxUrl $pfxPath $pfxPwd
    if(!$pfx){
      return $null
    }
    $cert = Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Thumbprint -eq ($pfx.Thumbprint) } 
    if (!$cert) {
      $cert = Import-PfxCertificate -FilePath $pfxPath `
      -Password (ConvertTo-SecureString $pfxPwd -AsPlainText -Force) `
      -CertStoreLocation Cert:\LocalMachine\My

      Write-Host "Certificate imported with thumbprint:" $newCert.Thumbprint
    }else{
      Write-Host "Certificate existing with thumbprint:" $cert.Thumbprint
    }

    return $cert.Thumbprint
  } catch {
    return $null
  }
}


function Get-AppIdByBinding($binding) {
    $items = netsh http show sslcert
    $lines = $items -split "`n"
    for ($i = 0; $i -lt $lines.Length; $i++) {
      if ($lines[$i] -match "Hostname:port\s*:\s*$binding") {
        $webBinding = [PSCustomObject]@{
          AppId = ""
          CertHash = ""
        }
        for ($j = $i; $j -lt $i+10; $j++) {
          # Capture Cert Hash
          if ($lines[$j] -match "Certificate Hash\s*:\s*(.+)") {
              $webBinding.CertHash = $matches[1].Trim()
          }

          # Capture AppId (always in {GUID} format)
          if ($lines[$j] -match "Application ID\s*:\s*(\{[0-9a-fA-F\-]+\})") {
              $webBinding.AppId = $matches[1].Trim()
              return $webBinding
          }
        }
      }
    }
    return $null
}

function Add-NewWebBinding ([string]$site, [string]$domain, [Int32]$port) {
  # Import-Module WebAdministration
  New-WebBinding -Name $site -Protocol "https" -Port $port -HostHeader $domain
  Set-WebBinding -Name $site -BindingInformation "*:${port}:${domain}" -PropertyName "sslFlags" -Value 1
}

function Initialize {
  $port=443;
  
  $domain = $env.DOMAIN
  $site = $env.IIS_SITE
  $pfxPwd = $env.PFX_PWD
  $pfxPath = $env.PFX_TMP + "\$domain.pfx"
  $binding = "${domain}:${port}"

  if (((Get-DayExpiredHttp $domain) -le 0) -or ($Force)) {
    $thumbprint = Import-PfxCert $domain $pfxPath $pfxPwd

    if ($thumbprint) {
      $webBinding = Get-AppIdByBinding $binding
      if($webBinding){
        $appId = $webBinding.AppId
        if($webBinding.CertHash -ne $thumbprint){
          netsh http update sslcert hostnameport=$binding certhash=$thumbprint appid=$appId certstorename=My
          Write-Host "Updated binding [$binding], $thumbprint, $appId"
        }
        Write-Host "No update found"
      }else{
        $appId = "{$([guid]::NewGuid())}"
        Add-NewWebBinding $site $domain $port
        Write-Host "Add IIS web binding  [$binding], $thumbprint, $appId"
        # netsh http delete sslcert hostnameport=$binding
        netsh http add sslcert hostnameport=$binding certhash=$thumbprint appid=$appId certstorename=My
        Write-Host "Add certificate and assign {$appId} to [$binding], certhash:$thumbprint"
      }
    }else{
      Write-Host "No new pfx certificate imported and updated"
    }

    # Cleanup - Remove expired certificate
    Get-ChildItem Cert:\LocalMachine\My | Where-Object { ($_.NotAfter -lt (Get-Date)) } | Remove-Item
  } else {
    Write-Host "Certificate does not expire today. Use -Force to re-import."
  }
}

Initialize