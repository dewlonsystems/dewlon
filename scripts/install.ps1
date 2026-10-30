# Installs the dewlon CLI on Windows.
#
#   irm https://raw.githubusercontent.com/dewlonsystems/dewlon/main/scripts/install.ps1 | iex
#
# Environment:
#   DEWLON_VERSION      version to install, e.g. v0.2.0 (default: latest)
#   DEWLON_INSTALL_DIR  install folder (default: %LOCALAPPDATA%\Programs\dewlon)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$repo = 'dewlonsystems/dewlon'
$bin = 'dewlon'

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$version = if ($env:DEWLON_VERSION) { $env:DEWLON_VERSION } else { 'latest' }
if ($version -eq 'latest') {
    $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest" `
        -Headers @{ 'User-Agent' = 'dewlon-installer' }
    $version = $release.tag_name
}
if (-not $version.StartsWith('v')) { $version = "v$version" }
$number = $version.TrimStart('v')

$archive = "${bin}_${number}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/$version"

$tmp = Join-Path ([IO.Path]::GetTempPath()) ("dewlon-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "Downloading $bin $version for windows/$arch..."
    Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $tmp $archive) -UseBasicParsing
    Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing

    $line = Get-Content (Join-Path $tmp 'checksums.txt') |
        Where-Object { $_ -match "\s$([regex]::Escape($archive))$" } |
        Select-Object -First 1
    if (-not $line) { throw "No checksum listed for $archive" }

    $expected = ($line -split '\s+')[0]
    $actual = (Get-FileHash (Join-Path $tmp $archive) -Algorithm SHA256).Hash
    if ($expected -ne $actual) {
        throw "Checksum mismatch for $archive (expected $expected, got $actual)"
    }

    Expand-Archive (Join-Path $tmp $archive) -DestinationPath $tmp -Force

    $dir = if ($env:DEWLON_INSTALL_DIR) { $env:DEWLON_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\dewlon' }
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    Copy-Item (Join-Path $tmp "$bin.exe") (Join-Path $dir "$bin.exe") -Force

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (($userPath -split ';') -notcontains $dir) {
        [Environment]::SetEnvironmentVariable('Path', "$userPath;$dir", 'User')
        Write-Host "Added $dir to your PATH. Open a new terminal to use it."
    }

    Write-Host "Installed $bin $version to $dir"
}
finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}