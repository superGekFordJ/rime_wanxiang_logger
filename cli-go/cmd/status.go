package cmd

import (
	"os"

	"rime-wanxiang-logger-go/internal/manager"
	"rime-wanxiang-logger-go/internal/ui"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the current installation status.",
	Long: `This command checks if the logger scripts are correctly installed,
if the Rime schema is properly configured, and reports the location of the log file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkStatus()
	},
}

// checkStatus performs comprehensive status checking (matching Python check_status method)
func checkStatus() error {
	ui.Section("Rime 日志记录器状态检查")

	// Initialize RimeManager
	rimeManager, err := manager.NewRimeManager()
	if err != nil {
		ui.Errorf("未找到 Rime 用户目录。请问 Rime 是否已安装？")
		return nil
	}

	ui.Successf("找到 Rime 用户目录: %s", rimeManager.UserDirectory)

	// Check Lua script files (matching Python loop)
	loggerInstalled, configInstalled := rimeManager.CheckScriptsInstalled()

	if loggerInstalled {
		ui.Successf("找到脚本: %s", rimeManager.GetLuaDirectory()+"/"+manager.LoggerLuaFile)
	} else {
		ui.Errorf("未找到脚本: %s", rimeManager.GetLuaDirectory()+"/"+manager.LoggerLuaFile)
	}

	if configInstalled {
		ui.Successf("找到脚本: %s", rimeManager.GetLuaDirectory()+"/"+manager.ConfigLuaFile)
	} else {
		ui.Errorf("未找到脚本: %s", rimeManager.GetLuaDirectory()+"/"+manager.ConfigLuaFile)
	}

	// Check schema configuration (matching Python logic)
	schemaPath := rimeManager.GetSchemaPath()
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		ui.Errorf("未找到输入方案文件: %s", schemaPath)
	} else {
		configured, err := rimeManager.CheckSchemaConfigured()
		if err != nil {
			ui.Warnf("无法读取输入方案文件。错误: %v", err)
		} else if configured {
			ui.Successf("输入方案 '%s' 已为日志记录器正确配置。", manager.SchemaYamlFile)
		} else {
			ui.Errorf("输入方案 '%s' 尚未配置。", manager.SchemaYamlFile)
		}
	}

	// Check log file existence (matching Python logic)
	logFilePath, err := rimeManager.GetLogFilePath()
	if err != nil {
		ui.Warnf("无法确定日志文件路径。错误: %v", err)
	} else if rimeManager.LogFileExists() {
		ui.Successf("找到日志文件: %s", logFilePath)
	} else {
		ui.Errorf("在 '%s' 未找到日志文件。请打字以生成日志。", logFilePath)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
