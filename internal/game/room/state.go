package room

// RoomState 房间状态
type RoomState int

const (
	RoomStateWaiting RoomState = iota
	RoomStateReady
	RoomStateBidding
	RoomStatePlaying
	RoomStateEnded
)
