#################################### Local build ##############################################################################
local:
	set ENV=development
	go run main.go

#################################### QA build ##############################################################################
qa:
	set ENV=qa
	go run main.go

build-app-qa:
	ENV=qa GOOS=windows GOARCH=amd64 go build -o app-launch/kafka-consumer-qa.exe -ldflags="-s -w" main.go

build-gui-qa:
	ENV=qa GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o app-launch/kafka-consumer-gui-qa.exe main.go

#################################### Production build ###########################################################################
prod:
	set ENV=production
	go run main.go

build-app-prod:
	ENV=production GOOS=windows GOARCH=amd64 go build -o app-launch/kafka-consumer-prod.exe -ldflags="-s -w" main.go

build-gui-prod:
	ENV=production GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o app-launch/kafka-consumer-gui-prod.exe main.go

##################################################################################################################################

# Build Icon
rsrc:
	rsrc -ico icon.ico


# RUN
run-application:
	./app-launch/kafka-consumer-.exe

run-application-windows-logs:
	.\app-launch\kafka-consumer.exe > .\app-launch\logs\kafka-consumer.log 2>&1


build:
	set GOOS=windows
	set GOARCH=amd64
	set ENV=qa
	go build -o app-launch/kafka-consumer-qc.exe -ldflags="-s -w" main.go

