-- Rime Input Habit Logger Configuration (Version 3.1 - Self-Contained Presets)

--[[-----------------------------------------------------------------------
-- PRESET SELECTION
---------------------------------------------------------------------------]]
-- Set your desired logging level here. The script will use the settings
-- from the matching preset defined below.
--
-- Available options:
-- "normal"    - Minimal logging for calculating basic typing accuracy.
-- "developer" - Good for debugging; focuses on non-first-choice commits.
-- "advanced"  - Logs almost everything for deep analysis.
-- "custom"    - A fully editable preset for your own settings.
--
local preset_choice = "normal"

--[[-----------------------------------------------------------------------
-- PRESET DEFINITIONS
---------------------------------------------------------------------------]]
-- The configurations for the presets selected above.
-- To make changes, edit the "custom" preset below and set preset_choice = "custom".
local presets = {
    -- ==================================================================
    -- PRESET: NORMAL
    -- Minimal logging for calculating basic predicting accuracy.
    -- ==================================================================
    normal = {
        enabled = true,
        log_only_non_first_choice = false,
        log_events = {
            session_start = true,
            session_end = true,
            text_committed = true,
            input_state_changed = false, --disable to prevent too much noise
            error = true
        },
        log_fields = {
            text_committed = {
                selected_candidate_rank = true, -- Essential for accuracy
                committed_text = true
            }
        }
    },

    -- ==================================================================
    -- PRESET: DEVELOPER
    -- Good for debugging. Focuses on non-first-choice selections.
    -- ==================================================================
    developer = {
        enabled = true,
        log_only_non_first_choice = true,
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
    -- PRESET: ADVANCED
    -- Logs almost everything for deep analysis.
    -- ==================================================================
    advanced = {
        enabled = true,
        log_only_non_first_choice = false,
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
    -- PRESET: CUSTOM
    -- This section is for your personal configuration.
    -- The original comments are preserved here to help you understand each setting.
    -- To use these settings, set 'preset_choice' at the top of the file to "custom".
    -- ==================================================================
    custom = {
        -- Master switch. If set to false, no logging will occur at all.
        enabled = true,

        -- (Placeholder) In the future, you can override the log file path here.
        -- Example: log_file_path = "C:\\Users\\YourUser\\Desktop\\rime_log.jsonl"
        log_file_path = nil,

        -- If true, only log commits where the selected candidate was NOT the first one offered.
        -- This is useful for focusing on prediction errors or manual selections.
        log_only_non_first_choice = false,

        -- Granular control over which event types to log.
        -- This is useful for reducing log file size or for privacy.
        log_events = {
            session_start = true,
            session_end = true,
            text_committed = true,
            -- NOTE: 'input_state_changed' is very verbose. It logs every keypress
            -- that changes the input buffer. Disable this if you only care about
            -- the final committed text.
            input_state_changed = false,
            error = true
        },

        -- Advanced control over which fields are logged for each event type.
        -- This gives you maximum control over privacy and data granularity.
        log_fields = {
            text_committed = {
                -- For example, to stop logging the exact text you type, set:
                --   committed_text = false,
                committed_text = true,
                input_sequence_at_commit = true,
                selection_method = true,
                selected_candidate_rank = true,
                source_input_buffer = true,
                source_first_candidate = true,
                -- To stop logging the list of candidates you saw, set:
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

-- This line selects the chosen preset table and returns it to the main script.
-- If the choice is invalid, it safely defaults to the "custom" preset.
return presets[preset_choice] or presets.custom
