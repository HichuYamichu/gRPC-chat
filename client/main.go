package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/HichuYamichu/gRPC-chat/proto"
	"google.golang.org/grpc"
)

func readStream(stream proto.Chat_ChatClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		m := fmt.Sprintf("<%v> %v", msg.Author, msg.Content)
		_, err = fmt.Fprintln(os.Stdout, m)
		if err != nil {
			log.Fatal(err)
		}

	}
}

var addr = flag.String("addr", "localhost:3000", "gRPC server address")
var name = flag.String("name", "jan", "username")

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := proto.NewChatClient(conn)
	stream, err := c.Chat(context.Background())
	if err != nil {
		return
	}

	stream.Send(&proto.Message{Author: *name, Content: ""})

	go readStream(stream)
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", 1)
		msg := &proto.Message{Author: *name, Content: text}
		stream.Send(msg)
	}
}
