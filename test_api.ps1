# Red Packet API Test Script (PowerShell)
# Usage: .\test_api.ps1

$BASE_URL = "http://localhost:8080/api/v1"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Red Packet API Test Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

function Test-API {
    param($Step, $Method, $Path, $Body)

    Write-Host "[$Step] Method=$Method Path=$Path" -ForegroundColor Yellow
    try {
        $params = @{
            Uri = "$BASE_URL$Path"
            Method = $Method
        }
        if ($Body) {
            $params["ContentType"] = "application/json"
            $params["Body"] = $Body
        }
        $result = Invoke-RestMethod @params
        $result | ConvertTo-Json -Depth 10
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
    }
    Write-Host ""
}

Write-Host "[1/7] Creating Red Packet Activity #1..." -ForegroundColor Green
$body1 = '{"name":"端午节红包活动","total_amount":10000,"total_count":10,"min_amount":500,"max_amount":2000,"start_time":"2026-01-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z"}'
Test-API "1/7" "POST" "/activities" $body1

Write-Host "[2/7] Creating Red Packet Activity #2..." -ForegroundColor Green
$body2 = '{"name":"新人专享红包","total_amount":5000,"total_count":5,"min_amount":800,"max_amount":1200,"start_time":"2026-01-01T00:00:00Z","end_time":"2026-12-31T23:59:59Z"}'
Test-API "2/7" "POST" "/activities" $body2

Write-Host "[3/7] Get Activity List..." -ForegroundColor Green
Test-API "3/7" "GET" "/activities?page=1&page_size=10"

Write-Host "[4/7] User1 Grab Red Packet (Activity 1)..." -ForegroundColor Green
$grab1 = '{"activity_id":1,"user_id":"user_001"}'
Test-API "4/7" "POST" "/redpacket/grab" $grab1

Write-Host "[5/7] User2 Grab Red Packet (Activity 1)..." -ForegroundColor Green
$grab2 = '{"activity_id":1,"user_id":"user_002"}'
Test-API "5/7" "POST" "/redpacket/grab" $grab2

Write-Host "[6/7] User1 Duplicate Grab (Should FAIL)..." -ForegroundColor Green
Test-API "6/7" "POST" "/redpacket/grab" $grab1

Write-Host "[7/7] Get Grab Records for Activity #1..." -ForegroundColor Green
Test-API "7/7" "GET" "/activities/1/records?page=1&page_size=20"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Additional Test Commands:" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Get Activity Detail: " -NoNewline
Write-Host "Invoke-RestMethod '$BASE_URL/activities/1'" -ForegroundColor Gray
Write-Host "  Get User RedPackets: " -NoNewline
Write-Host "Invoke-RestMethod '$BASE_URL/users/user_001/redpackets'" -ForegroundColor Gray
Write-Host ""
