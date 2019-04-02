all: run

install:
	go install

run:
	go run *.go

build:
	go build -o qt

demo: build
	# Torrenting Sintel.
	./qt

clean:
	rm -f qt
	# Removing test torrent leftovers.
	rm -f .torrent*
	rm -rf Sintel
	rm -rf There*
