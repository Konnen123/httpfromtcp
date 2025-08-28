build-tcp:
	go build -o bin/app-tcp cmd/tcplistener/main.go

build-udp:
	go build -o bin/app-udp cmd/udpsender/main.go

run-tcp: build-tcp
	./bin/app-tcp

run-udp: build-udp
	./bin/app-udp