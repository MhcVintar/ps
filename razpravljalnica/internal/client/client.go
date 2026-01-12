package client

//https://github.com/BekBrace/inv-tui-go/blob/main/main.go
import (
	"context"
	"fmt"
	"log"
	"razpravljalnica/internal/api"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	sidebarFocus = 0
	mainFocus    = 1
)

var (
	app                               *tview.Application
	topicView                         *tview.TreeView
	topicTree                         *tview.TreeNode
	exitButton                        *tview.Button
	sidebar                           *tview.Flex
	msgs                              *tview.List
	textField                         *tview.TextArea
	sendButton                        *tview.Button
	pages                             *tview.Pages
	server                            string
	port                              int
	sidebarOrMainFocus                int
	conn                              *grpc.ClientConn
	err                               error
	client                            api.MessageBoardClient
	controlClient                     api.ControlPlaneClient
	UUID                              string
	idOfClient                        int64
	listOfTopics                      []*api.Topic
	listOfCurrentMessages             []*api.Message
	globalCurrentTopic                int64
	head                              *api.NodeInfo
	tail                              *api.NodeInfo
	currentChild                      int
	hashmapOfTopicToTopicStreamStruct map[int64]topicStreamStruct
	topicIdToName                     map[int64]string
)

type topicStreamStruct struct {
	ListOfMessagesInTopic  []*api.Message
	Topic                  *api.Topic
	StreamOfMsg            api.MessageBoard_SubscribeTopicClient
	SubscriptionHandleNode *api.NodeInfo
}

func updateMessageView() {
	updateMessageViewWithOffset(int64(0), hashmapOfTopicToTopicStreamStruct[globalCurrentTopic])
}

func updateMessageViewWithOffset(offset int64, updateTopicVar topicStreamStruct) {
	tmpTopic := updateTopicVar.Topic.Id
	if tmpTopic != globalCurrentTopic {
		return
	}
	msgs.Clear()

	for _, message := range listOfCurrentMessages {
		if globalCurrentTopic == message.TopicId {
			msgs.AddItem(
				message.Text,
				fmt.Sprintf("Likes: %d  Id:[%d] Created by id:%d", message.Likes, message.Id, message.UserId),
				0,
				nil,
			)

		}

	}
	msgs.SetCurrentItem(int(offset))

}

// General housekeeping
func handleStop() {
	app.Stop()
}

func handleStopWithError(err error) {
	app.Stop()
	fmt.Println(err)
}
func handleStopWithErrorString(err string) {
	app.Stop()
	fmt.Println(err)
}

// should be called once per wheneverthefuck
func updateTopicsOnSidebar() {
	topicTree.ClearChildren()
	for _, topic := range listOfTopics {
		tmp := tview.NewTreeNode(topic.Name).SetReference(topic)
		topicTree.AddChild(tmp)
		if hashmapOfTopicToTopicStreamStruct[topic.Id].StreamOfMsg != nil {
			tmp.SetColor(tcell.ColorRed)
		}
	}
	if len(topicTree.GetChildren()) > 0 {
		topicView.SetCurrentNode(topicTree.GetChildren()[0])
	}
	time.Sleep(10 * time.Second)
}

