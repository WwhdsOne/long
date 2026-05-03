package httpapi

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/core"
)

type roomStore interface {
	ListRooms(context.Context, string) (core.RoomList, error)
	SwitchPlayerRoom(context.Context, string, string) (core.RoomSwitchResult, error)
}

func registerRoomRoutes(router route.IRouter, options Options) {
	store, ok := options.Store.(roomStore)
	if !ok {
		return
	}

	router.GET("/api/rooms", func(ctx context.Context, c *app.RequestContext) {
		nickname := resolvedPlayerNicknameForRead(ctx, c, options.PlayerAuthenticator)
		rooms, err := store.ListRooms(ctx, nickname)
		if err != nil {
			if writeNicknameError(c, err) {
				return
			}
			writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ROOM_LIST_FAILED"})
			return
		}
		writeJSON(c, consts.StatusOK, rooms)
	})

	router.POST("/api/rooms/join", func(ctx context.Context, c *app.RequestContext) {
		var body struct {
			Nickname string `json:"nickname"`
			RoomID   string `json:"roomId"`
		}
		if !bindJSON(c, &body, map[string]string{"error": "INVALID_REQUEST"}) {
			return
		}
		nickname, ok := resolvedPlayerNickname(ctx, c, options.PlayerAuthenticator, body.Nickname)
		if !ok {
			return
		}
		result, err := store.SwitchPlayerRoom(ctx, nickname, body.RoomID)
		if err != nil {
			switch {
			case errors.Is(err, core.ErrRoomNotFound):
				writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "ROOM_NOT_FOUND"})
			case errors.Is(err, core.ErrRoomNotJoinable):
				writeJSON(c, consts.StatusBadRequest, map[string]string{"error": "ROOM_NOT_JOINABLE"})
			case errors.Is(err, core.ErrRoomSwitchCooldown):
				writeJSON(c, consts.StatusTooManyRequests, map[string]string{"error": "ROOM_SWITCH_COOLDOWN"})
			case writeNicknameError(c, err):
			default:
				writeJSON(c, consts.StatusInternalServerError, map[string]string{"error": "ROOM_JOIN_FAILED"})
			}
			return
		}

		publishChange(ctx, options.ChangePublisher, core.StateChange{
			Type:             core.StateChangeBossChanged,
			Nickname:         nickname,
			RoomID:           result.CurrentRoomID,
			BroadcastUserAll: true,
		})
		writeJSON(c, consts.StatusOK, result)
	})
}
