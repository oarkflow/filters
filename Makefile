build:
	GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o ./wasm/filters.wasm ./wasm/main.go
	cp "$(GOROOT)/misc/wasm/wasm_exec.js" ./wasm/

build-go:
	tinygo build -o ./wasm/filters.wasm -target=wasm -gc=leaking -no-debug ./wasm/main.go
	cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js ./wasm/

cp-exec:
	echo $(tinygo env TINYGOROOT)