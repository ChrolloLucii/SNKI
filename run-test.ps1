param([string]$ApiUrl = "http://localhost:8080")

$slotId = "55555555-5555-5555-5555-555555555555"

# Get slot info
$slot = (Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId" -UseBasicParsing).Content | ConvertFrom-Json
$capacity = $slot.data.capacity
$current = $slot.data.participants
Write-Host "Slot capacity: $capacity, Current: $current"

# Test users
$users = @(
    "22222222-2222-2222-2222-222222222222",
    "33333333-3333-3333-3333-333333333333",
    "44444444-4444-4444-4444-444444444444",
    "55555555-5555-5555-5555-555555555556",
    "66666666-6666-6666-6666-666666666666",
    "77777777-7777-7777-7777-777777777777",
    "88888888-8888-8888-8888-888888888888",
    "99999999-9999-9999-9999-999999999999"
)

$success = 0
$failed = 0

foreach ($uid in $users) {
    $body = "{`"user_id`":`"$uid`"}"
    
    try {
        $r = Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId/join" -Method POST -ContentType "application/json" -Body $body -UseBasicParsing -ErrorAction Stop
        $data = $r.Content | ConvertFrom-Json
        Write-Host "[OK] $uid - $($data.message)"
        $success++
    }
    catch {
        $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
        $body = $reader.ReadToEnd()
        $errorData = $body | ConvertFrom-Json
        Write-Host "[FAIL] $uid - $($errorData.error)"
        $failed++
    }
}

$final = (Invoke-WebRequest -Uri "$ApiUrl/slots/$slotId" -UseBasicParsing).Content | ConvertFrom-Json
Write-Host ""
Write-Host "Results: Joined=$success Failed=$failed Final=$($final.data.participants)/$capacity"

if ($failed -gt 0 -and $final.data.participants -eq $capacity) {
    Write-Host "SUCCESS: Overbooking protection is working!"
} else {
    Write-Host "FAILED: Overbooking protection not working!"
}
