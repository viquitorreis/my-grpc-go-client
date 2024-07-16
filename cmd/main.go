package main

import (
	"context"
	"log"

	"github.com/viquitorreis/my-grpc-go-client/internal/adapter/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	// Create a new gRPC client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:9090", opts...)
	if err != nil {
		log.Fatalln("Erro ao conectar com o servidor gRPC, err:", err)
	}
	defer conn.Close()

	// Create a new adapter
	helloAdapter, err := hello.NewHelloAdapter(conn)
	if err != nil {
		log.Fatal("Erro ao criar o adapter de hello, err:", err)
	}

	runSayHelloContinuous(helloAdapter, []string{"Víctor", "Reis", "Cícero", "César", "Pompeu"})
}

func runSayHello(adapter *hello.HelloAdapter, name string) {
	greet, err := adapter.SayHello(context.Background(), name)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
	}

	log.Println("Resposta do serviço de hello:", greet.Message)
}

func runManyHello(adapter *hello.HelloAdapter, name string) {
	greet, err := adapter.SayManyHello(context.Background(), name)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
	}

	log.Println("Resposta do serviço de hello:", greet.Message)
}

func runSayHelloToEveryone(adapter *hello.HelloAdapter, names []string) {
	adapter.SayHelloToEveryone(context.Background(), names)
}

func runSayHelloContinuous(adapter *hello.HelloAdapter, names []string) {
	adapter.SayHelloContinuous(context.Background(), names)
}
