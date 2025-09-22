# 项目架构 (`rime-logger-go`)

本文档概述了 Rime 输入习惯记录器项目 **Go 语言版本** 的架构。

## 概述

本项目是一个基于 Go 语言的命令行工具（CLI），旨在替代原有的 Python 版本，用于安装、管理和分析 Rime 输入法引擎的 Lua 日志脚本。项目的核心目标保持不变：收集用户输入习惯数据，以帮助改善 Rime 的预测准确性并提供输入模式的分析洞见。

Go 版本的实现旨在提供更高的性能、更简便的部署方式（零依赖的单一可执行文件）以及更强的类型安全性。

## 目录结构

```
cli-go/
├── main.go                    # 应用程序入口
├── go.mod                     # Go 模块定义文件
├── cmd/                       # 存放所有 CLI 命令
│   ├── root.go                # 根命令 (rime-logger-go)
│   ├── install.go             # 'install' 命令实现
│   ├── uninstall.go           # 'uninstall' 命令实现
│   ├── status.go              # 'status' 命令实现
│   ├── analyze.go             # 'analyze' 命令实现
│   └── export-misses.go       # 'export-misses' 命令实现
├── internal/
│   ├── manager/               # 核心管理逻辑
│   │   └── manager.go         # RimeManager 的 Go 实现
│   ├── assets/                # 内嵌(embedded)的 Lua 脚本资源
│   │   ├── assets.go          # 使用 go:embed 指令加载 Lua 脚本
│   │   ├── input_habit_logger.lua
│   │   └── input_habit_logger_config.lua
│   └── analyzer/              # 数据分析逻辑
│       └── analyzer.go        # JSONL 解析和统计分析功能
└── rime-logger-go.exe         # (构建产物) 最终的可执行文件
```

## 核心组件

### 1. **`cmd` 包：CLI 接口**

这是应用的 CLI 逻辑层，使用 `github.com/spf13/cobra` 库构建，提供了与原 Python 版本 (`click`) 类似的功能。

- **`root.go`**: 定义根命令 `rime-logger-go`。
- **`install.go`**:
  - 实现 `install` 命令。
  - **交互式预设选择**: 使用 `github.com/manifoldco/promptui` 库，提供与原版 `questionary` 相同的交互式菜单，让用户选择日志记录模式（Normal, Developer, Advanced）。
  - **脚本安装**: 调用 `internal/manager` 组件，将内嵌的 Lua 脚本复制到 Rime 用户目录的 `lua` 子目录中。
  - **配置文件修改**: 在复制 `input_habit_logger_config.lua` 之前，通过字符串替换修改其内容，以激活用户选择的预设。
  - **Schema 自动修改**: 调用 `internal/manager` 组件，自动在 `wanxiang.schema.yaml`（或其他 schema）文件中添加 `lua_processor` 配置，并在此之前创建备份。
- **`uninstall.go`**: 实现 `uninstall` 命令，负责移除 Lua 脚本并从 schema 文件中清理配置。
- **`status.go`**: 实现 `status` 命令，全面检查脚本安装状态、schema 配置状态和日志文件的存在情况。
- **`analyze.go`** & **`export-misses.go`**: 实现数据分析和报告导出命令，它们依赖 `internal/analyzer` 包来执行核心的数据处理。

### 2. **`internal/manager` 包：核心管理逻辑**

这个包是原 Python 版本 `RimeManager` 类的 Go 语言等价实现。

- **`manager.go`**:
  - **`RimeManager` 结构体**: 封装了与 Rime 用户目录交互的所有核心逻辑。
  - **跨平台目录检测**: `GetRimeUserDirectory()` 函数通过检查操作系统 (`runtime.GOOS`) 和环境变量 (`%APPDATA%`, `~/Library`, `~/.config`) 来自动定位 Rime 用户目录，完全兼容 Windows、macOS 和多种 Linux 发行版。
  - **日志文件路径解析**: `GetLogFilePath()` 方法实现了对 `input_habit_logger_config.lua` 文件的解析。它通过正则表达式查找当前激活的 `preset`，然后在对应的配置块中寻找 `log_file_path`，从而精确地定位日志文件，即使它被用户自定义过。
  - **文件操作**: 提供了对 Lua 脚本和 schema 文件的复制、删除、备份和修改功能。Schema 修改逻辑能够精确定位到 `punctuator` 之后插入 `lua_processor`，与原版行为一致。

### 3. **`internal/analyzer` 包：数据分析引擎**

这个包是原 Python 版本中 `pandas` 数据处理部分的 Go 语言替代实现。

