package wamp3bridge

import (
	"errors"
	"log"

	wamp "github.com/wamp3hub/wamp3go"
	routerShared "github.com/wamp3hub/wamp3router/shared"
)

type Link struct {
	URI        string
	ResourceID string
	Clients    *routerShared.Set[string]
}

type Bridge struct {
	left              *wamp.Session
	right             *wamp.Session
	subscriptionLinks map[string]*Link
}

func (bridge *Bridge) subscribe(
	subscription *wamp.Subscription,
) error {
	link, found := bridge.subscriptionLinks[subscription.URI]
	if found {
		link.Clients.Add(subscription.AuthorID)
		return nil
	}

	subscription, e := wamp.Subscribe(
		bridge.right,
		subscription.URI,
		subscription.Options,
		func(publishEvent wamp.PublishEvent) {
			e := wamp.Publish(bridge.left, publishEvent.Features(), publishEvent.Content())
			if e == nil {
				log.Printf("[bridge] publish forward sucess URI=%s", subscription.URI)
			} else {
				log.Printf("[bridge] publish forward error=%s URI=%s", e, subscription.URI)
			}
		},
	)
	if e == nil {
		link = &Link{
			URI:        subscription.URI,
			ResourceID: subscription.ID,
			Clients: routerShared.NewSet[string](
				[]string{subscription.AuthorID},
			),
		}
		bridge.subscriptionLinks[subscription.URI] = link
	}

	return e
}

func (bridge *Bridge) unsubscribe(
	subscription *wamp.Subscription,
) (e error) {
	link, found := bridge.subscriptionLinks[subscription.URI]
	if found {
		link.Clients.Delete(subscription.AuthorID)
		if link.Clients.Size() == 0 {
			delete(bridge.subscriptionLinks, subscription.URI)
			e = wamp.Unsubscribe(bridge.right, link.ResourceID)
		}
	} else {
		e = errors.New("LinkNotFound")
	}
	return e
}

func Unite(
	left *wamp.Session,
	right *wamp.Session,
) <-chan struct{} {
	bridge := Bridge{left, right, make(map[string]*Link)}

	wamp.Subscribe(
		left,
		"wamp.subscription.new",
		&wamp.SubscribeOptions{},
		func(publishEvent wamp.PublishEvent) {
			subscription := new(wamp.Subscription)
			e := publishEvent.Payload(subscription)
			if e == nil {
				bridge.subscribe(subscription)
			}
		},
	)

	wamp.Subscribe(
		left,
		"wamp.subscription.gone",
		&wamp.SubscribeOptions{},
		func(publishEvent wamp.PublishEvent) {
			subscription := new(wamp.Subscription)
			e := publishEvent.Payload(subscription)
			if e == nil {
				bridge.unsubscribe(subscription)
			}
		},
	)

	quit := make(chan struct{})
	return quit
}
