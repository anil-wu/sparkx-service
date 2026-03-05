param(
	[string]$EnvFile = (Join-Path $PSScriptRoot "..\\deploy\\local\\.env"),
	[string]$ConfigFile = (Join-Path $PSScriptRoot "etc\\sparkx-api-dev.yaml"),
	[string]$HostIp = "127.0.0.1"
)

function Load-DotEnv([string]$Path) {
	if (-not (Test-Path -LiteralPath $Path)) {
		throw "env file not found: $Path"
	}

	foreach ($raw in (Get-Content -LiteralPath $Path)) {
		$line = $raw.Trim()
		if (-not $line) { continue }
		if ($line.StartsWith("#")) { continue }

		$idx = $line.IndexOf("=")
		if ($idx -lt 1) { continue }

		$name = $line.Substring(0, $idx).Trim()
		$value = $line.Substring($idx + 1).Trim()
		if ($value.Length -ge 2) {
			if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
				$value = $value.Substring(1, $value.Length - 2)
			}
		}

		[System.Environment]::SetEnvironmentVariable($name, $value, "Process")
	}
}

Load-DotEnv -Path $EnvFile

if (-not $env:MYSQL_DSN) {
	$mysqlUser = if ($env:MYSQL_USER) { $env:MYSQL_USER } else { "sparkplay" }
	$mysqlPassword = if ($env:MYSQL_PASSWORD) { $env:MYSQL_PASSWORD } else { "sparkplay" }
	$mysqlDatabase = if ($env:MYSQL_DATABASE) { $env:MYSQL_DATABASE } else { "sparkplay" }
	$mysqlPort = if ($env:MYSQL_PORT) { $env:MYSQL_PORT } else { "6003" }
	$env:MYSQL_DSN = "${mysqlUser}:${mysqlPassword}@tcp(${HostIp}:${mysqlPort})/${mysqlDatabase}?charset=utf8mb4&parseTime=true&loc=Local"
}

if (-not $env:AUTH_ACCESS_SECRET) { $env:AUTH_ACCESS_SECRET = "dev-user-secret" }
if (-not $env:ADMIN_AUTH_ACCESS_SECRET) { $env:ADMIN_AUTH_ACCESS_SECRET = "dev-admin-secret" }
if (-not $env:GOOGLE_CLIENT_ID) { $env:GOOGLE_CLIENT_ID = "" }

if (-not $env:STORAGE_PROVIDER) { $env:STORAGE_PROVIDER = "s3" }

if (-not $env:S3_ENDPOINT) {
	$minioPort = if ($env:MINIO_PORT) { $env:MINIO_PORT } else { "9000" }
	$env:S3_ENDPOINT = "$HostIp`:$minioPort"
}
if (-not $env:S3_ACCESS_KEY_ID) {
	$env:S3_ACCESS_KEY_ID = if ($env:MINIO_ROOT_USER) { $env:MINIO_ROOT_USER } else { "minioadmin" }
}
if (-not $env:S3_ACCESS_KEY_SECRET) {
	$env:S3_ACCESS_KEY_SECRET = if ($env:MINIO_ROOT_PASSWORD) { $env:MINIO_ROOT_PASSWORD } else { "minioadmin" }
}
if (-not $env:S3_BUCKET) { $env:S3_BUCKET = "sparkplay" }
if (-not $env:S3_REGION) { $env:S3_REGION = "us-east-1" }
if (-not $env:S3_USE_SSL) { $env:S3_USE_SSL = "false" }

if (-not $env:OSS_ENDPOINT) { $env:OSS_ENDPOINT = "" }
if (-not $env:OSS_ACCESS_KEY_ID) { $env:OSS_ACCESS_KEY_ID = "" }
if (-not $env:OSS_ACCESS_KEY_SECRET) { $env:OSS_ACCESS_KEY_SECRET = "" }
if (-not $env:OSS_BUCKET) { $env:OSS_BUCKET = "" }

if ($MyInvocation.InvocationName -ne ".") {
	go run sparkx.go -f $ConfigFile
}
