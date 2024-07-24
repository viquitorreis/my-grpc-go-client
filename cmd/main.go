package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/viquitorreis/my-grpc-go-client/internal/adapter/bank"
	"github.com/viquitorreis/my-grpc-go-client/internal/adapter/hello"
	domainBank "github.com/viquitorreis/my-grpc-go-client/internal/application/domain/bank"
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

	bankAdapter, err := bank.NewBankAdapter(conn)
	if err != nil {
		log.Fatal("Erro ao criar o adapter de bank, err:", err)
	}

	runTransferMultiple(bankAdapter, "7835697001zzzz", "7835697002", 5)
}

func runGetCurrentBalance(adapter *bank.BankAdapter, account string) {
	bal, err := adapter.GetCurrentBalance(context.Background(), account)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço de bank, err:", err)
	}

	log.Println("Saldo atual da conta:", bal)
}

func runFetchExchangeRates(adapter *bank.BankAdapter, fromCur, toCur string) {
	ctx := context.Background()
	log.Println("Fetching exchange rates from", fromCur, "to", toCur)
	log.Println("ctx:", ctx)
	adapter.FetchExchangeRates(ctx, fromCur, toCur)
}

func runSummarizeTransactions(adapter *bank.BankAdapter, account string, dummyTransactions int) {
	var tx []*domainBank.Transaction

	for i := 1; i <= dummyTransactions; i++ {
		tranType := domainBank.TransactionTypeIn

		if i%3 == 0 {
			tranType = domainBank.TransactionTypeOut
		}

		t := &domainBank.Transaction{
			Amount:          float64(rand.Intn(500) + 10),
			TransactionType: tranType,
			Notes:           fmt.Sprintf("Transação de teste %d", i),
		}

		tx = append(tx, t)
	}

	adapter.SummarizeTransactions(context.Background(), account, tx)
}

func runSayHello(adapter *hello.HelloAdapter, name string) {
	greet, err := adapter.SayHello(context.Background(), name)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
	}

	log.Println("Resposta do serviço de hello:", greet.Message)
}

func runTransferMultiple(adapter *bank.BankAdapter, fromAcc, toAcc string, numDummyTransactions int) {
	var trf []domainBank.TransferTransaction

	for i := 1; i <= numDummyTransactions; i++ {
		t := domainBank.TransferTransaction{
			FromAccountNumber: fromAcc,
			ToAccountNumber:   toAcc,
			Currency:          "BRL",
			Amount:            float64(rand.Intn(200) + 10),
		}

		trf = append(trf, t)
	}

	adapter.TransferMultiple(context.Background(), trf)
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
