#!/bin/bash

# ==========================================
# 批量发放原石/宝石脚本
# 使用方法: bash grant_gems.sh
# ==========================================

# ---------- 在这里修改配置 ----------
# 指定发给某个玩家（填玩家的标识，比如 Wwhds）。如果填 "ALL"，则发给所有拥有 vote:user:* 的玩家
TARGET_USER="ALL"  

# 发放的原石数量
GEMS_AMOUNT=100

# Redis 配置（一般不用动）
REDIS_PASS="****"
REDIS_DB=666
# --------------------------------------

REDIS_CMD="redis-cli -a ${REDIS_PASS} -n ${REDIS_DB}"

if [ "$TARGET_USER" = "ALL" ]; then
    echo "🚀 开始向【所有玩家】发放 ${GEMS_AMOUNT} 原石..."
    
    # 扫描所有 vote:user:* 并提取后缀生成 vote:user-gems:xxx
    $REDIS_CMD --scan --pattern 'vote:user:*' | while read key; do
        nickname="${key#vote:user:}"
        $REDIS_CMD SET "vote:user-gems:${nickname}" $GEMS_AMOUNT > /dev/null
        
        # 打印进度（避免刷屏，只打印成功状态）
        echo "✅ 已处理: vote:user-gems:${nickname}"
    done
    
    echo "🎉 全部玩家发放完毕！"

else
    echo "🚀 开始向【指定玩家 ${TARGET_USER}】发放 ${GEMS_AMOUNT} 原石..."
    
    # 检查该玩家是否存在基础数据
    check=$($REDIS_CMD EXISTS "vote:user:${TARGET_USER}")
    if [ "$check" -eq 0 ]; then
        echo "❌ 错误：找不到玩家 ${TARGET_USER} 的基础数据 (vote:user:${TARGET_USER} 不存在)"
        exit 1
    fi
    
    # 给指定玩家设置
    $REDIS_CMD SET "vote:user-gems:${TARGET_USER}" $GEMS_AMOUNT
    echo "🎉 玩家 ${TARGET_USER} 发放完毕！"
fi
