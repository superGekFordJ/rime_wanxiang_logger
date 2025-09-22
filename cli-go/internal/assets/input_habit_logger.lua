-- 输入习惯记录器 (Version 14.1 - V2.2 Incremental Update)
-- This version incrementally adds subtype classification and field-level filtering
-- to the stable V14.1 base, preserving all core state logic for commit handling.
local rime = require "lib"

--[[-----------------------------------------------------------------------
-- Configuration Loading
---------------------------------------------------------------------------]]
local config = {
    enabled = true,
    log_only_non_first_choice = false,
    log_file_path = nil,
    log_events = {
        session_start = true,
        session_end = true,
        text_committed = true,
        input_state_changed = false,
        error = true
    },
    log_fields = {
        text_committed = {},
        input_state_changed = { event_subtype = {} }
    }
}
local config_ok, user_config = pcall(require, "input_habit_logger_config")
if config_ok and type(user_config) == 'table' then
    local function merge(base, new)
        for k, v in pairs(new) do
            if type(v) == 'table' and type(base[k]) == 'table' then
                merge(base[k], v)
            else
                base[k] = v
            end
        end
    end
    merge(config, user_config)
end

--[[-----------------------------------------------------------------------
-- JSON Encoder (V14.1 Stable)
---------------------------------------------------------------------------]]
local function escape_str(s)
    return '"' .. s:gsub('[\\"]', { ['\\'] = '\\\\', ['"'] = '\\"'
    }):gsub('\n', '\\n'):gsub('\r', '\\r'):gsub('\t', '\\t'):gsub('[\001-\031]', '') .. '"'
end
local to_json_value_func
local function table_to_json(val, is_array_hint)
    local parts = {}
    local is_array = is_array_hint
    if is_array == nil then
        is_array = true
        local expected_idx = 1
        for k, _ in pairs(val) do
            if type(k) ~= 'number' or k ~= expected_idx then
                is_array = false; break
            end
            expected_idx = expected_idx + 1
        end
        if expected_idx - 1 ~= #val then is_array = false end
        if #val == 0 and next(val) ~= nil and not is_array then is_array = false end
    end
    if is_array then
        for i = 1, #val do table.insert(parts, to_json_value_func(val[i])) end
        return '[' .. table.concat(parts, ',') .. ']'
    else
        for k, v in pairs(val) do
            if type(k) == 'string' then
                table.insert(parts, escape_str(k) .. ':' .. to_json_value_func(v))
            else
                table.insert(parts, escape_str(tostring(k)) .. ':' .. to_json_value_func(v))
            end
        end
        return '{' .. table.concat(parts, ',') .. '}'
    end
end
to_json_value_func = function(val)
    local t = type(val)
    if t == 'string' then
        return escape_str(val)
    elseif t == 'number' then
        if val ~= val or val == math.huge or val == -math.huge then return 'null' end
        return tostring(val)
    elseif t == 'boolean' then
        return tostring(val)
    elseif t == 'nil' then
        return 'null'
    elseif t == 'table' then
        return table_to_json(val)
    else
        return '"[unsupported_type:' .. t .. ']"'
    end
end
local function json_encode(tbl)
    if type(tbl) ~= 'table' then return to_json_value_func(tbl) end
    local is_array_root = true
    local expected_idx_root = 1
    for k, _ in pairs(tbl) do
        if type(k) ~= 'number' or k ~= expected_idx_root then
            is_array_root = false; break
        end
        expected_idx_root = expected_idx_root + 1
    end
    if expected_idx_root - 1 ~= #tbl then is_array_root = false end
    if #tbl == 0 and next(tbl) ~= nil and not is_array_root then is_array_root = false end
    return table_to_json(tbl, is_array_root)
end

--[[-----------------------------------------------------------------------
-- Logging Core (Upgraded to V2.2)
---------------------------------------------------------------------------]]
local function get_log_file_path()
    if config.log_file_path and type(config.log_file_path) == 'string' and config.log_file_path ~= "" then
        return config.log_file_path
    end
    if os.getenv("OS") == "Windows_NT" then
        return os.getenv("APPDATA") .. "\\Rime\\input_habit_log_structured.jsonl"
    end
    if os.execute("test -d '" .. (os.getenv("HOME") or "") .. "/Library/Rime'") == 0 then
        return (os.getenv("HOME") or "") .. "/Library/Rime/input_habit_log_structured.jsonl"
    end
    return (os.getenv("HOME") or "") .. "/.config/rime/input_habit_log_structured.jsonl"
