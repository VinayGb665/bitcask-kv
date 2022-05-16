# Bitcask-KV
## Help me if you've already worked on things like this or just wanna hangout

[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://travis-ci.org/joemccann/dillinger)



## Installation

Requires golang, preferrably the latest version to build and change
If you dont have golang, the binary would be the way to go, just to get started right away

### Build
```sh
go build -o bitcask-kv
```
### Run

```sh
make start-server
or
./bitcask-kv -server=true -port=8080 -storage=/tmp/bitcask -max-file-size=104857600
```
### Start client
```sh
make start-client
or
./bitcask-kv -address=localhost:8080
```

## Simply explained
- You will be able to set a key and get a key basically
- Index file is maintained to keep track of keys and even updates to then
- Index file just has the CRC, timestamp along with details related to location of the value like fileID, offset and value size
- Writes are an append to the activeFile and an entry regarding the same to indexFile
- Activefile changes if the max-file-size exceeds
- All the reads are just one file seek and read so the read throughput is good
- Crash recovery is excellent because the actual data files themselves are the commit logs too, on startup all the keys are loaded back without anything going wrong drastically
- Index file records are maintained in a base64 + gob encodings, timestamps are used to preserve the latest version of values
- All the keys reside in memory but none of the values do
- Loadup times are quite high around 90s in case of 20M keys
- Older values are not tombstoned yet, so some useless data will exist if keys are updated/deleted
- 
## Notes
-   Bro do no DO NOT use this for prod or even anything just some experiments here
-   Implementation does not exactly adhere to the paper, some modifications have been done to improvise loadup and efficiency
