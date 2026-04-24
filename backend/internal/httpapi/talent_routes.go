package httpapi

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/route"

	"long/internal/vote"
)

type talentSelectRequest struct {
	Tree    string `json:"tree"`
	SubTree string `json:"subTree,omitempty"`
}

type talentLearnRequest struct {
	TalentID string `json:"talentId"`
}

type talentAPI struct {
	store ButtonStore
	auth  PlayerAuthenticator
}

func registerTalentRoutes(router route.IRouter, options Options) {
	h := &talentAPI{
		store: options.Store,
		auth:  options.PlayerAuthenticator,
	}

	talentGroup := router.Group("/api/talents")
	talentGroup.Use(requireTalentAuth(h.auth))
	{
		talentGroup.POST("/select", h.selectTree)
		talentGroup.GET("/state", h.getState)
		talentGroup.POST("/learn", h.learn)
		talentGroup.POST("/reset", h.reset)
		talentGroup.GET("/defs", h.getDefs)
	}
}

func requireTalentAuth(auth PlayerAuthenticator) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		nickname, ok := requireAuthenticatedPlayerNickname(ctx, c, auth)
		if !ok {
			c.Abort()
			return
		}
		c.Set("nickname", nickname)
	}
}

func (h *talentAPI) selectTree(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	var req talentSelectRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	tree := vote.TalentTree(req.Tree)
	subTree := vote.TalentTree(req.SubTree)

	if err := h.store.SelectTalentTree(ctx, nickStr, tree, subTree); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
}

func (h *talentAPI) getState(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	state, err := h.store.GetTalentState(ctx, nickStr)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	treeDefs := vote.GetTreeTalents(state.Tree)
	subDefs := []vote.TalentDef{}
	if state.SubTree != "" {
		subDefs = vote.GetTreeTalents(state.SubTree)
	}

	c.JSON(consts.StatusOK, map[string]any{
		"tree":     state.Tree,
		"subTree":  state.SubTree,
		"talents":  state.Talents,
		"treeDefs": treeDefs,
		"subDefs":  subDefs,
	})
}

func (h *talentAPI) learn(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	var req talentLearnRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if err := h.store.LearnTalent(ctx, nickStr, req.TalentID); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
}

func (h *talentAPI) reset(ctx context.Context, c *app.RequestContext) {
	nickname, _ := c.Get("nickname")
	nickStr, _ := nickname.(string)

	if err := h.store.ResetTalents(ctx, nickStr); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, map[string]string{"status": "ok"})
}

func (h *talentAPI) getDefs(ctx context.Context, c *app.RequestContext) {
	result := map[string]any{
		"trees": map[string]any{
			"normal": map[string]any{
				"name":    "均衡攻势",
				"talents": defsToMap(vote.GetTreeTalents(vote.TalentTreeNormal)),
			},
			"armor": map[string]any{
				"name":    "碎盾攻坚",
				"talents": defsToMap(vote.GetTreeTalents(vote.TalentTreeArmor)),
			},
			"crit": map[string]any{
				"name":    "致命洞察",
				"talents": defsToMap(vote.GetTreeTalents(vote.TalentTreeCrit)),
			},
		},
	}
	c.JSON(consts.StatusOK, result)
}

func defsToMap(defs []vote.TalentDef) []map[string]any {
	result := make([]map[string]any, 0, len(defs))
	for _, d := range defs {
		result = append(result, map[string]any{
			"id":           d.ID,
			"tree":         d.Tree,
			"tier":         d.Tier,
			"name":         d.Name,
			"effectType":   d.EffectType,
			"effectValue":  d.EffectValue,
			"prerequisite": d.Prerequisite,
		})
	}
	return result
}
