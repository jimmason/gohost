param (
    [string]$Version = "latest"
)

$repo = "jimmason/gohost"
$installPath = "$env:ProgramFiles\gohost"
$exePath = "$installPath\gohost.exe"

function Get-Platform {
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } elseif ($env:PROCESSOR_ARCHITECTURE -like "*ARM64*") { "arm64" } else {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
    return "windows_$arch"
}

function Get-Version {
    if ($Version -ne "latest") { return $Version }
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    return $latest.tag_name
}

function Install-Gohost {
    $platform = Get-Platform
    $version = Get-Version
    $filename = "gohost_${platform}.zip"
    $url = "https://github.com/$repo/releases/download/$version/$filename"

    Write-Host "Downloading $url..."
    $tmp = New-TemporaryFile
    Invoke-WebRequest -Uri $url -OutFile $tmp.FullName

    Write-Host "Extracting..."
    Expand-Archive -Path $tmp.FullName -DestinationPath $installPath -Force

    Write-Host "Installed to $installPath"

    # Add to PATH if needed
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$installPath*") {
        Write-Host "Adding $installPath to PATH"
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installPath", "User")
    }

    Write-Host "$exePath" --help
}

Install-Gohost
