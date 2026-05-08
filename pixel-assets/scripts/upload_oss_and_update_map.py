#!/usr/bin/env python3
"""调用 ossutil 上传文件，并回写 oss-url-map.json。"""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

import oss2


REPO_ROOT = Path("/Users/Learning/web/long")
DEFAULT_SOURCE_ROOT = REPO_ROOT / "frontend" / "public" / "effects"
DEFAULT_MAP_PATH = REPO_ROOT / "pixel-assets" / "oss-url-map.json"
DEFAULT_BUCKET = "hai-world2"
DEFAULT_ENDPOINT = "oss-cn-beijing.aliyuncs.com"
DEFAULT_PREFIX = "effects"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="上传本地特效资源到阿里云 OSS，并更新 URL 映射文件")
    parser.add_argument(
        "files",
        nargs="+",
        help="要上传的文件路径，支持直接传文件路径，也支持 shell 展开的通配结果",
    )
    parser.add_argument(
        "--source-root",
        default=str(DEFAULT_SOURCE_ROOT),
        help="资源根目录，用于推导相对路径，默认 frontend/public/effects",
    )
    parser.add_argument(
        "--map-path",
        default=str(DEFAULT_MAP_PATH),
        help="要更新的 URL 映射文件，默认 pixel-assets/oss-url-map.json",
    )
    parser.add_argument(
        "--bucket",
        default=os.environ.get("OSS_BUCKET", DEFAULT_BUCKET),
        help="OSS bucket 名称，默认 hai-world2，可被环境变量 OSS_BUCKET 覆盖",
    )
    parser.add_argument(
        "--endpoint",
        default=os.environ.get("OSS_ENDPOINT", DEFAULT_ENDPOINT),
        help="OSS endpoint，默认 oss-cn-beijing.aliyuncs.com，可被环境变量 OSS_ENDPOINT 覆盖",
    )
    parser.add_argument(
        "--prefix",
        default=os.environ.get("OSS_PREFIX", DEFAULT_PREFIX),
        help="OSS 对象前缀，默认 effects，可被环境变量 OSS_PREFIX 覆盖",
    )
    parser.add_argument(
        "--public-base-url",
        default=os.environ.get("OSS_PUBLIC_BASE_URL", "").strip(),
        help="公开访问前缀；不传时自动按 bucket + endpoint 推导",
    )
    parser.add_argument(
        "--ossutil-bin",
        default=os.environ.get("OSSUTIL_BIN", "ossutil"),
        help="ossutil 可执行文件名，默认 ossutil，可被环境变量 OSSUTIL_BIN 覆盖",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="只打印将执行的上传动作，不真正调用 ossutil，也不改映射文件",
    )
    return parser.parse_args()


