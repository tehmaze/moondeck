package moonraker

import (
	"log"
	"sync"
)

type eventType int

const (
	eventOnline eventType = iota
	eventOffline
)

type eventDispatcher struct {
	sync.RWMutex
	Listeners []EventListener
}

func (d *eventDispatcher) Dispatch(c *Client, err error) {
	d.RLock()
	for _, l := range d.Listeners {
		go func(l EventListener, c *Client, err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("moonraker: PANIC during execution of event listener %T: %v", l, err)
				}
			}()
			l(c, err)
		}(l, c, err)
	}
	d.RUnlock()
}

func (d *eventDispatcher) Subscribe(l EventListener) {
	d.RLock()
	d.Listeners = append(d.Listeners, l)
	d.RUnlock()
}

var subscribers = make(map[eventType]*eventDispatcher)

func subscribe(t eventType, l EventListener) {
	d, ok := subscribers[t]
	if !ok {
		d = new(eventDispatcher)
		subscribers[t] = d
	}
	d.Subscribe(l)
}

func dispatch(t eventType, c *Client, err error) {
	if d, ok := subscribers[t]; ok {
		d.Dispatch(c, err)
	}
}

type EventListener func(*Client, error)

func IsOnline(l EventListener)  { subscribe(eventOnline, l) }
func IsOffline(l EventListener) { subscribe(eventOffline, l) }
