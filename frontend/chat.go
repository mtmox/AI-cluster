
package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"
	"strconv"
)

var selectedThreadID int = -1
var currentThreadIndex int = 0
var threadCounter int = 1
var sendToAllThreads bool = false

func createChatTab() fyne.CanvasObject {
	if len(conversations) == 0 {
		conversations = append(conversations, Conversation{
			ID:            "1",
			Threads:       []Thread{},
			ThreadCounter: 0,
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
		if selectedConversation != nil && selectedConversation.ID == conversations[id].ID {
			conversationList.Unselect(id)
			selectedConversation = nil
			updateThreadsList(threadsList, nil)
			chatOutput.SetText("")
		} else {
			selectedConversation = &conversations[id]
			updateThreadsList(threadsList, selectedConversation.Threads)
			currentThreadIndex = 0
			if len(selectedConversation.Threads) > 0 {
				updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
			} else {
				chatOutput.SetText("")
			}
		}
	}

	threadsList = widget.NewList(
		func() int {
			if selectedConversation == nil {
				return 0
			}
			return len(selectedConversation.Threads)
		},
		func() fyne.CanvasObject { return widget.NewLabel("Thread") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if selectedConversation != nil && int(id) < len(selectedConversation.Threads) {
				item.(*widget.Label).SetText(strconv.Itoa(selectedConversation.Threads[id].ID))
			}
		},
	)
	threadsList.OnSelected = func(id widget.ListItemID) {
		if selectedConversation == nil {
			return
		}
		if selectedThreadID == int(id) {
			threadsList.Unselect(id)
			selectedThreadID = -1
			chatOutput.SetText("")
		} else {
			selectedThreadID = int(id)
			currentThreadIndex = int(id)
			if id < len(selectedConversation.Threads) {
				updateChatOutput(chatOutput, selectedConversation.Threads[id].Messages)
			}
		}
	}

	killToggle := widget.NewRadioGroup([]string{"Conversation", "Thread"}, func(value string) {})
	killToggle.SetSelected("Conversation")

	killButton := widget.NewButton("Kill", func() {
		if killToggle.Selected == "Conversation" {
			if selectedConversation != nil {
				for i, conv := range conversations {
					if conv.ID == selectedConversation.ID {
						conversations = append(conversations[:i], conversations[i+1:]...)
						break
					}
				}
				updateConversationList(conversationList, conversations)
				selectedConversation = nil
				updateThreadsList(threadsList, nil)
				chatOutput.SetText("")
			}
		} else if killToggle.Selected == "Thread" {
			if selectedConversation != nil && selectedThreadID != -1 {
				selectedConversation.Threads = append(selectedConversation.Threads[:selectedThreadID], selectedConversation.Threads[selectedThreadID+1:]...)
				updateThreadsList(threadsList, selectedConversation.Threads)
				selectedThreadID = -1
				if len(selectedConversation.Threads) > 0 {
					currentThreadIndex = 0
					updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
				} else {
					chatOutput.SetText("")
				}
			}
		}
	})

	newConversationButton := widget.NewButton("New Conversation", func() {
		newConversation := Conversation{
			ID:            strconv.Itoa(len(conversations) + 1),
			Threads:       []Thread{},
			ThreadCounter: 0,
		}
		conversations = append(conversations, newConversation)
		updateConversationList(conversationList, conversations)
		selectedConversation = &conversations[len(conversations)-1]
		updateThreadsList(threadsList, selectedConversation.Threads)
		chatOutput.SetText("")
		currentThreadIndex = 0
	})

	threadCounterEntry := widget.NewEntry()
	threadCounterEntry.SetText("1")

	newThreadButton := widget.NewButton("New Thread", func() {
		if selectedConversation != nil {
			count, err := strconv.Atoi(threadCounterEntry.Text)
			if err != nil {
				count = 1
			}
			for i := 0; i < count; i++ {
				selectedConversation.ThreadCounter++
				newThread := Thread{
					ID:       selectedConversation.ThreadCounter,
					Messages: []Message{},
				}
				selectedConversation.Threads = append(selectedConversation.Threads, newThread)
			}
			updateThreadsList(threadsList, selectedConversation.Threads)
			chatOutput.SetText("")
			currentThreadIndex = len(selectedConversation.Threads) - 1
		}
	})

	copyThreadButton := widget.NewButton("Copy Thread", func() {
		if selectedConversation == nil || selectedThreadID == -1 {
			return
		}
		if selectedThreadID >= len(selectedConversation.Threads) {
			return
		}
		count, err := strconv.Atoi(threadCounterEntry.Text)
		if err != nil {
			count = 1
		}
		for i := 0; i < count; i++ {
			selectedConversation.ThreadCounter++
			copiedThread := selectedConversation.Threads[selectedThreadID]
			newThread := Thread{
				ID:       selectedConversation.ThreadCounter,
				Messages: make([]Message, len(copiedThread.Messages)),
			}
			copy(newThread.Messages, copiedThread.Messages)
			selectedConversation.Threads = append(selectedConversation.Threads, newThread)
		}
		updateThreadsList(threadsList, selectedConversation.Threads)
		currentThreadIndex = len(selectedConversation.Threads) - 1
	})

	settingsContainer := container.NewVBox(
		container.NewHBox(killToggle),
		killButton,
		newConversationButton,
		container.NewHBox(widget.NewLabel("Thread Count:"), threadCounterEntry),
		newThreadButton,
		copyThreadButton,
	)

	chatOutput = widget.NewMultiLineEntry()
	chatOutput.Disable()

	leftArrow := widget.NewButton("<", func() {
		if selectedConversation != nil && len(selectedConversation.Threads) > 0 {
			currentThreadIndex--
			if currentThreadIndex < 0 {
				currentThreadIndex = len(selectedConversation.Threads) - 1
			}
			updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
			threadsList.Select(currentThreadIndex)
		}
	})
	rightArrow := widget.NewButton(">", func() {
		if selectedConversation != nil && len(selectedConversation.Threads) > 0 {
			currentThreadIndex++
			if currentThreadIndex >= len(selectedConversation.Threads) {
				currentThreadIndex = 0
			}
			updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
			threadsList.Select(currentThreadIndex)
		}
	})

	sendToAllThreadsToggle := widget.NewCheck("Send to All Threads", func(value bool) {
		sendToAllThreads = value
	})

	sendButton := widget.NewButton("Send", func() {
		sendMessage(messageBar, conversationList, threadsList)
	})

	messageBar.OnSubmitted = func(s string) {
		sendMessage(messageBar, conversationList, threadsList)
	}

	topContainer := container.NewBorder(nil, nil, nil, container.NewHBox(sendToAllThreadsToggle, sendButton), messageBar)
	
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

func sendMessage(messageBar *widget.Entry, conversationList, threadsList *widget.List) {
	message := messageBar.Text
	if message != "" {
		if selectedConversation == nil {
			newConversation := Conversation{
				ID:            strconv.Itoa(len(conversations) + 1),
				Threads:       []Thread{},
				ThreadCounter: 0,
			}
			conversations = append(conversations, newConversation)
			selectedConversation = &conversations[len(conversations)-1]
			updateConversationList(conversationList, conversations)
		}

		if len(selectedConversation.Threads) == 0 {
			selectedConversation.ThreadCounter++
			newThread := Thread{
				ID:       selectedConversation.ThreadCounter,
				Messages: []Message{},
			}
			selectedConversation.Threads = append(selectedConversation.Threads, newThread)
			updateThreadsList(threadsList, selectedConversation.Threads)
			currentThreadIndex = 0
		}

		newMessage := Message{Sender: "User", Content: message}
		if sendToAllThreads {
			for i := range selectedConversation.Threads {
				selectedConversation.Threads[i].Messages = append(selectedConversation.Threads[i].Messages, newMessage)
			}
			updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
		} else {
			if currentThreadIndex < len(selectedConversation.Threads) {
				selectedConversation.Threads[currentThreadIndex].Messages = append(selectedConversation.Threads[currentThreadIndex].Messages, newMessage)
				updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
			}
		}
		messageBar.SetText("")
	}
}

// Setup a receiveMessage function which creates a consumer and listens to the NATS
// queue to consume messages and then populate them in the right spots.
