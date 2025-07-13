@echo off
title Check Auto Start Status
echo Checking Kafka Consumer Auto-Start Status...
echo.

:: Set the task name
set "TASK_NAME=KafkaConsumerAutoStart"

echo ========================================
echo Checking Scheduled Task Status
echo ========================================

:: Check if task exists and get its status
schtasks /query /tn "%TASK_NAME%" /fo table 2>nul
if %errorLevel% == 0 (
    echo.
    echo ✅ SUCCESS: Auto-start task "%TASK_NAME%" exists and is configured
    echo.
    echo ========================================
    echo Detailed Task Information
    echo ========================================
    schtasks /query /tn "%TASK_NAME%" /fo list
) else (
    echo.
    echo ❌ ERROR: Auto-start task "%TASK_NAME%" not found
    echo.
    echo This means the auto-start has not been set up yet.
    echo Please run setup-auto-start.bat as Administrator first.
)

echo.
echo ========================================
echo Additional Checks
echo ========================================

:: Check if start.bat exists
if exist "start.bat" (
    echo ✅ start.bat exists
) else (
    echo ❌ start.bat not found
)

:: Check if kafka-consumer.exe exists
if exist "app-launch\kafka-consumer.exe" (
    echo ✅ kafka-consumer.exe exists in app-launch directory
) else (
    echo ❌ kafka-consumer.exe not found in app-launch directory
)

:: Check logs directory
if exist "logs" (
    echo ✅ logs directory exists
    echo   Recent log files:
    dir /b /o-d logs\startup_*.log 2>nul | findstr /r "startup_.*\.log" | head -5
) else (
    echo ❌ logs directory not found
)

echo.
echo ========================================
echo Manual Verification Steps
echo ========================================
echo To manually verify in Windows Task Scheduler:
echo 1. Press Win + R, type "taskschd.msc" and press Enter
echo 2. Look for task named "KafkaConsumerAutoStart"
echo 3. Check if it's enabled and configured to run at startup
echo.
echo To test the start script manually:
echo 1. Run: start.bat
echo 2. Check if the application starts correctly
echo.
echo Press any key to exit...
pause >nul 