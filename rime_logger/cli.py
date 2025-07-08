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
        """检查并报告当前日志记录器的安装状态。"""
        click.echo("--- Rime 日志记录器状态检查 ---")
        if not self.rime_user_dir:
            click.secho("❌ 未找到 Rime 用户目录。请问 Rime 是否已安装？", fg="red")
            return

        click.secho(f"✅ 找到 Rime 用户目录: {self.rime_user_dir}", fg="green")

        # 检查 Lua 脚本文件
        for file_name in [LOGGER_LUA_FILE, CONFIG_LUA_FILE]:
            path = self.rime_lua_dir / file_name
            if path.exists():
                click.secho(f"✅ 找到脚本: {path}", fg="green")
            else:
                click.secho(f"❌ 未找到脚本: {path}", fg="red")

        # 检查输入方案配置
        schema_path = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_path.exists():
            click.secho(f"❌ 未找到输入方案文件: {schema_path}", fg="red")
        else:
            try:
                with open(schema_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                if LINE_TO_ADD_IN_SCHEMA.strip() in content:
                    click.secho(f"✅ 输入方案 '{SCHEMA_YAML_FILE}' 已为日志记录器正确配置。", fg="green")
                else:
                    click.secho(f"❌ 输入方案 '{SCHEMA_YAML_FILE}'尚未配置。", fg="yellow")
            except Exception as e:
                click.secho(f"❓ 无法读取输入方案文件。错误: {e}", fg="yellow")

        if self.log_file_path and self.log_file_path.exists():
            click.secho(f"✅ 找到日志文件: {self.log_file_path}", fg="green")
        else:
            click.secho(f"❌ 在 '{self.log_file_path}' 未找到日志文件。请打字以生成日志。", fg="yellow")


    def install(self, preset: str):
        """安装日志记录器脚本并修改输入方案。"""
        click.echo("--- 开始安装日志记录器 ---")
        if not self.rime_user_dir:
            click.secho("❌ 错误: 未找到 Rime 用户目录，无法安装。", fg="red")
            sys.exit(1)

        click.echo(f"找到 Rime 目录: {self.rime_user_dir}")

        # 步骤 1: 根据选择的预设安装 Lua 脚本
        click.echo("\n--> 步骤 1: 复制 Lua 脚本...")
        self.rime_lua_dir.mkdir(exist_ok=True)

        src_logger = self.assets_dir / LOGGER_LUA_FILE
        dest_logger = self.rime_lua_dir / LOGGER_LUA_FILE
        shutil.copy(src_logger, dest_logger)
        click.echo(f"    [+] 已安装: {dest_logger}")

        # 修改并安装配置文件
        src_config_path = self.assets_dir / CONFIG_LUA_FILE
        dest_config_path = self.rime_lua_dir / CONFIG_LUA_FILE

        with open(src_config_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # 替换预设值
        new_content = re.sub(
            r'local preset_choice\s*=\s*".*"',
            f'local preset_choice = "{preset}"',
            content
        )

        with open(dest_config_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        click.echo(f"    [+] 已安装配置文件，预设为 '{preset}': {dest_config_path}")

        # 步骤 2: 修改输入方案文件
        click.echo("\n--> 步骤 2: 修改输入方案文件...")
        self._modify_schema_for_install()

        click.secho("\n--- ✅ 安装成功！ ---", fg="green", bold=True)
        click.secho("\n重要提示: 您必须立即“重新部署”Rime才能使更改生效。", fg="yellow")

    def _modify_schema_for_install(self):
        """封装安装过程中的输入方案修改操作。"""
        schema_file = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_file.exists():
            click.secho(f"    ❌ 错误: 未找到 '{schema_file}'。", fg="red")
            click.echo("    请确保已安装“万象”输入方案并已至少部署过一次。")
            sys.exit(1)

        try:
            with open(schema_file, 'r', encoding='utf-8') as f: lines = f.readlines()
            if any(LINE_TO_ADD_IN_SCHEMA.strip() in line for line in lines):
                click.echo("    [*] 输入方案已配置，无需更改。")
                return

            punctuator_index = -1
            for i, line in enumerate(lines):
                if 'punctuator' in line and line.strip().startswith('-'):
                    punctuator_index = i
                    break

            if punctuator_index == -1:
                click.secho("    ❌ 错误: 在输入方案中未找到 'punctuator' 入口。", fg="red")
                sys.exit(1)

            indentation = lines[punctuator_index][:lines[punctuator_index].find('-')]
            line_with_indent = f"{indentation}{LINE_TO_ADD_IN_SCHEMA.strip()}\n"
            lines.insert(punctuator_index + 1, line_with_indent)

            backup_file = schema_file.with_suffix('.yaml.bak')
            shutil.copy(schema_file, backup_file)
            click.echo(f"    [*] 已将原始输入方案备份至: {backup_file}")

            with open(schema_file, 'w', encoding='utf-8') as f: f.writelines(lines)
            click.secho(f"    [+] 已成功配置 '{SCHEMA_YAML_FILE}'。", fg="green")
        except Exception as e:
            click.secho(f"    ❌ 发生意外错误: {e}", fg="red")
            sys.exit(1)


    def uninstall(self):
        """卸载日志记录器组件。"""
        click.echo("--- 开始卸载日志记录器 ---")
        if not self.rime_user_dir:
            click.secho("❌ 错误: 未找到 Rime 用户目录。", fg="red")
            sys.exit(1)

        # 移除 Lua 脚本
        click.echo("\n--> 步骤 1: 移除 Lua 脚本...")
        logger_path = self.rime_lua_dir / LOGGER_LUA_FILE
        if logger_path.exists():
            logger_path.unlink()
            click.echo(f"    [-] 已移除: {logger_path}")

        if click.confirm(f"    您是否也想移除配置文件 '{CONFIG_LUA_FILE}'？"):
             config_path = self.rime_lua_dir / CONFIG_LUA_FILE
             if config_path.exists():
                config_path.unlink()
                click.echo(f"    [-] 已移除: {config_path}")

        # 恢复输入方案文件
        click.echo("\n--> 步骤 2: 恢复输入方案文件...")
        self._revert_schema_for_uninstall()

        click.secho("\n--- ✅ 卸载成功！ ---", fg="green", bold=True)
        click.secho("\n重要提示: 您必须立即“重新部署”Rime才能使更改生效。", bold=True, fg="yellow")

    def _revert_schema_for_uninstall(self):
        """封装卸载过程中的输入方案修改操作。"""
        schema_file = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_file.exists():
            click.echo("    [*] 未找到输入方案文件，跳过。")
            return
        try:
            with open(schema_file, 'r', encoding='utf-8') as f: lines = f.readlines()
            lines_to_keep = [line for line in lines if LINE_TO_ADD_IN_SCHEMA.strip() not in line]

            if len(lines) != len(lines_to_keep):
                with open(schema_file, 'w', encoding='utf-8') as f: f.writelines(lines_to_keep)
                click.secho(f"    [+] 已从 '{SCHEMA_YAML_FILE}' 中移除日志记录器配置。", fg="green")
            else:
                click.echo("    [*] 在输入方案中未找到日志记录器配置，无需更改。")
        except Exception as e:
             click.secho(f"    ❌ 发生意外错误: {e}", fg="red")

    def analyze(self):
        """分析收集到的日志数据。"""
        click.echo("--- 输入习惯分析 ---")
        if not self.log_file_path or not self.log_file_path.exists():
            click.secho(f"❌ 未找到日志文件: {self.log_file_path}", fg="red")
            return

        try:
            df = pd.read_json(self.log_file_path, lines=True)
            df_commit = df[df['event_type'] == 'text_committed'].copy()

            if df_commit.empty:
                click.secho("日志文件中未找到“text_committed”事件。", fg="yellow")
                return

            # --- 准确度指标 ---
            click.secho("\n## 预测准确度指标", bold=True)
            df_selections = df_commit[df_commit['selected_candidate_rank'] >= 0].copy()

            if df_selections.empty:
                 click.secho("未找到可供分析的有效候选词选择。", fg="yellow")
            else:
                total_selections = len(df_selections)
                first_choice_count = (df_selections['selected_candidate_rank'] == 0).sum()
                top_3_count = (df_selections['selected_candidate_rank'] < 3).sum()
                df_selections['accuracy_score'] = 1 / (df_selections['selected_candidate_rank'] + 1)
                overall_accuracy_score = df_selections['accuracy_score'].mean()

                click.echo(f"  - 总候选词选择数: {total_selections}")
                click.echo(f"  - 首选命中率:      {first_choice_count / total_selections:.2%}")
                click.echo(f"  - 前三候选命中率:   {top_3_count / total_selections:.2%}")
                click.echo(f"  - 平均选择排名:     {df_selections['selected_candidate_rank'].mean():.2f}")
                click.secho(f"  - 综合预测得分:   {overall_accuracy_score:.3f} / 1.000", fg="green")

            # --- 其他统计 ---
            click.secho("\n## 常规统计", bold=True)
            total_commits = len(df_commit)
            raw_input_commits = (df_commit['selected_candidate_rank'] == -1).sum()

            click.echo(f"  - 总上屏次数 (包括直接上屏): {total_commits}")
            if total_commits > 0:
                click.echo(f"  - 直接上屏率 (非候选词): {raw_input_commits / total_commits:.2%}")

        except Exception as e:
            click.secho(f"❌ 分析过程中发生错误: {e}", fg="red")


    def export_misses(self):
        """将预测错误数据导出到CSV文件。"""
        click.echo("--- 导出预测错误报告 ---")
        if not self.log_file_path or not self.log_file_path.exists():
            click.secho(f"❌ 未找到日志文件: {self.log_file_path}", fg="red")
            return

        try:
            df = pd.read_json(self.log_file_path, lines=True)
            df_commit = df[df['event_type'] == 'text_committed'].copy()

            # 筛选错误记录 (rank > 0)
            df_misses = df_commit[df_commit['selected_candidate_rank'] > 0].copy()

            if df_misses.empty:
                click.secho("✅ 未发现预测错误 (selected_candidate_rank > 0)。做得好！", fg="green")
                return

            # 为清晰起见，选择并重命名列
            report_cols = {
                'source_input_buffer': '用户输入',
                'committed_text': '实际选择',
                'source_first_candidate': '程序预测',
                'selected_candidate_rank': '选择排名'
            }
            # 确保在尝试选择列之前所有列都存在
            cols_to_select = [col for col in report_cols.keys() if col in df_misses.columns]
            df_report = df_misses[cols_to_select].rename(columns=report_cols)


            # 计算每个错误的频率并据此排序
            if '实际选择' in df_report.columns:
                df_report['错误频率'] = df_report.groupby('实际选择')['实际选择'].transform('count')
                df_report.sort_values(by=['错误频率', '用户输入'], ascending=[False, True], inplace=True)

            # 导出到 CSV
            export_path = Path.home() / "rime_mispredictions_report.csv"
            df_report.to_csv(export_path, index=False, encoding='utf-8-sig')

            click.secho(f"✅ 成功导出 {len(df_report)} 条预测错误记录。", fg="green")
            click.echo("最常见的错误位于文件顶部。")
            click.echo(f"报告已保存至: {export_path}")

        except Exception as e:
            click.secho(f"❌ 导出过程中发生错误: {e}", fg="red")

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
        "✅ 普通模式 (Normal) - 推荐，用于计算输入法预测准确率": "normal",
        "👩‍💻 词库贡献者模式 (Developer) - 用于调试，关注非首选上屏": "developer",
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
