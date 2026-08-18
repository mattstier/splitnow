package db

import "splitnow/internal/types"

type Store struct{}

func (Store) RoomExists(roomID int) (bool, error) {
	return RoomExists(roomID)
}

func (Store) CreateMessage(roomID int, sender, content string) (types.Message, error) {
	return CreateMessage(roomID, sender, content)
}

func (Store) GetMessagesByRoom(roomID, messageID, limit int) ([]types.Message, error) {
	return GetMessagesByRoom(roomID, messageID, limit)
}

func (Store) DeleteMessage(messageID, roomID int, sender string) (types.Message, error) {
	return DeleteMessage(messageID, roomID, sender)
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

func (Store) AddMember(userID, roomID int) (types.Membership, error) {
	return AddMember(userID, roomID)
}

func (Store) IsMember(userID, roomID int) (bool, error) {
	return IsMember(userID, roomID)
}

func (Store) GetUserRooms(userID int) ([]types.Room, error) {
	return GetUserRooms(userID)
}

func (Store) RemoveMember(userID, roomID int) (types.Membership, error) {
	return RemoveMember(userID, roomID)
}
