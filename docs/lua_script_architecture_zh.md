# Rime 输入习惯记录器 Lua 脚本架构与配置指南

本文档旨在说明 Rime 输入习惯记录器中 Lua 脚本的工作方式，特别是用户如何通过 `input_habit_logger_config.lua` 文件进行配置。

## 概述

本记录器包含两个核心 Lua 脚本：

1.  `input_habit_logger.lua`: 主逻辑脚本，负责捕获 Rime 输入法引擎的事件，并根据配置将数据记录到日志文件中。
2.  `input_habit_logger_config.lua`: 用户配置文件，允许用户自定义日志记录的行为，例如选择日志级别、指定日志路径、以及控制记录哪些具体信息。

当 `rime-logger install` 命令执行时，这两个文件会被复制到 Rime 用户目录下的 `lua` 文件夹中。主脚本 `input_habit_logger.lua` 会被 Rime 的 `lua_processor` 加载。

## `input_habit_logger_config.lua` 配置文件详解

此文件是用户与 Lua 记录器交互的主要途径。它允许用户在不直接修改主逻辑脚本的情况下，调整日志记录的详细程度和行为。

### 1. 预设选择 (`preset_choice`)

这是配置文件的核心。用户可以通过修改 `preset_choice` 变量来选择一个预定义的日志记录方案。

```lua
local preset_choice = "normal" -- 在此修改为你想要的预设
```

可用的预设包括：

*   `"normal"`: **普通模式 (推荐)**
    *   目标：最小化日志记录，主要用于计算基本的输入法预测准确率。
    *   特点：仅记录会话开始/结束、文本上屏事件和错误。对于上屏事件，主要记录选择的候选词排名、上屏的文本以及输入法预测的首选词。
    *   禁用 `input_state_changed` 事件以避免日志文件过大。
*   `"developer"`: **词库贡献者/开发者模式**
    *   目标：适合调试，特别关注非首选上屏的情况。
    *   特点：会开启 `log_only_non_first_choice = true`，意味着只记录那些用户没有选择第一个候选词就上屏的情况。记录的字段比 `normal` 模式更详细，包括按键、输入缓冲区状态等。
*   `"advanced"`: **高级模式**
    *   目标：记录几乎所有可获取的信息，用于深度分析和问题排查。
    *   特点：记录所有定义的事件类型和每个事件类型下的所有可用字段。日志文件可能会非常大。
*   `"custom"`: **自定义模式**
    *   目标：允许用户完全自定义所有日志记录参数。当 `preset_choice` 设置为 `"custom"` 时，脚本会使用 `presets.custom` 表中的设置。

### 2. 预设定义 (`presets` 表)

`presets` 表包含了上述每种预设的具体配置。每种预设都是一个独立的 Lua 表，包含以下可配置项：

#### a. `enabled` (布尔值)

*   **作用**: 总开关。如果设置为 `false`，则该预设下不会进行任何日志记录。
*   **示例**: `enabled = true`

#### b. `log_file_path` (字符串 或 `nil`)

*   **作用**: 指定日志文件的绝对路径。如果设置为 `nil` 或空字符串，脚本会尝试自动侦测并使用默认路径。
    *   Windows: `%APPDATA%\Rime\input_habit_log_structured.jsonl`
    *   macOS: `~/Library/Rime/input_habit_log_structured.jsonl`
    *   Linux: `~/.config/rime/input_habit_log_structured.jsonl`
