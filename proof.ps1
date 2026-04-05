$ErrorActionPreference = "Stop"

$slot="55555555-5555-5555-5555-555555555555"
$user="12345678-1234-1234-1234-123456789012"
$body="{\`"user_id\`":\`"$user\`",\`"amount\`":500}"
$idempotencyKey1 = [guid]::NewGuid().ToString()
$idempotencyKey2 = [guid]::NewGuid().ToString()

Write-Host "==========================="
Write-Host "API DOUBLE PAYMENT TEST"
Write-Host "==========================="

Write-Host "`n1. Joining slot..." -ForegroundColor Yellow
$joinResp = curl.exe -s -X POST -H "Content-Type: application/json" -d "{\`"user_id\`":\`"$user\`"}" http://localhost:8080/slots/$slot/join
Write-Host "Response: $joinResp" -ForegroundColor Green

Write-Host "`n2. First Payment Attempt (Key: $idempotencyKey1)..." -ForegroundColor Yellow
$pay1Resp = curl.exe -s -X POST -H "X-Idempotency-Key: $idempotencyKey1" -H "Content-Type: application/json" -d $body http://localhost:8080/slots/$slot/pay
Write-Host "Response: $pay1Resp" -ForegroundColor Green

Write-Host "`n3. Simulating Network Retry (Same Key)..." -ForegroundColor Yellow
$pay2Resp = curl.exe -s -X POST -H "X-Idempotency-Key: $idempotencyKey1" -H "Content-Type: application/json" -d $body http://localhost:8080/slots/$slot/pay
Write-Host "Response: $pay2Resp" -ForegroundColor Green

Write-Host "`n4. Second Payment Attempt / Double Charge (New Key: $idempotencyKey2)..." -ForegroundColor Yellow
$pay3Resp = curl.exe -s -X POST -H "X-Idempotency-Key: $idempotencyKey2" -H "Content-Type: application/json" -d $body http://localhost:8080/slots/$slot/pay
Write-Host "Response: $pay3Resp" -ForegroundColor Red
