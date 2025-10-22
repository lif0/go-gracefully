package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"github.com/lif0/go-gracefully"
)

func main() {
	// Configure gracefully trigger
	gracefully.SetShutdownTrigger(context.Background(),
		gracefully.WithSysSignal(),
	)

	// init service
	serverEventCollector := NewBatcher("server_events.log")
	userEventCollector := NewBatcher("user_events.log")
	gracefully.MustRegister(serverEventCollector, userEventCollector)
	/* OR
	serverEventCollector := gracefully.NewInstance(func() *eventBatcher { return NewBatcher("server_events.log") })
	userEventCollector := gracefully.NewInstance(func() *eventBatcher { return NewBatcher("user_events.log") })
	*/

	go serverEventCollector.Run()
	go userEventCollector.Run()

	go runServer(serverEventCollector, userEventCollector)

	gracefully.WaitShutdown()
	if !gracefully.GlobalError().IsEmpty() {
		log.Println(gracefully.GlobalError().MaybeUnwrap().Error())
	}
	log.Println("app is done...")
}

func runServer(serverEventCollector, userEventCollector *eventBatcher) {
	http.HandleFunc("/user/event", func(w http.ResponseWriter, r *http.Request) {
		if gracefully.GetStatus() != gracefully.StatusRunning {
			w.Write([]byte("service is shutting down, try be later."))
			return
		}

		events, err := toStringArr(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		userEventCollector.Store(events)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/server/event", func(w http.ResponseWriter, r *http.Request) {
		if gracefully.GetStatus() != gracefully.StatusRunning {
			w.Write([]byte("service is shutting down, try be later."))
			return
		}

		events, err := toStringArr(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		serverEventCollector.Store(events)
		w.Write([]byte("OK"))
	})

	if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
		log.Fatal(err)
	}
}

func toStringArr(r *http.Request) ([]string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	lines := bytes.Split(body, []byte("\n"))

	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = string(line)
	}

	return result, nil
}
