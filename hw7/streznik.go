package main

import (
	"context"
	"fmt"
	"hw7/protobufStorage"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Server(url string) {
	// pripravimo strežnik gRPC
	grpcServer := grpc.NewServer()

	// pripravimo strukturo za streženje metod CRUD na shrambi TodoStorage
	crudServer := NewServerCRUD()

	// streženje metod CRUD na shrambi TodoStorage povežemo s strežnikom gRPC
	protobufStorage.RegisterCRUDServer(grpcServer, crudServer)

	// izpišemo ime strežnika
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	// odpremo vtičnico
	listener, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}
	fmt.Printf("gRPC server listening at %v%v\n", hostName, url)
	// začnemo s streženjem
	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
}

// struktura za strežnik CRUD za shrambo TodoStorage
type serverCRUD struct {
	protobufStorage.UnimplementedCRUDServer
	todoStore *TodoStorage
	observer  *Observer
}

// pripravimo nov strežnik CRUD za shrambo TodoStorage
func NewServerCRUD() *serverCRUD {
	return &serverCRUD{
		todoStore: NewTodoStorage(),
		observer:  NewObserver(),
	}
}

// metode strežnika CRUD za shrambo TodoStorage
// odtis metode (argumenti, vrnjene vrednosti) so podane tako, kot zahtevajo paketi, ki jih je ustvaril prevajalnik protoc
// vse metode so samo ovojnice za klic metod v TodoStorage
// metode se izvajajo v okviru omejitev izvajalnega okolju ctx
func (s *serverCRUD) Create(ctx context.Context, in *protobufStorage.Todo) (*emptypb.Empty, error) {
	var ret struct{}
	todo := &Todo{Task: in.Task, Completed: in.Completed}
	if err := s.todoStore.Create(todo, &ret); err != nil {
		return nil, err
	}
	s.observer.Notify(&protobufStorage.TodoEvent{
		T: &protobufStorage.Todo{
			Task:      todo.Task,
			Completed: todo.Completed,
		},
		Action: "create",
	})
	return &emptypb.Empty{}, nil
}

func (s *serverCRUD) Read(ctx context.Context, in *protobufStorage.Todo) (*protobufStorage.TodoStorage, error) {
	dict := make(map[string](Todo))
	err := s.todoStore.Read(&Todo{Task: in.Task, Completed: in.Completed}, &dict)
	pbDict := protobufStorage.TodoStorage{}
	for k, v := range dict {
		pbDict.Todos = append(pbDict.Todos, &protobufStorage.Todo{Task: k, Completed: v.Completed})
	}
	return &pbDict, err
}

func (s *serverCRUD) Update(ctx context.Context, in *protobufStorage.Todo) (*emptypb.Empty, error) {
	var ret struct{}
	todo := &Todo{Task: in.Task, Completed: in.Completed}
	if err := s.todoStore.Update(todo, &ret); err != nil {
		return nil, err
	}
	s.observer.Notify(&protobufStorage.TodoEvent{
		T: &protobufStorage.Todo{
			Task:      todo.Task,
			Completed: todo.Completed,
		},
		Action: "update",
	})
	return &emptypb.Empty{}, nil
}

func (s *serverCRUD) Delete(ctx context.Context, in *protobufStorage.Todo) (*emptypb.Empty, error) {
	var ret struct{}
	todo := &Todo{Task: in.Task, Completed: in.Completed}
	if err := s.todoStore.Delete(todo, &ret); err != nil {
		return nil, err
	}
	s.observer.Notify(&protobufStorage.TodoEvent{
		T: &protobufStorage.Todo{
			Task:      todo.Task,
			Completed: todo.Completed,
		},
		Action: "delete",
	})
	return &emptypb.Empty{}, nil
}

func (s *serverCRUD) Subscribe(_ *emptypb.Empty, stream grpc.ServerStreamingServer[protobufStorage.TodoEvent]) error {
	ctx := stream.Context()
	observable, cancel := s.observer.Observe()
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event := <-observable:
			if err := stream.Send(event); err != nil {
				return err
			}
		}
	}
}
