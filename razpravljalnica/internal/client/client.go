package client
//https://github.com/BekBrace/inv-tui-go/blob/main/main.go
import(
	"fmt"
	"github.com/rivo/tview"
	//"github.com/gdamore/tcell/v2"

)

func Bootstrap() error{
	
	fmt.Println("Client bootstrap")
	app := tview.NewApplication()

	chooseTopic := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		})
	chooseTopic.SetBorder(true).SetTitle(" Choose topic ").SetTitleAlign(tview.AlignLeft)
	
	sidebar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chooseTopic, 0, 1, true).
		AddItem(
			tview.NewForm().
				AddButton("Exit", func() {
					app.Stop()
				}), 3, 0, true,
			)

	
	msgs := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText("YIPEEE")
	
	// Lower part
	textField := tview.NewInputField().
			SetLabel("TODO").
			SetFieldWidth(20)
	
	sendText := tview.NewForm().
			//SetDirection(tview.FlexRow).
			AddFormItem(textField).
			AddButton("Send", func() {
				//TODO: Send current text
				app.Stop()
			})




	main := tview.NewFlex().
		AddItem(msgs, 0, 1, false).
		AddItem(sendText, 3, 0, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(sidebar, 20, 1, true).
		AddItem(main, 0, 0, true)

	
	return app.SetRoot(layout, true).Run()
}

func changeFooterToStop(footer *tview.TextView){
	
}