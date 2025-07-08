import os
import sys
import platform
import shutil
import re
from pathlib import Path

import click
import questionary
import pandas as pd

# --- Constants ---
LOGGER_LUA_FILE = "input_habit_logger.lua"
CONFIG_LUA_FILE = "input_habit_logger_config.lua"
SCHEMA_YAML_FILE = "wanxiang.schema.yaml"
DEFAULT_LOG_JSONL_FILE = "input_habit_log_structured.jsonl"
LINE_TO_ADD_IN_SCHEMA = "      - lua_processor@*input_habit_logger #输入习惯记录器 - 记录用户输入习惯"

class RimeManager:
    """
    Manages the installation, uninstallation, and analysis of the Rime logger.
    """
    def __init__(self):
        self.rime_user_dir = self._get_rime_user_dir()
        self.script_dir = Path(__file__).parent.resolve()
        self.assets_dir = self.script_dir / "assets"
        self.rime_lua_dir = self.rime_user_dir / "lua" if self.rime_user_dir else None
        self.log_file_path = self._get_log_file_path()

    def _get_rime_user_dir(self) -> Path | None:
        """Detects the Rime user directory based on the operating system."""
        system = platform.system()
        home = Path.home()
        path_map = {
            "Windows": Path(os.getenv("APPDATA", "")) / "Rime",
            "Darwin": home / "Library" / "Rime",
            "Linux": [
                home / ".config" / "rime",
                home / ".config" / "fcitx" / "rime",
                home / ".config" / "fcitx5" / "rime",
                home / ".config" / "ibus" / "rime",
            ]
        }
        if system == "Linux":
            for path in path_map["Linux"]:
                if path.is_dir(): return path
            return None
        path = path_map.get(system)
        return path if path and path.is_dir() else None

    def _get_log_file_path(self) -> Path | None:
        """
        Gets the log file path by correctly parsing the active preset in the config.
        """
        if not self.rime_user_dir:
            return None

        default_path = self.rime_user_dir / DEFAULT_LOG_JSONL_FILE
        config_file = self.rime_lua_dir / CONFIG_LUA_FILE

        if not config_file.exists():
            return default_path

        try:
            with open(config_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # 1. Find the active preset choice
            preset_choice_match = re.search(r'local\s+preset_choice\s*=\s*"([^"]+)"', content)
            if not preset_choice_match:
                return default_path

            active_preset = preset_choice_match.group(1)

            # 2. Find the configuration block for the active preset
            # This regex looks for `preset_name = {` and captures everything until the matching `}`
            preset_block_regex = re.compile(
                r'^\s*' + re.escape(active_preset) + r'\s*=\s*\{([\s\S]*?)\n\s*\}',
                re.MULTILINE
            )
            preset_block_match = preset_block_regex.search(content)
            if not preset_block_match:
                return default_path

            preset_content = preset_block_match.group(1)

            # 3. Search for a non-commented log_file_path within that block
            log_path_match = re.search(r'^\s*log_file_path\s*=\s*"([^"]+)"', preset_content, re.MULTILINE)

            if log_path_match:
                # Path must be un-escaped (e.g., "C:\\Users" -> "C:\Users")
                custom_path_str = log_path_match.group(1).replace('\\\\', '\\')
                if custom_path_str: # Ensure it's not an empty string
                    click.echo(f"Found active custom log path: {custom_path_str}")
                    return Path(custom_path_str)

        except Exception as e:
            click.secho(f"Warning: Could not parse config to find custom log path. Using default. Error: {e}", fg="yellow")
            pass

        return default_path

    def check_status(self):
        """Checks and reports the current installation status of the logger."""
        click.echo("--- Rime Logger Status ---")
        if not self.rime_user_dir:
            click.secho("❌ Rime user directory not found. Is Rime installed?", fg="red")
            return

        click.secho(f"✅ Rime user directory found: {self.rime_user_dir}", fg="green")

        # Check for Lua files
        for file_name in [LOGGER_LUA_FILE, CONFIG_LUA_FILE]:
            path = self.rime_lua_dir / file_name
            if path.exists():
                click.secho(f"✅ Found script: {path}", fg="green")
            else:
                click.secho(f"❌ Not found: {path}", fg="red")

        # Check schema configuration
        schema_path = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_path.exists():
            click.secho(f"❌ Schema file not found: {schema_path}", fg="red")
        else:
            try:
                with open(schema_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                if LINE_TO_ADD_IN_SCHEMA.strip() in content:
                    click.secho(f"✅ Schema '{SCHEMA_YAML_FILE}' is configured for logger.", fg="green")
                else:
                    click.secho(f"❌ Schema '{SCHEMA_YAML_FILE}' is not configured.", fg="yellow")
            except Exception as e:
                click.secho(f"❓ Could not read schema file. Error: {e}", fg="yellow")

        if self.log_file_path and self.log_file_path.exists():
            click.secho(f"✅ Log file found: {self.log_file_path}", fg="green")
        else:
            click.secho(f"❌ Log file not found at '{self.log_file_path}'. Type something to generate it.", fg="yellow")


    def install(self, preset: str):
        """Installs the logger scripts and modifies the schema."""
        click.echo("--- Starting Logger Installation ---")
        if not self.rime_user_dir:
            click.secho("❌ ERROR: Rime user directory not found. Cannot install.", fg="red")
            sys.exit(1)

        click.echo(f"Found Rime directory: {self.rime_user_dir}")

        # Step 1: Install Lua scripts with selected preset
        click.echo("\n--> Step 1: Copying Lua scripts...")
        self.rime_lua_dir.mkdir(exist_ok=True)

        src_logger = self.assets_dir / LOGGER_LUA_FILE
        dest_logger = self.rime_lua_dir / LOGGER_LUA_FILE
        shutil.copy(src_logger, dest_logger)
        click.echo(f"    [+] Installed: {dest_logger}")

        # Modify and install config file
        src_config_path = self.assets_dir / CONFIG_LUA_FILE
        dest_config_path = self.rime_lua_dir / CONFIG_LUA_FILE

        with open(src_config_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Replace the preset value
        new_content = re.sub(
            r'local preset_choice\s*=\s*".*"',
            f'local preset_choice = "{preset}"',
            content
        )

        with open(dest_config_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        click.echo(f"    [+] Installed config with '{preset}' preset: {dest_config_path}")

        # Step 2: Modify schema file
        click.echo("\n--> Step 2: Modifying schema file...")
        self._modify_schema_for_install()

        click.secho("\n--- ✅ Installation successful! ---", fg="green", bold=True)
        click.secho("\nIMPORTANT: You must 'Re-deploy' Rime now for changes to take effect.", fg="yellow")

    def _modify_schema_for_install(self):
        """Helper to encapsulate schema modification during installation."""
        schema_file = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_file.exists():
            click.secho(f"    ❌ ERROR: '{schema_file}' not found.", fg="red")
            click.echo("    Please ensure the Wanxiang schema is installed and deploy Rime once.")
            sys.exit(1)

        try:
            with open(schema_file, 'r', encoding='utf-8') as f: lines = f.readlines()
            if any(LINE_TO_ADD_IN_SCHEMA.strip() in line for line in lines):
                click.echo("    [*] Schema already configured. No changes needed.")
                return

            punctuator_index = -1
            for i, line in enumerate(lines):
                if 'punctuator' in line and line.strip().startswith('-'):
                    punctuator_index = i
                    break

            if punctuator_index == -1:
                click.secho("    ❌ ERROR: Could not find 'punctuator' entry in schema.", fg="red")
                sys.exit(1)

            indentation = lines[punctuator_index][:lines[punctuator_index].find('-')]
            line_with_indent = f"{indentation}{LINE_TO_ADD_IN_SCHEMA.strip()}\n"
            lines.insert(punctuator_index + 1, line_with_indent)

            backup_file = schema_file.with_suffix('.yaml.bak')
            shutil.copy(schema_file, backup_file)
            click.echo(f"    [*] Backed up original schema to: {backup_file}")

            with open(schema_file, 'w', encoding='utf-8') as f: f.writelines(lines)
            click.secho(f"    [+] Successfully configured '{SCHEMA_YAML_FILE}'.", fg="green")
        except Exception as e:
            click.secho(f"    ❌ An unexpected error occurred: {e}", fg="red")
            sys.exit(1)


    def uninstall(self):
        """Removes the logger components."""
        click.echo("--- Starting Logger Uninstallation ---")
        if not self.rime_user_dir:
            click.secho("❌ ERROR: Rime user directory not found.", fg="red")
            sys.exit(1)

        # Remove Lua scripts
        click.echo("\n--> Step 1: Removing Lua scripts...")
        logger_path = self.rime_lua_dir / LOGGER_LUA_FILE
        if logger_path.exists():
            logger_path.unlink()
            click.echo(f"    [-] Removed: {logger_path}")

        if click.confirm(f"    Do you want to remove the config file '{CONFIG_LUA_FILE}' as well?"):
             config_path = self.rime_lua_dir / CONFIG_LUA_FILE
             if config_path.exists():
                config_path.unlink()
                click.echo(f"    [-] Removed: {config_path}")

        # Revert schema file
        click.echo("\n--> Step 2: Reverting schema file...")
        self._revert_schema_for_uninstall()

        click.secho("\n--- ✅ Uninstallation successful! ---", fg="green", bold=True)
        click.secho("\nIMPORTANT: You must 'Re-deploy' Rime now for changes to take effect.", fg="yellow")

    def _revert_schema_for_uninstall(self):
        """Helper to encapsulate schema modification during uninstallation."""
        schema_file = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_file.exists():
            click.echo("    [*] Schema file not found, skipping.")
            return
        try:
            with open(schema_file, 'r', encoding='utf-8') as f: lines = f.readlines()
            lines_to_keep = [line for line in lines if LINE_TO_ADD_IN_SCHEMA.strip() not in line]

            if len(lines) != len(lines_to_keep):
                with open(schema_file, 'w', encoding='utf-8') as f: f.writelines(lines_to_keep)
                click.secho(f"    [+] Removed logger configuration from '{SCHEMA_YAML_FILE}'.", fg="green")
            else:
                click.echo("    [*] Logger configuration not found in schema. No changes needed.")
        except Exception as e:
             click.secho(f"    ❌ An unexpected error occurred: {e}", fg="red")

    def analyze(self):
        """Analyzes the collected log data."""
        click.echo("--- Analyzing Typing Data ---")
        if not self.log_file_path or not self.log_file_path.exists():
            click.secho(f"❌ Log file not found at: {self.log_file_path}", fg="red")
            return

        try:
            df = pd.read_json(self.log_file_path, lines=True)
            df_commit = df[df['event_type'] == 'text_committed'].copy()

            if df_commit.empty:
                click.secho("No 'text_committed' events found in the log file.", fg="yellow")
                return

            # --- Accuracy Metrics ---
            click.secho("\n## Prediction Accuracy Metrics", bold=True)
            df_selections = df_commit[df_commit['selected_candidate_rank'] >= 0].copy()

            if df_selections.empty:
                 click.secho("No valid candidate selections found to analyze.", fg="yellow")
            else:
                total_selections = len(df_selections)
                first_choice_count = (df_selections['selected_candidate_rank'] == 0).sum()
                top_3_count = (df_selections['selected_candidate_rank'] < 3).sum()
                df_selections['accuracy_score'] = 1 / (df_selections['selected_candidate_rank'] + 1)
                overall_accuracy_score = df_selections['accuracy_score'].mean()

                click.echo(f"  - Total Candidate Selections: {total_selections}")
                click.echo(f"  - First-Choice Accuracy:      {first_choice_count / total_selections:.2%}")
                click.echo(f"  - Top-3 Accuracy:             {top_3_count / total_selections:.2%}")
                click.echo(f"  - Average Candidate Rank:     {df_selections['selected_candidate_rank'].mean():.2f}")
                click.secho(f"  - Overall Prediction Score:   {overall_accuracy_score:.3f} / 1.000", fg="green")

            # --- Other Stats ---
            click.secho("\n## General Statistics", bold=True)
            total_commits = len(df_commit)
            raw_input_commits = (df_commit['selected_candidate_rank'] == -1).sum()

            click.echo(f"  - Total Commits (incl. raw): {total_commits}")
            if total_commits > 0:
                click.echo(f"  - Raw Input Rate (non-candidate): {raw_input_commits / total_commits:.2%}")

        except Exception as e:
            click.secho(f"❌ An error occurred during analysis: {e}", fg="red")

    def export_misses(self):
        """Exports mis-prediction data to a CSV file."""
        click.echo("--- Exporting Mis-prediction Report ---")
        if not self.log_file_path or not self.log_file_path.exists():
            click.secho(f"❌ Log file not found at: {self.log_file_path}", fg="red")
            return

        try:
            df = pd.read_json(self.log_file_path, lines=True)
            df_commit = df[df['event_type'] == 'text_committed'].copy()

            # Filter for misses (rank > 0)
            df_misses = df_commit[df_commit['selected_candidate_rank'] > 0].copy()

            if df_misses.empty:
                click.secho("✅ No mis-predictions found (selected_candidate_rank > 0). Great job!", fg="green")
                return

            # Select and rename columns for clarity
            report_cols = {
                'source_input_buffer': 'User_Input',
                'committed_text': 'Selected_Text',
                'source_first_candidate': 'Predicted_Text',
                'selected_candidate_rank': 'Selected_Rank'
            }
            # Ensure all columns exist before trying to select them
            cols_to_select = [col for col in report_cols.keys() if col in df_misses.columns]
            df_report = df_misses[cols_to_select].rename(columns=report_cols)


            # Calculate frequency of each mistake and sort by it
            if 'Selected_Text' in df_report.columns:
                df_report['miss_frequency'] = df_report.groupby('Selected_Text')['Selected_Text'].transform('count')
                df_report.sort_values(by=['miss_frequency', 'User_Input'], ascending=[False, True], inplace=True)

            # Export to CSV
            export_path = Path.home() / "rime_mispredictions_report.csv"
            df_report.to_csv(export_path, index=False, encoding='utf-8-sig')

            click.secho(f"✅ Successfully exported {len(df_report)} mis-predictions.", fg="green")
            click.echo("The most common mistakes are at the top of the file.")
            click.echo(f"Report saved to: {export_path}")

        except Exception as e:
            click.secho(f"❌ An error occurred during export: {e}", fg="red")

# --- CLI using Click ---
@click.group(context_settings=dict(help_option_names=['-h', '--help']))
@click.pass_context
def main(ctx):
    """A tool to manage the Rime Input Habit Logger."""
    ctx.obj = RimeManager()

@main.command()
@click.pass_context
def install(ctx):
    """Install the logger with an interactive preset selection."""
    preset_map = {
        "✅ 普通模式 (Normal) - 推荐，用于计算基本打字准确率": "normal",
        "👩‍💻 开发者模式 (Developer) - 用于调试，关注非首选上屏": "developer",
        "🔬 高级模式 (Advanced) - 记录几乎所有信息，用于深度分析": "advanced",
        "⚙️ 自定义 (Custom) - (需要手动修改配置文件)": "custom"
    }

    choice = questionary.select(
        "请选择一个日志记录预设模式:",
        choices=list(preset_map.keys())
    ).ask()

    if choice is None:
        click.echo("Installation cancelled.")
        return

    selected_preset = preset_map[choice]

    if selected_preset == 'custom':
        click.secho("您选择了 '自定义' 模式。", fg="yellow")
        click.echo("请先使用其他模式安装，然后手动编辑以下文件以符合您的需求:")
        manager = ctx.obj
        if manager.rime_lua_dir:
            click.secho(str(manager.rime_lua_dir / CONFIG_LUA_FILE), fg="cyan")
        else:
            click.secho("(Rime 目录未找到，请先安装 Rime)", fg="red")
        return

    ctx.obj.install(preset=selected_preset)


@main.command()
@click.pass_context
def uninstall(ctx):
    """Uninstall the logger and clean the schema."""
    ctx.obj.uninstall()

@main.command()
@click.pass_context
def status(ctx):
    """Check the current installation status."""
    ctx.obj.check_status()

@main.command()
@click.pass_context
def analyze(ctx):
    """Analyze the collected typing data."""
    ctx.obj.analyze()

@main.command(name='export-misses')
@click.pass_context
def export_misses(ctx):
    """Export a CSV report of typing mis-predictions."""
    ctx.obj.export_misses()

if __name__ == "__main__":
    main()
