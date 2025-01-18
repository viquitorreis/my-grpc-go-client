package main

import (
	"context"
	"log"
	"time"

	"github.com/viquitorreis/my-grpc-go-client/internal/adapter/resiliency"
	domainResiliency "github.com/viquitorreis/my-grpc-go-client/internal/application/domain/resiliency"
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

	resiliencyAdapter, err := resiliency.NewResiliencyAdapter(conn)
	if err != nil {
		log.Fatal("Erro ao criar o adapter de resiliency, err:", err)
	}

	// bankAdapter, err := bank.NewBankAdapter(conn)
	// if err != nil {
	// 	log.Fatal("Erro ao criar o adapter de bank, err:", err)
	// }

	// runTransferMultiple(bankAdapter, "7835697001zzzz", "7835697002", 5)

	runUnaryResiliencyWithTimeout(resiliencyAdapter, 4, 15, []uint32{domainResiliency.OK}, 5*time.Second)
	now := time.Now()
	defer func() {
		log.Println("Tempo total de execução:", time.Since(now))
	}()

	// runServerStreamingResiliencyWithTimeout(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK}, 15*time.Second)
	// runClientStreamingResiliencyWithTimeout(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK}, 10, 2*time.Second)
	runBiDirectionalStreamingResiliencyWithTimeout(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK}, 10, 60*time.Second)
}

// func runGetCurrentBalance(adapter *bank.BankAdapter, account string) {
// 	bal, err := adapter.GetCurrentBalance(context.Background(), account)
// 	if err != nil {
// 		log.Fatalln("Erro ao chamar o serviço de bank, err:", err)
// 	}

// 	log.Println("Saldo atual da conta:", bal)
// }

// func runFetchExchangeRates(adapter *bank.BankAdapter, fromCur, toCur string) {
// 	ctx := context.Background()
// 	log.Println("Fetching exchange rates from", fromCur, "to", toCur)
// 	log.Println("ctx:", ctx)
// 	adapter.FetchExchangeRates(ctx, fromCur, toCur)
// }

// func runSummarizeTransactions(adapter *bank.BankAdapter, account string, dummyTransactions int) {
// 	var tx []*domainBank.Transaction

// 	for i := 1; i <= dummyTransactions; i++ {
// 		tranType := domainBank.TransactionTypeIn

// 		if i%3 == 0 {
// 			tranType = domainBank.TransactionTypeOut
// 		}

// 		t := &domainBank.Transaction{
// 			Amount:          float64(rand.Intn(500) + 10),
// 			TransactionType: tranType,
// 			Notes:           fmt.Sprintf("Transação de teste %d", i),
// 		}

// 		tx = append(tx, t)
// 	}

// 	adapter.SummarizeTransactions(context.Background(), account, tx)
// }

// func runSayHello(adapter *hello.HelloAdapter, name string) {
// 	greet, err := adapter.SayHello(context.Background(), name)
// 	if err != nil {
// 		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
// 	}

// 	log.Println("Resposta do serviço de hello:", greet.Message)
// }

// func runTransferMultiple(adapter *bank.BankAdapter, fromAcc, toAcc string, numDummyTransactions int) {
// 	var trf []domainBank.TransferTransaction

// 	for i := 1; i <= numDummyTransactions; i++ {
// 		t := domainBank.TransferTransaction{
// 			FromAccountNumber: fromAcc,
// 			ToAccountNumber:   toAcc,
// 			Currency:          "BRL",
// 			Amount:            float64(rand.Intn(200) + 10),
// 		}

// 		trf = append(trf, t)
// 	}

// 	adapter.TransferMultiple(context.Background(), trf)
// }

// func runManyHello(adapter *hello.HelloAdapter, name string) {
// 	greet, err := adapter.SayManyHello(context.Background(), name)
// 	if err != nil {
// 		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
// 	}

// 	log.Println("Resposta do serviço de hello:", greet.Message)
// }

// func runSayHelloToEveryone(adapter *hello.HelloAdapter, names []string) {
// 	adapter.SayHelloToEveryone(context.Background(), names)
// }

// func runSayHelloContinuous(adapter *hello.HelloAdapter, names []string) {
// 	adapter.SayHelloContinuous(context.Background(), names)
// }

// func runUnaryResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, timeout time.Duration) {
// 	ctx, cancel := context.WithTimeout(context.Background(), timeout) // contexto vai esperar apenas o timeout específicado no parâmetro
// 	defer cancel()

// 	res, err := adapter.UnaryResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes)
// 	if err != nil {
// 		log.Fatalln("Erro ao chamar o serviço de resiliency, err:", err)
// 	}

// 	log.Println("Resposta do serviço de resiliency:", res)
// }

// func runServerStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, timeout time.Duration) {
// 	ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 	defer cancel()

// 	adapter.ServerStreamingResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes)
// }

// func runClientStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int, timeout time.Duration) {
// 	ctx, _ := context.WithTimeout(context.Background(), timeout)

// 	adapter.ClientStreamResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes, count)
// }

// func runBiDirectionalStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int, timeout time.Duration) {
// 	ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 	defer cancel()

// 	adapter.BidirectionalStreamingResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes, count)
// }

func runUnaryResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := adapter.UnaryResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes)
	if err != nil {
		log.Fatalln("Failed to call UnaryResiliency: ", err)
	}

	log.Println(res.DummyString)
}

func runServerStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32, timeout time.Duration) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	adapter.ServerStreamingResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes)
}

func runClientStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int, timeout time.Duration) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	adapter.ClientStreamResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes, count)
}

func runBiDirectionalStreamingResiliencyWithTimeout(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32, count int, timeout time.Duration) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	adapter.BidirectionalStreamingResiliency(ctx, minDelaySecond, maxDelaySecond, statusCodes, count)
}
