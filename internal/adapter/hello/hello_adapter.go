package hello

import (
	"context"
	"log"

	"github.com/viquitorreis/my-grpc-go-client/internal/port"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/hello"
	"google.golang.org/grpc"
)

type HelloAdapter struct {
	helloClient port.HelloClientPort
}

func NewHelloAdapter(conn *grpc.ClientConn) (*HelloAdapter, error) {
	return &HelloAdapter{
		helloClient: hello.NewHelloServiceClient(conn),
	}, nil
}

func (a *HelloAdapter) SayHello(ctx context.Context, name string) (*hello.HelloResponse, error) {
	helloRequest := &hello.HelloRequest{
		Name: name,
	}

	greet, err := a.helloClient.SayHello(ctx, helloRequest)
	if err != nil {
		log.Fatal("Erro ao chamar o serviço de hello, err:", err)
	}

	// greet = hello response
	return greet, nil
}
