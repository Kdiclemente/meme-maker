package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"net/http"

	"github.com/fogleman/gg"
	"github.com/paked/configure"
	"github.com/paked/messenger"
)

type MessageState int

var (
	conf        = configure.New()
	verifyToken = conf.String("verify-token", "mad-skrilla", "The token used to verify facebook")
	verify      = conf.Bool("should-verify", false, "Whether or not the app should verify itself")
	pageToken   = conf.String("page-token", "not skrilla", "The token that is used to verify the page on facebook")
	font        = conf.String("font", "fonts/Economica-Bold.ttf", "The font you want the meme maker to use")

	states map[int64]MessageState
	memes  map[int64]*Meme

	client *messenger.Messenger
)

const (
	NoAction MessageState = iota
	MakingMeme

	fontSize = 20
)

func init() {
	conf.Use(configure.NewFlag())
	conf.Use(configure.NewEnvironment())
	conf.Use(configure.NewJSONFromFile("config.json"))
}

func main() {
	conf.Parse()

	memes = make(map[int64]*Meme)
	states = make(map[int64]MessageState)

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

	fmt.Println(m.Sender.ID)

	state := messageState(m.Sender)

	switch state {
	case NoAction:
		r.Text(fmt.Sprintf("Greetings, %v? You're here to make a meme?", from.FirstName))
		r.Text("If so, you are in just the right place.")
		r.Text("All you need to do is send me a picture and a line of text to put on that picture!")

		setState(m.Sender, MakingMeme)
	case MakingMeme:
		meme := messageMeme(m.Sender)

		if len(m.Attachments) > 0 {
			a := m.Attachments[0]
			if a.Type != "image" {
				r.Text("Sorry to be a sad pepe. Unfortunately you're going to need to send an image")
			}

			meme.ImageURL = a.Payload.URL
		}

		if m.Text != "" {
			meme.Text = m.Text
		}

		if meme.Ready() {
			err = r.Image(meme.Make())
			if err != nil {
				fmt.Println("error encoding image:", err)
				return
			}

			fmt.Println("Done!")
		}
	}
}

func messageState(s messenger.Sender) MessageState {
	return states[s.ID]
}

func setState(s messenger.Sender, state MessageState) {
	states[s.ID] = state
}

func messageMeme(s messenger.Sender) *Meme {
	meme := memes[s.ID]
	if meme == nil {
		meme = &Meme{}
		memes[s.ID] = meme
	}

	return meme
}

type Meme struct {
	ImageURL string
	Text     string
}

func (m Meme) Ready() bool {
	return m.ImageURL != "" && m.Text != ""
}

func (m Meme) Make() image.Image {
	res, err := http.Get(m.ImageURL)
	if err != nil {
		fmt.Println("error downloading image:", err)
		return nil
	}

	defer res.Body.Close()

	background, _, err := image.Decode(res.Body)
	if err != nil {
		fmt.Println("error decoding image:", err)
	}

	r := background.Bounds()
	w := r.Dx()
	h := r.Dy()

	final := gg.NewContext(w, h)
	final.DrawImage(background, 0, 0)
	final.LoadFontFace(*font, fontSize)

	final.SetHexColor("#000")
	strokeSize := 6
	for dy := -strokeSize; dy <= strokeSize; dy++ {
		for dx := -strokeSize; dx <= strokeSize; dx++ {
			// give it rounded corners
			if dx*dx+dy*dy >= strokeSize*strokeSize {
				continue
			}

			x := float64(w/2 + dx)
			y := float64(h - fontSize + dy)
			final.DrawStringAnchored(m.Text, x, y, 0.5, 0.5)
		}
	}

	final.SetHexColor("#FFF")
	final.DrawStringAnchored(m.Text, float64(w)/2, float64(h)-fontSize, 0.5, 0.5)

	return final.Image()
}
