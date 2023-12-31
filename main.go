package main

import (
	"context"
	"log"
	"os"

	"github.com/georgecpp/mimir/handler"
	"github.com/georgecpp/mimir/misc"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {	
	config, err := misc.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}
	
	// Create a new client to slack by giving token
	// Set debug to true while developing
	client := slack.New(config.SlackAuthToken, slack.OptionDebug(true), slack.OptionAppLevelToken(config.SlackAppToken))

	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())

	// make this cancel called properly in a real program, graceful shutdown etc
	defer cancel()

	go listen(ctx,client,socketClient)
	go misc.RunSpotifyAuthServer();

	socketClient.Run()
}


func listen(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
	// Create a for loop that selects either the context cancellation or the events incoming
	for {
		select {
		// case context cancel is called exit the goroutine
		case <-ctx.Done():
			log.Println("Shutting down socketmode listener")
			return
		case event := <-socketClient.Events:
			// We have a new Events, let's type switch the event
			// Add more use cases here if you want to listen to other events.
			switch event.Type {
			// handle EventAPI events
			case socketmode.EventTypeEventsAPI:
				// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
				eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
					continue
				}
				// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
				payload, err := handler.HandleEventMessage(eventsAPIEvent, client)
				if err != nil {
					// TODO: replace with actual error handling
					log.Fatal(err)
				}
				// Don't forget to acknowledge the request
				// The payload is the response
				socketClient.Ack(*event.Request, payload)

			// handle Slash Commmands Events
			case socketmode.EventTypeSlashCommand:
				// Just like before, type cast to the correct event type, this time a SlashEvent
				command, ok := event.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
					continue
				}
		
				// handleSlashCommand will take care of the command
				payload, err := handler.HandleSlashCommand(command, client)
				if err != nil {
					log.Fatal(err)
				}
				// Don't forget to acknowledge the request
				// The payload is the response
				socketClient.Ack(*event.Request, payload)

			// handle Inreraction Events				
			case socketmode.EventTypeInteractive:
				interaction, ok := event.Data.(slack.InteractionCallback)
				if !ok {
					log.Printf("Could not type cast the message to a Interaction callback: %v\n", interaction)
					continue				
				}

				payload, err := handler.HandleInteractionEvent(interaction, client)
				if err != nil {
					log.Fatal(err)
				}
				socketClient.Ack(*event.Request, payload)
			}
		}
	}
}