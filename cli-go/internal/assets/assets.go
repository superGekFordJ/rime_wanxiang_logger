// Package assets provides access to the embedded Lua scripts.
package assets

import _ "embed"

//go:embed input_habit_logger.lua
var LoggerScript []byte

//go:embed input_habit_logger_config.lua
var ConfigScript []byte