end

local log_file_path = get_log_file_path()

local function log_json_event(event_data)
    if not config.enabled then return end
    local event_type = event_data.event_type

    -- 1. Master switch for the event type
    if not (config.log_events and config.log_events[event_type]) then return end

    -- 2. Special filter for non-first choice on commit
    if event_type == "text_committed" and config.log_only_non_first_choice then
        if event_data.selected_candidate_rank == nil or event_data.selected_candidate_rank < 1 then return end
    end

    -- 3. Get the rules for the fields of this event type.
    local field_rules = config.log_fields[event_type]
    if not field_rules then return end -- No rules for this type, log nothing.

    -- 4. Special filter for input_state_changed SUBTYPE
    if event_type == "input_state_changed" then
        local subtype_rules = field_rules.event_subtype or {}
        if not subtype_rules[event_data.event_subtype] then
            return -- This specific subtype is disabled, skip logging.
        end
    end

    -- 5. Build the final data payload based on the field rules.
    local filtered_data = { event_type = event_type }
    for key, value in pairs(event_data) do
        if key ~= "event_type" and field_rules[key] then
            if key == "event_subtype" then                  -- The rule is a table
                filtered_data[key] = value
            elseif type(field_rules[key]) == 'boolean' then -- The rule is a boolean
                filtered_data[key] = value
            end
        end
    end

    -- 6. Final check: if only event_type was added, don't log an empty event.
    if not next(filtered_data, "event_type") then return end

    -- 7. Add timestamp and write to file.
    filtered_data.timestamp = os.date("!%Y-%m-%dT%H:%M:%S.") ..
        string.format("%03dZ", math.floor((os.clock() * 1000) % 1000))
    local json_string = json_encode(filtered_data)

    local file = io.open(log_file_path, "a")
    if file then
        file:write(json_string .. "\n")
        file:close()
    elseif rime and rime.log_info then
        rime.log_info("LUA_LOGGER_V2.2 :: ERROR: Could not open log file: " .. log_file_path)
    end
end


--[[-----------------------------------------------------------------------
-- Rime Event Processing (V14.1 Stable Base)
---------------------------------------------------------------------------]]
-- State variables for manual pagination and input tracking
local last_input_state_for_commit = {}
local current_page_index = 0
local last_seen_input_buffer = ""

local function get_current_candidate_info(context, max_to_display)
    local info = { candidates_list = {}, first_candidate_text = nil, has_menu = false }
    if not context then return info end
    info.has_menu = context:has_menu()
    local sel_cand_success, sel_cand_val = pcall(function() return context:get_selected_candidate() end)
    if sel_cand_success and sel_cand_val and sel_cand_val.text then info.first_candidate_text = sel_cand_val.text end
    if not info.has_menu then return info end
    local composition = context.composition
    if not composition or composition:empty() then return info end
    local segment = composition:back()
    if not segment or not segment.menu or segment.menu:empty() then return info end
    local current_menu = segment.menu
    local count_success, cand_count = pcall(function() return current_menu:candidate_count() end)
    if not count_success or not cand_count or cand_count == 0 then return info end
    local display_limit = math.min(max_to_display or 5, cand_count)
    for i = 0, display_limit - 1 do
        local cand_obj_success, cand_obj_val = pcall(function() return current_menu:get_candidate_at(i) end)
        if cand_obj_success and cand_obj_val then
            table.insert(info.candidates_list,
                cand_obj_val.text or (cand_obj_val.comment and cand_obj_val.comment .. " (comment)") or "?")
        else
            table.insert(info.candidates_list, "?err")
        end
    end
    return info
end

