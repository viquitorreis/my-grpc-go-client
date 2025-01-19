package resiliency

import (
	"context"
	"io"
	"log"

	"github.com/viquitorreis/my-grpc-go-client/internal/port"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
)

type ResiliencyAdapter struct {
	resiliencyClientPort         port.ResiliencyClientPort
	resiliencyClientWithMetadata port.ResiliencyWithMetadataServiceClientPort
}

func NewResiliencyAdapter(conn *grpc.ClientConn) (*ResiliencyAdapter, error) {
	return &ResiliencyAdapter{
		resiliencyClientPort:         resiliency.NewResiliencyServiceClient(conn),
		resiliencyClientWithMetadata: resiliency.NewResiliencyWithMetadataServiceClient(conn),
	}, nil
}

func (a *ResiliencyAdapter) UnaryResiliency(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) (*resiliency.ResiliencyReponse, error) {
	resiliencyRequest := &resiliency.ResiliencyRequest{
		MinDelaySecond: minDelaySecond,
		MaxDelaySecond: maxDelaySecond,
		StatusCodes:    statusCodes,
	}

	res, err := a.resiliencyClientPort.UnaryResiliency(ctx, resiliencyRequest)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (a *ResiliencyAdapter) ServerStreamingResiliency(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) {
	resiliencyRequest := &resiliency.ResiliencyRequest{
		MinDelaySecond: minDelaySecond,
		MaxDelaySecond: maxDelaySecond,
		StatusCodes:    statusCodes,
	}

	resilStream, err := a.resiliencyClientPort.ServerStreamResiliency(ctx, resiliencyRequest)
	if err != nil {
		log.Fatalln("Error on ServerStreamingResiliency: ", err)
	}

	for {
		res, err := resilStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln("Error on ServerStreamingResiliency: ", err)
		}

		log.Println("ServerStreamingResiliency: ", res.DummyString)
	}
}

func (a *ResiliencyAdapter) ClientStreamResiliency(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int) {
	resilStream, err := a.resiliencyClientPort.ClientStreamResiliency(ctx)
	if err != nil {
		log.Fatalln("Error on ClientStreamResiliency 1: ", err)
	}

	for i := 0; i < count; i++ {
		// streaming request para o servidor

		test := &resiliency.ResiliencyRequest{
			MinDelaySecond: minDelaySecond,
			MaxDelaySecond: maxDelaySecond,
			StatusCodes:    statusCodes,
		}
		resilStream.Send(test)
	}

	res, err := resilStream.CloseAndRecv()
	if err != nil {
		log.Fatalln("Error on ClientStreamResiliency 2: ", err)
	}

	log.Println("ClientStreamResiliency: ", res.DummyString)
}

func (a *ResiliencyAdapter) BidirectionalStreamingResiliency(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int) {
	resilStream, err := a.resiliencyClientPort.BidirectionalStreamResiliency(ctx)
	if err != nil {
		log.Fatalln("Error on BidirectionalStreamingResiliency: ", err)
	}

	resilChan := make(chan struct{})

	// primeira goroutine vai enviar o número espec´ficiado de requisições resilientes para o servidor
	go func() {
		for i := 0; i < count; i++ {
			// streaming request para o servidor
			resiliencyRequest := &resiliency.ResiliencyRequest{
				MinDelaySecond: minDelaySecond,
				MaxDelaySecond: maxDelaySecond,
				StatusCodes:    statusCodes,
			}

			resilStream.Send(resiliencyRequest)
		}

		resilStream.CloseSend()
	}()

	// segunda goroutine vai receber as respostas do servidor
	go func() {
		for {
			res, err := resilStream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalln("Error on BidirectionalStreamingResiliency: ", err)
			}

			// se não tiver erro, imprime a resposta
			log.Println("BidirectionalStreamingResiliency: ", res.DummyString)
		}

		close(resilChan)
	}()

	// bloqueando até receber um valor do go channel
	<-resilChan
}
