-- User configuration for the Rime Input Habit Logger
-- You can enable or disable features by changing 'true' to 'false'.

return {
  -- Master switch. If set to false, no logging will occur at all.
  enabled = true,

  -- (Placeholder) In the future, you can override the log file path here.
  -- Example: log_file_path = "C:\\Users\\YourUser\\Desktop\\rime_log.jsonl"
  log_file_path = nil,

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

  -- (Placeholder for Pandas Analysis)
  -- Advanced control over which fields are logged for each event type.
  -- This gives you maximum control over privacy and data granularity.
  -- For example, to stop logging the exact text you type:
  --   committed_text = false,
  -- To stop logging the list of candidates you saw:
  --   source_candidates_list = false,
  log_fields = {
    text_committed = {
      committed_text = true,
      input_sequence_at_commit = true,
      selection_method = true,
      selected_candidate_rank = true,
      source_input_buffer = true,
      source_first_candidate = true,
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
