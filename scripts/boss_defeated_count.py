import os
import socket
from collections import defaultdict
import argparse

def redis_connect(host, port, password, db):
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

    pattern = f"{KEY_PREFIX}:*-*:*"

    while True:
        send_cmd("SCAN", cursor, "MATCH", pattern, "COUNT", 5000)
        cursor, batch = read_resp()
        keys.extend(batch)

        if cursor == "0":
            break

    return keys


def parse_boss_key(key):
    prefix = f"{KEY_PREFIX}:"

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

    if num > 10_000_000:
        return None

    return boss_id, num


def hget(send_cmd, read_resp, key, field):
    send_cmd("HGET", key, field)
    return read_resp()


def get_boss_name(send_cmd, read_resp, boss_id):
    pool_key = f"{KEY_PREFIX}:pool:{boss_id}"
    name = hget(send_cmd, read_resp, pool_key, "name")

    if name:
        return name.strip()

    return boss_id


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
    parser = argparse.ArgumentParser()

    parser.add_argument("--host", required=True)
    parser.add_argument("--port", type=int, default=6379)
    parser.add_argument("--password", default="")
    parser.add_argument("--db", type=int, default=0)
    parser.add_argument("--prefix", default="vote:boss")

    args = parser.parse_args()

    global KEY_PREFIX
    KEY_PREFIX = args.prefix

    conn, send_cmd, read_resp = redis_connect(
        args.host,
        args.port,
        args.password,
        args.db
    )

    keys = scan_all_boss_keys(send_cmd, read_resp)
    counts = get_boss_defeated_counts(keys)

    print("在大家的努力下，已经击败了 :")
    for boss_id, count in sorted(counts.items(), key=lambda x: x[1], reverse=True):
        boss_name = get_boss_name(send_cmd, read_resp, boss_id)
        print(f"    {count} 只「{boss_name}」")

    conn.close()