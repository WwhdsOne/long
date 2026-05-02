// Package publicfs 提供编译时嵌入的前端静态资源(由 vite build 输出到 backend/public/)。
package publicfs

import "embed"

// FS 包含前端构建产物。
//
//go:embed public
var FS embed.FS
