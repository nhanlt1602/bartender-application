@echo off
title Kafka Consumer Auto Start
echo Starting Kafka Consumer Application...

:: Set the working directory to the script location
cd /d "%~dp0"

:: Create logs directory if it doesn't exist
if not exist "logs" mkdir logs

:: Get current timestamp for log file
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YY=%dt:~2,2%" & set "YYYY=%dt:~0,4%" & set "MM=%dt:~4,2%" & set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%" & set "Min=%dt:~10,2%" & set "Sec=%dt:~12,2%"
set "timestamp=%YYYY%%MM%%DD%_%HH%%Min%%Sec%"

:: Set log file path
set "logfile=app-launch\logs\startup_%timestamp%.log"

:: Log startup information
echo [%date% %time%] Starting Kafka Consumer Application >> "%logfile%"
echo [%date% %time%] Working Directory: %CD% >> "%logfile%"

:: Check if executable exists
if not exist "app-launch\kafka-consumer.exe" (
    echo [%date% %time%] ERROR: kafka-consumer.exe not found in app-launch directory >> "%logfile%"
    echo ERROR: kafka-consumer.exe not found in app-launch directory
    pause
    exit /b 1
)

:: Start the application
echo [%date% %time%] Launching kafka-consumer.exe... >> "%logfile%"
echo Launching kafka-consumer.exe...

:: Change to app-launch directory and run the executable
cd /d "%~dp0app-launch"
start /wait "Kafka Consumer" kafka-consumer.exe

:: Log if application exits
echo [%date% %time%] Application exited with code %errorlevel% >> "..\%logfile%"
echo Application exited with code %errorlevel%

:: If application crashes, wait a bit and restart
if %errorlevel% neq 0 (
    echo [%date% %time%] Application crashed, restarting in 10 seconds... >> "..\%logfile%"
    echo Application crashed, restarting in 10 seconds...
    timeout /t 10 /nobreak > nul
    goto :restart
)

:restart
:: Restart the application
echo [%date% %time%] Restarting application... >> "..\%logfile%"
echo Restarting application...
goto :eof
