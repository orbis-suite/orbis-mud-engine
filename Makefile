.PHONY: test proto engine game run clean

test:
	go test ./...

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       plugin/proto/orbis.proto

engine:
	go build -o orbis .

game:
	go build -o game-binary ./game

run: proto engine game
	./orbis

clean:
	rm -f orbis game-binary
