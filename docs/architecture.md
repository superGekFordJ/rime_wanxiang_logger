# Project Architecture

This document outlines the architecture of the Rime Input Habit Logger project.

## Overview

The project is a Python-based command-line tool designed to install, manage, and analyze data from a Lua-based logger for the Rime input method engine. The primary goal is to collect data on user input habits to help improve Rime's prediction accuracy and provide insights into typing patterns.

## Directory Structure

```
.
├── .gitignore
├── LICENSE
├── README.md
├── docs/                        # Documentation files
│   └── architecture.md          # This file
│   └── lua_script_architecture_zh.md # Lua script architecture (Chinese)
├── rime_logger/                 # Main Python package
│   ├── __init__.py
│   ├── assets/                  # Lua scripts and other assets
│   │   ├── input_habit_logger.lua
│   │   └── input_habit_logger_config.lua
│   ├── cli.py                   # Command-line interface logic
├── setup.py                     # Packaging and distribution script
```

## Core Components

### 1. Python Package: `rime_logger`

This is the main application logic written in Python.

#### a. `rime_logger.cli.py`

*   **Purpose:** Provides the command-line interface (CLI) for users to interact with the logger.
*   **Key Features:**
    *   Installation of Lua scripts into the Rime user directory.
    *   Uninstallation of Lua scripts.
    *   Checking the status of the logger.
    *   Analyzing collected log data.
    *   Exporting misprediction reports.
*   **Core Class: `RimeManager`**
    *   Handles detection of the Rime user directory across different operating systems (Windows, macOS, Linux).
    *   Manages file operations for copying Lua scripts and modifying Rime schema files.
    *   Parses the Lua configuration to find the active log file path.
    *   Uses `pandas` for data analysis of the JSONL log files.
*   **Dependencies:** `click` (for CLI creation), `questionary` (for interactive prompts), `pandas` (for data analysis).

#### b. `rime_logger.assets/`

*   **Purpose:** Contains the Lua scripts that are deployed to the Rime user's directory.
    *   `input_habit_logger.lua`: The main Lua script that integrates with Rime's `lua_processor` to capture input events.
    *   `input_habit_logger_config.lua`: A configuration file for the Lua logger, allowing users to customize logging behavior (e.g., log file path, logging level/preset).

### 2. `setup.py`

*   **Purpose:** Standard Python script for packaging and distributing the `rime-wanxiang-logger` tool.
*   **Key Functions:**
    *   Defines package metadata (name, version, author, etc.).
    *   Specifies dependencies (`pandas`, `click`, `questionary`).
    *   Includes the `rime_logger/assets` directory (containing Lua scripts) in the distributable package using `package_data`.
    *   Defines the entry point for the command-line script `rime-logger`, which maps to the `main` function in `rime_logger.cli`.

## Workflow

1.  **Installation (`rime-logger install`):**
    *   The user runs the install command.
    *   `RimeManager` detects the Rime user directory.
    *   The Lua scripts (`input_habit_logger.lua` and a preset-configured `input_habit_logger_config.lua`) are copied from `rime_logger/assets` to the Rime user's `lua` directory.
    *   The Rime schema file (e.g., `wanxiang.schema.yaml`) is modified to include a `lua_processor` entry that points to `input_habit_logger.lua`. A backup of the original schema is created.
    *   The user is prompted to "redeploy" Rime for changes to take effect.

2.  **Logging (by Rime):**
    *   As the user types, Rime's `lua_processor` invokes `input_habit_logger.lua`.
    *   The Lua script, based on its configuration in `input_habit_logger_config.lua`, captures relevant input events (e.g., text committed, candidate selected).
    *   These events are written to a JSONL (JSON Lines) file specified in the Lua config.

3.  **Analysis (`rime-logger analyze`):**
    *   The user runs the analyze command.
    *   `RimeManager` locates the log file (parsing `input_habit_logger_config.lua` if necessary).
    *   The JSONL log data is loaded into a `pandas` DataFrame.
    *   Various metrics are calculated, such as prediction accuracy (first choice hit rate, top-3 hit rate) and general typing statistics.
    *   The results are displayed to the user.

4.  **Export Misses (`rime-logger export-misses`):**
    *   The user runs the export command.
    *   Log data is processed similarly to the analysis step.
    *   Entries where the selected candidate was not the first one predicted by Rime (i.e., `selected_candidate_rank > 0`) are filtered.
    *   A CSV report of these "mispredictions" is generated, showing the user's input, the actual selection, and what Rime predicted as the first candidate.

5.  **Uninstallation (`rime-logger uninstall`):**
    *   The Lua scripts are removed from the Rime user's `lua` directory.
    *   The modifications made to the Rime schema file are reverted (the added `lua_processor` line is removed).
    *   The user is prompted to "redeploy" Rime.

## Key Design Decisions

*   **Python for CLI & Management:** Python is well-suited for CLI tools, file system operations, and data analysis (with `pandas`).
*   **Lua for Rime Integration:** Rime uses Lua for scripting and extending its functionality via `lua_processor`.
*   **JSONL for Logging:** JSON Lines format is chosen for its simplicity and ease of parsing, where each line is a valid JSON object. This is efficient for appending logs and for stream processing.
*   **Configuration Presets:** The Lua config offers presets to simplify setup for users with different needs (normal, developer, advanced), while still allowing custom configurations.
*   **Cross-Platform Support:** The `RimeManager` attempts to automatically detect Rime's user directory on Windows, macOS, and common Linux setups.
*   **Schema Backup:** Important user configuration files like the Rime schema are backed up before modification.

## Future Considerations

*   More sophisticated data analysis and visualization.
*   GUI for easier management and analysis.
*   Support for more Rime schemas beyond "wanxiang".
*   Automatic Rime redeployment (if feasible and safe).
