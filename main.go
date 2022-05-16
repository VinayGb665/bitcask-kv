package main

import (
	"flag"
	"fmt"
	"os"
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
		serverPort := strconv.Itoa(portPtr)
		fmt.Println("Starting server at 0.0.0.0:", portPtr)
		envPort, err := strconv.Atoi(os.Getenv("PORT"))

		if err == nil {
			serverPort = strconv.Itoa(envPort)
		}

		Server.Start(serverPort, *storageDir, int64(*maxFileSize))
	} else {
		// Start client
		fmt.Println("Starting client to ", (addressPtr))
		Client.Start(string(addressPtr))
	}
}
