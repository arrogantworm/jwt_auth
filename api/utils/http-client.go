package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Notifications struct

type NewIpNotification struct {
	UserID int    `json:"userId"`
	NewIp  string `json:"newIp"`
	OldIp  string `json:"oldIp"`
}

// HTTP Client

type Client struct {
	client *http.Client
}

func NewClient(timeout time.Duration) (*Client, error) {
	if timeout == 0 {
		return nil, errors.New("timeout can't be 0")
	}

	return &Client{
		client: &http.Client{
			Timeout: timeout,
			Transport: &loggingRoundTripper{
				logger: os.Stdout,
				next:   http.DefaultTransport,
			},
		},
	}, nil
}

// RoundTripper

type loggingRoundTripper struct {
	logger io.Writer
	next   http.RoundTripper
}

func (l loggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	log.Printf("[NEW IP NOTIFICATION] sent to %s\n", r.URL)
	return l.next.RoundTrip(r)
}

func ChangedIPRequest(reqStruct NewIpNotification) {

	c, err := NewClient(time.Second * 30)
	if err != nil {
		log.Printf("error sending notification: %v", err)
		return
	}

	url := viper.GetString("notifications.newIpURL")

	if url == "" {
		log.Println("error sending notification: url not defined")
		return
	}

	req, err := json.Marshal(reqStruct)
	if err != nil {
		log.Printf("error sending notification: %v", err)
		return
	}

	res, err := c.client.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		log.Printf("error sending notification: %v", err)
		return
	}

	log.Printf("[NEW IP NOTIFICATION] %s\n", res.Status)
}
