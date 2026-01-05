package client
//https://github.com/BekBrace/inv-tui-go/blob/main/main.go
import(
	"fmt"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"

)
const(
	sidebarFocus = 0
	mainFocus = 1
)
func Bootstrap() error{
	//sidebarOrMainFocus := sidebarFocus
	fmt.Println("Client bootstrap")
	app := tview.NewApplication()

	chooseTopic := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText("YIPEEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE")

	chooseTopic.SetBorder(true).SetTitle(" Choose topic ").SetTitleAlign(tview.AlignLeft)
	
	sidebar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chooseTopic, 0, 1, false).
		AddItem(
			tview.NewForm().
				AddButton("Exit", func() {
					app.Stop()
				}), 3, 0, true,
			)

	
	msgs := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText("YIPEEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE\nYIPEE")
	
	// Lower part of main section, meant to represent input
	textField := tview.NewTextArea().
		SetLabel("textField")
		//SetFieldWidth(20)
	sendText := tview.NewForm().
			//SetDirection(tview.FlexRow).
			AddFormItem(textField).
			AddButton("Send", func() {
				//TODO: Send current text
				app.Stop()
			})




	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(msgs, 0, 1, false).
		AddItem(sendText, 0, 1, true)
	
	layout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(sidebar, 0, 1, false).
		AddItem(main, 0, 1, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
			case tcell.KeyLeft:
				app.SetFocus(sidebar)
			case tcell.KeyRight:
				app.SetFocus(main)
		}
		return event
	})
	
	return app.SetRoot(layout, true).Run()
}

func changeFooterToStop(footer *tview.TextView){
	
}