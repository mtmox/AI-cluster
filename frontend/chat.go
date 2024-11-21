
package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
	"strconv"
)

func createChatTab() fyne.CanvasObject {
	if len(conversations) == 0 {
		conversations = append(conversations, Conversation{
			ID: "1",
			Threads: []Thread{
				{
					ID:       1,
					Messages: []Message{},
				},
			},
		})
	}

	if selectedConversation == nil {
		selectedConversation = &conversations[0]
	}

	messageBar := widget.NewEntry()
	messageBar.SetPlaceHolder("Type your message here...")

	conversationList := widget.NewList(
		func() int { return len(conversations) },
		func() fyne.CanvasObject { return widget.NewLabel("Conversation") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(conversations[id].ID)
		},
	)
	conversationList.OnSelected = func(id widget.ListItemID) {
		selectedConversation = &conversations[id]
		updateThreadsList(threadsList, selectedConversation.Threads)
		if len(selectedConversation.Threads) > 0 {
			updateChatOutput(chatOutput, selectedConversation.Threads[0].Messages)
		}
	}

	threadsList = widget.NewList(
		func() int { return len(selectedConversation.Threads) },
		func() fyne.CanvasObject { return widget.NewLabel("Thread") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(strconv.Itoa(selectedConversation.Threads[id].ID))
		},
	)

	killButton := widget.NewButton("Kill", func() {})
	newConversationButton := widget.NewButton("New Conversation", func() {
		newConversation := Conversation{
			ID: strconv.Itoa(len(conversations) + 1),
			Threads: []Thread{
				{
					ID:       1,
					Messages: []Message{},
				},
			},
		}
		conversations = append(conversations, newConversation)
		conversationList.Refresh()
	})
	settingsContainer := container.NewVBox(killButton, newConversationButton)

	chatOutput = widget.NewMultiLineEntry()
	chatOutput.Disable()

	leftArrow := widget.NewButton("<", func() {})
	rightArrow := widget.NewButton(">", func() {})

	sendButton := widget.NewButton("Send", func() {
		sendMessage(messageBar)
	})

	messageBar.OnSubmitted = func(s string) {
		sendMessage(messageBar)
	}

	topContainer := container.NewBorder(nil, nil, nil, sendButton, messageBar)
	
	conversationsScroll := container.NewVScroll(conversationList)
	threadsScroll := container.NewVScroll(threadsList)
	
	leftSide := container.NewHSplit(
		container.NewBorder(widget.NewLabel("Conversations"), nil, nil, nil, conversationsScroll),
		container.NewBorder(widget.NewLabel("Threads"), nil, nil, nil, threadsScroll),
	)
	leftSide.SetOffset(0.7)
	
	rightSide := settingsContainer
	middleContent := container.NewBorder(nil, nil, leftArrow, rightArrow, container.NewScroll(chatOutput))
	
	content := container.New(layout.NewBorderLayout(topContainer, nil, leftSide, rightSide),
		topContainer,
		leftSide,
		rightSide,
		middleContent,
	)

	return content
}

func sendMessage(messageBar *widget.Entry) {
	message := messageBar.Text
	if message != "" {
		if selectedConversation == nil {
			newConversation := Conversation{
				ID: strconv.Itoa(len(conversations) + 1),
				Threads: []Thread{
					{
						ID:       1,
						Messages: []Message{},
					},
				},
			}
			conversations = append(conversations, newConversation)
			selectedConversation = &conversations[len(conversations)-1]
		}

		newMessage := Message{Sender: "User", Content: message}
		selectedConversation.Threads[0].Messages = append(selectedConversation.Threads[0].Messages, newMessage)
		updateChatOutput(chatOutput, selectedConversation.Threads[0].Messages)
		messageBar.SetText("")
	}
}