// Creates a subscription on current topic
func createSubscription(from int64) {
	// 1. Request a node to which a subscription can be opened
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx := context.WithoutCancel(context.Background())
	//defer cancel()
	response, err := client.GetSubscriptionNode(ctx, &api.SubscriptionNodeRequest{
		UserId:  idOfClient,
		TopicId: []int64{globalCurrentTopic},
	})
	if err != nil {
		handleStopWithError(err)
	}
	//nodeToSubscribeTo := response.Node
	subscribeToken := response.SubscribeToken
	// 2. Subscribe to topic given by previous request

	//ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	ctx2 := context.WithoutCancel(context.Background())
	//defer cancel2()
	stream, err := client.SubscribeTopic(ctx2, &api.SubscribeTopicRequest{
		TopicId:        []int64{globalCurrentTopic},
		UserId:         idOfClient,
		FromMessageId:  from,
		SubscribeToken: subscribeToken,
	})
	if err != nil {
		handleStopWithError(err)
	}
	tmp := hashmapOfTopicToTopicStreamStruct[globalCurrentTopic]
	tmp.StreamOfMsg = stream
	tmp.SubscriptionHandleNode = response.Node
	hashmapOfTopicToTopicStreamStruct[globalCurrentTopic] = tmp
	// Start a goroutine to read from the stream
	go func() {
		for {
			event, err := stream.Recv()
			if err != nil {
				log.Println("Stream error:", err)
				break
			}
			postedMsg := event.Message
			if event.Op == api.OpType_OP_POST {
				tmp := hashmapOfTopicToTopicStreamStruct[postedMsg.TopicId]
				tmp.ListOfMessagesInTopic = append(tmp.ListOfMessagesInTopic, postedMsg)
				hashmapOfTopicToTopicStreamStruct[postedMsg.TopicId] = tmp
			} else if event.Op == api.OpType_OP_LIKE {
				tmp := hashmapOfTopicToTopicStreamStruct[postedMsg.TopicId]
				tmp.ListOfMessagesInTopic[postedMsg.Id].Likes = tmp.ListOfMessagesInTopic[postedMsg.Id].Likes + 1
				hashmapOfTopicToTopicStreamStruct[postedMsg.TopicId] = tmp
			} else if event.Op == api.OpType_OP_DELETE {
				deleteMessage(postedMsg.Id)
			} else if event.Op == api.OpType_OP_UPDATE {
				updateMessage(postedMsg.Id, postedMsg.Text, postedMsg.TopicId)
			}
			updateMessageViewWithOffset(postedMsg.Id, hashmapOfTopicToTopicStreamStruct[postedMsg.TopicId])
			// TODO: Update UI or state based on event
		}
	}()
	updateTopicsOnSidebar()
}

func createUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.CreateUser(ctx, &api.CreateUserRequest{
		Name: UUID,
	})
	if err != nil {
		handleStopWithError(err)
	}
	idOfClient = response.Id
}

func likeMessage(sporociloId int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.LikeMessage(ctx, &api.LikeMessageRequest{
		TopicId:   globalCurrentTopic,
		MessageId: sporociloId,
		UserId:    idOfClient,
	})
	if err != nil {
		handleStopWithError(err)
	}
}

func sendMessage(sporocilo string, topicId int64) {
	if len(sporocilo) >= 1 {

		//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		ctx := context.WithoutCancel(context.Background())
		//defer cancel()
		postedMsg, err := client.PostMessage(ctx, &api.PostMessageRequest{
			TopicId: topicId,
			UserId:  idOfClient,
			Text:    sporocilo,
		})
		if err != nil {
			app.Stop()
			fmt.Println(err)
			//handleStopWithError(err)
		}
		tmp := hashmapOfTopicToTopicStreamStruct[topicId]
		tmp.ListOfMessagesInTopic = append(tmp.ListOfMessagesInTopic, postedMsg)
		hashmapOfTopicToTopicStreamStruct[topicId] = tmp
		updateMessageViewWithOffset(int64(len(tmp.ListOfMessagesInTopic)-1), hashmapOfTopicToTopicStreamStruct[topicId])
	}
}

func deleteMessage(sporociloId int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.DeleteMessage(ctx, &api.DeleteMessageRequest{
		TopicId:   globalCurrentTopic,
		UserId:    idOfClient,
		MessageId: sporociloId,
	})
	if err != nil {
		handleStopWithError(err)
	}
	//updateMessageView()
}

func createTopic(topicName string) {
	if len(topicName) >= 1 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.CreateTopic(ctx, &api.CreateTopicRequest{
			Name: topicName,
		})
		if err != nil {
			handleStopWithError(err)
		}
		updateTopicsOnSidebar()
	}
}

func updateMessage(messageId int64, newText string, topicId int64) {
	if len(newText) >= 1 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.UpdateMessage(ctx, &api.UpdateMessageRequest{
			TopicId:   topicId,
			UserId:    idOfClient,
			MessageId: messageId,
			Text:      newText,
		})
		if err != nil {
			handleStopWithError(err)
		}
		getMsgs(int64(0), int32(50), topicId)
	}
}

