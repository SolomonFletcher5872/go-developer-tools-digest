package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Message struct {
	MessageID string          `json:"message_id"`
	Payload   json.RawMessage `json:"payload"`
}
type consumeResult struct {
	Messages []Message `json:"messages"`
}

func runDigest(c *Infrai, to string) error {
	var got consumeResult
	if err := c.Consume(100, 60, &got); err != nil {
		return err
	}
	events := make([]Event, 0, len(got.Messages))
	for _, m := range got.Messages {
		var e Event
		if err := json.Unmarshal(m.Payload, &e); err != nil {
			return err
		}
		events = append(events, e)
	}
	if err := c.SendEmail(to, "Developer tools digest", Digest(events)); err != nil {
		return err
	}
	for _, m := range got.Messages {
		if err := c.Ack(m.MessageID); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	c, err := NewInfrai()
	if err != nil {
		log.Fatal(err)
	}
	if err = c.Schedule("https://example.com/digest"); err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/digest", func(w http.ResponseWriter, r *http.Request) {
		if err := runDigest(c, "devtools@example.com"); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintln(w, "digest sent")
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