- **`analyzer.go`**:
  - **`LogEvent` 结构体**: 定义了与 Lua 脚本输出的 JSONL 日志格式完全匹配的 Go 结构体，利用 `json` 标签进行高效、类型安全的解析。
  - **`AnalysisResult` 结构体**: 用于存储所有分析指标的结果，如首选命中率、前三命中率、平均选择排名等。
  - **`ReadLogFile()`**: 使用 `bufio.Scanner` 流式读取 JSONL 文件，逐行解析 JSON。这种方式内存效率极高，即使面对非常大的日志文件也能轻松处理。
  - **`PerformAnalysis()`**: 实现了与 Python 版本完全相同的统计分析逻辑。它迭代处理 `LogEvent` 切片，计算所有核心指标，包括“综合预测得分”。
  - **`ExportMisses()`**: 筛选出所有 `selected_candidate_rank > 0` 的误预测事件，并使用 `encoding/csv` 包将它们写入 CSV 文件。报告的列名和排序逻辑（按错误频率降序）也与原版保持一致。

### 4. **`internal/assets` 包：内嵌资源**

- **`assets.go`**:
  - 使用 Go 1.16+ 的 `//go:embed` 指令。
  - 在编译时，`input_habit_logger.lua` 和 `input_habit_logger_config.lua` 文件被直接读取并嵌入到最终的可执行文件中。
  - 这使得整个工具成为一个**单一的二进制文件**，分发和使用都极为方便，无需像 Python 版本那样关心 `assets` 文件夹的相对路径问题。

## 工作流程

Go 版本的工作流程与 Python 版本基本相同，但底层实现更高效：

1.  **安装 (`rime-logger-go install`)**:
    - `promptui` 显示预设菜单。
    - `RimeManager` 定位 Rime 目录。
    - 从内嵌的 `assets` 中读取 Lua 脚本内容。
    - `RimeManager` 修改 `config.lua` 内容以匹配预设，然后将两个脚本写入 Rime 的 `lua` 目录。
    - `RimeManager` 备份并修改 `wanxiang.schema.yaml`。
    - 提示用户重新部署 Rime。

2.  **分析 (`rime-logger-go analyze`)**:
    - `RimeManager` 解析 `config.lua` 找到日志文件路径。
    - `analyzer.ReadLogFile` 流式读取并解析 JSONL 数据到 `[]LogEvent`。
    - `analyzer.PerformAnalysis` 计算所有指标。
    - 在终端打印格式化的结果。

## 关键设计决策与优势

- **Go 替代 Python**:
  - **性能**: Go 是编译型语言，CLI 启动速度和执行效率远超解释型的 Python。
  - **零依赖部署**: 最终产物是一个独立的二进制文件，用户无需安装 Go 环境或任何包依赖（如 `pandas`, `click`），下载即可运行。
  - **跨平台编译**: Go 可以轻松地交叉编译为 Windows, macOS, Linux 的可执行文件。

- **`cobra` & `promptui`**:
  - `cobra` 是 Go 生态中最成熟的 CLI 构建库，功能强大，是 `click` 的理想替代品。
  - `promptui` 提供了美观且用户友好的交互式提示，完美替代了 `questionary`。

- **标准库进行数据分析**:
  - **放弃 `pandas`**: 虽然 Go 社区有一些 DataFrame 库（如 `dataframe-go`），但它们远不如 `pandas` 成熟。对于本项目的需求（JSONL 解析和基本统计），Go 的标准库 `encoding/json`、`bufio` 和 `encoding/csv` 结合自定义结构体和函数，不仅功能足够，而且性能更高、依赖更少。
  - **类型安全**: 使用强类型的 `LogEvent` 结构体解析 JSON，避免了 `pandas` 中可能出现的数据类型错误或 `KeyError`。

- **`go:embed` 嵌入资源**:
  - 这是处理静态资源（如 Lua 脚本）的现代 Go 解决方案。它简化了文件路径管理，并确保了应用的独立性和可移植性。

## 未来展望

Go 版本为项目带来了坚实的基础，未来的扩展可以考虑：

- **并发分析**: 利用 Go 的协程（goroutines）来并发处理超大规模的日志文件或多个日志文件。
- **完善 Schema 管理**: 扫描 Rime 目录下的所有 `*.schema.yaml` 文件，让用户交互式选择要修改哪一个。
- **自动重新部署**: 尝试调用 Rime 的命令行工具（如果存在）来触发自动重新部署，进一步提升用户体验。
- **构建图形用户界面 (GUI)**: 使用 Go 的 GUI 库（如 `Fyne` 或 `Wails`）为该工具创建一个跨平台的图形界面。
