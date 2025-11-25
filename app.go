package main

import (
	"context"
	"os"
	"path/filepath"
	"pulse/backend"
)

type App struct {
	ctx         context.Context
	dataDir     string
	grpcManager *backend.GrpcStreamManager
	wsManager   *backend.WebSocketManager
	sseManager  *backend.SSEManager
	httpHandler *backend.HTTPHandler
}

func NewApp() *App {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".pulse")
	os.MkdirAll(dataDir, 0755)

	app := &App{}
	app.dataDir = dataDir
	app.grpcManager = backend.NewGrpcStreamManager(app)
	app.wsManager = backend.NewWebSocketManager(app)
	app.sseManager = backend.NewSSEManager(app)
	app.httpHandler = backend.NewHTTPHandler(app, dataDir)

	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.ensureDataDirectories()
}

// SSE handler functions

func (a *App) SSEConnect(req backend.SSEConnectRequest) (string, error) {
	return a.sseManager.Connect(req)
}

func (a *App) SSEDisconnect(connectionID string) error {
	return a.sseManager.Disconnect(connectionID)
}

// WebSocket handler functions

func (a *App) WebSocketConnect(req backend.WebSocketConnectRequest) (string, error) {
	return a.wsManager.Connect(req)
}

func (a *App) WebSocketSendMessage(req backend.WebSocketSendRequest) error {
	return a.wsManager.SendMessage(req)
}

func (a *App) WebSocketDisconnect(connectionID string) error {
	return a.wsManager.Disconnect(connectionID)
}

// gRPC handler functions

func (a *App) GrpcParseProtoFiles(req backend.ProtoFileUploadRequest) (*backend.ParsedProtoResponse, error) {
	return a.grpcManager.ParseProtoFiles(req)
}

func (a *App) GrpcUseReflection(serverURL string, useTLS bool) (*backend.ParsedProtoResponse, error) {
	return a.grpcManager.UseReflection(serverURL, useTLS)
}

func (a *App) GrpcConnect(req backend.GrpcConnectRequest) (string, error) {
	return a.grpcManager.Connect(req)
}

func (a *App) GrpcSendMessage(req backend.GrpcSendMessageRequest) error {
	return a.grpcManager.SendMessage(req)
}

func (a *App) GrpcDisconnect(connectionID string) error {
	return a.grpcManager.Disconnect(connectionID)
}

// HTTP handler functions

func (a *App) SendRequest(req backend.RequestData) (*backend.ResponseData, error) {
	return a.httpHandler.SendRequest(req)
}

func (a *App) SaveWorkspaces(workspaces []backend.Workspace) error {
	return a.httpHandler.SaveWorkspaces(workspaces)
}

func (a *App) LoadWorkspaces() ([]backend.Workspace, error) {
	return a.httpHandler.LoadWorkspaces()
}

func (a *App) SaveCollections(collections []backend.Collection) error {
	return a.httpHandler.SaveCollections(collections)
}

func (a *App) LoadCollections() ([]backend.Collection, error) {
	return a.httpHandler.LoadCollections()
}

func (a *App) SaveEnvironments(environments []backend.Environment) error {
	return a.httpHandler.SaveEnvironments(environments)
}

func (a *App) LoadEnvironments() ([]backend.Environment, error) {
	return a.httpHandler.LoadEnvironments()
}

func (a *App) SaveHistory(items []backend.HistoryItem) error {
	return a.httpHandler.SaveHistory(items)
}

func (a *App) LoadHistory() ([]backend.HistoryItem, error) {
	return a.httpHandler.LoadHistory()
}

func (a *App) SaveSettings(settings backend.Settings) error {
	return a.httpHandler.SaveSettings(settings)
}

func (a *App) LoadSettings() (*backend.Settings, error) {
	return a.httpHandler.LoadSettings()
}

func (a *App) KafkaConnect(config backend.KafkaConfig) (string, error) {
	return backend.KafkaConnect(a, config)
}

func (a *App) KafkaDisconnect(connectionID string) error {
	return backend.KafkaDisconnect(a, connectionID)
}

func (a *App) KafkaListTopics(connectionID string) ([]backend.TopicInfo, error) {
	return backend.KafkaListTopics(a, connectionID)
}

func (a *App) KafkaStartConsumer(config backend.ConsumerConfig) (string, error) {
	return backend.KafkaStartConsumer(a, config)
}

func (a *App) KafkaStopConsumer(connectionID string, consumerID string) error {
	return backend.KafkaStopConsumer(a, connectionID, consumerID)
}

func (a *App) KafkaProduceMessage(config backend.ProducerConfig) error {
	return backend.KafkaProduceMessage(a, config)
}

func (a *App) EmitStreamMessage(connectionID, direction, protocol, payload string) {
	backend.EmitStreamMessage(a, connectionID, direction, protocol, payload)
}

func (a *App) GetDataDirectory() string {
	return a.httpHandler.GetDataDirectory()
}

func (a *App) GetCtx() context.Context {
	return a.ctx
}

func (a *App) ensureDataDirectories() {
	os.MkdirAll(filepath.Join(a.dataDir, "workspaces"), 0755)
	os.MkdirAll(filepath.Join(a.dataDir, "collections"), 0755)
	os.MkdirAll(filepath.Join(a.dataDir, "environments"), 0755)
	os.MkdirAll(filepath.Join(a.dataDir, "history"), 0755)
	os.MkdirAll(filepath.Join(a.dataDir, "settings"), 0755)
}






