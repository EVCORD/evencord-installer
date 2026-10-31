$link = "https://github.com/EVCORD/evencord-installer/releases/latest/download/EvenCordInstallerCli.exe"

$outfile = "$env:TEMP\EvenCordInstallerCli.exe"

Write-Output "Downloading installer to $outfile"

Invoke-WebRequest -Uri "$link" -OutFile "$outfile"

Write-Output ""

Start-Process -Wait -NoNewWindow -FilePath "$outfile"

# Cleanup
Remove-Item -Force "$outfile"
