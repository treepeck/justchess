// Package ws implements the WebSocket server.
//
// # Overview
//
// The Manager type represents a WebSocket server or hub. A server application
// calls the NewManager method to create a new Manager.
//
// The client type stores a connection. All interactions with the
// front-end are done by the client methods.
package ws
