package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// GrpcStreamManager handles gRPC connections and streaming
type GrpcStreamManager struct {
	app           AppInterface
	connections   map[string]*GrpcConnection
	mu            sync.RWMutex
	protoRegistry *ProtoRegistry
}

// GrpcConnection holds an active gRPC connection
type GrpcConnection struct {
	ID         string
	ServerURL  string
	Conn       *grpc.ClientConn
	Stub       grpcdynamic.Stub
	MethodDesc *desc.MethodDescriptor
	StreamType string // "unary", "server", "client", "bidi"
	Context    context.Context
	Cancel     context.CancelFunc
	Metadata   metadata.MD
}

// ProtoRegistry keeps track of proto files and descriptors
type ProtoRegistry struct {
	files    map[string]*desc.FileDescriptor
	services map[string]*desc.ServiceDescriptor
	mu       sync.RWMutex
}


type ProtoFileUploadRequest struct {
	Files []ProtoFile `json:"files"`
}

type ProtoFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ParsedProtoResponse struct {
	Services []ServiceInfo `json:"services"`
}

type ServiceInfo struct {
	Name    string       `json:"name"`
	Methods []MethodInfo `json:"methods"`
}

type MethodInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // "unary", "server", "client", "bidi"
	InputType  string `json:"inputType"`
	OutputType string `json:"outputType"`
}

type GrpcConnectRequest struct {
	ServerURL   string            `json:"serverUrl"`
	Service     string            `json:"service"`
	Method      string            `json:"method"`
	UseTLS      bool              `json:"useTLS"`
	Deadline    int               `json:"deadline"` // milliseconds
	Compression string            `json:"compression"`
	Metadata    map[string]string `json:"metadata"`
}

type GrpcSendMessageRequest struct {
	ConnectionID string `json:"connectionId"`
	Message      string `json:"message"` // JSON string
}

func NewGrpcStreamManager(app AppInterface) *GrpcStreamManager {
	return &GrpcStreamManager{
		app:         app,
		connections: make(map[string]*GrpcConnection),
		protoRegistry: &ProtoRegistry{
			files:    make(map[string]*desc.FileDescriptor),
			services: make(map[string]*desc.ServiceDescriptor),
		},
	}
}

func (g *GrpcStreamManager) ParseProtoFiles(req ProtoFileUploadRequest) (*ParsedProtoResponse, error) {
	g.protoRegistry.mu.Lock()
	defer g.protoRegistry.mu.Unlock()

	parser := protoparse.Parser{
		ImportPaths:           []string{},
		IncludeSourceCodeInfo: true,
	}

	var protoFileNames []string
	fileContents := make(map[string]string)

	for _, file := range req.Files {
		protoFileNames = append(protoFileNames, file.Name)
		fileContents[file.Name] = file.Content
	}

	parser.Accessor = protoparse.FileContentsFromMap(fileContents)

	fileDescriptors, err := parser.ParseFiles(protoFileNames...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto files: %w", err)
	}

	for _, fd := range fileDescriptors {
		g.protoRegistry.files[fd.GetName()] = fd

		for _, svc := range fd.GetServices() {
			g.protoRegistry.services[svc.GetFullyQualifiedName()] = svc
		}
	}

	response := &ParsedProtoResponse{
		Services: []ServiceInfo{},
	}

	for _, svc := range g.protoRegistry.services {
		serviceInfo := ServiceInfo{
			Name:    svc.GetName(),
			Methods: []MethodInfo{},
		}

		for _, method := range svc.GetMethods() {
			methodType := getMethodType(method)

			serviceInfo.Methods = append(serviceInfo.Methods, MethodInfo{
				Name:       method.GetName(),
				Type:       methodType,
				InputType:  method.GetInputType().GetFullyQualifiedName(),
				OutputType: method.GetOutputType().GetFullyQualifiedName(),
			})
		}

		response.Services = append(response.Services, serviceInfo)
	}

	return response, nil
}