*   **注意**: 路径中的反斜杠 `\` 在 Lua 字符串中需要转义，例如 `"C:\\Users\\YourUser\\Desktop\\rime_log.jsonl"`。
*   **示例**: `log_file_path = nil` (使用默认路径) 或 `log_file_path = "/path/to/your/custom_log.jsonl"`

#### c. `log_only_non_first_choice` (布尔值)

*   **作用**: (仅对 `text_committed` 事件有效) 如果设置为 `true`，则只有当用户选择的候选词不是输入法预测的第一个候选词时，才会记录该上屏事件。这对于专注于分析预测失误或用户手动选择的场景很有用。
*   **示例**: `log_only_non_first_choice = false`

#### d. `log_events` (表)

*   **作用**: 精细控制要记录哪些类型的事件。键是事件类型名称，值是布尔值 (`true` 表示记录, `false` 表示不记录)。
*   **可用事件类型**:
    *   `session_start`: Rime 会话开始时记录。
    *   `session_end`: Rime 会话结束时记录。
    *   `text_committed`: 用户文本上屏时记录 (例如，按下空格或数字选择候选词)。
    *   `input_state_changed`: 每当输入状态（如输入缓冲区内容、候选词列表）发生变化时记录。**注意：此事件非常频繁，可能导致日志文件迅速增大。如果只关心最终上屏结果，建议禁用此项。**
    *   `error`: 当 Lua 脚本内部发生错误时记录。
*   **示例**:
    ```lua
    log_events = {
        session_start = true,
        session_end = true,
        text_committed = true,
        input_state_changed = false, -- 普通模式下通常关闭
        error = true
    }
    ```

#### e. `log_fields` (表)

*   **作用**: 对每种启用的事件类型，进一步控制具体记录哪些数据字段。这提供了最大程度的隐私保护和数据粒度控制。
*   `log_fields` 的键是事件类型名称 (必须与 `log_events` 中启用的事件对应)。
*   对应的值是另一个表，其键为该事件类型下可记录的字段名，值为布尔值 (`true` 表示记录该字段, `false` 表示不记录)。

*   **`text_committed` 事件的可选字段**:
    *   `committed_text`: 用户最终上屏的文本。
    *   `input_sequence_at_commit`: 上屏时，Rime 内部的原始输入序列 (如拼音串)。
    *   `selection_method`: 用户选择候选词的方式 (例如, `"first_choice_space"`, `"nth_choice_number_2"`)。
    *   `selected_candidate_rank`: 用户选择的候选词在其所在候选列表中的排名 (0 代表首选)。这是计算准确率的关键指标。
    *   `source_input_buffer`: 上屏前一刻的输入缓冲区内容。
    *   `source_first_candidate`: 上屏前一刻，输入法预测的第一个候选词是什么。
    *   `source_candidates_list`: 上屏前一刻，完整的候选词列表。

*   **`input_state_changed` 事件的可选字段**:
    *   `key_action`: 用户按下的按键的描述 (例如, `"a"`, `"space"`, `"BackSpace"`, `"Page_Down"`)。
    *   `input_buffer`: 当前输入缓冲区的内容。
    *   `candidates`: 当前候选词列表。
    *   `first_candidate`: 当前候选列表中的第一个候选词。
    *   `has_menu`: 当前是否有候选菜单显示。

*   **示例 (normal 预设下的 `text_committed` 事件字段配置)**:
    ```lua
    log_fields = {
        text_committed = {
            selected_candidate_rank = true, -- 对计算准确率至关重要
            committed_text = true,
            source_first_candidate = true
            -- 其他字段如 input_sequence_at_commit, source_candidates_list 等在此预设下为 false 或未定义，因此不记录
        }
        -- input_state_changed 的字段配置在此处省略，因为 log_events.input_state_changed = false
    }
    ```

### 3. 自定义配置 (`presets.custom`)

如果用户将 `preset_choice` 设置为 `"custom"`，则 `presets.custom` 表中的所有设置将被激活。用户可以自由修改此部分以满足特定的日志记录需求。配置文件中已为 `custom` 预设提供了所有可用选项的注释和示例。

## `input_habit_logger.lua` 主脚本逻辑简介

用户通常不需要直接修改此文件。其主要职责包括：

1.  **加载配置**: 启动时，读取 `input_habit_logger_config.lua` 文件，并根据 `preset_choice` 合并相应的预设配置。
2.  **JSON 编码**: 将捕获到的事件数据编码为 JSON 格式，以便于后续处理和分析。
3.  **日志写入**: 将 JSON 格式的日志条目追加到指定的日志文件中。如果文件无法打开，会尝试通过 Rime 的日志系统报告错误。
4.  **事件处理**:
    *   `on_commit_callback`: 当 Rime 引擎提交文本时被调用。它收集关于提交文本、选择方式、候选词排名等信息，并根据配置记录 `text_committed` 事件。
    *   `main_processor_func`: Rime 的主处理器函数，在每次按键时被调用。它负责：
        *   跟踪输入缓冲区的变化、候选词的翻页。
        *   根据配置记录 `input_state_changed` 事件，区分不同的子类型（如编辑、导航、拒绝等）。
        *   维护提交前的状态，供 `on_commit_callback` 使用。
5.  **会话管理**: 在 Rime 会话开始 (`init`) 和结束 (`fini`) 时记录相应的事件。

## 总结

通过精心设计的 `input_habit_logger_config.lua` 文件，用户可以方便地控制日志记录的粒度，平衡数据收集需求与隐私保护及性能考量。建议用户从一个标准预设开始，如果需要更精细的控制，再切换到 `custom` 预设进行调整。
