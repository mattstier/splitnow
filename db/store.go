package db

import "splitnow/internal/types"

type Store struct{}

func (Store) RoomExists(roomID int) (bool, error) {
	return RoomExists(roomID)
}

func (Store) CreateMessage(roomID int, sender, content string) (types.Message, error) {
	return CreateMessage(roomID, sender, content)
}

func (Store) GetMessagesByRoom(roomID int) ([]types.Message, error) {
	return GetMessagesByRoom(roomID)
}

func (Store) CreateRoom(name string, creator int) (types.Room, error) {
	return CreateRoom(name, creator)
}

func (Store) GetAllRooms() ([]types.Room, error) {
	return GetAllRooms()
}

func (Store) GetRoomsWithName(name string) ([]types.Room, error) {
	return GetRoomsWithName(name)
}
