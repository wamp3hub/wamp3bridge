package main

import (
	"flag"
	"log"

	wampSerializer "github.com/wamp3hub/wamp3go/serializers"

	bridge "github.com/wamp3hub/wamp3bridge"
	bridgeTransports "github.com/wamp3hub/wamp3bridge/transports"
)

func main() {
	leftAddress := flag.String("leftAddress", "", "")
	leftTicket := flag.String("leftTicket", "", "")
	rightAddress := flag.String("rightAddress", "", "")
	rightTicket := flag.String("rightTicket", "", "")
	flag.Parse()

	leftRouter, leftRouterJoinError := bridgeTransports.WebsocketJoin(*leftAddress, *leftTicket, wampSerializer.DefaultSerializer)
	rightRouter, rightRouterJoinError := bridgeTransports.WebsocketJoin(*rightAddress, *rightTicket, wampSerializer.DefaultSerializer)

	if leftRouterJoinError == nil && rightRouterJoinError == nil {
		log.Printf("Connected")
	} else {
		log.Printf("Failed to unite")
	}

	select {
	case <-bridge.Unite(leftRouter, rightRouter):
	case <-bridge.Unite(rightRouter, leftRouter):
	}
}
