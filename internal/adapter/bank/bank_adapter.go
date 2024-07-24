package bank

import (
	"context"
	"fmt"
	"io"
	"log"

	domainBank "github.com/viquitorreis/my-grpc-go-client/internal/application/domain/bank"
	"github.com/viquitorreis/my-grpc-go-client/internal/port"
	protoBank "github.com/viquitorreis/my-grpc-proto/protogen/go/bank"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BankAdapter struct {
	bankClient port.ClientPort
}

func NewBankAdapter(conn *grpc.ClientConn) (*BankAdapter, error) {
	if conn == nil {
		return nil, fmt.Errorf("gRPC connection is nil")
	}

	return &BankAdapter{
		bankClient: protoBank.NewBankServiceClient(conn),
	}, nil
}

func (a *BankAdapter) GetCurrentBalance(ctx context.Context, account string) (*protoBank.CurrentBalanceResponse, error) {
	bankrequest := &protoBank.CurrentBalanceRequest{
		AccountNumber: account,
	}

	bal, err := a.bankClient.GetCurrentBalance(ctx, bankrequest)
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalln("[FATAL] failed to get current balance: ", st)
	}

	return bal, nil
}

func (a *BankAdapter) FetchExchangeRates(ctx context.Context, fromCur, toCur string) {
	if a.bankClient == nil {
		log.Fatalln("[FATAL] bankClient is nil")
	}

	bankReq := &protoBank.ExchangeRateRequest{
		FromCurrency: fromCur,
		ToCurrency:   toCur,
	}

	exchangeRateStream, err := a.bankClient.FetchExchangeRates(ctx, bankReq)
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalln("[FATAL] failed to fetch exchange rates: ", st)
	}

	for {
		rate, err := exchangeRateStream.Recv()
		if err == io.EOF {
			// loop termina quando a stream termina ou algum erro ocorre
			break
		}

		if err != nil {
			st, _ := status.FromError(err)

			if st.Code() == codes.InvalidArgument {
				log.Fatalln("[FATAL] invalid argument: ", st)
			}
		}

		log.Printf("[INFO] time: %f exchange rate: from %s to %s\n", rate.Rate, rate.FromCurrency, rate.ToCurrency)
	}
}

func (a *BankAdapter) SummarizeTransactions(ctx context.Context, account string, tx []*domainBank.Transaction) {
	txStream, err := a.bankClient.SummarizeTransactions(ctx)
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalln("[FATAL] failed to summarize transactions: ", st)
	}

	for _, t := range tx {
		tranType := protoBank.TransactionType_TRANSACTION_TYPE_UNSPECIFIED

		if t.TransactionType == domainBank.TransactionTypeIn {
			tranType = protoBank.TransactionType_TRANSACTION_TYPE_IN
		}

		if t.TransactionType == domainBank.TransactionTypeOut {
			tranType = protoBank.TransactionType_TRANSACTION_TYPE_OUT
		}

		bankReq := &protoBank.Transaction{
			AccountNumber: account,
			Type:          tranType,
			Amount:        t.Amount,
			Notes:         t.Notes,
		}

		// enviando transação para o server gRPC
		txStream.Send(bankReq)
	}

	summary, err := txStream.CloseAndRecv()
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalln("[FATAL] failed to get transaction summary: ", st)
	}

	log.Println(summary)
}

func (a *BankAdapter) TransferMultiple(ctx context.Context, trf []domainBank.TransferTransaction) {
	trfStream, err := a.bankClient.TransferMultiple(ctx)
	if err != nil {
		st, _ := status.FromError(err)
		log.Fatalln("[FATAL] failed to transfer multiple: ", st)
	}

	// channel para receber mensagens
	trfChan := make(chan struct{})

	// função envia 2 goroutines simultaneas para enviar e receber mensagens
	// 1ª goroutine vai buildar e enviar cada transferência para o server
	go func() {
		for _, tt := range trf {
			req := &protoBank.TransferRequest{
				FromAccountNumber: tt.FromAccountNumber,
				ToAccountNumber:   tt.ToAccountNumber,
				Currency:          tt.Currency,
				Amount:            tt.Amount,
			}

			trfStream.Send(req)
		}

		// depois que todas requisições forem enviadas, precisamos fechar a stream
	}()

	// 2ª goroutine vai receber mensagens do server usando o método Recv
	go func() {
		for {
			resp, err := trfStream.Recv()
			if err == io.EOF {
				// loop termina quando a stream termina ou algum erro ocorre
				break
			}

			if err != nil {
				handleTransferErrorGrpc(err)
				break
			} else {
				log.Printf("[INFO] transfer satatus: %v time: %v\n", resp.Status, resp.Timestamp)
			}

			close(trfChan)
		}
	}()

	<-trfChan
}

func handleTransferErrorGrpc(err error) {
	st := status.Convert(err)
	log.Printf("[ERROR] code: %v message: %v\n", st.Code(), st.Message())

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.PreconditionFailure:
			for _, violation := range t.GetViolations() {
				log.Println("[VIOLATION]", violation)
			}
		case *errdetails.ErrorInfo:
			log.Printf("[ERROR_INFO] reason: %v domain: %v metadata: %v\n", t.Reason, t.Domain, t.Metadata)
			for k, v := range t.GetMetadata() {
				log.Printf("[METADATA] key: %v value: %v\n", k, v)
			}
		}
	}
}
