build:
	go build  -o pygoraftkv.bin
	
lib:
	mkdir -p dist && \
	rm -rf dist/* && \
	cd dist && gopy build  ../pygoraftkv 
	
	
