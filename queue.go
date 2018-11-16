package main

import (
	"time"
)

type Queue struct {
	Input                   chan *SecurityEventsContainer
	Options                 *Options
	NextChannel             chan []*SecurityEventsContainer
	_items                  []*SecurityEventsContainer
	_lastRun                time.Time
	_maxItems               int
	_maxSecurityEvents      int
	_containsCriticalEvent  bool
	_idleTimeout            time.Duration
	_idleTimeoutForCritical time.Duration
	_firstRun               bool
}

func (queue *Queue) Init() {
	queue._lastRun = time.Now()
	queue._maxItems = 10
	queue._maxSecurityEvents = 1000
	queue._idleTimeout = queue.Options.IdleTimeout
	queue._idleTimeoutForCritical = time.Second * 20
	queue._containsCriticalEvent = false
	queue._firstRun = true
}

func (queue *Queue) Flush() {
	itemsToSend := queue._items
	queue.NextChannel <- itemsToSend

	queue._items = nil
	queue._containsCriticalEvent = false
	queue._lastRun = time.Now()
	queue._firstRun = false
}

func (queue *Queue) Run() {

	// time ticker to flush events
	ticker := time.NewTicker(queue._idleTimeout)
	go func() {
		for _ = range ticker.C {
			// push fake nil to input to run reprocessing queue
			queue.Input <- nil
		}
	}()

	// speed up the first sending request
	time.AfterFunc(20*time.Second, func() {
		// push fake nil to input to run reprocessing queue
		queue.Input <- nil
	})

	for eventsContainer := range queue.Input {

		// debugJson(eventsContainer)
		if eventsContainer != nil {
			queue._items = append(queue._items, eventsContainer)

			// check if eventsContainer contains critical event, if yes, it should be send on server faster as usual
			if !queue._containsCriticalEvent {
				for _, securityEvent := range eventsContainer.SecurityEvents {
					if securityEvent.Critical {
						queue._containsCriticalEvent = true
						break
					}
				}
			}

		}

		currentTime := time.Now()
		// check time for flash
		if currentTime.Sub(queue._lastRun) > queue._idleTimeout {
			// debug("FLUSH by time. queue size: %d", len(queue._items))
			queue.Flush()
		} else if len(queue._items) >= queue._maxItems {
			// debug("FLUSH by maxItems. queue size: %d", len(queue._items))
			queue.Flush()
		} else if queue._containsCriticalEvent && currentTime.Sub(queue._lastRun) > queue._idleTimeoutForCritical {
			// debug("FLUSH by critical event. queue size: %d", len(queue._items))
			queue.Flush()
		} else if queue._firstRun {
			// debug("first run after start. queue size: %d", len(queue._items))
			queue.Flush()
		} else {
			totalEvents := 0
			for _, eventsContainer := range queue._items {
				totalEvents += len(eventsContainer.SecurityEvents)
			}

			if totalEvents >= queue._maxSecurityEvents {
				// debug("FLUSH by max security events. size: %d, max: %d", totalEvents, queue._maxSecurityEvents)
				queue.Flush()
			}
		}
	}
}
