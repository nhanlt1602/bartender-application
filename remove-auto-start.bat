@echo off
title Remove Auto Start for Kafka Consumer
echo Removing auto-start for Kafka Consumer...

:: Check if running as administrator
net session >nul 2>&1
if %errorLevel% == 0 (
    echo Running as administrator - proceeding with removal
) else (
    echo ERROR: This script must be run as Administrator
    echo Please right-click on this file and select "Run as administrator"
    pause
    exit /b 1
)

:: Set the task name
set "TASK_NAME=KafkaConsumerAutoStart"

:: Check if task exists
schtasks /query /tn "%TASK_NAME%" >nul 2>&1
if %errorLevel% == 0 (
    echo Found existing task: %TASK_NAME%
    echo Removing task...
    schtasks /delete /tn "%TASK_NAME%" /f
    
    if %errorLevel% == 0 (
        echo.
        echo SUCCESS: Auto-start task removed successfully!
        echo The Kafka Consumer will no longer start automatically on Windows restart.
    ) else (
        echo.
        echo ERROR: Failed to remove scheduled task
        echo Error code: %errorLevel%
    )
) else (
    echo.
    echo INFO: No auto-start task found with name: %TASK_NAME%
    echo The task may have already been removed or never existed.
)

echo.
echo Press any key to exit...
pause >nul 