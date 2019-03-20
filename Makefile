all: run

run:
	go run *.go

build:
	go build -o qt

demo: build
	# Torrenting Sintel.
	./qt

clean:
	rm -f qt
	rm -rf Sintel
	rm -rf There*