local function on_commit_callback(context)
    local committed_text_val = "N/A"
    local ct_succ, ct_val = pcall(function() return context:get_commit_text() end)
    if ct_succ and ct_val then committed_text_val = ct_val end

    local input_sequence_at_commit = last_input_state_for_commit.input_buffer or "N/A"

    local selected_rank = -1
    local page_size = 6 -- As specified by the user
    local key_action = last_input_state_for_commit.key_action_for_selection
    local page_index = last_input_state_for_commit.page_index or 0

    if key_action then
        if key_action == "space" then
            -- For spacebar commits, we need to find the actual selected candidate
            -- within the last known candidate list to respect Up/Down key highlighting.
            local local_index = -1
            if last_input_state_for_commit.candidates then
                for i, cand_text in ipairs(last_input_state_for_commit.candidates) do
                    if cand_text == committed_text_val then
                        local_index = i - 1 -- 0-indexed
                        break
                    end
                end
            end

            if local_index ~= -1 then
                selected_rank = (page_index * page_size) + local_index
            else
                -- Fallback for safety: assume first candidate if not found
                selected_rank = page_index * page_size
            end
        elseif tonumber(key_action) then
            -- For number commits, the calculation is direct.
            local local_index = tonumber(key_action) - 1
            selected_rank = (page_index * page_size) + local_index
        end
    end

    local selection_method = "unknown"
    if key_action then
        if key_action == "space" then
            selection_method = (selected_rank == 0) and "first_choice_space" or "nth_choice_space"
        elseif tonumber(key_action) then
            selection_method = "nth_choice_number_" .. key_action
        end
    elseif not last_input_state_for_commit.input_buffer then
        selection_method = "direct_commit_no_menu"
    end

    log_json_event({
        event_type = "text_committed",
        committed_text = committed_text_val,
        input_sequence_at_commit = input_sequence_at_commit,
        selection_method = selection_method,
        selected_candidate_rank = selected_rank,
        source_input_buffer = last_input_state_for_commit.input_buffer,
        source_candidates_list = last_input_state_for_commit.candidates,
        source_first_candidate = last_input_state_for_commit.candidates[1],
        source_event_timestamp = last_input_state_for_commit.timestamp -- Timestamp of pre-commit state
    })

    last_input_state_for_commit.key_action_for_selection = nil
end

