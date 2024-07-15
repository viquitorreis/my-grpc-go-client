package hello

import (
	"context"
	"io"
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

func (a *HelloAdapter) SayManyHello(ctx context.Context, name string) (*hello.HelloResponse, error) {
	helloRequest := &hello.HelloRequest{
		Name: name,
	}

	greetStream, err := a.helloClient.SayManyHello(ctx, helloRequest)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
	}

	// loop infinito para receber as mensagens do servidor
	for {
		greet, err := greetStream.Recv()
		if err != nil {
			log.Fatalln("Erro ao receber a mensagem do servidor, err:", err)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln("Erro ao receber a mensagem do servidor, err:", err)
		}

		log.Println("Mensagem do servidor:", greet.Message)
	}

	return nil, nil
}
