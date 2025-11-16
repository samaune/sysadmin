
Import-Module WebAdministration
function Get-DotEnv() {
    $envFile = Get-Content ".env" | Where-Object { $_ -match '=' }
    $envVars = $envFile -replace '"', '' -join "`n" | ConvertFrom-StringData
    return $envVars
}

$env = Get-DotEnv
$port=443;
$dnsName = $env.DNS_NAME
$pfxPwd = $env.PFX_PWD
$pfxPath = $env.PFX_TMP + "\$dnsName.pfx"
$binding = "${dnsName}:${port}"

function Add-NewIISBinding(){
  $site = "forms"
  $domain = "forms.example.com"
  $port = 443

  # $cert = Get-ChildItem -Path Cert:\LocalMachine\My | Where-Object { $_.Subject -like "*$domain*" }
  # $thumb = $cert.Thumbprint

  # Create binding
  New-WebBinding -Name "forms" -Protocol "https" -Port 443 -HostHeader "forms.example.com"
  Set-WebBinding -Name "forms" -BindingInformation "*:443:forms.example.com" -PropertyName "sslFlags" -Value 1

  # Push-Location IIS:\SslBindings
  # Get-Item "0.0.0.0!$port!$domain" | Remove-Item -ErrorAction SilentlyContinue
  # New-Item "0.0.0.0!$port!$domain" -Thumbprint $thumb -SSLFlags 1
  # Pop-Location
}

Add-NewIISBinding

Get-ChildItem IIS:\Sites | Where-Object { $_.Bindings.Collection | Where-Object { $_.HostHeader -eq "forms.example.com"} }