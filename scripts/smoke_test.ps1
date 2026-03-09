param(
  [string]$BaseUrl = "http://127.0.0.1:8080",
  [string]$Username = "admin",
  [string]$Password = "Admin@123456"
)

$ErrorActionPreference = "Stop"

Write-Host "==> Health check"
$health = Invoke-RestMethod -Method Get -Uri "$BaseUrl/health"
if ($health.status -ne "ok") {
  throw "health check failed: $($health | ConvertTo-Json -Compress)"
}
Write-Host "health ok"

Write-Host "==> Login"
$loginBody = @{
  username = $Username
  password = $Password
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/user/login" -ContentType "application/json" -Body $loginBody
if ($loginResp.code -ne 0 -or -not $loginResp.data.token) {
  throw "login failed: $($loginResp | ConvertTo-Json -Compress)"
}
$token = $loginResp.data.token
Write-Host "login ok"

Write-Host "==> Get current user"
$selfResp = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/v1/sys/user/getSelfInfo" -Headers @{ "x-token" = $token }
if ($selfResp.code -ne 0) {
  throw "getSelfInfo failed: $($selfResp | ConvertTo-Json -Compress)"
}
Write-Host "user ok: $($selfResp.data.username)"

Write-Host "==> Get menu"
$menuResp = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/v1/sys/menu/getMenu" -Headers @{ "x-token" = $token }
if ($menuResp.code -ne 0) {
  throw "getMenu failed: $($menuResp | ConvertTo-Json -Compress)"
}
Write-Host "menu ok, count=$($menuResp.data.Count)"

Write-Host "Smoke test passed."
