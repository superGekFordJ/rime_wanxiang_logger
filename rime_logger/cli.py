import os
import sys
import platform
import shutil
import argparse
from pathlib import Path

# --- Constants ---
# This makes it easy to refer to the script files
LOGGER_LUA_FILE = "input_habit_logger.lua"
CONFIG_LUA_FILE = "input_habit_logger_config.lua"
SCHEMA_YAML_FILE = "wanxiang.schema.yaml"
LINE_TO_ADD_IN_SCHEMA = "      - lua_processor@*input_habit_logger #输入习惯记录器 - 记录用户输入习惯"

class RimeManager:
    """
    Manages the installation, uninstallation, and status check
    of the Rime Input Habit Logger.
    """
    def __init__(self):
        self.rime_user_dir = self._get_rime_user_dir()
        # The script's own path is needed to find the 'assets' directory
        self.script_dir = Path(__file__).parent.resolve()
        self.assets_dir = self.script_dir / "assets"
        self.rime_lua_dir = self.rime_user_dir / "lua" if self.rime_user_dir else None

    def _get_rime_user_dir(self) -> Path | None:
        """Detects the Rime user directory based on the operating system."""
        system = platform.system()
        home = Path.home()

        path_map = {
            "Windows": Path(os.getenv("APPDATA", "")) / "Rime",
            "Darwin": home / "Library" / "Rime",
            # Common Linux paths, we will check them in order
            "Linux": [
                home / ".config" / "rime",
                home / ".config" / "fcitx" / "rime",
                home / ".config" / "fcitx5" / "rime",
                home / ".config" / "ibus" / "rime",
            ]
        }

        if system == "Linux":
            for path in path_map["Linux"]:
                if path.is_dir():
                    return path
            return None # None found

        path = path_map.get(system)
        if path and path.is_dir():
            return path

        return None

    def check_status(self):
        """Checks and reports the current installation status of the logger."""
        print("--- Rime Logger Status ---")
        if not self.rime_user_dir:
            print("❌ Rime user directory not found. Is Rime installed?")
            return

        print(f"✅ Rime user directory found: {self.rime_user_dir}")

        # Check for Lua files
        logger_path = self.rime_lua_dir / LOGGER_LUA_FILE
        config_path = self.rime_lua_dir / CONFIG_LUA_FILE

        if logger_path.exists():
            print(f"✅ Logger script found: {logger_path}")
        else:
            print(f"❌ Logger script not found.")

        if config_path.exists():
            print(f"✅ Config script found: {config_path}")
        else:
            print(f"❌ Config script not found.")

        # Check schema configuration
        schema_path = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_path.exists():
            print(f"❌ Schema file not found: {schema_path}")
        else:
            try:
                with open(schema_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                if LINE_TO_ADD_IN_SCHEMA.strip() in content:
                    print(f"✅ Schema '{SCHEMA_YAML_FILE}' is configured for logger.")
                else:
                    print(f"❌ Schema '{SCHEMA_YAML_FILE}' is not configured.")
            except Exception as e:
                print(f"❓ Could not read schema file. Error: {e}")

    def install(self):
        """Installs the logger scripts and modifies the schema."""
        print("--- Starting Logger Installation ---")
        if not self.rime_user_dir:
            print("❌ ERROR: Rime user directory not found. Cannot install.")
            sys.exit(1)

        print(f"Found Rime directory: {self.rime_user_dir}")

        # 1. Install Lua scripts
        print("\n--> Step 1: Copying Lua scripts...")
        self.rime_lua_dir.mkdir(exist_ok=True)

        src_logger = self.assets_dir / LOGGER_LUA_FILE
        dest_logger = self.rime_lua_dir / LOGGER_LUA_FILE
        shutil.copy(src_logger, dest_logger)
        print(f"    [+] Installed: {dest_logger}")

        src_config = self.assets_dir / CONFIG_LUA_FILE
        dest_config = self.rime_lua_dir / CONFIG_LUA_FILE
        if dest_config.exists():
            print(f"    [*] Config file already exists. Skipping to preserve user settings.")
        else:
            shutil.copy(src_config, dest_config)
            print(f"    [+] Installed: {dest_config}")

        # 2. Modify schema file
        print("\n--> Step 2: Modifying schema file...")
        schema_file = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_file.exists():
            print(f"    ❌ ERROR: '{schema_file}' not found.")
            print("    Please ensure the Wanxiang schema is installed and deploy Rime once.")
            sys.exit(1)

        try:
            with open(schema_file, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            # Idempotency check
            if any(LINE_TO_ADD_IN_SCHEMA.strip() in line for line in lines):
                print("    [*] Schema already configured. No changes needed.")
            else:
                # Find the 'punctuator' line to insert after
                punctuator_index = -1
                for i, line in enumerate(lines):
                    if 'punctuator' in line and line.strip().startswith('-'):
                        punctuator_index = i
                        break

                if punctuator_index == -1:
                    print("    ❌ ERROR: Could not find 'punctuator' entry in schema.")
                    print("    Please add the following line manually under `engine/processors`:")
                    print(f"    {LINE_TO_ADD_IN_SCHEMA}")
                    sys.exit(1)

                # Insert the line
                indentation = lines[punctuator_index][:lines[punctuator_index].find('-')]
                line_with_indent = f"{indentation}{LINE_TO_ADD_IN_SCHEMA.strip()}\n"
                lines.insert(punctuator_index + 1, line_with_indent)

                # Backup and write
                backup_file = schema_file.with_suffix('.yaml.bak')
                shutil.copy(schema_file, backup_file)
                print(f"    [*] Backed up original schema to: {backup_file}")

                with open(schema_file, 'w', encoding='utf-8') as f:
                    f.writelines(lines)
                print(f"    [+] Successfully configured '{SCHEMA_YAML_FILE}'.")

            print("\n--- ✅ Installation successful! ---")
            print("\nIMPORTANT: You must 'Re-deploy' Rime now for changes to take effect.")

        except Exception as e:
            print(f"    ❌ An unexpected error occurred: {e}")
            sys.exit(1)

    def uninstall(self):
        """Removes the logger scripts and cleans the schema."""
        print("--- Starting Logger Uninstallation ---")
        if not self.rime_user_dir:
            print("❌ ERROR: Rime user directory not found. Cannot uninstall.")
            sys.exit(1)

        # 1. Remove Lua scripts
        print("\n--> Step 1: Removing Lua scripts...")
        logger_path = self.rime_lua_dir / LOGGER_LUA_FILE
        config_path = self.rime_lua_dir / CONFIG_LUA_FILE

        if logger_path.exists():
            logger_path.unlink()
            print(f"    [-] Removed: {logger_path}")
        else:
            print(f"    [*] Logger script not found, skipping.")

        if config_path.exists():
            config_path.unlink()
            print(f"    [-] Removed: {config_path}")
        else:
            print(f"    [*] Config script not found, skipping.")

        # 2. Revert schema file
        print("\n--> Step 2: Reverting schema file...")
        schema_file = self.rime_user_dir / SCHEMA_YAML_FILE
        if not schema_file.exists():
            print(f"    [*] Schema file not found, skipping.")
        else:
            try:
                with open(schema_file, 'r', encoding='utf-8') as f:
                    lines = f.readlines()

                lines_to_keep = [line for line in lines if LINE_TO_ADD_IN_SCHEMA.strip() not in line]

                if len(lines) == len(lines_to_keep):
                    print("    [*] Logger configuration not found in schema. No changes needed.")
                else:
                    backup_file = schema_file.with_suffix('.yaml.uninstall.bak')
                    shutil.copy(schema_file, backup_file)
                    print(f"    [*] Backed up schema to: {backup_file}")

                    with open(schema_file, 'w', encoding='utf-8') as f:
                        f.writelines(lines_to_keep)
                    print(f"    [+] Removed logger configuration from '{SCHEMA_YAML_FILE}'.")

                print("\n--- ✅ Uninstallation successful! ---")
                print("\nIMPORTANT: You must 'Re-deploy' Rime now for changes to take effect.")

            except Exception as e:
                print(f"    ❌ An unexpected error occurred: {e}")
                sys.exit(1)

def main():
    """Main entry point for the command-line interface."""
    parser = argparse.ArgumentParser(
        description="A command-line tool to manage the Rime Input Habit Logger."
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    # Install command
    parser_install = subparsers.add_parser("install", help="Install the logger.")
    parser_install.set_defaults(func=lambda mgr: mgr.install())

    # Uninstall command
    parser_uninstall = subparsers.add_parser("uninstall", help="Uninstall the logger.")
    parser_uninstall.set_defaults(func=lambda mgr: mgr.uninstall())

    # Status command
    parser_status = subparsers.add_parser("status", help="Check the installation status.")
    parser_status.set_defaults(func=lambda mgr: mgr.check_status())

    args = parser.parse_args()

    manager = RimeManager()
    args.func(manager)

if __name__ == "__main__":
    main()
