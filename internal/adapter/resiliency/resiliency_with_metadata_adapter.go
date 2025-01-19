package resiliency

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func sampleRequestMetadata() metadata.MD {
	md := map[string]string{
		"grpc-client-time":  fmt.Sprintf(time.Now().Format("15:04:05")),
		"grpc-client-os":    runtime.GOOS,
		"grpc-request-uuid": uuid.New().String(),
	}

	return metadata.New(md)
}

func sampleResponseMetadata(md metadata.MD) {
	if md.Len() == 0 {
		log.Println("Response metadata not found")
	} else {
		log.Println("Response metadata: ")
		for k, v := range md {
			log.Printf(" %v: %v\n", k, v)
		}
	}
}

func (a *ResiliencyAdapter) UnaryResiliencyWithMetadata(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) (*resiliency.ResiliencyReponse, error) {
	ctx = metadata.NewOutgoingContext(ctx, sampleRequestMetadata())

	resiliencyRequest := &resiliency.ResiliencyRequest{
		MinDelaySecond: minDelaySecond,
		MaxDelaySecond: maxDelaySecond,
		StatusCodes:    statusCodes,
	}

	var responseMetadata metadata.MD
	res, err := a.resiliencyClientWithMetadata.UnaryResiliencyWithMetadata(ctx, resiliencyRequest, grpc.Header(&responseMetadata))
	if err != nil {
		log.Println("Error on UnaryResiliencyWithMetadata: ", err)
		return nil, err
	}

	sampleResponseMetadata(responseMetadata)

	return res, err
}

func (a *ResiliencyAdapter) ServerStreamingResiliencyWithMetadata(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32) {
	ctx = metadata.NewOutgoingContext(ctx, sampleRequestMetadata())

	resiliencyRequest := &resiliency.ResiliencyRequest{
		MinDelaySecond: minDelaySecond,
		MaxDelaySecond: maxDelaySecond,
		StatusCodes:    statusCodes,
	}

	resilStream, err := a.resiliencyClientWithMetadata.ServerStreamResiliencyWithMetadata(ctx, resiliencyRequest)
	if err != nil {
		log.Fatalln("Error on ServerStreamResiliencyWithMetadata: ", err)
	}
	if err != nil {
		log.Fatalln("Error on ServerStreamingResiliencyWithMetadata: ", err)
	}

	if responseMetada, err := resilStream.Header(); err == nil {
		sampleResponseMetadata(responseMetada)
	}

	for {
		res, err := resilStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln("Error on ServerStreamResiliencyWithMetadata: ", err)
		}

		log.Println("ServerStreamResiliencyWithMetadata: ", res.DummyString)
	}
}

func (a *ResiliencyAdapter) ClientStreamResiliencyWithMetadata(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int) {
	resilStream, err := a.resiliencyClientWithMetadata.ClientStreamResiliencyWithMetadata(ctx)
	if err != nil {
		log.Fatalln("Error on ClientStreamResiliencyWithMetadata 1: ", err)
	}

	for i := 0; i < count; i++ {
		// streaming request para o servidor
		ctx = metadata.NewOutgoingContext(ctx, sampleRequestMetadata())

		test := &resiliency.ResiliencyRequest{
			MinDelaySecond: minDelaySecond,
			MaxDelaySecond: maxDelaySecond,
			StatusCodes:    statusCodes,
		}

		resilStream.Send(test)
	}

	res, err := resilStream.CloseAndRecv()
	if err != nil {
		log.Fatalln("Error on ClientStreamResiliencyWithMetadata 2: ", err)
	}

	if responseMetadata, err := resilStream.Header(); err == nil {
		sampleResponseMetadata(responseMetadata)
	}

	log.Println("ClientStreamResiliencyWithMetadata: ", res.DummyString)
}

func (a *ResiliencyAdapter) BidirectionalStreamResiliencyWithMetadata(ctx context.Context, minDelaySecond int32, maxDelaySecond int32, statusCodes []uint32, count int) {
	resilStream, err := a.resiliencyClientWithMetadata.BidirectionalStreamResiliencyWithMetadata(ctx)
	if err != nil {
		log.Fatalln("Error on BidirectionalStreamResiliencyWithMetadata: ", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if responseMetadata, err := resilStream.Header(); err == nil {
		sampleResponseMetadata(responseMetadata)
	}

	resilChan := make(chan struct{})

	// primeira goroutine vai enviar o número espec´ficiado de requisições resilientes para o servidor
	go func() {
		for i := 0; i < count; i++ {
			// streaming request para o servidor
			ctx = metadata.NewOutgoingContext(ctx, sampleRequestMetadata())

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
				log.Fatalln("Error on BidirectionalStreamResiliencyWithMetadata: ", err)
			}

			// se não tiver erro, imprime a resposta
			log.Println("BidirectionalStreamResiliencyWithMetadata: ", res.DummyString)
		}

		close(resilChan)
	}()

	// bloqueando até receber um valor do go channel
	<-resilChan
}
