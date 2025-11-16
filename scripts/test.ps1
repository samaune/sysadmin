
$port=443;
$DomainName = "forms..com"
$binding = "${DomainName}:${port}"
$pfxPath = "uploads\$DomainName.pfx"

function Get-AppIdByBinding($binding) {
    $items = netsh http show sslcert
    $lines = $items -split "`n"
    for ($i = 0; $i -lt $lines.Length; $i++) {
        if ($lines[$i] -match "Hostname:port\s*:\s*$binding") {
            for ($j = $i; $j -lt $i+10; $j++) {
                if ($lines[$j] -match "Application ID\s*:\s*(.+)") {
                    return $matches[1]
                }
            }
        }
    }
}

$appId = Get-AppIdByBinding $binding
$thumbprint = ""
netsh http update sslcert hostnameport=$binding certhash=$thumbprint certstorename=My appid=$appId