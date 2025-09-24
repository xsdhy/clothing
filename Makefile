



build:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o clothing



web:
	cp ./web/dist/index.html cmd/server/web/dist