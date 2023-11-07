package bridgeTransports

import (
	wamp "github.com/wamp3hub/wamp3go"
	wampTransports "github.com/wamp3hub/wamp3go/transports"
	routerShared "github.com/wamp3hub/wamp3router/shared"
)

func WebsocketJoin(
	address string,
	ticket string,
	serializer wamp.Serializer,
) (*wamp.Session, error) {
	claims, e := routerShared.JWTParse(nil, ticket)
	if e == nil {
		wsAddress := "ws://" + address + "/wamp3/websocket?ticket=" + ticket
		transport, e := wampTransports.WebsocketConnect(wsAddress, serializer)
		if e == nil {
			peer := wamp.SpawnPeer(claims.Subject, transport)
			session := wamp.NewSession(peer)
			return session, nil
		}
	}
	return nil, e
}
