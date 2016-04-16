package main

import (
	"fmt"
	"net/http"

	"github.com/paked/configure"
	"github.com/paked/messenger"
)

type MessageState int

var (
	conf        = configure.New()
	verifyToken = conf.String("verify-token", "mad-skrilla", "The token used to verify facebook")
	verify      = conf.Bool("should-verify", false, "Whether or not the app should verify itself")
	pageToken   = conf.String("page-token", "not skrilla", "The token that is used to verify the page on facebook")

	states map[int64]MessageState

	client *messenger.Messenger
)

const (
	NoAction MessageState = iota
	MakingMeme
)

func main() {
	conf.Use(configure.NewFlag())
	conf.Use(configure.NewEnvironment())
	conf.Use(configure.NewJSONFromFile("config.json"))

	conf.Parse()

	client = messenger.New(messenger.Options{
		Verify:      *verify,
		VerifyToken: *verifyToken,
		Token:       *pageToken,
	})

	client.HandleMessage(messages)

	fmt.Println("Serving messenger bot on localhost:8080")

	http.ListenAndServe("localhost:8080", client.Handler())
}

func messages(m messenger.Message, r *messenger.Response) {
	from, err := client.ProfileByID(m.Sender.ID)
	if err != nil {
		fmt.Println("error getting profile:", err)
		return
	}

	state := messageState(m.Sender)

	switch state {
	case NoAction:
		r.Text(fmt.Sprintf("Greetings, %v? You're here to make a meme?", from.FirstName))
		r.Text("If so, you are in just the right place.")
		r.Text("All you need to do is send me a picture and a line of text to put on that picture!")
	}
}

func messageState(s messenger.Sender) MessageState {
	return states[s.ID]
}
