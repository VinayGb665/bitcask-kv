package server

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	Utils "github.com/vinaygb665/bitcask-kv/utils"

	Bitcask "github.com/vinaygb665/bitcask-kv/bitcask"
)

type Server struct {
	Storage *Bitcask.Storage
}

func (s *Server) Get(req *Utils.GetRequest, res *Utils.GetResponse) error {
	fmt.Println("Yoooo ")
	res.Value, res.Success = s.Storage.Read(req.Key)
	return nil
}
func (s *Server) Set(req *Utils.SetRequest, res *Utils.SetResponse) error {
	err := s.Storage.Write(req.Key, req.Value)
	if err != nil {
		res.Success = false
		return nil
	}
	res.Success = true
	return nil
}

func Start(port string, storageDir string, maxFileSize int64) {
	// Start a rpc server
	server := &Server{}
	server.Storage = &Bitcask.Storage{}
	server.Storage.Init(storageDir, false, maxFileSize)

	// Start rpc server
	rpc.Register(server)
	rpc.HandleHTTP()

	// Start http server
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	http.Serve(l, nil)

}
