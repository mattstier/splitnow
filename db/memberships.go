package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"splitnow/internal/types"
)

// adds a member to a given room, joining is idempotent
func AddMember(userID, roomID int) (types.Membership, error) {
	row := Pool.QueryRow(context.Background(),
		`INSERT INTO memberships (room_id, user_id)
         VALUES ($1, $2)
		 ON CONFLICT DO NOTHING
         RETURNING room_id, user_id`,
		roomID, userID)

	var m types.Membership
	err := row.Scan(&m.RoomID, &m.UserID)
	// already a member: ON CONFLICT DO NOTHING returns no row, treat as a no-op
	if errors.Is(err, pgx.ErrNoRows) {
		return types.Membership{RoomID: roomID, UserID: userID}, types.ErrRoomAlreadyJoined 
	}
	return m, err
}