func getTopics() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	listOfTopics2, err := client.ListTopics(ctx, &emptypb.Empty{})

	if err != nil {
		handleStopWithError(err)
	}
	listOfTopics = listOfTopics2.Topics
	for _, topic := range listOfTopics {
		if _, b := hashmapOfTopicToTopicStreamStruct[topic.Id]; b == true {
			continue
		}
		topicIdToName[topic.Id] = topic.Name

		hashmapOfTopicToTopicStreamStruct[topic.Id] = topicStreamStruct{
			ListOfMessagesInTopic:  []*api.Message{},
			Topic:                  topic,
			StreamOfMsg:            nil,
			SubscriptionHandleNode: nil,
		}
	}
	updateTopicsOnSidebar()

}

func getMsgs(from int64, limit int32, idOfTopic int64) {
	//ctx, cancel := context.WithTimeout(context.Background(), 5000000*time.Second)
	ctx := context.WithoutCancel(context.Background())
	//defer cancel()
	tmpListOfMessages, err := client.GetMessages(ctx, &api.GetMessagesRequest{
		TopicId:       idOfTopic,
		FromMessageId: from,
		Limit:         limit,
	})
	if err != nil {
		handleStopWithError(err)
	}
	listOfCurrentMessages = tmpListOfMessages.Messages
	tmp := hashmapOfTopicToTopicStreamStruct[idOfTopic]
	tmp.ListOfMessagesInTopic = tmpListOfMessages.Messages
	hashmapOfTopicToTopicStreamStruct[idOfTopic] = tmp
	updateMessageViewWithOffset(0, hashmapOfTopicToTopicStreamStruct[idOfTopic])
	//msgs.SetBackgroundColor(tcell.ColorRed)

}

