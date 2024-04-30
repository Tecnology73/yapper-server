package main

type WelcomePacket struct {
	Name string `json:"name"`
}

type WelcomeResponsePacket struct {
	Id uint64 `json:"id"`
}

type UserConnectedPacket struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type ChatMessagePacket struct {
	Sender  uint64 `json:"sender"`
	Message string `json:"message"`
}
