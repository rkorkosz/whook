package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	addr := flag.String("addr", ":8000", "Listen address")
	flag.Parse()
	eb := &PubSub{
		subs: make(map[string]chan []byte),
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	for _, topic := range flag.Args() {
		go handleTopic(ctx, eb, topic)
	}
	srv := &http.Server{
		Addr:    *addr,
		Handler: http.HandlerFunc(webhook(eb)),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	srv.Shutdown(context.Background())
}

func handleTopic(ctx context.Context, eb *PubSub, topic string) {
	ch := make(chan []byte)
	eb.Subscribe(topic, ch)
	for {
		select {
		case <-ctx.Done():
		case d := <-ch:
			os.Stdout.Write(d)
			os.Stdout.Write([]byte("\n"))
		}
	}

}

func webhook(eb *PubSub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Path[1:]
		defer r.Body.Close()
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		eb.Publish(topic, data)
		w.WriteHeader(204)
	}
}

type PubSub struct {
	rm   sync.RWMutex
	subs map[string]chan []byte
}

func (eb *PubSub) Subscribe(topic string, ch chan []byte) {
	eb.rm.Lock()
	defer eb.rm.Unlock()
	eb.subs[topic] = ch
}

func (eb *PubSub) Publish(topic string, data []byte) {
	eb.rm.RLock()
	defer eb.rm.RUnlock()
	if ch, found := eb.subs[topic]; found {
		ch <- data
	}
}
