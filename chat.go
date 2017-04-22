package upframy

import (
	"log"
	"net/http"

	"github.com/nlopes/slack"
	"golang.org/x/net/websocket"
)

const chatChannel = "C50SNCVSR"

var chatConnections = make(map[string]*websocket.Conn)

func startChat() {
	http.Handle("/", websocket.Handler(chatHandler))

	if err := http.ListenAndServe(":4256", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func chatHandler(ws *websocket.Conn) {
	var err error
	var ts string

	thread := make(chan string)

	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			break
		}

		if ts != "" {
			msg := rtm.NewOutgoingMessage(reply, chatChannel)
			msg.ThreadTimestamp = ts

			rtm.SendMessage(msg)
			continue
		}

		msg := rtm.NewOutgoingMessage(reply, chatChannel)

		pendingTimestamps = append(pendingTimestamps, &pendingTimestamp{
			ID:   msg.ID,
			Chan: thread,
		})

		rtm.SendMessage(msg)

		ts = <-thread
		close(thread)

		chatConnections[ts] = ws
	}

	delete(chatConnections, ts)

	msg := rtm.NewOutgoingMessage("*My fan just left.*", chatChannel)
	msg.ThreadTimestamp = ts
	rtm.SendMessage(msg)
}

func chatSlackHandler(ev *slack.MessageEvent) bool {
	if ev.ThreadTimestamp == "" {
		return false
	}

	conn, ok := chatConnections[ev.ThreadTimestamp]
	if !ok {
		return false
	}

	err := websocket.Message.Send(conn, ev.Text)
	if err != nil {
		logger.Printf("Error on Chat Slack Handler: %v", err)

		msg := rtm.NewOutgoingMessage("*There was a problem delivering the message.*", chatChannel)
		msg.ThreadTimestamp = ev.ThreadTimestamp
		rtm.SendMessage(msg)
		return false
	}

	return true
}