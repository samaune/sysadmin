# $pfxPwd = ConvertTo-SecureString $Password -AsPlainText -Force
  # if (!$Password) {
  #   $pfxPwd = Read-Host -Prompt "Enter PFX password" -AsSecureString | ConvertFrom-SecureString
  #   Read-Host "Enter password" -AsSecureString | ConvertFrom-SecureString
  # }