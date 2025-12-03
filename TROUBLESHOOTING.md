# ⚠️ 常见问题解决方案

## 问题 1: Rust 链接器错误 (link.exe extra operand)

### 错误信息

```
error: linking with `link.exe` failed: exit code: 1
= note: link: extra operand '...'
Try 'link --help' for more information.
```

### 原因

Git Bash 的 `/usr/bin/link` 在 PATH 中优先于 Visual Studio 的 `link.exe`，导致 Rust 使用了错误的链接器。

### 解决方案

#### 方案 1: 使用 PowerShell 或 CMD（推荐）

不要使用 Git Bash，改用 PowerShell 或 CMD：

**PowerShell:**

```powershell
cd c:\Users\mayuh\.gemini\antigravity\playground\plasma-pinwheel\audio-track-switcher
npm run build:backend:win
npm run tauri dev
```

**CMD:**

```cmd
cd c:\Users\mayuh\.gemini\antigravity\playground\plasma-pinwheel\audio-track-switcher
npm run build:backend:win
npm run tauri dev
```

#### 方案 2: 使用 Visual Studio Developer Command Prompt

1. 打开 "Developer Command Prompt for VS" 或 "x64 Native Tools Command Prompt"
2. 运行：

```cmd
cd c:\Users\mayuh\.gemini\antigravity\playground\plasma-pinwheel\audio-track-switcher
npm run build:backend:win
npm run tauri dev
```

#### 方案 3: 切换到 GNU Rust 工具链

如果你想继续使用 Git Bash：

```bash
# 安装 GNU 工具链
rustup toolchain install stable-x86_64-pc-windows-gnu
rustup default stable-x86_64-pc-windows-gnu

# 然后正常运行
npm run tauri dev
```

**注意**: 使用 GNU 工具链需要安装 MinGW-w64。

#### 方案 4: 修改 PATH（临时）

在 Git Bash 中临时修改 PATH：

```bash
# 找到 Visual Studio 的 link.exe 路径
# 通常在: C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Tools\MSVC\{version}\bin\Hostx64\x64

# 临时添加到 PATH 前面（每次打开终端都需要运行）
export PATH="/c/Program Files/Microsoft Visual Studio/2022/Community/VC/Tools/MSVC/{version}/bin/Hostx64/x64:$PATH"

# 然后运行
npm run tauri dev
```

---

## 问题 2: 找不到 Go 后端

### 错误信息

```
Failed to execute Go backend: No such file or directory
```

### 解决方案

确保已构建 Go 后端：

```bash
npm run build:backend:win  # Windows
```

检查文件是否存在：

```bash
ls src-tauri/audio-track-backend.exe
```

---

## 问题 3: FFmpeg 未找到

### 错误信息

```
failed to execute ffprobe: No such file or directory
```

### 解决方案

1. 下载 FFmpeg: https://www.gyan.dev/ffmpeg/builds/
2. 解压到 `C:\ffmpeg`
3. 添加到 PATH:
   - 打开"系统属性" → "环境变量"
   - 编辑 "Path" 变量
   - 添加 `C:\ffmpeg\bin`
4. 重启终端并验证:

```bash
ffmpeg -version
```

---

## 问题 4: Go 未安装

### 解决方案

1. 访问: https://golang.org/dl/
2. 下载 Windows 安装包 (go1.21+)
3. 运行安装程序
4. 重启终端并验证:

```bash
go version
```

---

## 问题 5: Rust 未安装

### 解决方案

1. 访问: https://www.rust-lang.org/tools/install
2. 下载并运行 `rustup-init.exe`
3. 按照提示完成安装
4. 重启终端并验证:

```bash
rustc --version
cargo --version
```

---

## 推荐的开发环境设置

### Windows 用户推荐使用:

1. **PowerShell** 或 **CMD** 作为主要终端
2. 或者使用 **Visual Studio Developer Command Prompt**
3. 避免在 Git Bash 中编译 Rust 项目（除非切换到 GNU 工具链）

### 完整的工具安装顺序:

1. ✅ Visual Studio Build Tools (包含 MSVC)
2. ✅ Rust (使用 rustup)
3. ✅ Go
4. ✅ FFmpeg
5. ✅ Node.js

---

## 快速诊断命令

运行以下命令检查所有工具是否正确安装：

```bash
node --version
npm --version
rustc --version
cargo --version
go version
ffmpeg -version
```

如果所有命令都能正常输出版本号，说明环境配置正确。
