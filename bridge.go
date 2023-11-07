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

type Router = wamp.Session

type Bridge struct {
	leftRouter        *Router
	rightRouter       *Router
	subscriptionLinks map[string]*Link
	registrationLinks map[string]*Link
}

func (bridge *Bridge) subscribe(
	subscription *wamp.Subscription,
) error {
	if subscription.Options.Entrypoint() == bridge.leftRouter.ID() {
		return errors.New("CircularSubscribeDetected")
	}

	link, found := bridge.subscriptionLinks[subscription.URI]
	if found {
		link.Clients.Add(subscription.AuthorID)
		return nil
	}

	mySubscription, e := wamp.Subscribe(
		bridge.rightRouter,
		subscription.URI,
		subscription.Options,
		func(publishEvent wamp.PublishEvent) {
			e := wamp.Publish(bridge.leftRouter, publishEvent.Features(), publishEvent.Content())
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
			ResourceID: mySubscription.ID,
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
			e = wamp.Unsubscribe(bridge.rightRouter, link.ResourceID)
		}
	} else {
		e = errors.New("LinkNotFound")
	}
	return e
}

func (bridge *Bridge) register(
	registration *wamp.Registration,
) error {
	if registration.Options.Entrypoint() == bridge.leftRouter.ID() {
		return errors.New("CircularRegisterDetected")
	}

	link, found := bridge.registrationLinks[registration.URI]
	if found {
		link.Clients.Add(registration.AuthorID)
		return nil
	}

	myRegistration, e := wamp.Register(
		bridge.rightRouter,
		registration.URI,
		registration.Options,
		func(callEvent wamp.CallEvent) wamp.ReplyEvent {
			pendingResponse := wamp.Call[any](bridge.leftRouter, callEvent.Features(), callEvent.Content())
			replyEvent, _, _ := pendingResponse.Await()
			return replyEvent
		},
	)
	if e == nil {
		link = &Link{
			URI:        registration.URI,
			ResourceID: myRegistration.ID,
			Clients: routerShared.NewSet[string](
				[]string{registration.AuthorID},
			),
		}
		bridge.registrationLinks[registration.URI] = link
	}

	return e
}

func (bridge *Bridge) unregister(
	registration *wamp.Registration,
) (e error) {
	link, found := bridge.registrationLinks[registration.URI]
	if found {
		link.Clients.Delete(registration.AuthorID)
		if link.Clients.Size() == 0 {
			delete(bridge.registrationLinks, registration.URI)
			e = wamp.Unregister(bridge.rightRouter, link.ResourceID)
		}
	} else {
		e = errors.New("LinkNotFound")
	}
	return e
}

func Unite(
	leftRouter *Router,
	rightRouter *Router,
) <-chan struct{} {
	bridge := Bridge{leftRouter, rightRouter, make(map[string]*Link), make(map[string]*Link)}

	wamp.Subscribe(
		leftRouter,
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
		leftRouter,
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

	wamp.Subscribe(
		leftRouter,
		"wamp.registration.new",
		&wamp.SubscribeOptions{},
		func(publishEvent wamp.PublishEvent) {
			registration := new(wamp.Registration)
			e := publishEvent.Payload(registration)
			if e == nil {
				bridge.subscribe(registration)
			}
		},
	)

	wamp.Subscribe(
		leftRouter,
		"wamp.registration.gone",
		&wamp.SubscribeOptions{},
		func(publishEvent wamp.PublishEvent) {
			registration := new(wamp.Registration)
			e := publishEvent.Payload(registration)
			if e == nil {
				bridge.unsubscribe(registration)
			}
		},
	)

	quit := make(chan struct{})
	return quit
}
