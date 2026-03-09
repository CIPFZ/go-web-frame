param(
  [string]$BaseUrl = "http://127.0.0.1:8080",
  [string]$AdminUsername = "admin",
  [string]$AdminPassword = "Admin@123456"
)

$ErrorActionPreference = "Stop"

function Assert-CodeZero($resp, $name) {
  if ($resp.code -ne 0) {
    throw "$name failed: $($resp | ConvertTo-Json -Compress)"
  }
}

Write-Host "==> Health check"
$health = Invoke-RestMethod -Method Get -Uri "$BaseUrl/health"
if ($health.status -ne "ok") {
  throw "health check failed: $($health | ConvertTo-Json -Compress)"
}
Write-Host "health ok"

Write-Host "==> Admin login"
$adminLoginBody = @{ username = $AdminUsername; password = $AdminPassword } | ConvertTo-Json
$adminLogin = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/user/login" -ContentType "application/json" -Body $adminLoginBody
Assert-CodeZero $adminLogin "admin login"
$adminToken = $adminLogin.data.token
if (-not $adminToken) { throw "admin login failed: missing token" }
$adminHeaders = @{ "x-token" = $adminToken }
Write-Host "admin login ok"

Write-Host "==> Base APIs"
$selfResp = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/v1/sys/user/getSelfInfo" -Headers $adminHeaders
Assert-CodeZero $selfResp "getSelfInfo"
$menuResp = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/v1/sys/menu/getMenu" -Headers $adminHeaders
Assert-CodeZero $menuResp "getMenu"
$stateResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/sys/system/getServerInfo" -Headers $adminHeaders -ContentType "application/json" -Body "{}"
Assert-CodeZero $stateResp "getServerInfo"
Write-Host "base api ok"

Write-Host "==> Notice flow"
$stamp = Get-Date -Format "yyyyMMddHHmmss"
$testUser = "smoke_user_$stamp"
$testPass = "Smoke@123456"

$addUserBody = @{
  username = $testUser
  password = $testPass
  nickName = "Smoke User"
  authorityIds = @(888)
  status = 1
} | ConvertTo-Json
$addUserResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/sys/user/addUser" -Headers $adminHeaders -ContentType "application/json" -Body $addUserBody
Assert-CodeZero $addUserResp "addUser"

$userListBody = @{ page = 1; pageSize = 20; username = $testUser } | ConvertTo-Json
$userListResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/sys/user/getUserList" -Headers $adminHeaders -ContentType "application/json" -Body $userListBody
Assert-CodeZero $userListResp "getUserList"
$userId = ($userListResp.data.list | Select-Object -First 1).ID
if (-not $userId) { throw "getUserList failed: user not found" }

$noticeTitle = "Smoke notice $stamp"
$createNoticeBody = @{
  title = $noticeTitle
  content = "Smoke test directed notice"
  level = "info"
  targetType = "users"
  targetIds = @($userId)
  isPopup = $false
  needConfirm = $true
} | ConvertTo-Json
$createNoticeResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/sys/notice/createNotice" -Headers $adminHeaders -ContentType "application/json" -Body $createNoticeBody
Assert-CodeZero $createNoticeResp "createNotice"

$userLoginBody = @{ username = $testUser; password = $testPass } | ConvertTo-Json
$userLoginResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/user/login" -ContentType "application/json" -Body $userLoginBody
Assert-CodeZero $userLoginResp "test user login"
$userToken = $userLoginResp.data.token
if (-not $userToken) { throw "test user login failed: missing token" }
$userHeaders = @{ "x-token" = $userToken }

$myNotices = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/v1/sys/notice/getMyNotices?page=1&pageSize=10" -Headers $userHeaders
Assert-CodeZero $myNotices "getMyNotices"
$notice = $myNotices.data.list | Where-Object { $_.title -eq $noticeTitle } | Select-Object -First 1
if (-not $notice) { throw "getMyNotices failed: created notice not found" }

$markReadBody = @{ noticeId = $notice.ID } | ConvertTo-Json
$markReadResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/v1/sys/notice/markRead" -Headers $userHeaders -ContentType "application/json" -Body $markReadBody
Assert-CodeZero $markReadResp "markRead"

Write-Host "notice flow ok"
Write-Host "All checks passed."
