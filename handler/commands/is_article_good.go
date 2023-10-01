package commands

import "github.com/slack-go/slack"

// handleIsArticleGood will trigger a Yes or No question to the initializer
func HandleIsArticleGood(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	// Create the checkbox element
	checkbox := slack.NewCheckboxGroupsBlockElement("answer",
		slack.NewOptionBlockObject("yes", &slack.TextBlockObject{Text: "Yes", Type: slack.MarkdownType}, &slack.TextBlockObject{Text: "Did you enjoy it?", Type: slack.MarkdownType}),
		slack.NewOptionBlockObject("no", &slack.TextBlockObject{Text: "No", Type: slack.MarkdownType}, &slack.TextBlockObject{Text: "Did you Dislike it?", Type: slack.MarkdownType}),
	)	
	// Create the Accessory that will be included in the Block and add the checkbox to it
	accessory := slack.NewAccessory(checkbox)
	// Add blocks to the attachment
	attachment.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			// Create a new section block element and add some text and the accessory to it
			slack.NewSectionBlock(
				&slack.TextBlockObject{
					Type: slack.MarkdownType,
					Text: "Did you think this article was helpful?",
				},
				nil,
				accessory,
			),
		},
	}

	attachment.Color = "#4af030"
	return attachment, nil
}