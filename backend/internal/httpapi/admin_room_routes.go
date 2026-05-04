package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

type adminRoomStore interface {
	ListRooms(context.Context, string) (core.RoomList, error)
	SetRoomDisplayName(context.Context, string, string) error
}

func registerAdminRoomRoutes(router route.IRouter, options Options) {
	store, ok := options.Store.(adminRoomStore)
	if !ok {
		return
	}

	router.GET("/api/admin/rooms", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}
		rooms, err := store.ListRooms(ctx, "")
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ROOM_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, adminRoomViews(rooms.Rooms))
	})

	router.PUT("/api/admin/rooms/:roomId", func(ctx context.Context, c *app.RequestContext) {
		if !isAdminAuthenticated(c, options.AdminAuthenticator) {
			writeJSON(c, consts.StatusUnauthorized, map[string]string{"error": "UNAUTHORIZED"})
			return
		}

		var body struct {
			DisplayName string `json:"displayName"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}

		roomID := c.Param("roomId")
		if err := store.SetRoomDisplayName(ctx, roomID, body.DisplayName); err != nil {
			switch {
			case errors.Is(err, core.ErrRoomNotFound):
				writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "ROOM_NOT_FOUND"})
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ROOM_SAVE_FAILED"})
			}
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			RoomID:           roomID,
			BroadcastUserAll: true,
		})

		rooms, err := store.ListRooms(ctx, "")
		if err != nil {
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ROOM_LIST_FAILED"})
			return
		}
		for _, room := range rooms.Rooms {
			if room.ID == roomID {
				writeJSON(c, consts.StatusOK, adminRoomView{
					ID:          room.ID,
					DisplayName: room.DisplayName,
				})
				return
			}
		}

		writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "ROOM_NOT_FOUND"})
	})
}

type adminRoomView struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	CurrentBossName string `json:"currentBossName,omitempty"`
	CycleEnabled bool `json:"cycleEnabled"`
}

func adminRoomViews(rooms []core.RoomInfo) []adminRoomView {
	items := make([]adminRoomView, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, adminRoomView{
			ID:              room.ID,
			DisplayName:     room.DisplayName,
			CurrentBossName: room.CurrentBossName,
			CycleEnabled:    room.CycleEnabled,
		})
	}
	return items
}
