package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/hichuyamichu/gRPC-chat/proto"
	"github.com/jroimartin/gocui"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", "localhost:3000", "gRPC server address")
var name = flag.String("name", "jan", "username")
var stream proto.Chat_ChatClient

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	typerY := maxY / 8
	chatY := maxY - (typerY) - 1

	if v, err := g.SetView("typer", 0, chatY, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Editable = true

		if _, err := g.SetCurrentView("typer"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("chatBox", 0, 0, maxX-1, chatY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Wrap = true
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func send(g *gocui.Gui, v *gocui.View) error {
	typer, err := g.View("typer")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(typer)
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	text = strings.Replace(text, "\r", "", 1)
	msg := &proto.Message{Author: *name, Content: text}
	stream.Send(msg)
	if err != nil {
		return err
	}
	typer.Clear()
	typer.SetCursor(0, 0)
	return nil
}

func readStream(stream proto.Chat_ChatClient, g *gocui.Gui) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		g.Update(func(g *gocui.Gui) error {
			chatBox, err := g.View("chatBox")
			if err != nil {
				return err
			}
			m := fmt.Sprintf("<%v> %v", msg.Author, msg.Content)
			_, err = fmt.Fprintln(chatBox, m)
			if err != nil {
				return err
			}
			return nil
		})
	}
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := proto.NewChatClient(conn)
	stream, err = c.Chat(context.Background())
	if err != nil {
		return
	}
	stream.Send(&proto.Message{Author: *name, Content: ""})

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorRed
	g.SetManagerFunc(layout)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatal(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, send); err != nil {
		log.Fatal(err)
	}

	go readStream(stream, g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
	}
}
