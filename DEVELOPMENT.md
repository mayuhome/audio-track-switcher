# Audio Track Switcher - 开发说明

## 开发环境设置

### 必需工具

- Node.js 18+
- Rust (最新稳定版)
- Go 1.21+
- FFmpeg

### 首次设置

1. 克隆项目后，安装前端依赖：

```bash
npm install
```

2. 构建 Go 后端：

```bash
# Windows
npm run build:backend:win

# Linux/macOS
npm run build:backend
```

3. 启动开发服务器：

```bash
npm run tauri dev
```

## 开发工作流

### 前端开发

- 修改 `src/` 目录下的文件
- Vite 会自动热重载
- 使用 TypeScript 进行类型检查

### Rust 后端开发

- 修改 `src-tauri/src/` 目录下的文件
- Tauri 会自动重新编译
- 添加新命令需要在 `lib.rs` 中注册

### Go 后端开发

- 修改 `go-backend/main.go`
- 每次修改后需要重新构建：
  ```bash
  npm run build:backend:win  # Windows
  npm run build:backend      # Linux/macOS
  ```
- 重启 Tauri 开发服务器以使用新版本

## 调试技巧

### 前端调试

- 在 Tauri 窗口中按 F12 打开开发者工具
- 使用 `console.log()` 输出调试信息

### Rust 调试

- 查看终端输出的 Rust 日志
- 使用 `println!()` 或 `dbg!()` 宏

### Go 后端调试

- Go 程序的输出会通过 Rust 返回
- 可以在 Go 代码中使用 `fmt.Println()` 输出到 stderr
- 查看 Tauri 终端的错误输出

## 构建发布版本

```bash
# 1. 确保 Go 后端已构建
npm run build:backend:win  # 或对应平台的命令

# 2. 构建 Tauri 应用
npm run tauri build
```

输出位置：`src-tauri/target/release/`

## 项目架构

### 数据流

```
UI (React)
  ↓ invoke()
Tauri Commands (Rust)
  ↓ Command::new()
Go Backend
  ↓ exec.Command()
FFmpeg
```

### 关键文件

- `src/App.tsx` - 主 UI 组件
- `src-tauri/src/lib.rs` - Tauri 命令定义
- `go-backend/main.go` - FFmpeg 集成逻辑

## 添加新功能

### 添加新的 FFmpeg 功能

1. 在 `go-backend/main.go` 中添加新函数
2. 在 `main()` 的 switch 中添加新命令
3. 在 `src-tauri/src/lib.rs` 中添加对应的 Tauri 命令
4. 在 `src/App.tsx` 中调用新命令

### 示例：添加获取视频时长功能

**Go 后端 (go-backend/main.go):**

```go
func GetVideoDuration(videoPath string) (float64, error) {
    // FFmpeg 实现
}

// 在 main() 中添加
case "get-duration":
    // 处理逻辑
```

**Rust (src-tauri/src/lib.rs):**

```rust
#[tauri::command]
async fn get_video_duration(video_path: String) -> Result<f64, String> {
    // 调用 Go 后端
}
```

**React (src/App.tsx):**

```typescript
const duration = await invoke<number>('get_video_duration', {
  videoPath: path
});
```

## 常见问题

### Go 后端找不到

- 确保已运行构建脚本
- 检查 `src-tauri/` 目录下是否有 `audio-track-backend.exe` (Windows) 或 `audio-track-backend` (Linux/macOS)

### FFmpeg 命令失败

- 检查 FFmpeg 是否在 PATH 中
- 在终端运行 `ffmpeg -version` 验证

### Tauri 编译错误

- 运行 `cargo clean` 清理缓存
- 删除 `src-tauri/target` 目录后重新构建
