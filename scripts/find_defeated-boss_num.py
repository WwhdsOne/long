import os
import socket
from collections import defaultdict

# boss标识 -> 展示名称
BOSS_NAMES = {
    "41-basic": "41姐",
    "Rachel-basic": "Rachel老师",
    "liuyiou-basic": "刘屹鸥"
    # 自己继续加
}

def redis_connect():
    host = os.getenv("REDIS_HOST", "localhost")
    port = int(os.getenv("REDIS_PORT", 6379))
    password = os.getenv("REDIS_PASSWORD", "")
    db = int(os.getenv("REDIS_DB", 0))

    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(15)
    s.connect((host, port))
    f = s.makefile("rb")

    def send_cmd(*args):
        parts = [f"*{len(args)}\r\n".encode()]
        for arg in args:
            arg = str(arg).encode()
            parts.append(f"${len(arg)}\r\n".encode())
            parts.append(arg + b"\r\n")
        s.sendall(b"".join(parts))

    def read_resp():
        prefix = f.read(1)

        if prefix == b"+":
            return f.readline()[:-2].decode()
        if prefix == b"-":
            raise Exception(f.readline()[:-2].decode())
        if prefix == b":":
            return int(f.readline()[:-2])
        if prefix == b"$":
            length = int(f.readline()[:-2])
            if length == -1:
                return None
            data = f.read(length)
            f.read(2)
            return data.decode()
        if prefix == b"*":
            count = int(f.readline()[:-2])
            return [read_resp() for _ in range(count)]

        raise Exception(f"Unknown RESP prefix: {prefix}")

    if password:
        send_cmd("AUTH", password)
        read_resp()

    send_cmd("SELECT", db)
    read_resp()

    return s, send_cmd, read_resp


def scan_all_boss_keys(send_cmd, read_resp):
    cursor = "0"
    keys = []

    while True:
        send_cmd("SCAN", cursor, "MATCH", "vote:boss:*", "COUNT", 5000)
        cursor, batch = read_resp()
        keys.extend(batch)

        if cursor == "0":
            break

    return keys


def parse_boss_key(key):
    prefix = "vote:boss:"
    if not key.startswith(prefix):
        return None

    body = key[len(prefix):]

    if ":" not in body:
        return None

    boss_part, _ = body.split(":", 1)

    if "-" not in boss_part:
        return None

    boss_id, num_str = boss_part.rsplit("-", 1)

    if not num_str.isdigit():
        return None

    num = int(num_str)

    # 防止把时间戳、随机ID当成击败数量
    if num > 10_000_000:
        return None

    return boss_id, num


def get_boss_defeated_counts(keys):
    result = defaultdict(int)

    for key in keys:
        parsed = parse_boss_key(key)
        if not parsed:
            continue

        boss_id, num = parsed
        result[boss_id] = max(result[boss_id], num)

    return result


if __name__ == "__main__":
    conn, send_cmd, read_resp = redis_connect()

    keys = scan_all_boss_keys(send_cmd, read_resp)
    counts = get_boss_defeated_counts(keys)

    print('在大家的努力下\n')
    for boss_id, count in sorted(counts.items(), key=lambda x: x[1], reverse=True):
        boss_name = BOSS_NAMES.get(boss_id, boss_id)
        print(f"已经击败了 {count} 只「{boss_name.strip()}」")

    conn.close()