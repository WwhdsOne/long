package vote

import (
	"context"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

const bossClickLuaSource = `
local delta = tonumber(ARGV[1])
local nickname = ARGV[2]
local now = ARGV[3]
local expectedBossID = ARGV[4]
local defeatedAt = ARGV[5]

local buttonCount = redis.call("HINCRBY", KEYS[1], "count", delta)
local userCount = redis.call("HINCRBY", KEYS[2], "click_count", delta)
redis.call("HSET", KEYS[2], "nickname", nickname, "updated_at", now)
redis.call("ZINCRBY", KEYS[3], delta, nickname)
redis.call("ZADD", KEYS[4], now, nickname)

local bossID = redis.call("HGET", KEYS[5], "id")
local bossStatus = redis.call("HGET", KEYS[5], "status")
if not bossID or bossID ~= expectedBossID or bossStatus ~= "active" then
  return {0, buttonCount, userCount}
end

local bossTemplateID = redis.call("HGET", KEYS[5], "template_id") or ""
local bossName = redis.call("HGET", KEYS[5], "name") or expectedBossID
local maxHP = tonumber(redis.call("HGET", KEYS[5], "max_hp") or "0")
local currentHP = tonumber(redis.call("HGET", KEYS[5], "current_hp") or "0")
local startedAt = redis.call("HGET", KEYS[5], "started_at") or ""

currentHP = currentHP - delta
if currentHP < 0 then
  currentHP = 0
end
if currentHP == 0 then
  bossStatus = "defeated"
end

redis.call("ZINCRBY", KEYS[6], delta, nickname)

 local updateArgs = {"id", bossID, "name", bossName, "status", bossStatus, "max_hp", tostring(maxHP), "current_hp", tostring(currentHP)}
 if bossTemplateID ~= "" then
  table.insert(updateArgs, "template_id")
  table.insert(updateArgs, bossTemplateID)
 end
if startedAt ~= "" then
  table.insert(updateArgs, "started_at")
  table.insert(updateArgs, startedAt)
end
if bossStatus == "defeated" then
  table.insert(updateArgs, "defeated_at")
  table.insert(updateArgs, defeatedAt)
end

redis.call("HSET", KEYS[5], unpack(updateArgs))
return {1, buttonCount, userCount, bossID, bossTemplateID, bossName, bossStatus, maxHP, currentHP, startedAt, bossStatus == "defeated" and defeatedAt or ""}
`

type luaScriptRunner interface {
	EvalSha(context.Context, string, []string, ...any) (any, error)
	ScriptLoad(context.Context, string) (string, error)
}

type redisLuaRunner struct {
	client redis.UniversalClient
}

func (r redisLuaRunner) EvalSha(ctx context.Context, sha string, keys []string, args ...any) (any, error) {
	return r.client.EvalSha(ctx, sha, keys, args...).Result()
}

func (r redisLuaRunner) ScriptLoad(ctx context.Context, script string) (string, error) {
	return r.client.ScriptLoad(ctx, script).Result()
}

type luaScriptCache struct {
	mu   sync.RWMutex
	shas map[string]string
}

func newLuaScriptCache() *luaScriptCache {
	return &luaScriptCache{
		shas: make(map[string]string),
	}
}

func (c *luaScriptCache) get(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sha, ok := c.shas[name]
	return sha, ok
}

func (c *luaScriptCache) set(name string, sha string) {
	c.mu.Lock()
	c.shas[name] = sha
	c.mu.Unlock()
}

type cachedLuaScript struct {
	name  string
	body  string
	cache *luaScriptCache
}

func newCachedLuaScript(name string, body string, cache *luaScriptCache) *cachedLuaScript {
	return &cachedLuaScript{
		name:  name,
		body:  body,
		cache: cache,
	}
}

func (s *cachedLuaScript) Run(ctx context.Context, runner luaScriptRunner, keys []string, args ...any) (any, error) {
	sha, ok := s.cache.get(s.name)
	if !ok || sha == "" {
		loadedSHA, err := runner.ScriptLoad(ctx, s.body)
		if err != nil {
			return nil, err
		}
		sha = loadedSHA
		s.cache.set(s.name, sha)
	}

	value, err := runner.EvalSha(ctx, sha, keys, args...)
	if !isNoScriptError(err) {
		return value, err
	}

	loadedSHA, loadErr := runner.ScriptLoad(ctx, s.body)
	if loadErr != nil {
		return nil, loadErr
	}
	s.cache.set(s.name, loadedSHA)
	return runner.EvalSha(ctx, loadedSHA, keys, args...)
}

func isNoScriptError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "NOSCRIPT")
}
