// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: quizsession.proto

package gen

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	QuizSessionService_CreateQuizSession_FullMethodName  = "/QuizSessionService/CreateQuizSession"
	QuizSessionService_GetQuizSessionById_FullMethodName = "/QuizSessionService/GetQuizSessionById"
)

// QuizSessionServiceClient is the client API for QuizSessionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type QuizSessionServiceClient interface {
	CreateQuizSession(ctx context.Context, in *CreateQuizSessionRequest, opts ...grpc.CallOption) (*CreateQuizSessionResponse, error)
	GetQuizSessionById(ctx context.Context, in *GetQuizSessionByIdRequest, opts ...grpc.CallOption) (*GetQuizSessionByIdResponse, error)
}

type quizSessionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewQuizSessionServiceClient(cc grpc.ClientConnInterface) QuizSessionServiceClient {
	return &quizSessionServiceClient{cc}
}

func (c *quizSessionServiceClient) CreateQuizSession(ctx context.Context, in *CreateQuizSessionRequest, opts ...grpc.CallOption) (*CreateQuizSessionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateQuizSessionResponse)
	err := c.cc.Invoke(ctx, QuizSessionService_CreateQuizSession_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *quizSessionServiceClient) GetQuizSessionById(ctx context.Context, in *GetQuizSessionByIdRequest, opts ...grpc.CallOption) (*GetQuizSessionByIdResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetQuizSessionByIdResponse)
	err := c.cc.Invoke(ctx, QuizSessionService_GetQuizSessionById_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QuizSessionServiceServer is the server API for QuizSessionService service.
// All implementations must embed UnimplementedQuizSessionServiceServer
// for forward compatibility.
type QuizSessionServiceServer interface {
	CreateQuizSession(context.Context, *CreateQuizSessionRequest) (*CreateQuizSessionResponse, error)
	GetQuizSessionById(context.Context, *GetQuizSessionByIdRequest) (*GetQuizSessionByIdResponse, error)
	mustEmbedUnimplementedQuizSessionServiceServer()
}

// UnimplementedQuizSessionServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedQuizSessionServiceServer struct{}

func (UnimplementedQuizSessionServiceServer) CreateQuizSession(context.Context, *CreateQuizSessionRequest) (*CreateQuizSessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQuizSession not implemented")
}
func (UnimplementedQuizSessionServiceServer) GetQuizSessionById(context.Context, *GetQuizSessionByIdRequest) (*GetQuizSessionByIdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQuizSessionById not implemented")
}
func (UnimplementedQuizSessionServiceServer) mustEmbedUnimplementedQuizSessionServiceServer() {}
func (UnimplementedQuizSessionServiceServer) testEmbeddedByValue()                            {}

// UnsafeQuizSessionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QuizSessionServiceServer will
// result in compilation errors.
type UnsafeQuizSessionServiceServer interface {
	mustEmbedUnimplementedQuizSessionServiceServer()
}

func RegisterQuizSessionServiceServer(s grpc.ServiceRegistrar, srv QuizSessionServiceServer) {
	// If the following call pancis, it indicates UnimplementedQuizSessionServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&QuizSessionService_ServiceDesc, srv)
}

func _QuizSessionService_CreateQuizSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateQuizSessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QuizSessionServiceServer).CreateQuizSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QuizSessionService_CreateQuizSession_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QuizSessionServiceServer).CreateQuizSession(ctx, req.(*CreateQuizSessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _QuizSessionService_GetQuizSessionById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetQuizSessionByIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QuizSessionServiceServer).GetQuizSessionById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: QuizSessionService_GetQuizSessionById_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QuizSessionServiceServer).GetQuizSessionById(ctx, req.(*GetQuizSessionByIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// QuizSessionService_ServiceDesc is the grpc.ServiceDesc for QuizSessionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var QuizSessionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "QuizSessionService",
	HandlerType: (*QuizSessionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateQuizSession",
			Handler:    _QuizSessionService_CreateQuizSession_Handler,
		},
		{
			MethodName: "GetQuizSessionById",
			Handler:    _QuizSessionService_GetQuizSessionById_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "quizsession.proto",
}