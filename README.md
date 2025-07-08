# Rime 输入习惯记录器

该项目提供了一个 Lua 脚本和一个基于 Python 的命令行工具，用于安装和管理 [Rime 输入法引擎](https://rime.im/) 的数据记录器。它旨在捕获有关您打字习惯的详细信息，并将其保存到结构化的 JSONL 文件中，以便后续分析。

主要目标是收集可用于分析打字效率、识别常见错误以及可能微调您的 Rime 输入方案以获得更好、更个性化输入体验的数据。

## 功能特性

-   **详细事件记录:** 捕获按键、候选词列表和最终文本选择。
-   **丰富上下文:** 记录您*如何*选择一个词（空格键或数字键）以及所选候选词的排名。
-   **可配置:** 您可以通过简单的 Lua 配置文件完全控制记录内容。
-   **跨平台:** 安装程序会自动检测 Windows、macOS 和 Linux 上的 Rime 用户目录。
-   **易于管理:** 一个简单的命令行界面 (`rime-logger`) 用于安装、卸载和检查记录器的状态。
-   **便于分析:** 输出结构化的 JSONL 数据，非常适合使用 Python 的 Pandas 库等工具进行处理。

## 安装

您需要系统上安装 Python 3.7+。

1.  克隆此仓库或下载源代码:
    ```bash
    git clone https://github.com/superGekFordJ/rime_wanxiang_logger.git
    cd rime-wanxiang-logger
    ```

2.  使用 pip 安装此软件包。这将安装必要的文件并使 `rime-logger` 命令在系统范围内可用。
    ```bash
    pip install .
    ```

## 使用方法

安装程序提供了一个单独的命令行工具 `rime-logger` 来管理安装。

**重要提示:** 运行 `install` 或 `uninstall` 后，您**必须**重新部署 Rime 才能使更改生效。您通常可以通过单击系统菜单/任务栏中的 Rime 图标并选择“部署”来完成此操作。

### 安装记录器

此命令现在会启动一个交互式会话，允许您直接从命令行选择日志记录预设。

```bash
rime-logger install
```

您将看到一个选择模式的提示，例如“普通”、“开发者”或“高级”。

### 分析您的打字习惯

`analyze` 命令会读取您的日志文件，并提供您打字准确性的统计摘要，包括首选命中率、前三候选命中率和整体预测得分。

```bash
rime-logger analyze
```

### 导出开发者数据

如果您想帮助改进输入方案，可以使用 `export-misses` 命令。它会在您的用户主目录中创建一个名为 `rime_mispredictions_report.csv` 的文件，其中包含所有您未选择首选候选词的情况。此文件可以轻松与开发者共享。

```bash
rime-logger export-misses
```

### 卸载记录器

这将从您的输入方案文件中删除记录器并删除主 Lua 脚本。它将保留配置文件不变，以保留您的设置。

```bash
rime-logger uninstall
```

### 检查状态

这将报告必要文件是否到位以及输入方案是否配置正确。

```bash
rime-logger status
```

## 配置

当您安装记录器时，会在您的 Rime `lua` 目录中创建一个名为 `input_habit_logger_config.lua` 的配置文件。您可以编辑此文件来控制记录哪些数据。

`input_habit_logger_config.lua` 示例:
```lua
return {
  -- Master switch. If false, no logging will occur.
  enabled = true,

  -- Control which events to log.
  log_events = {
    session_start = true,
    session_end = true,
    text_committed = true,
    -- 'input_state_changed' is very noisy (logs every keypress).
    -- Set to false if you only care about the final result.
    input_state_changed = false,
    error = true
  },

  -- Future-proofing for advanced data filtering.
  log_fields = { ... }
}
```

## 日志文件

记录器将在您的主要 Rime 用户目录中创建一个名为 `input_habit_log_structured.jsonl` 的文件。此文件中的每一行都是一个独立的 JSON 对象，表示一个事件。

此格式非常适合使用数据分析工具进行处理。例如，您可以轻松将其加载到 Python 的 Pandas DataFrame 中：

```python
import pandas as pd

# 日志文件通常位于 Rime 用户目录的根目录
log_file_path = 'path/to/your/Rime/input_habit_log_structured.jsonl'
df = pd.read_json(log_file_path, lines=True)

# 开始您的分析！
print(df.info())
```

## 许可证

本项目采用 GPL-3.0 许可证。有关详细信息，请参阅 `LICENSE` 文件。

