package db

type Store struct{}

func (Store) CreateMessage(roomID int, sender, content string) (Message, error) {
	return CreateMessage(roomID, sender, content)
}

func (Store) GetMessagesByRoom(roomID int) ([]Message, error) {
	return GetMessagesByRoom(roomID)
}

func (Store) CreateRoom(name string, creator int) (Room, error) {
	return CreateRoom(name, creator)
}

func (Store) GetAllRooms() ([]Room, error) {
	return GetAllRooms()
}

func (Store) GetRoomsWithName(name string) ([]Room, error) {
	return GetRoomsWithName(name)
}
