package room

import (
	"github.com/palemoky/fight-the-landlord/internal/server/storage"
)

// ToRoomData 将 Room 转换为可序列化的 RoomData
func (r *Room) ToRoomData() *storage.RoomData {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := &storage.RoomData{
		Code:        r.Code,
		State:       int(r.State),
		Players:     make([]storage.PlayerData, 0, len(r.Players)),
		PlayerOrder: r.PlayerOrder,
		CreatedAt:   r.CreatedAt.Unix(),
	}

	for _, player := range r.Players {
		data.Players = append(data.Players, storage.PlayerData{
			ID:         player.Client.GetID(),
			Name:       player.Client.GetName(),
			Seat:       player.Seat,
			Ready:      player.Ready,
			IsLandlord: player.IsLandlord,
		})
	}

	// TODO: 如果需要保存游戏会话数据，在这里添加

	return data
}
