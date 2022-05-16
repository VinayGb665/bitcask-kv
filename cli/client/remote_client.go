package client_cli

import (
	"fmt"
	"net/rpc"
	"time"

	Utils "github.com/vinaygb665/bitcask-kv/utils"

	"github.com/abiosoft/ishell"
)

func Start(address string) {
	// RPC client
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		panic(err)
	}
	shell := ishell.New()
	fmt.Println("Welcome to Bitcask-KV remote client")
	// RPC call
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "get <key>",
		Func: func(c *ishell.Context) {
			// c.Println("Hello", strings.Join(c.Args, " "))
			start := time.Now()
			var response Utils.GetResponse
			err = client.Call("Server.Get", &Utils.GetRequest{Key: c.Args[0]}, &response)
			if err != nil {
				c.Err(err)
			}
			c.Println(string(response.Value))
			c.Println("Time taken: ", time.Since(start))

		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "set <key> <value>",
		Func: func(c *ishell.Context) {
			// c.Println("Hello", strings.Join(c.Args, " "))
			start := time.Now()
			if len(c.Args) != 2 {
				c.Println("Invalid arguments")
				return
			}

			var response Utils.SetResponse
			err = client.Call(
				"Server.Set",
				&Utils.SetRequest{
					Key: c.Args[0], Value: []byte(c.Args[1]),
				},
				&response,
			)
			if err != nil {
				c.Err(err)
			}
			c.Println(response.Success)
			c.Println("Time taken: ", time.Since(start))

		},
	})

	shell.Run()
}