--[[-----------------------------------------------------------------------
-- Main Processor (V14.1 Base with V2.2 Logic Injected)
---------------------------------------------------------------------------]]
local function main_processor_func(key, env)
    if not key then return rime.process_results.kNoop end
    local key_name_val
    local kn_succ, kn_val = pcall(function() return key:repr() end)
    if not kn_succ then return rime.process_results.kNoop end
    key_name_val = kn_val
    if key_name_val == "0x0000" or key:release() then return rime.process_results.kNoop end
    local context = env.engine and env.engine.context
    if not context then return rime.process_results.kNoop end

    local pcall_success, err_msg = pcall(function()
        local current_input_buffer = "N/A"
        local st_succ_main, st_val_main = pcall(function() return context:get_script_text() end)
        if st_succ_main and st_val_main and st_val_main ~= "" then
            current_input_buffer = st_val_main
        else -- Fallback logic to get segmented pinyin
            local sel_cand_obj
            local sel_c_succ, sel_c_val = pcall(function() return context:get_selected_candidate() end)
            if sel_c_succ and sel_c_val then sel_cand_obj = sel_c_val end
            if sel_cand_obj and sel_cand_obj.preedit and sel_cand_obj.preedit ~= "" then
                current_input_buffer = sel_cand_obj.preedit
            elseif context.input and context.input ~= "" then
                current_input_buffer = context.input
            end
        end
        -- Page tracking logic
        if current_input_buffer ~= last_seen_input_buffer then
            current_page_index = 0 -- Reset page index on new input
            last_seen_input_buffer = current_input_buffer
        end

        local nav_keys = { Page_Down = 1, Next = 1, Page_Up = -1, Prev = -1 }
        if nav_keys[key_name_val] then
            current_page_index = math.max(0, current_page_index + nav_keys[key_name_val])
        end

        local cand_info = get_current_candidate_info(context)

        -- ========== START: INJECTED V2.2 LOGIC ==========
        -- This block determines the subtype and decides whether to log,
        -- replacing the original, unconditional log_json_event call.
        if cand_info.has_menu and cand_info.candidates_list and #cand_info.candidates_list > 0 then
            local event_subtype = "other_key" -- Default subtype
            local key_name_for_log = key_name_val

            local menu_keys = { Up = true, Down = true, Page_Up = true, Page_Down = true, Next = true }
            local seg_keys = { Control_Left = true, Control_Right = true }

            if menu_keys[key_name_val] then
                event_subtype = "menu_navigation"
            elseif key_name_val == "Escape" then
                event_subtype = "input_rejected"
            elseif seg_keys[key_name_val] and key:modifier() == rime.key_modifier.kControlMask then
                event_subtype = "manual_segmentation"
            elseif key_name_val:len() == 1 or key_name_val == "BackSpace" then
                event_subtype = "buffer_edit"
            end

            -- The ONLY call to log this event type. log_json_event handles all filtering.
            -- NOTE: We do NOT filter for the space key here. The logic is passed through,
            -- but the log_json_event call for it will be skipped by the subtype check if configured.
            log_json_event({
                event_type = "input_state_changed",
                event_subtype = event_subtype,
                key_action = key_name_for_log,
                input_buffer = current_input_buffer,
                candidates = cand_info.candidates_list,
                first_candidate = cand_info.candidates_list[1],
                has_menu = cand_info.has_menu
            })
        end
        -- ========== END: INJECTED V2.2 LOGIC ==========

        -- ========== START: PRESERVED V14.1 CORE LOGIC ==========
        -- This logic MUST remain untouched to ensure on_commit works correctly.
        if cand_info.has_menu then
            last_input_state_for_commit.input_buffer = current_input_buffer
            last_input_state_for_commit.first_candidate = cand_info.first_candidate_text
            last_input_state_for_commit.candidates = cand_info.candidates_list
            last_input_state_for_commit.page_index = current_page_index -- Store the page index
            last_input_state_for_commit.timestamp = os.date("!%Y-%m-%dT%H:%M:%S.") ..
                string.format("%03dZ", math.floor((os.clock() * 1000) % 1000))
        else
            last_input_state_for_commit.input_buffer = current_input_buffer
            last_input_state_for_commit.first_candidate = nil
            last_input_state_for_commit.candidates = nil
            last_input_state_for_commit.page_index = 0 -- Reset on menu close
            last_input_state_for_commit.timestamp = os.date("!%Y-%m-%dT%H:%M:%S.") ..
                string.format("%03dZ", math.floor((os.clock() * 1000) % 1000))
        end
        last_input_state_for_commit.key_action_for_selection = nil
        if cand_info.has_menu then
            if key_name_val == "space" then
                last_input_state_for_commit.key_action_for_selection = "space"
            elseif key_name_val:match("^[1-9]$") then
                last_input_state_for_commit.key_action_for_selection = key_name_val
            end
        end
        -- ========== END: PRESERVED V14.1 CORE LOGIC ==========
    end)

    if not pcall_success then
        log_json_event({
            event_type = "error",
            component = "main_processor",
            message = tostring(err_msg),
            key_repr =
                key_name_val
        })
    end

    return rime.process_results.kNoop
end

--[[-----------------------------------------------------------------------
-- Module Definition (V14.1 Stable)
---------------------------------------------------------------------------]]
return {
    init = function(env)
        log_json_event({
            event_type = "session_start",
            schema_id = (env.engine and env.engine.context and env.engine.context.schema and env.engine.context.schema.id) or
                "N/A"
        })
        if env.engine and env.engine.context and env.engine.context.commit_notifier then
            env.commit_notifier_connection = env.engine.context.commit_notifier:connect(on_commit_callback)
        else
            log_json_event({ event_type = "error", component = "init", message = "Could not connect to commit_notifier." })
        end
        last_input_state_for_commit = {
            input_buffer = nil,
            first_candidate = nil,
            candidates = nil,
            key_action_for_selection = nil,
            timestamp = nil
        }
    end,
    fini = function(env)
        if env.commit_notifier_connection then
            env.commit_notifier_connection:disconnect()
        end
        log_json_event({ event_type = "session_end" })
    end,
    func = main_processor_func
}
