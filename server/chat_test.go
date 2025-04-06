package main

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"grpc_chatapp.com/chatproto"
)

var bufferSize int = 1024 * 1024

func startTestServer() (*grpc.Server, *bufconn.Listener) {
	// This function would start the test server
	// and listen for incoming connections.
	lis := bufconn.Listen(bufferSize)
	grpcServer := grpc.NewServer()
	chatproto.RegisterChatServiceServer(grpcServer, &chatServer{
		clients: make(map[chatproto.ChatService_ChatServer]string),
	})

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	return grpcServer, lis
}

func dialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}
}

// Test1: Test client connect and send message
func TestChatServer_SingleClient(t *testing.T) {
	grpcSer, lis := startTestServer()
	defer grpcSer.Stop()
	ctx := context.Background()

	// create client connection
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(lis)), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := chatproto.NewChatServiceClient(conn)
	stream, err := client.Chat(ctx)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Send a message
	stream.Send(&chatproto.ChatMessage{
		User:    "testuser",
		Message: "Hello, World!",
	})

	// receive a message
	msgFromServer, err := stream.Recv()
	if err != nil {
		t.Fatalf("Failed to receive message: %v", err)
	}

	if msgFromServer.User != "testuser" || msgFromServer.Message != "Hello, World!" {
		t.Fatalf("Expected message from server to be 'Hello, World!', got: %s", msgFromServer.Message)
	}
}

func TestChatServer_BroadCast(t *testing.T) {
	grpcSer, lis := startTestServer()
	defer grpcSer.Stop()
	ctx := context.Background()
	// create client connection
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer(lis)), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	var clients []chatproto.ChatServiceClient = make([]chatproto.ChatServiceClient, 5)
	var streams []chatproto.ChatService_ChatClient = make([]chatproto.ChatService_ChatClient, 5)
	for i := 0; i < 5; i++ {
		cln := chatproto.NewChatServiceClient(conn)
		clients[i] = cln
		stm, err := cln.Chat(ctx)
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}
		streams[i] = stm
	}

	// 1st client sends a message
	streams[0].Send(&chatproto.ChatMessage{
		User:    "testuser1",
		Message: "Hi everyone!",
	})
	// receive a message from all clients
	for j := 0; j < 5; j++ {
		msgServer, err := streams[j].Recv()
		if err != nil {
			t.Fatalf("Failed to receive message: %v", err)
		}
		if msgServer.User != "testuser1" || msgServer.Message != "Hi everyone!" {
			if msgServer.User != "testuser1" {
				t.Fatalf("Expected message from server to be 'testuser1', exist stream got: %s", msgServer.User)
			} else {

				t.Fatalf("Expected message from server to be 'Hi everyone!',exist stream got: %s", msgServer.Message)
			}
		}
	}
}

func TestMain(m *testing.M) {
	m.Run()
}
