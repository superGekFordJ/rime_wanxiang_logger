# Rime Input Habit Logger

This project provides a Lua script and a Python-based command-line tool to install and manage a data logger for the [Rime Input Method Engine](https://rime.im/). It is designed to capture detailed information about your typing habits, saving it to a structured JSONL file for later analysis.

The primary goal is to gather data that can be used to analyze typing efficiency, identify common errors, and potentially fine-tune your Rime schema for a better, more personalized input experience.

## Features

- **Detailed Event Logging:** Captures keypresses, candidate lists, and final text selections.
- **Rich Context:** Records *how* you select a word (space vs. number key) and the rank of your chosen candidate.
- **Configurable:** You have full control over what gets logged via a simple Lua configuration file.
- **Cross-Platform:** The installer automatically detects the Rime user directory on Windows, macOS, and Linux.
- **Easy Management:** A simple command-line interface (`rime-logger`) to install, uninstall, and check the status of the logger.
- **Ready for Analysis:** Outputs structured JSONL data, perfect for processing with tools like Python's Pandas library.

## Installation

You will need Python 3.7+ installed on your system.

1.  Clone this repository or download the source code:
    ```bash
    git clone https://github.com/your-username/rime-wanxiang-logger.git
    cd rime-wanxiang-logger
    ```

2.  Install the package using pip. This will install the necessary files and make the `rime-logger` command available system-wide.
    ```bash
    pip install .
    ```

## Usage

The installer provides a single command-line tool, `rime-logger`, to manage the installation.

**IMPORTANT:** After running `install` or `uninstall`, you **must** re-deploy Rime for the changes to take effect. You can usually do this by clicking the Rime icon in your system's menu/taskbar and selecting "Deploy" (部署).

### To Install the Logger

This command now starts an interactive session, allowing you to choose a logging preset directly from the command line.

```bash
rime-logger install
```

You will be prompted to select a mode like "Normal", "Developer", or "Advanced".

### To Analyze Your Typing Habits

The `analyze` command reads your log file and provides a statistical summary of your typing accuracy, including first-choice accuracy, top-3 accuracy, and an overall prediction score.

```bash
rime-logger analyze
```

### To Export Data for Developers

If you want to help improve the input schema, you can use the `export-misses` command. It creates a `rime_mispredictions_report.csv` file in your home directory, containing all instances where you didn't choose the first candidate. This file can be easily shared with developers.

```bash
rime-logger export-misses
```

### To Uninstall the Logger

This will remove the logger from your schema file and delete the main Lua script. It will leave the configuration file untouched to preserve your settings.

```bash
rime-logger uninstall
```

### To Check the Status

This will report whether the necessary files are in place and if the schema is configured correctly.

```bash
rime-logger status
```

## Configuration

When you install the logger, a configuration file named `input_habit_logger_config.lua` is created in your Rime `lua` directory. You can edit this file to control what data is logged.

Example `input_habit_logger_config.lua`:
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

## The Log File

The logger will create a file named `input_habit_log_structured.jsonl` in your main Rime user directory. Each line in this file is a separate JSON object representing a single event.

This format is ideal for processing with data analysis tools. You can easily load it into a Pandas DataFrame in Python, for example:

```python
import pandas as pd

# The log file is typically in the root of the Rime user directory
log_file_path = 'path/to/your/Rime/input_habit_log_structured.jsonl'
df = pd.read_json(log_file_path, lines=True)

# Start your analysis!
print(df.info())
```

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.