package main

import (
	"flag"
	"fmt"
	"strconv"

	Server "github.com/vinaygb665/bitcask-kv/server"

	Client "github.com/vinaygb665/bitcask-kv/cli/client"
)

func main() {
	// Get args
	// args := os.Args[1:]
	// if len(args) < 2 {
	// 	fmt.Println("Usage: ./bitcask-kv server <address> or ./bitcask-kv client <address>")
	// 	return
	// }

	serverPtr := flag.Bool("server", false, "Run as server")

	var addressPtr string
	flag.StringVar(&addressPtr, "address", "localhost:1234", "address")

	var portPtr int
	flag.IntVar(&portPtr, "port", 1234, "port number")

	storageDir := flag.String("storage", "/tmp/bitcask/", "storage directory")
	maxFileSize := flag.Int64("max-file-size", 1024*1024*1024, "max file size")

	flag.Parse()

	if *serverPtr {
		// Start server
		fmt.Println("Starting server at 0.0.0.0:", portPtr)
		Server.Start(strconv.Itoa(portPtr), *storageDir, int64(*maxFileSize))
	} else {
		// Start client
		fmt.Println("Starting client to ", (addressPtr))
		Client.Start(string(addressPtr))
	}
}
