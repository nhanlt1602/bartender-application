@echo off
echo 🔧 Creating rsrc.syso from hasaki.ico...
rsrc -ico hasaki.ico

echo 🚀 Building Go app...
go build -ldflags="-H windowsgui" -o app-launch\kafka-consumer-gui.exe main.go

echo ✅ Build completed!
pause
