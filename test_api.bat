@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo   Red Packet API Test Script
echo ========================================
echo.

set BASE_URL=http://localhost:8080/api/v1

echo [1/6] Creating Red Packet Activity...
echo.
curl -s -X POST "%BASE_URL%/activities" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"端午节红包活动\",\"total_amount\":10000,\"total_count\":10,\"min_amount\":500,\"max_amount\":2000,\"start_time\":\"2026-01-01T00:00:00Z\",\"end_time\":\"2026-12-31T23:59:59Z\"}"
echo.
echo.

echo [2/6] Creating Second Red Packet Activity...
echo.
curl -s -X POST "%BASE_URL%/activities" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"新人专享红包\",\"total_amount\":5000,\"total_count\":5,\"min_amount\":800,\"max_amount\":1200,\"start_time\":\"2026-01-01T00:00:00Z\",\"end_time\":\"2026-12-31T23:59:59Z\"}"
echo.
echo.

echo [3/6] Get Activity List...
echo.
curl -s "%BASE_URL%/activities?page=1&page_size=10"
echo.
echo.

echo [4/6] User1 Grab Red Packet (Activity 1)...
echo.
curl -s -X POST "%BASE_URL%/redpacket/grab" ^
  -H "Content-Type: application/json" ^
  -d "{\"activity_id\":1,\"user_id\":\"user_001\"}"
echo.
echo.

echo [5/6] User2 Grab Red Packet (Activity 1)...
echo.
curl -s -X POST "%BASE_URL%/redpacket/grab" ^
  -H "Content-Type: application/json" ^
  -d "{\"activity_id\":1,\"user_id\":\"user_002\"}"
echo.
echo.

echo [6/6] Get Grab Records for Activity 1...
echo.
curl -s "%BASE_URL%/activities/1/records?page=1&page_size=20"
echo.
echo.

echo ========================================
echo   Test Complete!
echo ========================================
echo.
echo Additional test commands you can run:
echo   - Get Activity Detail: curl "%BASE_URL%/activities/1"
echo   - Get User RedPackets: curl "%BASE_URL%/users/user_001/redpackets"
echo   - Duplicate Grab (should fail): curl -X POST "%BASE_URL%/redpacket/grab" -H "Content-Type: application/json" -d "{\"activity_id\":1,\"user_id\":\"user_001\"}"
echo.

endlocal
