all: build

build:
	go build -o qt

demo: build
	# Torrenting Sintel.
	./qt

# run: build
# 	./qt

clean:
	rm -f qt
	rm -rf Sintel
	rm -rf There*
