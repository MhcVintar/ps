package client
//https://github.com/BekBrace/inv-tui-go/blob/main/main.go
import(
	"fmt"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
	"log"
	"razpravljalnica/internal/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/google/uuid"
	"time"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
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
	pages *tview.Pages
	server string
	port int
	sidebarOrMainFocus int
	conn *grpc.ClientConn
	err error
	client api.MessageBoardClient
	UUID string
	idOfClient int64
	listOfTopics []*api.Topic
	listOfMessages []*api.Message
	globalCurrentTopic int64
	head *api.NodeInfo
	tail *api.NodeInfo
	
)

type topicStreamStruct struct{
	topic *api.Topic
	streamOfMsg api.MessageBoard_SubscribeTopicClient
	subscriptionHandleNode *api.NodeInfo
}

/*	
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.(ctx, &api.Request{
		
	})
	if err != nil {
		log.Fatal(err)
	}
*/

func handleStop(){

	app.Stop()
}

// TODO
func updateTopicsOnSidebar(){
	for indexOfTopic, topic := range listOfTopics{

	}
}

// Creates a subscription on current topic
func createSubscription(from int64){
	// 1. Request a node to which a subscription can be opened
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.GetSubscriptionNode(ctx, &api.SubscriptionNodeRequest{
		UserId: idOfClient,
		TopicId: []int64{globalCurrentTopic},
	})
	if err != nil {
		log.Fatal(err)
	}
	nodeToSubscribeTo := response.Node
	subscribeToken := response.SubscribeToken
	// 2. Subscribe to topic given by previous request

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	stream, err := client.SubscribeTopic(ctx2, &api.SubscribeTopicRequest{
		TopicId: []int64{globalCurrentTopic},
		UserId: idOfClient,
		FromMessageId: from,
		SubscribeToken: subscribeToken,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Start a goroutine to read from the stream
	go func() {
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("Stream error:", err)
				break
			}
			// Handle the event
			log.Printf("Received event: %v on message %d", event.Op, event.Message.Id)
			// TODO: Update UI or state based on event
		}
	}()
}


func createUser(){
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.CreateUser(ctx, &api.CreateUserRequest{
		Name: UUID,
	})
	if err != nil {
		log.Fatal(err)
	}
	idOfClient = response.Id
}

func likeMessage(sporociloId int64){
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.LikeMessage(ctx, &api.LikeMessageRequest{
		TopicId: globalCurrentTopic,
		MessageId: sporociloId,
		UserId: idOfClient,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func sendMessage(sporocilo string){
	if(len(sporocilo) >= 1){
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.PostMessage(ctx, &api.PostMessageRequest{
			TopicId: globalCurrentTopic,
			UserId: idOfClient,
			Text: sporocilo,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func deleteMessage(sporociloId int64){
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.DeleteMessage(ctx, &api.DeleteMessageRequest{
		TopicId: globalCurrentTopic,
		UserId: idOfClient,
		MessageId: sporociloId,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createTopic(topicName string){
	if(len(topicName) >= 1){
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTopic(ctx, &api.CreateTopicRequest{
			Name: topicName,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func updateMessage(messageId int64, newText string){
	if(len(newText) >= 1){
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.UpdateMessage(ctx, &api.UpdateMessageRequest{
			TopicId: globalCurrentTopic,
			UserId: idOfClient,
			MessageId: messageId,
			Text: newText,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getTopics(){
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	listOfTopics2, err := client.ListTopics(ctx, &emptypb.Empty{})
	
	if err != nil {
		log.Fatal(err)
	}
	listOfTopics = listOfTopics2.Topics
	updateTopicsOnSidebar()
}

func getMsgs(from int64, limit int32){
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tmpListOfMessages, err := client.GetMessages(ctx, &api.GetMessagesRequest{
		TopicId: globalCurrentTopic,
		FromMessageId: from,
		Limit: limit,
	})
	if err != nil {
		log.Fatal(err)
	}
	listOfMessages = tmpListOfMessages.Messages
}



func Bootstrap(serverName string, port int){
	UUID = uuid.NewString()
	address := fmt.Sprintf("%s:%d", serverName, port)
    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to %s: %v", address, err)
    }
    defer conn.Close()

	client = api.NewMessageBoardClient(conn)

	


	if x:=runGUI(); x!= nil{
		return
	}
}

func handlePopout(){
	sidebarOrMainFocus = 3

 	areaForTopic := tview.NewInputField().SetLabel("Topic: ")


	newFlexLine := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(areaForTopic, 100, 10, true)

	newFlexLine.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if (ev.Key() == tcell.KeyEsc || ev.Key() == tcell.KeyEnter){
			pages.RemovePage("popup")
			if(tcell.KeyEnter == ev.Key()){
				createTopic(areaForTopic.GetText())
			}
		}

		return ev
	})
	pages.AddAndSwitchToPage("popup", newFlexLine, true)
	//app.SetFocus(newFlexLine)
}

func runGUI() error{
	sidebarOrMainFocus = mainFocus
	fmt.Println("Client bootstrap")
	app = tview.NewApplication()

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
						handleStop()
					})
	addTopicButton := tview.NewButton("Add topic").
					SetSelectedFunc(func() {
						handlePopout()
					})
	rowButtons := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(exitButton, 0, 1, true).
		AddItem(addTopicButton, 0, 1, false)
	sidebar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chooseTopic, 0, 10, true).
		AddItem(rowButtons, 0, 1, false)

	
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
						sendMessage(textToSend)
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
						handleStop()
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
						handleStop()
						log.Fatal("Fatal error 101: tried getting focus in main, did not get msgs or sendText")
					}
				}
				return event
		}
		switch event.Rune(){
			case 'd', 'D':{
				if (app.GetFocus() == exitButton){
					app.SetFocus(addTopicButton)
				}		
			}
			case 'a', 'A':{
				if(app.GetFocus() == addTopicButton){
					app.SetFocus(exitButton)
				}
			}
		}

		return event
	})
	pages = tview.NewPages()
	pages.AddPage("main", layout, true, true)
	return app.SetRoot(pages, true).Run()
}
