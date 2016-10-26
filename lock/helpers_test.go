package lock

import (
	"log"
	"net"

	ect "github.com/coreos/etcd/client"
)

func init() {
	s, err := net.Dial("tcp", "127.0.0.1:2379")
	if err != nil {
		log.Fatalln("etcd is not running on localhost", err)
	}
	s.Close()
}

func client() ect.Client {
	client, _ := ect.New(ect.Config{Endpoints: []string{"http://localhost:2379"}})
	return client
}
