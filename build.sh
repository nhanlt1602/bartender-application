#!/bin/bash

echo "🔧 Tạo file rsrc.syso từ icon.ico..."
rsrc -ico hasaki.ico

echo "🚀 Build app thành .exe với icon + GUI..."
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o app-launch/kafka-consumer-gui.exe main.go

echo "✅ Done! File đã build tại app-launch/kafka-consumer-gui/"
