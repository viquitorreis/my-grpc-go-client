package hello

import (
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
