
package frontend

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/layout"

	"github.com/nats-io/nats.go"
	
	"github.com/mtmox/AI-cluster/constants"
	"github.com/mtmox/AI-cluster/node"
)

var selectedThreadID int = -1
var currentThreadIndex int = 0
var threadCounter int = 1
var sendToAllThreads bool = false
var modelSelector *widget.Select
var promptSelector *widget.Select

func createChatTab(js nats.JetStreamContext) fyne.CanvasObject {
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

	modelSelector = widget.NewSelect([]string{}, func(selected string) {
		// Handle model selection here if needed
		if selected != "" {
			// You can use the selected model name here
		}
	})
	
	// Create the prompt selector
	var promptNames []string
	for name := range constants.SystemPrompts {
		promptNames = append(promptNames, name)
	}
	promptSelector = widget.NewSelect(promptNames, func(selected string) {
		// Handle prompt selection here if needed
		if selected != "" {
			// You can use the selected prompt here
			// constants.SystemPrompts[selected] will give you the prompt content
		}
	})
	
	// Function to update the model selector options
	updateModelSelector := func() {
		var names []string
		for name := range modelNames {
			names = append(names, name)
		}
		modelSelector.Options = names
		modelSelector.Refresh()
	}
	
	// Initial update of the selector
	updateModelSelector()

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
		sendMessage(js, messageBar, conversationList, threadsList)
	})

	messageBar.OnSubmitted = func(s string) {
		sendMessage(js, messageBar, conversationList, threadsList)
	}

	// Create a container for both selectors
	selectorsContainer := container.NewHBox(
		container.NewHBox(widget.NewLabel("Model:"), modelSelector),
		container.NewHBox(widget.NewLabel("Prompt:"), promptSelector),
	)

	topContainer := container.NewVBox(
		selectorsContainer,
		container.NewBorder(
			nil, 
			nil, 
			nil,
			container.NewHBox(sendToAllThreadsToggle, sendButton), 
			messageBar,
		),
	)
	
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

func sendMessage(js nats.JetStreamContext, messageBar *widget.Entry, conversationList, threadsList *widget.List) {
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

        newMessage := Message{Role: "User", Content: message}
        if sendToAllThreads {
            for i := range selectedConversation.Threads {
                selectedConversation.Threads[i].Messages = append(selectedConversation.Threads[i].Messages, newMessage)
                
                // Send message to NATS for each thread
                if modelSelector.Selected != "" && promptSelector.Selected != "" {
                    natsMsg, err := formatMessageForNATS(
                        selectedConversation,
                        selectedConversation.Threads[i],
                        modelSelector.Selected,
                        promptSelector.Selected,
                    )
                    if err != nil {
                        node.HandleError(err, node.ERROR, "Error formatting message for NATS")
                        continue
                    }
                    
                    if js != nil {
                        err = sendMessageToNATS(js, natsMsg)
                        if err != nil {
                            node.HandleError(err, node.ERROR, "Error sending message to NATS")
                        } else {
                            node.HandleError(nil, node.SUCCESS, "Message successfully sent to NATS")
                        }
                    }
                }
            }
            updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
        } else {
            if currentThreadIndex < len(selectedConversation.Threads) {
                selectedConversation.Threads[currentThreadIndex].Messages = append(
                    selectedConversation.Threads[currentThreadIndex].Messages,
                    newMessage,
                )
                
                // Send message to NATS for current thread
                if modelSelector.Selected != "" && promptSelector.Selected != "" {
                    natsMsg, err := formatMessageForNATS(
                        selectedConversation,
                        selectedConversation.Threads[currentThreadIndex],
                        modelSelector.Selected,
                        promptSelector.Selected,
                    )
                    if err != nil {
                        node.HandleError(err, node.ERROR, "Error formatting message for NATS")
                    } else if js != nil {
                        err = sendMessageToNATS(js, natsMsg)
                        if err != nil {
                            node.HandleError(err, node.ERROR, "Error sending message to NATS")
                        } else {
                            node.HandleError(nil, node.SUCCESS, "Message successfully sent to NATS")
                        }
                    }
                }
                
                updateChatOutput(chatOutput, selectedConversation.Threads[currentThreadIndex].Messages)
            }
        }
        messageBar.SetText("")
    }
}

