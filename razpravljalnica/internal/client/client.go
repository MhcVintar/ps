package client

import(
	"fmt"
	"github.com/rivo/tview"
)

func Bootstrap(){
	
	fmt.Println("client")
	client := tview.NewApplication()
	clientView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			client.Draw()
		})
	clientView.SetBorder(true).SetTitle(" Vremenska postaja ")
}