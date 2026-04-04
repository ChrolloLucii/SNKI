[CmdletBinding()]
param([string]$ApiUrl = "http://localhost:8080")

$slotId = "55555555-5555-5555-5555-555555555555"
$userId = "12345678-1234-1234-1234-123456789012" 
$idempotencyKey = [guid]::NewGuid().ToString()

Write-Host "
=== Test: Double Payment Protection ===
" -ForegroundColor Cyan

# 1. Join slot first (if not already joined)
Write-Host "1. Ensure user is in slot..." -ForegroundColor Yellow
$joinBody = "{`"user_id`":`"$userId`"}"
Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId/join" -Method POST -ContentType "application/json" -Headers @{"X-Demo-Token"=$userId} -Body $joinBody -UseBasicParsing -ErrorAction SilentlyContinue | Out-Null
Write-Host "   User joined (or was already joined)." -ForegroundColor Green

# 2. First Payment Attempt (Should Succeed)
Write-Host "
2. First Payment Attempt (Should Succeed)..." -ForegroundColor Yellow
$payBody = "{
    `"user_id`":`"$userId`",
    `"amount`": 500
}"

$successCount = 0
$failCount = 0
$duplicateCount = 0

try {
    $response = Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId/pay" -Method POST -Headers @{ "X-Idempotency-Key" = $idempotencyKey } -ContentType "application/json" -Body $payBody -UseBasicParsing -ErrorAction Stop
    $data = $response.Content | ConvertFrom-Json
    Write-Host "   [OK] Initial payment: $($data.message)" -ForegroundColor Green
    $successCount++
} catch {
    $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
    $errorData = $reader.ReadToEnd() | ConvertFrom-Json
    Write-Host "   [FAIL] Failed: $($errorData.error) - $($errorData.code)" -ForegroundColor Red
}

# 3. Second Payment Attempt with SAME Idempotency Key (Should return cached success)
Write-Host "
3. Second Payment Attempt with SAME Idempotency Key (Should return idempotency hit)..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId/pay" -Method POST -Headers @{ "X-Idempotency-Key" = $idempotencyKey } -ContentType "application/json" -Body $payBody -UseBasicParsing -ErrorAction Stop
    $data = $response.Content | ConvertFrom-Json
    if ($data.message -match "idempotent") {
        Write-Host "   [OK] Idempotency worked: $($data.message)" -ForegroundColor Green
        $successCount++
    } else {
        Write-Host "   [FAIL] Paid again instead of using cache!" -ForegroundColor Red
    }
} catch {
    Write-Host "   [FAIL] Failed unexpectedly on idempotent request" -ForegroundColor Red
}

# 4. Third Payment Attempt with NEW Idempotency Key (Should Fail - already paid)
Write-Host "
4. Third Payment Attempt with NEW Idempotency Key (Should Fail - already paid)..." -ForegroundColor Yellow
$newKey = [guid]::NewGuid().ToString()
try {
    $response = Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId/pay" -Method POST -Headers @{ "X-Idempotency-Key" = $newKey } -ContentType "application/json" -Body $payBody -UseBasicParsing -ErrorAction Stop
    Write-Host "   [FAIL] Allowed double payment!" -ForegroundColor Red
} catch {
    $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
    $errorData = $reader.ReadToEnd() | ConvertFrom-Json
    if ($errorData.code -eq "ALREADY_PAID") {
        Write-Host "   [OK] Correctly blocked double payment: $($errorData.error)" -ForegroundColor Green
        $duplicateCount++
    } else {
        Write-Host "   [FAIL] Failed with wrong error: $($errorData.error) - $($errorData.code)" -ForegroundColor Red
    }
}

Write-Host "
=== Results ===" -ForegroundColor Yellow
if ($successCount -eq 2 -and $duplicateCount -eq 1) {
    Write-Host "[OK][OK][OK] DOUBLE PAYMENT PROTECTION IS WORKING! [OK][OK][OK]" -ForegroundColor Green
} else {
    Write-Host "[FAIL][FAIL][FAIL] DOUBLE PAYMENT PROTECTION FAILED! [FAIL][FAIL][FAIL]" -ForegroundColor Red
}