def load_credentials() -> tuple[str, str]:
    key_id = (
        os.environ.get("OSS_ACCESS_KEY_ID")
        or os.environ.get("ALIBABA_CLOUD_ACCESS_KEY_ID")
        or os.environ.get("ALIYUN_ACCESS_KEY_ID")
    )
    key_secret = (
        os.environ.get("OSS_ACCESS_KEY_SECRET")
        or os.environ.get("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
        or os.environ.get("ALIYUN_ACCESS_KEY_SECRET")
    )
    if not key_id or not key_secret:
        raise SystemExit(
            "缺少 OSS 密钥环境变量，请设置 OSS_ACCESS_KEY_ID / OSS_ACCESS_KEY_SECRET"
        )
    return key_id, key_secret


def normalize_endpoint(endpoint: str) -> str:
    value = endpoint.strip()
    value = value.removeprefix("https://")
    value = value.removeprefix("http://")
    return value.rstrip("/")


def resolve_public_base_url(bucket: str, endpoint: str, provided: str) -> str:
    if provided:
        return provided.rstrip("/")
    return f"https://{bucket}.{normalize_endpoint(endpoint)}"


def resolve_files(paths: list[str]) -> list[Path]:
    files: list[Path] = []
    for raw in paths:
        path = Path(raw).expanduser()
        if not path.exists():
            raise SystemExit(f"文件不存在：{path}")
        if path.is_dir():
            raise SystemExit(f"不支持直接上传目录：{path}")
        files.append(path.resolve())
    if not files:
        raise SystemExit("没有可上传的文件")
    return files


def load_url_map(map_path: Path) -> dict[str, str]:
    if not map_path.exists():
        return {}
    with map_path.open("r", encoding="utf-8") as f:
        data = json.load(f)
    if not isinstance(data, dict):
        raise SystemExit(f"映射文件格式非法：{map_path}")
    return {str(k): str(v) for k, v in data.items()}


def write_url_map(map_path: Path, data: dict[str, str]) -> None:
    map_path.parent.mkdir(parents=True, exist_ok=True)
    ordered = dict(sorted(data.items(), key=lambda item: item[0]))
    with map_path.open("w", encoding="utf-8") as f:
        json.dump(ordered, f, ensure_ascii=False, indent=2)
        f.write("\n")


def run_checked(cmd: list[str]) -> None:
    result = subprocess.run(cmd, check=False, capture_output=True, text=True)
    if result.returncode != 0:
        stderr = result.stderr.strip()
        stdout = result.stdout.strip()
        detail = stderr or stdout or f"退出码 {result.returncode}"
        raise SystemExit(f"命令执行失败：{' '.join(cmd)}\n{detail}")


def upload_with_oss2(
    file_path: Path,
    bucket: str,
    endpoint: str,
    object_key: str,
    key_id: str,
    key_secret: str,
) -> None:
    auth = oss2.Auth(key_id, key_secret)
    bucket_client = oss2.Bucket(auth, f"https://{normalize_endpoint(endpoint)}", bucket)
    headers = {"x-oss-object-acl": "public-read"}
    result = bucket_client.put_object_from_file(object_key, str(file_path), headers=headers)
    status = getattr(result, "status", 0)
    if status < 200 or status >= 300:
        raise SystemExit(f"oss2 上传失败：{file_path} -> {object_key}，HTTP {status}")


def build_object_key(file_path: Path, source_root: Path, prefix: str) -> tuple[str, str]:
    try:
        rel = file_path.relative_to(source_root)
    except ValueError:
        rel = Path(file_path.name)
    rel_posix = rel.as_posix()
    object_key = f"{prefix.strip('/')}/{rel_posix}" if prefix.strip("/") else rel_posix
    return rel_posix, object_key


def upload_files(args: argparse.Namespace) -> None:
    ossutil_path = shutil.which(args.ossutil_bin)
    use_ossutil = bool(ossutil_path)

    files = resolve_files(args.files)
    source_root = Path(args.source_root).resolve()
    map_path = Path(args.map_path).resolve()
    public_base_url = resolve_public_base_url(args.bucket, args.endpoint, args.public_base_url)
    url_map = load_url_map(map_path)

    key_id = ""
    key_secret = ""
    if not args.dry_run:
        key_id, key_secret = load_credentials()

    with tempfile.TemporaryDirectory(prefix="ossutil-config-") as tmpdir:
        config_path = Path(tmpdir) / "config"
        if not args.dry_run and use_ossutil:
            run_checked(
                [
                    ossutil_path,
                    "config",
                    "-c",
                    str(config_path),
                    "-e",
                    normalize_endpoint(args.endpoint),
                    "-i",
                    key_id,
                    "-k",
                    key_secret,
                    "-L",
                    "CH",
                ]
            )

        for file_path in files:
            _, object_key = build_object_key(file_path, source_root, args.prefix)
            target = f"oss://{args.bucket}/{object_key}"
            url = f"{public_base_url}/{object_key}"
            print(f"上传 {file_path} -> {target}")
            if not args.dry_run:
                if use_ossutil:
                    run_checked(
                        [
                            ossutil_path,
                            "cp",
                            str(file_path),
                            target,
                            "-c",
                            str(config_path),
                            "-u",
                        ]
                    )
                else:
                    upload_with_oss2(
                        file_path=file_path,
                        bucket=args.bucket,
                        endpoint=args.endpoint,
                        object_key=object_key,
                        key_id=key_id,
                        key_secret=key_secret,
                    )
                url_map[file_path.name] = url
            uploader = "ossutil" if use_ossutil else "oss2"
            print(f"  URL: {url} ({uploader})")

    if args.dry_run:
        print("dry-run 模式，不写映射文件")
        return

    write_url_map(map_path, url_map)
    print(f"已更新映射文件：{map_path}")


def main() -> None:
    args = parse_args()
    upload_files(args)


if __name__ == "__main__":
    main()
