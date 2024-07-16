package hello

import (
	"context"
	"io"
	"log"
	"time"

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

func (a *HelloAdapter) SayHelloToEveryone(ctx context.Context, names []string) {
	greetStream, err := a.helloClient.SayHelloToEveryone(ctx)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço SdayHelloToEveryone, err:", err)
	}

	// loop para enviar as mensagens para o servidor
	for _, name := range names {
		req := &hello.HelloRequest{
			Name: name,
		}

		greetStream.Send(req)
		time.Sleep(500 * time.Millisecond)
	}

	// fechando o stream
	res, err := greetStream.CloseAndRecv()
	if err != nil {
		log.Fatalln("Erro ao fechar o stream, err:", err)
	}

	log.Println("Mensagem do servidor:", res.Message)
}

func (a *HelloAdapter) SayHelloContinuous(ctx context.Context, names []string) {
	greetStream, err := a.helloClient.SayHelloContinuous(ctx)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço SayHelloContinuous, err:", err)
	}

	greetChan := make(chan struct{})

	// goroutine para fazer loop nos nomes e criar uma HelloRequest para enviar ao servidor para cada nome
	go func() {
		for _, name := range names {
			req := &hello.HelloRequest{
				Name: name,
			}

			// enviando a mensagem para o servidor
			greetStream.Send(req)
		}

		// fechando o stream
		greetStream.CloseSend()
	}()

	go func() {
		for {
			res, err := greetStream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalln("Erro ao receber a mensagem do servidor, err:", err)
			}

			log.Println("Mensagem do servidor:", res.Message)
		}
		close(greetChan)
	}()

	// aguardando o fechamento do canal
	<-greetChan
}
