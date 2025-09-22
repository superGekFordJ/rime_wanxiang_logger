-- Rime 输入习惯记录器配置 (版本 3.1 - 内置预设)

--[[-----------------------------------------------------------------------
-- 预设选择
---------------------------------------------------------------------------]]
-- 在此设置您想要的日志记录级别。脚本将使用下面定义的匹配预设中的设置。
--
-- 可用选项:
-- "normal"    - 最小化日志记录，用于计算基本的输入准确率。
-- "developer" - 适合调试；专注于记录非首选上屏。
-- "advanced"  - 记录几乎所有信息，用于深度分析。
-- "custom"    - 完全可编辑的预设，用于您自己的设置。
--
local preset_choice = "normal"

--[[-----------------------------------------------------------------------
-- 预设定义
---------------------------------------------------------------------------]]
-- 上面所选预设的配置。
-- 如需更改，请编辑下面的 "custom" 预设，并将 preset_choice 设置为 "custom"。
local presets = {
    -- ==================================================================
    -- 预设: NORMAL
    -- 最小化日志记录，用于计算基本的预测准确率。
    -- ==================================================================
    normal = {
        enabled = true,
        log_only_non_first_choice = false,
        log_events = {
            session_start = true,
            session_end = true,
            text_committed = true,
            input_state_changed = false, -- 禁用以防止过多的日志信息
            error = true
        },
        log_fields = {
            text_committed = {
                selected_candidate_rank = true, -- 对计算准确率至关重要
                committed_text = true,
                source_first_candidate = true
            }
        }
    },

    -- ==================================================================
    -- 预设: DEVELOPER
    -- 适合调试。专注于记录非首选上屏。
    -- ==================================================================
    developer = {
        enabled = true,
        log_only_non_first_choice = true,
        -- 您可以在此覆盖日志文件路径。
        -- 示例: log_file_path = "C:\\Users\\YourUser\\Desktop\\rime_log.jsonl"
        log_file_path = nil,
        log_events = {
            session_start = true,
            session_end = true,
            text_committed = true,
            input_state_changed = true,
            error = true
        },
        log_fields = {
            text_committed = {
                selected_candidate_rank = true,
                committed_text = true,
                input_sequence_at_commit = true,
                selection_method = true,
                source_input_buffer = true,
                source_first_candidate = true,
                source_candidates_list = false,
            },
            input_state_changed = {
                key_action = true,
                input_buffer = true,
                first_candidate = true,
                has_menu = true,
                candidates = false,
            }
        }
    },

    -- ==================================================================
    -- 预设: ADVANCED
    -- 记录几乎所有信息，用于深度分析。
    -- ==================================================================
    advanced = {
        enabled = true,
        log_only_non_first_choice = false,
        -- 您可以在此覆盖日志文件路径。
        -- 示例: log_file_path = "C:\\Users\\YourUser\\Desktop\\rime_log.jsonl"
        log_file_path = nil,
        log_events = {
            session_start = true,
            session_end = true,
            text_committed = true,
            input_state_changed = true,
            error = true
        },
        log_fields = {
            text_committed = {
                selected_candidate_rank = true,
                committed_text = true,
                input_sequence_at_commit = true,
                selection_method = true,
                source_input_buffer = true,
                source_first_candidate = true,
                source_candidates_list = true,
            },
            input_state_changed = {
                key_action = true,
                input_buffer = true,
                first_candidate = true,
                has_menu = true,
                candidates = true,
            }
        }
    },

    -- ==================================================================
    -- 预设: CUSTOM
    -- 此部分用于您的个人配置。
    -- 这里保留了原始注释，以帮助您理解每个设置。
    -- 要使用这些设置，请将文件顶部的 'preset_choice' 设置为 "custom"。
    -- ==================================================================
    custom = {
        -- 总开关。如果设置为 false，将完全不进行任何日志记录。
        enabled = true,

        -- 您可以在此覆盖日志文件路径。
        -- 示例: log_file_path = "C:\\Users\\YourUser\\Desktop\\rime_log.jsonl"
        log_file_path = nil,

        -- 如果为 true，则仅记录所选候选项不是首选的提交。
        -- 这对于专注于预测错误或手动选择非常有用。
        log_only_non_first_choice = false,

        -- 精细控制要记录的事件类型。
        -- 这对于减小日志文件大小或保护隐私非常有用。
        log_events = {
            session_start = true,
            session_end = true,
            text_committed = true,
            -- 注意: 'input_state_changed' 非常详细。它会记录每一次按键
            -- 导致输入缓冲区变化的事件。如果您只关心最终上屏的文本，
            -- 请禁用此项。
            input_state_changed = false,
            error = true
        },

        -- 对每种事件类型记录哪些字段进行高级控制。
        -- 这使您可以最大程度地控制隐私和数据粒度。
        log_fields = {
            text_committed = {
                -- 例如，要停止记录您输入的确切字符，请设置：
                --   committed_text = false,
                committed_text = true,
                input_sequence_at_commit = true,
                selection_method = true,
                selected_candidate_rank = true,
                source_input_buffer = true,
                source_first_candidate = true,
                -- 要停止记录您看到的候选词列表，请设置：
                --   source_candidates_list = false,
                source_candidates_list = true,
            },
            input_state_changed = {
                key_action = true,
                input_buffer = true,
                candidates = true,
                first_candidate = true,
                has_menu = true,
            }
        }
    }
}

-- 此行选择所选的预设表并将其返回给主脚本。
-- 如果选择无效，它将安全地默认为 "custom" 预设。
return presets[preset_choice] or presets.custom
