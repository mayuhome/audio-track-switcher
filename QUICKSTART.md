# 🚀 快速开始指南

> ⚠️ **重要提示**: 在 Windows 上请使用 **PowerShell** 或 **CMD** 运行命令，不要使用 Git Bash！Git Bash 会导致 Rust 链接器错误。详见 [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

## 📦 安装必需工具

### 1. Rust (Tauri 必需)

```bash
# Windows
# 访问 https://www.rust-lang.org/tools/install
# 下载并运行 rustup-init.exe
# 安装完成后重启终端
```

### 2. Go (后端编译必需)

```bash
# Windows
# 访问 https://golang.org/dl/
# 下载 Windows 安装包 (go1.21.windows-amd64.msi 或更高版本)
# 运行安装程序
# 安装完成后重启终端
```

### 3. FFmpeg (视频处理必需)

```bash
# Windows
# 访问 https://www.gyan.dev/ffmpeg/builds/
# 下载 "ffmpeg-release-essentials.zip"
# 解压到 C:\ffmpeg
# 添加 C:\ffmpeg\bin 到系统 PATH 环境变量
```

## ✅ 验证安装

打开新的终端窗口，运行以下命令：

```bash
rustc --version   # 应显示: rustc 1.x.x
cargo --version   # 应显示: cargo 1.x.x
go version        # 应显示: go version go1.21.x
ffmpeg -version   # 应显示: ffmpeg version x.x.x
```

## 🏗️ 构建项目

**打开 PowerShell 或 CMD（不要使用 Git Bash）**

```powershell
# 1. 进入项目目录
cd c:/Users/mayuh/.gemini/antigravity/playground/plasma-pinwheel/audio-track-switcher

# 2. 构建 Go 后端
npm run build:backend:win

# 3. 运行开发环境
npm run tauri dev
```

## 🎯 使用应用

1. 点击 "📁 Select Video File" 选择视频
2. 查看检测到的音轨列表
3. 选择想要设为默认的音轨
4. 点击 "✨ Switch Default Track" 开始处理
5. 完成后在原文件目录找到新文件

## 📝 输出文件命名

原文件: `video.mp4`  
输出文件: `video_track1.mp4` (数字为选择的音轨索引)

## ❓ 常见问题

**Q: 提示找不到 FFmpeg？**  
A: 确保 FFmpeg 已添加到 PATH 环境变量，重启终端后再试。

**Q: Go 后端编译失败？**  
A: 确保 Go 版本 ≥ 1.21，运行 `go version` 检查。

**Q: Tauri 启动失败？**  
A: 确保 Rust 已正确安装，运行 `cargo --version` 检查。

## 📚 更多信息

- 详细文档: [README.md](README.md)
- 开发指南: [DEVELOPMENT.md](DEVELOPMENT.md)
