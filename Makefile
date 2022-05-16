
clean:
	rm -rf /tmp/bitcask/*
	rm bitcask-kv

watch-storage:
	watch -n0 du -sh /tmp/bitcask/*.dat

build:
	go build -o bitcask-kv

start-server:
	./bitcask-kv -server=true -port=8080 -storage=/tmp/bitcask -max-file-size=1024

start-client:
	./bitcask-kv -address=localhost:8080 