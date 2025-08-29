build-tcp:
	go build -o bin/app-tcp cmd/tcplistener/main.go

build-udp:
	go build -o bin/app-udp cmd/udpsender/main.go

build-http:
	go build -o bin/app-http cmd/httpserver/main.go

run-tcp: build-tcp
	./bin/app-tcp

run-udp: build-udp
	./bin/app-udp

run-http: build-http
	./bin/app-http