func (g *GrpcStreamManager) UseReflection(serverURL string, useTLS bool) (*ParsedProtoResponse, error) {
	var opts []grpc.DialOption
	if useTLS {
		creds := credentials.NewTLS(nil)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(serverURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	refClient := grpcreflect.NewClient(context.Background(), grpc_reflection_v1alpha.NewServerReflectionClient(conn))
	defer refClient.Reset()

	services, err := refClient.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	response := &ParsedProtoResponse{
		Services: []ServiceInfo{},
	}

	g.protoRegistry.mu.Lock()
	defer g.protoRegistry.mu.Unlock()

	for _, serviceName := range services {
		if serviceName == "grpc.reflection.v1alpha.ServerReflection" {
			continue
		}

		svcDesc, err := refClient.ResolveService(serviceName)
		if err != nil {
			continue
		}

		g.protoRegistry.services[serviceName] = svcDesc

		serviceInfo := ServiceInfo{
			Name:    svcDesc.GetName(),
			Methods: []MethodInfo{},
		}

		for _, method := range svcDesc.GetMethods() {
			methodType := getMethodType(method)

			serviceInfo.Methods = append(serviceInfo.Methods, MethodInfo{
				Name:       method.GetName(),
				Type:       methodType,
				InputType:  method.GetInputType().GetFullyQualifiedName(),
				OutputType: method.GetOutputType().GetFullyQualifiedName(),
			})
		}

		response.Services = append(response.Services, serviceInfo)
	}

	return response, nil
}

func (g *GrpcStreamManager) Connect(req GrpcConnectRequest) (string, error) {
	var opts []grpc.DialOption
	if req.UseTLS {
		creds := credentials.NewTLS(nil)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(req.ServerURL, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	g.protoRegistry.mu.RLock()
	svc, ok := g.protoRegistry.services[req.Service]
	g.protoRegistry.mu.RUnlock()

	if !ok {
		conn.Close()
		return "", fmt.Errorf("service not found: %s", req.Service)
	}

	methodDesc := svc.FindMethodByName(req.Method)
	if methodDesc == nil {
		conn.Close()
		return "", fmt.Errorf("method not found: %s", req.Method)
	}

	ctx := context.Background()
	if req.Deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Deadline)*time.Millisecond)
		defer cancel()
	}

	md := metadata.New(req.Metadata)
	ctx = metadata.NewOutgoingContext(ctx, md)

	stub := grpcdynamic.NewStub(conn)

	connID := fmt.Sprintf("grpc-%d", time.Now().UnixNano())
	ctx, cancel := context.WithCancel(ctx)

	grpcConn := &GrpcConnection{
		ID:         connID,
		ServerURL:  req.ServerURL,
		Conn:       conn,
		Stub:       stub,
		MethodDesc: methodDesc,
		StreamType: getMethodType(methodDesc),
		Context:    ctx,
		Cancel:     cancel,
		Metadata:   md,
	}

	g.mu.Lock()
	g.connections[connID] = grpcConn
	g.mu.Unlock()

	g.emitMessage(StreamMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Direction: "system",
		Protocol:  "gRPC",
		Payload:   fmt.Sprintf("Connected to %s/%s", req.Service, req.Method),
		Timestamp: time.Now(),
	})

	return connID, nil
}

func (g *GrpcStreamManager) SendMessage(req GrpcSendMessageRequest) error {
	g.mu.RLock()
	conn, ok := g.connections[req.ConnectionID]
	g.mu.RUnlock()

	if !ok {
		return fmt.Errorf("connection not found: %s", req.ConnectionID)
	}

	var msgData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Message), &msgData); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	inputMsg := dynamic.NewMessage(conn.MethodDesc.GetInputType())
	if err := inputMsg.UnmarshalJSON([]byte(req.Message)); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	g.emitMessage(StreamMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Direction: "outbound",
		Protocol:  "gRPC",
		Payload:   req.Message,
		Timestamp: time.Now(),
	})

	switch conn.StreamType {
	case "unary":
		return g.handleUnary(conn, inputMsg)
	case "server":
		return g.handleServerStream(conn, inputMsg)
	case "client":
		return g.handleClientStreamSend(conn, inputMsg)
	case "bidi":
		return g.handleBidiStreamSend(conn, inputMsg)
	default:
		return fmt.Errorf("unknown stream type: %s", conn.StreamType)
	}
}

