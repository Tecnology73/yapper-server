package main

import (
	"errors"
	"time"
)

type Store struct {
	Users map[string]*StoreUser
	Rooms map[string]*StoreRoom
}

type StoreUser struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type StoreRoom struct {
	Name     string          `json:"name"`
	Messages []*StoreMessage `json:"messages"`
}

type StoreMessage struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

var userId = uint64(0)

func (s *Store) GetOrCreateUser(name string) *StoreUser {
	if user, ok := s.Users[name]; ok {
		return user
	}

	userId += 1
	s.Users[name] = &StoreUser{
		Id:   userId,
		Name: name,
	}

	return s.Users[name]
}

func (s *Store) NewRoom(name string) (*StoreRoom, error) {
	if _, ok := s.Rooms[name]; ok {
		return nil, errors.New("room already exists")
	}

	s.Rooms[name] = &StoreRoom{
		Name:     name,
		Messages: make([]*StoreMessage, 0),
	}

	return s.Rooms[name], nil
}

func (s *Store) GetRoom(name string) (*StoreRoom, error) {
	if room, ok := s.Rooms[name]; ok {
		return room, nil
	}

	return nil, errors.New("room does not exist")
}

func (s *Store) GetOrCreateRoom(name string) *StoreRoom {
	if room, ok := s.Rooms[name]; ok {
		return room
	}

	s.Rooms[name] = &StoreRoom{
		Name:     name,
		Messages: make([]*StoreMessage, 0),
	}

	return s.Rooms[name]
}

func (r *StoreRoom) NewMessage(sender string, content string) *StoreMessage {
	message := &StoreMessage{
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	r.Messages = append(r.Messages, message)
	return message
}
