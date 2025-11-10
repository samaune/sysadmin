param(
    [Parameter(Mandatory=$false, Position=0)]
    [string]$Path="C:\inetpub\wwwroot"
)

Get-ChildItem $Path -Force |
ForEach-Object {
    if ($_.PSIsContainer) {
        $size = (Get-ChildItem $_.FullName -Recurse -Force -ErrorAction SilentlyContinue |
                 Measure-Object -Property Length -Sum).Sum
    } else {
        $size = $_.Length
    }
    [PSCustomObject]@{
        Name = $_.FullName
        SizeGB = [math]::Round($size / 1GB, 2)
    }
} | Sort-Object SizeGB -Descending | Format-Table -AutoSize
