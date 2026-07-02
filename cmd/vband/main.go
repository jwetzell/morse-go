package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/jwetzell/morse-go/vband"
)

var ip string
var port uint

func init() {
	flag.StringVar(&ip, "ip", "0.0.0.0", "IP address to listen on")
	flag.UintVar(&port, "port", 7385, "Port to listen on")
}

func main() {
	flag.Parse()

	if port < 1 || port > 65535 {
		panic("Port must be between 1 and 65535")
	}

	server := vband.NewWebSocketServer()

	server.CreateChannel("Lobby")
	server.CreateChannel("Practice Channel")
	server.CreateChannel("Channel 1")
	server.CreateChannel("Channel 2")
	server.CreateChannel("Channel 3")
	server.CreateChannel("Channel 4")
	server.CreateChannel("Channel 5")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	server.Start(ctx, ip, port)
	<-ctx.Done()
	server.Stop()
}
