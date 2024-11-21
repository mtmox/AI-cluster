
package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Conversation struct {
	ID      string
	Threads []Thread
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

func updateThreadsList(grid *fyne.Container, threads []Thread) {
	grid.RemoveAll()
	for _, thread := range threads {
		grid.Add(createThreadBox(thread.ID))
	}
	grid.Refresh()
}

func updateChatOutput(output *widget.Entry, messages []Message) {
	var content string
	for _, msg := range messages {
		content += msg.Sender + ": " + msg.Content + "\n"
	}
	output.SetText(content)
}
