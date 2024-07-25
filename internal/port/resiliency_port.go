package port

import (
	"context"

	"github.com/viquitorreis/my-grpc-proto/protogen/go/resiliency"
	"google.golang.org/grpc"
)

type ResiliencyClientPort interface {
	UnaryResiliency(ctx context.Context, in *resiliency.ResiliencyRequest, opts ...grpc.CallOption) (*resiliency.ResiliencyReponse, error)
	ServerStreamResiliency(ctx context.Context, in *resiliency.ResiliencyRequest, opts ...grpc.CallOption) (resiliency.ResiliencyService_ServerStreamResiliencyClient, error)
	ClientStreamResiliency(ctx context.Context, opts ...grpc.CallOption) (resiliency.ResiliencyService_ClientStreamResiliencyClient, error)
	BidirectionalStreamResiliency(ctx context.Context, opts ...grpc.CallOption) (resiliency.ResiliencyService_BidirectionalStreamResiliencyClient, error)
}
