package main

import (
	"context"
	"log"
	"time"

	"github.com/sony/gobreaker"
	"github.com/viquitorreis/my-grpc-go-client/internal/adapter/hello"
	"github.com/viquitorreis/my-grpc-go-client/internal/adapter/resiliency"
	domainResiliency "github.com/viquitorreis/my-grpc-go-client/internal/application/domain/resiliency"
	"github.com/viquitorreis/my-grpc-go-client/internal/interceptor"
	reslProto "github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var circuitBreaker *gobreaker.CircuitBreaker

func initCircuitBreaker() {
	myBreaker := gobreaker.Settings{
		Name: "my-circuit-breaker",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)

			log.Printf("Circuit breaker failure is %v, requests is %v, means failure ratio: %v\n",
				counts.TotalFailures, counts.Requests, failureRatio,
			)

			// failure ratio de 60%
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		Timeout: 4 * time.Second,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("Circuit breaker %v changed state, from %v to %v\n\n", name, from, to)
		},
	}

	circuitBreaker = gobreaker.NewCircuitBreaker(myBreaker)
}

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	// Create a new gRPC client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// opts = append(
	// 	opts,
	// 	grpc.WithUnaryInterceptor(
	// 		grpcRetry.UnaryClientInterceptor(
	// 			// Vamos usar o retry apenas se for status code Unkown e internal error
	// 			grpcRetry.WithCodes(codes.Unknown, codes.Internal),
	// 			grpcRetry.WithMax(4),
	// 			grpcRetry.WithBackoff(grpcRetry.BackoffExponential(2*time.Second)),
	// 		),
	// 	),
	// )

	// opts = append(
	// 	opts,
	// 	grpc.WithStreamInterceptor(
	// 		grpcRetry.StreamClientInterceptor(
	// 			// Vamos usar o retry apenas se for status code Unkown e internal error
	// 			grpcRetry.WithCodes(codes.Unknown, codes.Internal),
	// 			grpcRetry.WithMax(4),
	// 			grpcRetry.WithBackoff(grpcRetry.BackoffLinear(3*time.Second)),
	// 		),
	// 	),
	// )

	initCircuitBreaker()

	opts = append(opts,
		grpc.WithChainUnaryInterceptor(
			interceptor.LogUnaryClientInterceptor(),
			interceptor.BasicUnaryServerInterceptor(),
			interceptor.TimeoutUnaryClientInterceptor(5*time.Second),
		),
	)

	opts = append(opts,
		grpc.WithChainStreamInterceptor(
			interceptor.LogStreamClientInterceptor(),
			interceptor.BasicClientStreamInterceptor(),
			interceptor.TimeoutStreamClientInterceptor(15*time.Second),
		),
	)

	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:9090", opts...)
	if err != nil {
		log.Fatalln("Erro ao conectar com o servidor gRPC, err:", err)
	}
	defer conn.Close()

	helloAdapter, err := hello.NewHelloAdapter(conn)
	if err != nil {
		log.Fatalln("Can not create Hello Adapter: ", err)
	}

	runSayHello(helloAdapter, "Victor Reis")

	resiliencyAdapter, err := resiliency.NewResiliencyAdapter(conn)
	if err != nil {
		log.Fatal("Erro ao criar o adapter de resiliency, err:", err)
	}

	// bankAdapter, err := bank.NewBankAdapter(conn)
	// if err != nil {
	// 	log.Fatal("Erro ao criar o adapter de bank, err:", err)
	// }

	// runTransferMultiple(bankAdapter, "7835697001zzzz", "7835697002", 5)

	// runUnaryResiliencyWithTimeout(resiliencyAdapter, 4, 15, []uint32{domainResiliency.OK}, 5*time.Second)
	// now := time.Now()
	// defer func() {
	// 	log.Println("Tempo total de execução:", time.Since(now))
	// }()

	// runServerStreamingResiliencyWithTimeout(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK}, 15*time.Second)
	// runClientStreamingResiliencyWithTimeout(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK}, 10, 2*time.Second)
	// runBiDirectionalStreamingResiliencyWithTimeout(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK}, 10, 60*time.Second)

	// === Retry Pattern ===
	// runUnaryResiliency(resiliencyAdapter, 0, 3, []uint32{domainResiliency.UNKNOWN, domainResiliency.OK})
	// runServerStreamingResiliency(resiliencyAdapter, 0, 3, []uint32{domainResiliency.UNKNOWN})
	// runClientStreamingResiliency(resiliencyAdapter, 0, 3, []uint32{domainResiliency.UNKNOWN}, 10)
	// runBiDirectionalStreamingResiliency(resiliencyAdapter, 0, 3, []uint32{domainResiliency.UNKNOWN}, 10)

	// === CIRCUIT BREAKER ===
	// for i := 0; i < 300; i++ {
	// 	runUnaryResiliencyWithCircuitBreaker(resiliencyAdapter, 0, 0, []uint32{domainResiliency.UNKNOWN, domainResiliency.OK})
	// 	time.Sleep(time.Second)
	// }

	// runUnaryResiliencyWithMetadata(resiliencyAdapter, 0, 3, []uint32{domainResiliency.OK})
	// runServerStreamingResiliencyWithMetadata(resiliencyAdapter, 0, 1, []uint32{domainResiliency.OK})
	// runClientStreamingResiliencyWithMetadata(resiliencyAdapter, 0, 1, []uint32{domainResiliency.OK}, 10)
	runBiDirectionalStreamResiliencyWithMetadata(resiliencyAdapter, 0, 1, []uint32{domainResiliency.OK}, 10)
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

