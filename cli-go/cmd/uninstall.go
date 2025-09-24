package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"rime-wanxiang-logger-go/internal/manager"
	"rime-wanxiang-logger-go/internal/ui"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the logger scripts from the Rime user directory.",
	Long: `This command removes the Lua scripts and warns the user to revert changes
made to the Rime schema file, effectively disabling the logger.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Section("开始卸载日志记录器")

		rimeManager, err := manager.NewRimeManager()
		if err != nil {
			return fmt.Errorf("could not initialize Rime manager: %w", err)
		}
		ui.Successf("检测到 Rime 用户目录: %s", rimeManager.UserDirectory)

		loggerScriptPath := filepath.Join(rimeManager.GetLuaDirectory(), "input_habit_logger.lua")
		configScriptPath := filepath.Join(rimeManager.GetLuaDirectory(), "input_habit_logger_config.lua")

		// 步骤 1: 移除 Lua 脚本...
		ui.Subsection("步骤 1: 移除 Lua 脚本...")

		if _, err := os.Stat(loggerScriptPath); err == nil {
			if err := os.Remove(loggerScriptPath); err != nil {
				return fmt.Errorf("failed to delete 'input_habit_logger.lua': %w", err)
			}
			ui.Successf("已移除: %s", loggerScriptPath)
		} else {
			ui.Warnf("未找到: %s", loggerScriptPath)
		}

		// 询问是否移除配置文件
		if _, err := os.Stat(configScriptPath); err == nil {
			confirm := promptui.Prompt{
				Label:     fmt.Sprintf("是否也移除配置文件 '%s'？", filepath.Base(configScriptPath)),
				IsConfirm: true,
			}
			if _, err := confirm.Run(); err == nil {
				if err := os.Remove(configScriptPath); err != nil {
					return fmt.Errorf("failed to delete 'input_habit_logger_config.lua': %w", err)
				}
				ui.Successf("已移除: %s", configScriptPath)
			} else {
				ui.Warnf("已保留配置文件: %s", configScriptPath)
			}
		} else {
			ui.Warnf("未找到: %s", configScriptPath)
		}

		// 步骤 2: 恢复输入方案文件...
		ui.Subsection("步骤 2: 恢复输入方案文件...")
		if err := rimeManager.RevertSchemaForUninstall(); err != nil {
			return fmt.Errorf("failed to revert schema file: %w", err)
		}
		ui.Warnf("重要提示: 您必须立即 '重新部署' Rime才能使更改生效。")

		ui.Section("卸载完成！")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
