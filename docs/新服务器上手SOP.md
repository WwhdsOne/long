# 新服务器上手 SOP

> 适用场景：同学赞助的服务器、新购 VPS、新增节点。
> 参照 x1sv（D-1581 6C6G 40G SSD, Ubuntu 24.04, Cockpit + NAT 穿透）经验编写。
> 每步完成后打 ✓，避免遗漏。

---

## Phase 0：确认基础信息

- [ ] **联系同学/服务商获取：**
  - 管理方式：Cockpit Web（端口____）/ IPMI / 直连 SSH
  - SSH 端口、用户名、密码
  - NAT 隧道地址和端口映射表
- [ ] **加入清单：**
  - 在 `~/servers.txt` 或飞书文档登记：IP/域名、端口映射、用途、联系人

---

## Phase 1：系统初始化

### 1.1 激活 SSH

- Cockpit → Terminal → 启动 sshd：`sudo systemctl start sshd && sudo systemctl enable sshd`
- 如果 Cockpit 也没有 → 联系服务商从管理面板恢复

### 1.2 配置本机 SSH

```sh
# ~/.ssh/config 添加：
Host <server-name>
    HostName <NAT-域名或IP>
    Port <SSH-端口>
    User <用户名>
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

### 1.3 设置免密登录

```sh
ssh-copy-id -p <端口> <用户>@<主机>
# 验证：
ssh <server-name> 'echo OK'
```

### 1.4 系统参数

```sh
# 时区
sudo timedatectl set-timezone Asia/Shanghai

# 更新系统
sudo apt update && sudo apt upgrade -y

# 常用工具
sudo apt install -y curl wget git htop iotop net-tools
```

---

## Phase 2：Docker 环境

### 2.1 安装 Docker

```sh
# Ubuntu 24.04 首选 apt 安装（绕过 get.docker.com 网络问题）
sudo apt install -y docker.io docker-compose-v2
sudo systemctl start docker
sudo systemctl enable docker
```

### 2.2 配置 Registry 镜像（中国大陆必需）

```python
# 用 Python 写 daemon.json，避免 sudo -S + heredoc 污染文件
sudo python3 -c "
import json
cfg = {
    'registry-mirrors': [
        'https://docker.m.daocloud.io',
        'https://docker.nju.edu.cn'
    ]
}
with open('/etc/docker/daemon.json', 'w') as f:
    json.dump(cfg, f)
"
sudo systemctl restart docker

# 验证
docker info 2>/dev/null | grep -A 2 'Registry Mirrors'
```

### 2.3 配置私有 Registry（用于 CI/CD 推送）

```sh
# 允许从腾讯云 Registry 拉取
sudo python3 -c "
import json
with open('/etc/docker/daemon.json') as f:
    cfg = json.load(f)
cfg['insecure-registries'] = ['82.157.208.173:5000']
with open('/etc/docker/daemon.json', 'w') as f:
    json.dump(cfg, f)
"
sudo systemctl restart docker

# 验证拉取
docker pull 82.157.208.173:5000/long:latest
```

---

## Phase 3：基础服务部署

### 3.1 创建共享网络

所有服务需要互相通信时：

```sh
docker network create docker-compose_app-net
```

### 3.2 Redis

```yaml
# docker-compose-redis.yaml
services:
  redis:
    image: redis:7-alpine
    container_name: my-redis
    restart: always
    ports:
      - "6379:6379"
    environment:
      - TZ=Asia/Shanghai
    volumes:
      - /home/data/redis:/data
    networks:
      - docker-compose_app-net
    command:
      - redis-server
      - "--requirepass"
      - "${REDIS_PASSWD}"
      - "--appendonly"
      - "yes"
      - "--appendfsync"
      - "everysec"
      - "--aof-use-rdb-preamble"
      - "yes"
      - "--save"
      - "900"
      - "1"
      - "--save"
      - "300"
      - "10"

networks:
  docker-compose_app-net:
    external: true
```

### 3.3 MongoDB

```yaml
# docker-compose-mongo.yaml
services:
  mongodb:
    image: mongo:7
    container_name: my-mongodb
    restart: always
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWD}
      TZ: Asia/Shanghai
    volumes:
      - /home/data/mongo:/data/db
    networks:
      - docker-compose_app-net

networks:
  docker-compose_app-net:
    external: true
```

### 3.4 Consul

```yaml
# docker-compose-consul.yaml
services:
  consul:
    image: hashicorp/consul:1.19
    container_name: my-consul
    restart: unless-stopped
    ports:
      - "8500:8500"
    volumes:
      - /home/data/consul:/consul/data
    networks:
      - docker-compose_app-net
    command:
      - agent
      - -server
      - -bootstrap-expect=1
      - -client=0.0.0.0
      - -data-dir=/consul/data
      - -ui

networks:
  docker-compose_app-net:
    external: true
```

### 3.5 容器启动失败排查

```sh
# 查看日志
docker logs <container-name>

# 检查网络
docker network inspect docker-compose_app-net

# 端口是否被占
ss -tlnp | grep <port>

# 配置是否正确（尤其是 volumes 目录权限）
ls -la /home/data/redis/

