package handler

import (
	"errors"
	"log"

	"github.com/slack-go/slack/slackevents"
)

// HandleEventMessage will take an event and handle it properly based on the type of event
func HandleEventMessage(event slackevents.EventsAPIEvent) error {
	switch event.Type {
		// first we check if this is an CallbackEvent
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		// Yet another type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			log.Println(ev)
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}