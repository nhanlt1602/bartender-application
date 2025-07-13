@echo off
title Setup Auto Start for Kafka Consumer
echo Setting up auto-start for Kafka Consumer...

:: Check if running as administrator
net session >nul 2>&1
if %errorLevel% == 0 (
    echo Running as administrator - proceeding with setup
) else (
    echo ERROR: This script must be run as Administrator
    echo Please right-click on this file and select "Run as administrator"
    pause
    exit /b 1
)

:: Get the current directory (where this script is located)
set "SCRIPT_DIR=%~dp0"
set "START_SCRIPT=%SCRIPT_DIR%start.bat"

:: Check if start.bat exists
if not exist "%START_SCRIPT%" (
    echo ERROR: start.bat not found in %SCRIPT_DIR%
    pause
    exit /b 1
)

:: Create the task name
set "TASK_NAME=KafkaConsumerAutoStart"

:: Delete existing task if it exists
echo Removing existing task if any...
schtasks /delete /tn "%TASK_NAME%" /f >nul 2>&1

:: Create new scheduled task
echo Creating scheduled task...
schtasks /create /tn "%TASK_NAME%" /tr "\"%START_SCRIPT%\"" /sc onstart /ru "SYSTEM" /rl highest /f

if %errorLevel% == 0 (
    echo.
    echo SUCCESS: Auto-start task created successfully!
    echo Task Name: %TASK_NAME%
    echo Trigger: On system startup
    echo Command: %START_SCRIPT%
    echo.
    echo The Kafka Consumer will now automatically start when Windows restarts.
    echo.
    echo To verify the task was created, you can run:
    echo schtasks /query /tn "%TASK_NAME%"
    echo.
    echo To remove the auto-start later, run:
    echo schtasks /delete /tn "%TASK_NAME%" /f
) else (
    echo.
    echo ERROR: Failed to create scheduled task
    echo Error code: %errorLevel%
    echo.
    echo Please check that:
    echo 1. You are running as Administrator
    echo 2. The start.bat file exists and is accessible
    echo 3. Windows Task Scheduler service is running
)

echo.
echo Press any key to exit...
pause >nul 