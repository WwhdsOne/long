#!/usr/bin/env bash
set -euo pipefail

# 批量迁移玩家房间（Redis 直改）
# 依赖：redis-cli
#
# 用法示例：
#   bash scripts/move_players_room.sh \
#     --host 127.0.0.1 --port 6379 --db 0 \
#     --prefix hai-world \
#     --room 2 \
#     --nicknames "阿明,小红,Tom"
#
# 可选：
#   --password xxx            Redis 密码
#   --username xxx            Redis 用户名（ACL）
#   --tls                     启用 TLS
#   --clear-cooldown          清理 player:room:cd:*（默认开启）
#   --keep-cooldown           保留原冷却
#   --sync-afk-room           同步挂机 afk_room_id
#   --dry-run                 只打印不写入

HOST=""
PORT="6379"
DB="0"
USERNAME=""
PASSWORD=""
USE_TLS="0"
PREFIX="hai-world"
TARGET_ROOM=""
NICKNAMES_RAW=""
CLEAR_COOLDOWN="1"
SYNC_AFK_ROOM="0"
DRY_RUN="0"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) HOST="${2:-}"; shift 2 ;;
    --port) PORT="${2:-}"; shift 2 ;;
    --db) DB="${2:-}"; shift 2 ;;
    --username) USERNAME="${2:-}"; shift 2 ;;
    --password) PASSWORD="${2:-}"; shift 2 ;;
    --tls) USE_TLS="1"; shift 1 ;;
    --prefix) PREFIX="${2:-}"; shift 2 ;;
    --room) TARGET_ROOM="${2:-}"; shift 2 ;;
    --nicknames) NICKNAMES_RAW="${2:-}"; shift 2 ;;
    --clear-cooldown) CLEAR_COOLDOWN="1"; shift 1 ;;
    --keep-cooldown) CLEAR_COOLDOWN="0"; shift 1 ;;
    --sync-afk-room) SYNC_AFK_ROOM="1"; shift 1 ;;
    --dry-run) DRY_RUN="1"; shift 1 ;;
    -h|--help)
      sed -n '1,40p' "$0"
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$HOST" ]]; then
  echo "缺少 --host" >&2
  exit 1
fi
if [[ -z "$TARGET_ROOM" ]]; then
  echo "缺少 --room" >&2
  exit 1
fi
if [[ -z "$NICKNAMES_RAW" ]]; then
  echo "缺少 --nicknames" >&2
  exit 1
fi

if ! command -v redis-cli >/dev/null 2>&1; then
  echo "未找到 redis-cli，请先安装" >&2
  exit 1
fi

if [[ "$PREFIX" != *":" ]]; then
  PREFIX="${PREFIX}:"
fi

IFS=',' read -r -a NICK_ARR <<< "$NICKNAMES_RAW"

trim() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
}

REDIS_ARGS=( -h "$HOST" -p "$PORT" -n "$DB" --raw )
if [[ -n "$USERNAME" ]]; then
  REDIS_ARGS+=( --user "$USERNAME" )
fi
if [[ -n "$PASSWORD" ]]; then
  REDIS_ARGS+=( -a "$PASSWORD" )
fi
if [[ "$USE_TLS" == "1" ]]; then
  REDIS_ARGS+=( --tls )
fi

run_redis() {
  if [[ "$DRY_RUN" == "1" ]]; then
    echo "[dry-run] redis-cli ${REDIS_ARGS[*]} $*"
    return 0
  fi
  redis-cli "${REDIS_ARGS[@]}" "$@"
}

echo "开始迁移：prefix=${PREFIX} target_room=${TARGET_ROOM} clear_cooldown=${CLEAR_COOLDOWN} sync_afk_room=${SYNC_AFK_ROOM} dry_run=${DRY_RUN}"

success=0
skip=0
fail=0

for raw in "${NICK_ARR[@]}"; do
  nickname="$(trim "$raw")"
  if [[ -z "$nickname" ]]; then
    ((skip+=1))
    continue
  fi

  room_key="${PREFIX}player:room:${nickname}"
  cd_key="${PREFIX}player:room:cd:${nickname}"
  afk_key="${PREFIX}afk:player:${nickname}"

  echo "迁移玩家: ${nickname}"
  if ! run_redis SET "$room_key" "$TARGET_ROOM" >/dev/null; then
    echo "  失败: 设置房间失败 key=${room_key}" >&2
    ((fail+=1))
    continue
  fi

  if [[ "$CLEAR_COOLDOWN" == "1" ]]; then
    run_redis DEL "$cd_key" >/dev/null || true
  fi

  if [[ "$SYNC_AFK_ROOM" == "1" ]]; then
    run_redis HSET "$afk_key" afk_room_id "$TARGET_ROOM" >/dev/null || true
  fi

  ((success+=1))
done

echo "完成: success=${success} skip=${skip} fail=${fail}"

