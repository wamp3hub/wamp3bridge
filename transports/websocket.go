package bridgeTransports

import (
	wamp "github.com/wamp3hub/wamp3go"
	wampTransport "github.com/wamp3hub/wamp3go/transport"
)

func WebsocketJoin(
	address string,
	peerID string,
	ticket string,
	serializer wamp.Serializer,
) (*wamp.Session, error) {
	wsAddress := "ws://" + address + "/wamp3/websocket?ticket=" + ticket
	_, transport, e := wampTransport.WebsocketConnect(wsAddress, serializer)
	if e == nil {
		peer := wamp.SpawnPeer(peerID, transport)
		session := wamp.NewSession(peer)
		return session, nil
	}
	return nil, e
}
