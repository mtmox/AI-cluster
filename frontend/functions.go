
package frontend

import (
	"strconv"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Conversation struct {
	ID           string
	Threads      []Thread
	ThreadCounter int
}

type Thread struct {
	ID       int
	Messages []Message
}

type Message struct {
	Sender  string
	Content string
}

func updateConversationList(list *widget.List, conversations []Conversation) {
	list.Length = func() int { return len(conversations) }
	list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		item.(*widget.Label).SetText(conversations[id].ID)
	}
	list.Refresh()
}

func updateThreadsList(list *widget.List, threads []Thread) {
	list.Length = func() int { return len(threads) }
	list.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		item.(*widget.Label).SetText(strconv.Itoa(threads[id].ID))
	}
	list.Refresh()
}

func updateChatOutput(output *widget.Entry, messages []Message) {
	var content string
	for _, msg := range messages {
		content += msg.Sender + ": " + msg.Content + "\n"
	}
	output.SetText(content)
}

func createThreadBox(number int) fyne.CanvasObject {
	return widget.NewLabel(strconv.Itoa(number))
}

func appendMessage(output *widget.Entry, sender, content string) {
	currentText := output.Text
	newMessage := sender + ": " + content + "\n"
	output.SetText(currentText + newMessage)
}
