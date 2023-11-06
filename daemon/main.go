package main

import (
	"flag"

	"github.com/rs/xid"

	wampSerializer "github.com/wamp3hub/wamp3go/serializer"

	bridge "github.com/wamp3hub/wamp3bridge"
	bridgeTransports "github.com/wamp3hub/wamp3bridge/transports"
)

func main() {
	leftAddress := flag.String("leftAddress", "", "")
	leftTicket := flag.String("leftTicket", "", "")
	rightAddress := flag.String("rightAddress", "", "")
	rightTicket := flag.String("rightTicket", "", "")
	flag.Parse()

	left, _ := bridgeTransports.WebsocketJoin(*leftAddress, xid.New().String(), *leftTicket, wampSerializer.DefaultSerializer)
	right, _ := bridgeTransports.WebsocketJoin(*rightAddress, xid.New().String(), *rightTicket, wampSerializer.DefaultSerializer)

	select {
	case <-bridge.Unite(left, right):
	case <-bridge.Unite(right, left):
	}
}
