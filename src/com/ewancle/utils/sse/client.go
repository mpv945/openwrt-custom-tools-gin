package sse

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	ID      string
	UserID  string
	Topics  map[string]bool
	Writer  http.ResponseWriter
	Flusher http.Flusher
	Ch      chan Event
	Done    chan struct{}
}

func NewClient(id string, user string, w http.ResponseWriter) *Client {

	flusher, _ := w.(http.Flusher)

	return &Client{
		ID:      id,
		UserID:  user,
		Topics:  map[string]bool{},
		Writer:  w,
		Flusher: flusher,
		Ch:      make(chan Event, 128),
		Done:    make(chan struct{}),
	}
}

func (c *Client) Send(e Event) error {

	if _, err := fmt.Fprintf(c.Writer, "id: %s\n", e.ID); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(c.Writer, "event: %s\n", e.Event); err != nil {
		return err
	}

	data := strings.ReplaceAll(e.Data, "\n", " ")
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
		return err
	}

	c.Flusher.Flush()

	return nil
}

func (c *Client) Loop(r *http.Request) {

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {

		select {

		case e := <-c.Ch:

			if err := c.Send(e); err != nil {
				close(c.Done)
				return
			}

		case <-heartbeat.C:

			_, err := fmt.Fprintf(c.Writer, ": ping\n\n")
			if err != nil {
				close(c.Done)
				return
			}

			c.Flusher.Flush()

		case <-r.Context().Done():

			close(c.Done)
			return
		}
	}
}
