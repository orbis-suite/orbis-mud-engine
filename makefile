.PHONY: run teal go clean

run: teal go

teal:
	cd data && cyan build

go:
	go run .

clean:
	cd data && cyan clean