# Redi s 权限问题（UID 999）
sudo chown -R 999:999 /home/data/redis

# MongoDB 初始化慢，第一次启动可能需要 30s+
# 耐心等待，不要急着重建
```

---

## Phase 4：监控接入

### 4.1 Beszel Agent

```yaml
# docker-compose-beszel-agent.yaml
services:
  beszel-agent:
    image: henrygd/beszel-agent:latest
    container_name: beszel-agent
    restart: unless-stopped
    network_mode: host
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      KEY: "<从 Beszel Hub 获取的 KEY>"
      PORT: 45876
```

启动后，在 Beszel Hub（腾讯云 82.157.208.173:8090）的"Add System"页面填入：
```
Name: <服务器名>
Host: <IP>:45876
```

🔴 **常见陷阱：**
- 如果多个服务器用了相同的 KEY/TOKEN，Beszel Hub 会只认先连的那个，后连的会 Connection refused
- 解决办法：在 Hub SQLite 里注册新机器的 KEY/TOKEN（参见 beszel-agent-db-registration.md）
- 指纹 fingerprint 由 Docker API 自动获取，不用手动填

### 4.2 Hermes 监控脚本（用于推送飞书卡片）

脚本在本地（82.157.208.173），通过 SSH 拉取远端数据。

- 如果监控指标要包含新服务器 → 改 `~/.hermes/scripts/monitor_collect.py`
- 如果新服务器承载 Redis → 加 SSH 获取 INFO/MEMORY/SLOWLOG 的代码段
- 如果新服务器承载 MongoDB → 加对应的 mongosh 调用

---

## Phase 5：NAT 端口映射（仅穿网机需要）

### 5.1 确认端口映射表

| 内部服务 | 内部端口 | NAT 端口 |
|---------|---------|---------|
| SSH | 22 | 60533 |
| long 前端(HTTPS) | 16002 | 60535 |
| Redis | 6379 | 61531 |
| MongoDB | 27017 | 61533 |
| Consul | 8500 | 61535 |

### 5.2 更新 SSH Config

```sh
Host x1sv
    HostName <NAT-域名>
    Port <NAT-SSH端口>
    User <用户名>
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

部署脚本中所有 `SERVER_IP`/`SERVER_PORT` 改为新服务器的值。

---

## Phase 6：备份策略

### 6.1 创建并部署备份脚本

```sh
# 复制 /home/backup-data.sh 到新服务器
scp backup-data.sh <server>:/home/backup-data.sh
```

脚本内容要点：
- 备份 Redis RDB：`cp /home/data/redis/dump.rdb /data/backup/redis/`
- 备份 MongoDB：`mongodump ...`
- 备份 Consul：`consul snapshot save ...`
- 保留最近 3 份，旧的自动删除

### 6.2 设置 Cron

```sh
# 每天凌晨 3 点执行
0 3 * * * /home/backup-data.sh >> /var/log/backup.log 2>&1
```

---

## Phase 7：安全配置

### 7.1 Fail2ban（防止 SSH 爆破）

```sh
sudo apt install -y fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

如果新服务器运行 GitHub Actions 的 SCP/SSH 部署，记得把 GitHub Actions Runner 的出口 IP 加入白名单：

```sh
sudo fail2ban-client set sshd unbanip <GitHub-Runner-IP>
sudo fail2ban-client set sshd addignoreip <GitHub-Runner-IP>
```

### 7.2 防火墙（如果需要）

```sh
# 检查当前规则
sudo ufw status

# 开放必要端口（禁止 22 以外的公网暴露）
sudo ufw allow 22/tcp
```

---

## Phase 8：验证清单

- [ ] SSH 免密登录正常
- [ ] Docker 安装正常，能拉取镜像
- [ ] 时区为 Asia/Shanghai
- [ ] Redis 启动成功，密码验证通过
- [ ] MongoDB 启动成功，密码验证通过
- [ ] Consul 启动成功，UI 可访问
- [ ] Beszel Agent 在 Hub 上显示在线
- [ ] 备份脚本能执行
- [ ] 监控脚本能采集新服务器数据
- [ ] 所有服务时间同步（ntp/chrony）

---

## 常见问题

| 问题 | 原因 | 解决 |
|------|------|------|
| SSH connection refused | sshd 未启动 | Cockpit 或管理面板启动 |
| Docker pull 超时 | registry mirror 失效 | 换 mirror 或走 Registry push |
| Redis 启动后鉴权失败 | `--requirepass` 没生效 | 检查 docker-compose 中 `${REDIS_PASSWD}` 环境变量是否 export |
| Redis 数据丢失 | appendonlydir 为空时跳过 RDB | `--appendonly no` 启动加载 RDB，再 CONFIG SET appendonly yes |
| MongoDB 初始化慢 | 第一次启动需要预分配 | 等 30s+，不要重建容器 |
| Beszel 连不上 | KEY/TOKEN 重复 | 在 Hub DB 注册新 KEY |
| dameon.json 损坏 | sudo tee + heredoc 密码泄漏 | 用 Python json.dump |
| 容器时区不对 | TZ env 在 restart 时不会重新注入 | `docker compose rm -f` + `up -d` |
