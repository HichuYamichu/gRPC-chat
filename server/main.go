package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/hichuyamichu/gRPC-chat/proto"
	"google.golang.org/grpc"
)

type server struct {
	Clients map[string]proto.Chat_ChatServer
}

func (s *server) broadcast(msg *proto.Message) {
	fmt.Println(s.Clients)
	go func(msg *proto.Message) {
		for _, stream := range s.Clients {
			stream.Send(msg)
		}
	}(msg)
}

func (s *server) Chat(stream proto.Chat_ChatServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	s.Clients[msg.Author] = stream
	defer delete(s.Clients, msg.Author)
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		s.broadcast(msg)
	}
}

var addr = flag.String("addr", "localhost:3000", "gRPC server address")

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *addr)

	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()

	clients := make(map[string]proto.Chat_ChatServer)
	proto.RegisterChatServer(s, &server{clients})

	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