func (g *GrpcStreamManager) handleUnary(conn *GrpcConnection, inputMsg *dynamic.Message) error {
	outputMsg, err := conn.Stub.InvokeRpc(conn.Context, conn.MethodDesc, inputMsg)
	if err != nil {
		g.emitMessage(StreamMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Direction: "error",
			Protocol:  "gRPC",
			Payload:   err.Error(),
			Timestamp: time.Now(),
		})
		return err
	}

	jsonData, err := outputMsg.(*dynamic.Message).MarshalJSONPB(&jsonpb.Marshaler{})
	if err != nil {
		jsonData = []byte(outputMsg.String())
	}

	g.emitMessage(StreamMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Direction: "inbound",
		Protocol:  "gRPC",
		Payload:   string(jsonData),
		Timestamp: time.Now(),
	})

	return nil
}

func (g *GrpcStreamManager) handleServerStream(conn *GrpcConnection, inputMsg *dynamic.Message) error {
	stream, err := conn.Stub.InvokeRpcServerStream(conn.Context, conn.MethodDesc, inputMsg)
	if err != nil {
		g.emitMessage(StreamMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Direction: "error",
			Protocol:  "gRPC",
			Payload:   err.Error(),
			Timestamp: time.Now(),
		})
		return err
	}

	go func() {
		for {
			outputMsg, err := stream.RecvMsg()

			if err == io.EOF {
				g.emitMessage(StreamMessage{
					ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
					Direction: "system",
					Protocol:  "gRPC",
					Payload:   "Server closed stream",
					Timestamp: time.Now(),
				})
				return
			}

			if err != nil {
				g.emitMessage(StreamMessage{
					ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
					Direction: "error",
					Protocol:  "gRPC",
					Payload:   err.Error(),
					Timestamp: time.Now(),
				})
				return
			}

			jsonData, err := outputMsg.(*dynamic.Message).MarshalJSONPB(&jsonpb.Marshaler{})
			if err != nil {
				jsonData = []byte(outputMsg.String())
			}

			g.emitMessage(StreamMessage{
				ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
				Direction: "inbound",
				Protocol:  "gRPC",
				Payload:   string(jsonData),
				Timestamp: time.Now(),
			})
		}
	}()

	return nil
}

func (g *GrpcStreamManager) handleClientStreamSend(conn *GrpcConnection, inputMsg *dynamic.Message) error {
	// TODO
	return fmt.Errorf("client streaming not yet implemented")
}

func (g *GrpcStreamManager) handleBidiStreamSend(conn *GrpcConnection, inputMsg *dynamic.Message) error {
	// TODO
	return fmt.Errorf("bidirectional streaming not yet implemented")
}

func (g *GrpcStreamManager) Disconnect(connectionID string) error {
	g.mu.Lock()
	conn, ok := g.connections[connectionID]
	if ok {
		delete(g.connections, connectionID)
	}
	g.mu.Unlock()

	if !ok {
		return fmt.Errorf("connection not found")
	}

	conn.Cancel()
	conn.Conn.Close()

	g.emitMessage(StreamMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Direction: "system",
		Protocol:  "gRPC",
		Payload:   "Disconnected",
		Timestamp: time.Now(),
	})

	return nil
}

func (g *GrpcStreamManager) emitMessage(msg StreamMessage) {
	runtime.EventsEmit(g.app.GetCtx(), "stream-message", msg)
}

func getMethodType(method *desc.MethodDescriptor) string {
	if !method.IsClientStreaming() && !method.IsServerStreaming() {
		return "unary"
	}
	if method.IsClientStreaming() && !method.IsServerStreaming() {
		return "client"
	}
	if !method.IsClientStreaming() && method.IsServerStreaming() {
		return "server"
	}
	return "bidi"
}