func runSayHello(adapter *hello.HelloAdapter, name string) {
	greet, err := adapter.SayHello(context.Background(), name)
	if err != nil {
		log.Fatalln("Erro ao chamar o serviço de hello, err:", err)
	}

	log.Println("Resposta do serviço de hello:", greet.Message)
}

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
		log.Fatalln("Failed to call runUnaryResiliencyWithTimeout: ", err)
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

// Retry Pattern
func runUnaryResiliency(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32) {
	res, err := adapter.UnaryResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
	if err != nil {
		log.Fatalln("Failed to call UnaryResiliency: ", err)
	}

	log.Println(res.DummyString)
}

func runServerStreamingResiliency(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32) {
	adapter.ServerStreamingResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
}

func runClientStreamingResiliency(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int) {
	adapter.ClientStreamResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes, count)
}

func runBiDirectionalStreamingResiliency(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32, count int) {
	adapter.BidirectionalStreamingResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes, count)
}

func runUnaryResiliencyWithCircuitBreaker(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32) {
	cBreakerRes, cBreakerErr := circuitBreaker.Execute(
		func() (interface{}, error) {
			return adapter.UnaryResiliency(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
		},
	)

	if cBreakerErr != nil {
		log.Println("Failed to call Circuit Breaker UnaryResiliency:", cBreakerErr)
	} else {
		log.Println(cBreakerRes.(*reslProto.ResiliencyReponse).DummyString)
	}
}

func runUnaryResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32) {
	res, err := adapter.UnaryResiliencyWithMetadata(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
	if err != nil {
		log.Fatalln("Failed to call runUnaryResiliencyWithMetadata: ", err)
	}

	log.Println(res.DummyString)
}

func runServerStreamingResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32) {
	adapter.ServerStreamingResiliencyWithMetadata(context.Background(), minDelaySecond, maxDelaySecond, statusCodes)
}

func runClientStreamingResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int) {
	adapter.ClientStreamResiliencyWithMetadata(context.Background(), minDelaySecond, maxDelaySecond, statusCodes, count)
}

func runBiDirectionalStreamResiliencyWithMetadata(adapter *resiliency.ResiliencyAdapter, minDelaySecond, maxDelaySecond int32, statusCodes []uint32, count int) {
	adapter.BidirectionalStreamResiliencyWithMetadata(context.Background(), minDelaySecond, maxDelaySecond, statusCodes, count)
}
