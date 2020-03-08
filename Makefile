build:
	mkdir -p dist && \
	rm -rf dist/* && \
	cd dist && gopy build  ../pygoraftkv 
	
	