func Bootstrap(serverName string, port int) {
	currentChild = -1
	globalCurrentTopic = -1
	UUID = uuid.NewString()
	hashmapOfTopicToTopicStreamStruct = make(map[int64]topicStreamStruct)
	topicIdToName = make(map[int64]string)
	addressToControl := fmt.Sprintf("%s:%d", serverName, port)
	controlClientConn, err := grpc.NewClient(addressToControl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	controlClient = api.NewControlPlaneClient(controlClientConn)
	if err != nil {
		log.Fatalf("Failed to connect to Control %s: %v", addressToControl, err)
	}
	defer controlClientConn.Close()

	setHeadAndTail()
	conn, err := grpc.NewClient(tail.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect to Tail %s: %v", addressToControl, err)
	}
	defer conn.Close()
	client = api.NewMessageBoardClient(conn)
	createUser()

	//Preiodically see other topics
	go func() {
		getTopics()
		time.Sleep(5 * time.Second)
	}()
	if x := runGUI(); x != nil {
		return
	}

}

func handlePopout() {
	sidebarOrMainFocus = 3

	areaForTopic := tview.NewInputField().SetLabel("Topic: ")

	newFlexLine := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(areaForTopic, 100, 10, true)

	newFlexLine.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc || ev.Key() == tcell.KeyEnter {
			pages.RemovePage("popup")
			if tcell.KeyEnter == ev.Key() {
				createTopic(areaForTopic.GetText())
			}
		}

		return ev
	})
	pages.AddAndSwitchToPage("popup", newFlexLine, true)
	//app.SetFocus(newFlexLine)
}
func setHeadAndTail() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	clusterResponse, err := controlClient.GetClusterState(ctx, &emptypb.Empty{})

	if err != nil {
		handleStopWithError(err)
	}
	head = clusterResponse.Head
	tail = clusterResponse.Tail

}
func runGUI() error {
	sidebarOrMainFocus = mainFocus
	fmt.Println("Client bootstrap")
	app = tview.NewApplication()

	//This will initialize into an empty Flex
	topicTree = tview.NewTreeNode("Topics")
	topicView = tview.NewTreeView().SetRoot(topicTree)

	topicView.SetBorder(true).SetTitle(" Choose topic ").SetTitleAlign(tview.AlignLeft)
	exitButton = tview.NewButton("Exit").
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
		AddItem(topicView, 0, 10, true).
		AddItem(rowButtons, 0, 1, false)

	/*msgs = tview.NewFlex().
	SetDirection(tview.FlexRow)*/

	msgs = tview.NewList().
		ShowSecondaryText(true)
	msgs.SetBorder(true).SetTitle("Messages")
	msgs.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			data := listOfCurrentMessages[msgs.GetCurrentItem()]
			if data == nil {
				return event
			}

			//TODO Like the msg if not the author, otherwise make a popup with editing MSG
			if data.UserId == idOfClient {
				updateMessage(data.Id, textField.GetText(), data.TopicId)
			} else {
				likeMessage(data.Id)
			}
		}

		return event

	})
	msgs.SetMainTextStyle(
		tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlue),
	)
	// Lower part of main section, meant to represent input
	textField = tview.NewTextArea()

	sendButton := tview.NewButton("Send").
		SetSelectedFunc(func() {
			textToSend := textField.GetText()
			sendMessage(textToSend, globalCurrentTopic)
			textField.SetText("", false)
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
		case tcell.KeyEnter:
			if app.GetFocus() == topicView {
				inter := topicView.GetCurrentNode().GetReference()
				var node *api.Topic
				if inter != nil {
					node = inter.(*api.Topic)
				} else {
					app.Stop()
					fmt.Println(inter)
					log.Fatal(inter)
				}
				tmpStruct, b := hashmapOfTopicToTopicStreamStruct[node.Id]
				if b == false {
					//app.Stop()
					fmt.Println(tmpStruct, " currentTopic", globalCurrentTopic)
				}
				if tmpStruct.Topic == nil {
					//getTopics()
				}
				if globalCurrentTopic == tmpStruct.Topic.Id {
					messages := tmpStruct.ListOfMessagesInTopic
					lastMessage := (len(tmpStruct.ListOfMessagesInTopic) - 1)
					if lastMessage < 0 {
						createSubscription(int64(0))
					} else {
						createSubscription(messages[lastMessage].Id)
					}
				} else {
					globalCurrentTopic = tmpStruct.Topic.Id
					getMsgs(0, 50, node.Id)
				}

			}
		case tcell.KeyLeft:
			if app.GetFocus() == sendButton {
				app.SetFocus(textField)
			} else if sidebarOrMainFocus != sidebarFocus {
				app.SetFocus(sidebar)
				sidebarOrMainFocus = sidebarFocus
			}
		case tcell.KeyRight:
			if sidebarOrMainFocus != mainFocus {
				app.SetFocus(main)
				sidebarOrMainFocus = mainFocus
			} else if app.GetFocus() == textField {
				app.SetFocus(sendButton)
			}
		case tcell.KeyTab:
			if sidebarOrMainFocus == sidebarFocus {
				if app.GetFocus() == topicView || app.GetFocus() == exitButton {
					if app.GetFocus() == exitButton {
						app.SetFocus(topicView)
					} else {
						app.SetFocus(exitButton)
					}
				} else {
					handleStopWithErrorString("Fatal error 100: tried getting focus in sidebar, did not get exitButton or topicView")
				}
			} else if sidebarOrMainFocus == mainFocus {
				if app.GetFocus() == msgs {
					app.SetFocus(textField)
				} else if app.GetFocus() == textField {
					app.SetFocus(sendButton)
				} else if app.GetFocus() == sendButton {
					app.SetFocus(msgs)
				} else {
					handleStopWithErrorString("Fatal error 101: tried getting focus in main, did not get msgs or sendText")
				}
			}
			return event
		}
		switch event.Rune() {
		case 'd', 'D':
			{
				if app.GetFocus() == exitButton {
					app.SetFocus(addTopicButton)
				}
			}
		case 'a', 'A':
			{
				if app.GetFocus() == addTopicButton {
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
