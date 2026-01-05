package client
//https://github.com/BekBrace/inv-tui-go/blob/main/main.go
import(
	"fmt"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"log"
)
const(
	sidebarFocus = 0
	mainFocus = 1
)

var(
	app *tview.Application
	chooseTopic *tview.TextView
	exitButton *tview.Button
	sidebar *tview.Flex
	msgs *tview.TextView
	textField *tview.TextArea
	sendButton *tview.Button
)
func SendMessageToServer(sporocilo string){
	fmt.Println(sporocilo)
}





func RunGUI() error{
	sidebarOrMainFocus := mainFocus
	fmt.Println("Client bootstrap")
	app := tview.NewApplication()

	chooseTopic := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText("1YIPEEE\n2YIPEE\n3YIPEE\n4YIPEE\n5YIPEE\n6YIPEE\n7YIPEE\n8YIPEE\n9YIPEE\n10YIPEE\n11YIPEE\n12YIPEE\n13YIPEE\n14YIPEE\n15YIPEE\n16YIPEE\n17YIPEE\n18YIPEE\n19YIPEE\n20YIPEE\n21YIPEE\n22YIPEE\n23YIPEE\n24YIPEE\n25YIPEE\n26YIPEE\n27YIPEE\n28YIPEE\n29YIPEE\n30YIPEE\n31YIPEE\n32YIPEE\n33YIPEE\n34YIPEE\n35YIPEE\n36YIPEE\n37YIPEE\n38YIPEE\n39YIPEE\n40YIPEE\n41YIPEE\n42YIPEE\n43YIPEE\n44YIPEE\n45YIPEE\n46YIPEE\n47YIPEE\n48YIPEE\n49YIPEE")

	chooseTopic.SetBorder(true).SetTitle(" Choose topic ").SetTitleAlign(tview.AlignLeft)
	exitButton := tview.NewButton("Exit").
					SetSelectedFunc(func() {
						app.Stop()
					})
	sidebar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chooseTopic, 0, 1, true).
		AddItem(exitButton, 1, 0, false)

	
	msgs := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText("1YIPEEE\n2YIPEE\n3YIPEE\n4YIPEE\n5YIPEE\n6YIPEE\n7YIPEE\n8YIPEE\n9YIPEE\n10YIPEE\n11YIPEE\n12YIPEE\n13YIPEE\n14YIPEE\n15YIPEE\n16YIPEE\n17YIPEE\n18YIPEE\n19YIPEE\n20YIPEE\n21YIPEE\n22YIPEE\n23YIPEE\n24YIPEE\n25YIPEE\n26YIPEE\n27YIPEE\n28YIPEE\n29YIPEE\n30YIPEE\n31YIPEE\n32YIPEE\n33YIPEE\n34YIPEE\n35YIPEE\n36YIPEE\n37YIPEE\n38YIPEE\n39YIPEE\n40YIPEE\n41YIPEE\n42YIPEE\n43YIPEE\n44YIPEE\n45YIPEE\n46YIPEE\n47YIPEE\n48YIPEE\n49YIPEE")
	// Lower part of main section, meant to represent input
	textField := tview.NewTextArea()
	sendButton := tview.NewButton("Send").
					SetSelectedFunc(func() {
						textToSend := textField.GetText()
						SendMessageToServer(textToSend)
					})
	sendRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(textField, 0, 1, true).
		AddItem(sendButton, 10, 0, false)




	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(msgs, 0, 1, false).
		AddItem(sendRow, 0, 1, true)
	
	layout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(sidebar, 0, 1, false).
		AddItem(main, 0, 3, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
			case tcell.KeyLeft:
				if(app.GetFocus() == sendButton){
					app.SetFocus(textField)
				} else if(sidebarOrMainFocus != sidebarFocus){
					app.SetFocus(sidebar)
					sidebarOrMainFocus = sidebarFocus
				}
			case tcell.KeyRight:
				if(sidebarOrMainFocus != mainFocus){	
					app.SetFocus(main)
					sidebarOrMainFocus = mainFocus
				}else if(app.GetFocus() == textField){
					app.SetFocus(sendButton)
				}
			case tcell.KeyTab:
				if (sidebarOrMainFocus == sidebarFocus){
					if(app.GetFocus() == chooseTopic || app.GetFocus() == exitButton){
						if(app.GetFocus() == exitButton){
							app.SetFocus(chooseTopic)
						}else{
							app.SetFocus(exitButton)
						}
					}else{
						app.Stop()
						log.Fatal("Fatal error 100: tried getting focus in sidebar, did not get exitButton or chooseTopic")
					}
				}else if (sidebarOrMainFocus == mainFocus){
					if(app.GetFocus() == msgs){
						app.SetFocus(textField)
					}else if(app.GetFocus() == textField){
						app.SetFocus(sendButton)
					}else if(app.GetFocus() == sendButton){
						app.SetFocus(msgs)
					}else{
						app.Stop()
						log.Fatal("Fatal error 101: tried getting focus in main, did not get msgs or sendText")
					}
				}
				return event
		}
		return event
	})
	return app.SetRoot(layout, true).Run()
}

func changeFooterToStop(footer *tview.TextView){
	
}