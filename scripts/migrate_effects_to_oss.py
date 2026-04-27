#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
将本地像素特效资源批量迁移到阿里云 OSS（同名覆盖）。

特点：
1. 默认读取 pixel-assets/output 下的 PNG。
2. 上传到 OSS 前缀 effects/ 下，同名直接覆盖。
3. 默认清理本地图片（保留 prompt/json/spec 等非图片文件）。
4. 生成迁移清单与 URL 映射，便于后续前端接入。
5. 追加模式：
   - oss-url-map.json 会读取旧内容并合并新内容。
   - oss-manifest.json 会保留历史 runs，不覆盖旧迁移记录。
"""

from __future__ import annotations

import datetime as dt
import hashlib
import json
import mimetypes
import os
import sys
from pathlib import Path
from types import SimpleNamespace
from typing import Any

DEFAULT_SOURCE_DIR = "pixel-assets/output"
DEFAULT_MANIFEST_PATH = "pixel-assets/oss-manifest.json"
DEFAULT_URL_MAP_PATH = "pixel-assets/oss-url-map.json"
DEFAULT_PREFIX = "effects"
DEFAULT_CLEANUP_DIRS = ("frontend/public/effects",)


def prompt_input(title: str, default: str) -> str:
    value = input(f"{title}（默认：{default}）：").strip()
    return value or default


def prompt_yes_no(title: str, default_yes: bool) -> bool:
    default = "Y/n" if default_yes else "y/N"
    raw = input(f"{title}（{default}）：").strip().lower()
    if not raw:
        return default_yes
    return raw in {"y", "yes", "1", "true"}


def interactive_options() -> SimpleNamespace:
    default_prefix = os.getenv("OSS_PREFIX", DEFAULT_PREFIX)

    print("")
    print("=== OSS 迁移菜单 ===")
    print("1. 正式迁移（上传 + 删除本地图片，推荐）")
    print("2. 正式迁移（上传 + 保留本地图片）")
    print("3. 仅预览（dry-run，不上传不删除）")
    print("4. 自定义配置")
    print("0. 退出")

    choice = input("请选择操作：").strip()

    base = SimpleNamespace(
        source_dirs=[DEFAULT_SOURCE_DIR],
        pattern="*.png",
        prefix=default_prefix,
        manifest=DEFAULT_MANIFEST_PATH,
        url_map=DEFAULT_URL_MAP_PATH,
        cache_control="public, max-age=300",
        cleanup_dirs=list(DEFAULT_CLEANUP_DIRS),
        dry_run=False,
        delete_local=True,
    )

    if choice == "0":
        print("已取消。")
        raise SystemExit(0)

    if choice == "1":
        return base

    if choice == "2":
        base.delete_local = False
        return base

    if choice == "3":
        base.dry_run = True
        return base

    if choice != "4":
        print("无效选项，默认按 1 执行。")
        return base

    source_dirs_raw = prompt_input("源目录（多个用逗号分隔）", DEFAULT_SOURCE_DIR)
    source_dirs = [p.strip() for p in source_dirs_raw.split(",") if p.strip()]

    cleanup_dirs_raw = prompt_input(
        "清理目录（多个用逗号分隔）",
        ",".join(DEFAULT_CLEANUP_DIRS),
    )
    cleanup_dirs = [p.strip() for p in cleanup_dirs_raw.split(",") if p.strip()]

    base.source_dirs = source_dirs or [DEFAULT_SOURCE_DIR]
    base.pattern = prompt_input("匹配文件模式", "*.png")
    base.prefix = prompt_input("OSS 对象前缀", default_prefix)
    base.manifest = prompt_input("迁移清单输出路径", DEFAULT_MANIFEST_PATH)
    base.url_map = prompt_input("URL 映射输出路径", DEFAULT_URL_MAP_PATH)
    base.cache_control = prompt_input("Cache-Control", "public, max-age=300")
    base.cleanup_dirs = cleanup_dirs or list(DEFAULT_CLEANUP_DIRS)
    base.dry_run = prompt_yes_no("是否 dry-run（仅预览）", False)
    base.delete_local = prompt_yes_no("上传后是否删除本地图片", True)

    return base


def require_env(name: str) -> str:
    value = os.getenv(name, "").strip()
    if not value:
        print(f"缺少环境变量：{name}", file=sys.stderr)
        raise SystemExit(2)
    return value


def normalize_endpoint(endpoint: str) -> str:
    raw = endpoint.strip()
    if not raw:
        return raw
    if raw.startswith("http://") or raw.startswith("https://"):
        return raw
    return f"https://{raw}"


def endpoint_host(endpoint: str) -> str:
    normalized = normalize_endpoint(endpoint)
    return normalized.split("://", 1)[-1].strip("/")


def derive_public_base_url(bucket: str, endpoint: str, explicit_base: str) -> str:
    if explicit_base.strip():
        return explicit_base.strip().rstrip("/")

    host = endpoint_host(endpoint)
    if host.startswith(f"{bucket}."):
        return f"https://{host}"

    return f"https://{bucket}.{host}"


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()

    with path.open("rb") as f:
        while True:
            chunk = f.read(1024 * 1024)
            if not chunk:
                break
            h.update(chunk)

    return h.hexdigest()


def collect_files(source_dirs: list[Path], pattern: str) -> list[tuple[Path, Path]]:
    collected: list[tuple[Path, Path]] = []

    for src in source_dirs:
        if not src.exists() or not src.is_dir():
            continue

        for file_path in sorted(src.rglob(pattern)):
            if file_path.is_file():
                rel = file_path.relative_to(src)
                collected.append((src, rel))

    return collected


def build_object_key(prefix: str, rel: Path) -> str:
    cleaned_prefix = prefix.strip("/")
    rel_posix = rel.as_posix().lstrip("/")

    if cleaned_prefix:
        return f"{cleaned_prefix}/{rel_posix}"

    return rel_posix


def ensure_parent(path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)


def read_json_file(path: Path, default: Any) -> Any:
    if not path.exists():
        return default

    try:
        with path.open("r", encoding="utf-8") as f:
            return json.load(f)
    except json.JSONDecodeError:
        print(f"警告：{path} 不是合法 JSON，将使用默认空结构。", file=sys.stderr)
        return default


def load_url_map(path: Path) -> dict[str, str]:
    data = read_json_file(path, {})

    if not isinstance(data, dict):
        print(f"警告：{path} 不是 JSON 对象，将重置为空映射。", file=sys.stderr)
        return {}

    result: dict[str, str] = {}

    for key, value in data.items():
        if isinstance(key, str) and isinstance(value, str):
            result[key] = value

    return result


def append_manifest(path: Path, current_manifest: dict[str, Any]) -> dict[str, Any]:
    old_manifest = read_json_file(path, None)

    if old_manifest is None:
        return {
            "mode": "append",
            "runs": [current_manifest],
        }

    if isinstance(old_manifest, dict) and isinstance(old_manifest.get("runs"), list):
        old_manifest["runs"].append(current_manifest)
        old_manifest["lastGeneratedAt"] = current_manifest["generatedAt"]
        old_manifest["totalRuns"] = len(old_manifest["runs"])
        return old_manifest

    return {
        "mode": "append",
        "runs": [old_manifest, current_manifest],
        "lastGeneratedAt": current_manifest["generatedAt"],
        "totalRuns": 2,
    }


def upload_file(
    bucket: Any,
    src_path: Path,
    object_key: str,
    cache_control: str,
) -> Any:
    content_type, _ = mimetypes.guess_type(str(src_path))

    headers = {
        "Cache-Control": cache_control,
    }

    if content_type:
        headers["Content-Type"] = content_type

    return bucket.put_object_from_file(object_key, str(src_path), headers=headers)


def delete_local_variants(base_file: Path, cleanup_dirs: list[Path]) -> list[str]:
    deleted: list[str] = []

    targets = {base_file.resolve()}

    for directory in cleanup_dirs:
        targets.add((directory / base_file.name).resolve())

    for target in sorted(targets):
        if target.exists() and target.is_file():
            target.unlink()
            deleted.append(str(target))

    return deleted


def main() -> int:
    args = interactive_options()

    access_key_id = require_env("OSS_ACCESS_KEY_ID")
    access_key_secret = require_env("OSS_ACCESS_KEY_SECRET")
    bucket_name = require_env("OSS_BUCKET")
    endpoint_raw = require_env("OSS_ENDPOINT")

    endpoint = normalize_endpoint(endpoint_raw)
    public_base_url = derive_public_base_url(
        bucket_name,
        endpoint,
        os.getenv("OSS_PUBLIC_BASE_URL", ""),
    )

    source_dirs = [Path(p) for p in (args.source_dirs or [DEFAULT_SOURCE_DIR])]
    cleanup_dirs = [Path(p) for p in (args.cleanup_dirs or list(DEFAULT_CLEANUP_DIRS))]

    manifest_path = Path(args.manifest)
    url_map_path = Path(args.url_map)

    files = collect_files(source_dirs, args.pattern)

    if not files:
        print("未找到可迁移文件。")
        print(f"已检查目录：{', '.join(str(p) for p in source_dirs)}，pattern={args.pattern}")
        return 0

    print(f"待迁移文件数：{len(files)}")
    print(f"目标桶：{bucket_name}")
    print(f"Endpoint：{endpoint}")
    print(f"对象前缀：{args.prefix.strip('/') or '(根目录)'}")
    print(f"上传后删除本地图片：{'是' if args.delete_local else '否'}")
    print("URL 映射模式：追加/合并")
    print("迁移清单模式：追加 runs")

    if args.dry_run:
        print("当前为 dry-run，仅预览，不执行上传。")

    try:
        import oss2
    except ImportError:
        print("缺少依赖 oss2，请先安装：pip install oss2", file=sys.stderr)
        return 2

    auth = oss2.Auth(access_key_id, access_key_secret)
    bucket = oss2.Bucket(auth, endpoint, bucket_name)

    manifest_records: list[dict[str, Any]] = []
    url_map = load_url_map(url_map_path)

    for idx, (src_root, rel) in enumerate(files, start=1):
        src_path = src_root / rel
        object_key = build_object_key(args.prefix, rel)
        url = f"{public_base_url}/{object_key}"
        size = src_path.stat().st_size
        sha256 = sha256_file(src_path)

        print(f"[{idx}/{len(files)}] {src_path} -> {object_key}")

        etag = ""
        deleted_paths: list[str] = []

        if not args.dry_run:
            result = upload_file(
                bucket=bucket,
                src_path=src_path,
                object_key=object_key,
                cache_control=args.cache_control,
            )
            etag = str(getattr(result, "etag", "") or "")

            if args.delete_local:
                deleted_paths = delete_local_variants(src_path, cleanup_dirs)

        record = {
            "sourceRoot": str(src_root),
            "relativePath": rel.as_posix(),
            "sourcePath": str(src_path),
            "objectKey": object_key,
            "url": url,
            "size": size,
            "sha256": sha256,
            "etag": etag,
            "deletedLocalPaths": deleted_paths,
        }

        manifest_records.append(record)

        # 追加/合并模式：
        # 用相对路径做 key，避免不同目录下同名文件互相覆盖。
        url_map[rel.as_posix()] = url

        # 兼容旧前端：如果文件名没有冲突，也额外保留 filename -> url。
        # 如果你不想要旧格式，可以删除下面这一行。
        url_map[rel.name] = url

    current_manifest = {
        "generatedAt": dt.datetime.now(dt.timezone.utc).isoformat(),
        "bucket": bucket_name,
        "endpoint": endpoint,
        "publicBaseURL": public_base_url,
        "prefix": args.prefix.strip("/"),
        "pattern": args.pattern,
        "deleteLocal": args.delete_local,
        "dryRun": args.dry_run,
        "totalFiles": len(manifest_records),
        "files": manifest_records,
    }

    ensure_parent(manifest_path)
    ensure_parent(url_map_path)

    manifest = append_manifest(manifest_path, current_manifest)

    with manifest_path.open("w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)

    with url_map_path.open("w", encoding="utf-8") as f:
        json.dump(url_map, f, ensure_ascii=False, indent=2, sort_keys=True)

    print("")
    print("迁移完成。")
    print(f"清单文件：{manifest_path}")
    print(f"URL 映射：{url_map_path}")
    print("URL 映射已合并旧内容。")
    print("迁移清单已追加到 runs。")

    if args.delete_local and not args.dry_run:
        print("本地图片已按规则清理，仅保留 prompt/json/spec 等文件。")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())