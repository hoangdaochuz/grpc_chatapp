package main

import (
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"grpc_chatapp.com/chatproto"
)

type chatServer struct {
	chatproto.UnimplementedChatServiceServer
	mux     sync.Mutex
	clients map[chatproto.ChatService_ChatServer]string
}

func (chatServer *chatServer) Chat(stream chatproto.ChatService_ChatServer) error {
	chatServer.mux.Lock()
	chatServer.clients[stream] = ""
	chatServer.mux.Unlock()

	var username string

	for {
		msg, err := stream.Recv()
		if err != nil {
			chatServer.mux.Lock()
			delete(chatServer.clients, stream)
			chatServer.mux.Unlock()
			return err
		}

		if username == "" {
			username = msg.User
			chatServer.mux.Lock()
			chatServer.clients[stream] = username
			chatServer.mux.Unlock()
		}
		chatServer.broadcast(msg)
	}

}

func (chatServer *chatServer) broadcast(msg *chatproto.ChatMessage) {
	chatServer.mux.Lock()
	defer chatServer.mux.Unlock()
	for client := range chatServer.clients {
		client.Send(msg)
	}
}

func main() {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	chatproto.RegisterChatServiceServer(grpcServer, &chatServer{
		clients: make(map[chatproto.ChatService_ChatServer]string),
	})
	fmt.Println("gRPC Chat server is running on port :50051")
	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}

}
