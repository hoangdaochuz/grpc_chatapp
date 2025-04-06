package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc_chatapp.com/chatproto"
)

func handleSendMsg(stream chatproto.ChatService_ChatClient) {
	defer stream.CloseSend()

	fmt.Println("Enter your username:")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	username = username[:len(username)-1] // Remove newline character
	stream.Send(&chatproto.ChatMessage{
		User:    username,
		Message: fmt.Sprintf("%s: %s", username, " has joined the chat")})

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			stream.Send(&chatproto.ChatMessage{
				User:    username,
				Message: fmt.Sprintf("%s: %s", username, " has left the chat"),
			})
			return
		}
		stream.Send(&chatproto.ChatMessage{
			User:    username,
			Message: fmt.Sprintf("%s: %s", username, msg), // Remove newline character
		})
	}
}

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := chatproto.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		panic(err)
	}
	go handleSendMsg(stream)

	for {
		msgFromServer, err := stream.Recv()
		if err != nil {
			fmt.Println("Error receiving message from server:", err)
			break
		}
		fmt.Printf("%s\n", msgFromServer.Message)
	}
}
