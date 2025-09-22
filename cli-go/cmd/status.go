package cmd

import (
	"fmt"
	"os"

	"rime-wanxiang-logger-go/internal/manager"

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
	fmt.Println("--- Rime 日志记录器状态检查 ---")

	// Initialize RimeManager
	rimeManager, err := manager.NewRimeManager()
	if err != nil {
		fmt.Println("❌ 未找到 Rime 用户目录。请问 Rime 是否已安装？")
		return nil
	}

	fmt.Printf("✅ 找到 Rime 用户目录: %s\n", rimeManager.UserDirectory)

	// Check Lua script files (matching Python loop)
	loggerInstalled, configInstalled := rimeManager.CheckScriptsInstalled()

	if loggerInstalled {
		fmt.Printf("✅ 找到脚本: %s\n", rimeManager.GetLuaDirectory()+"/"+manager.LoggerLuaFile)
	} else {
		fmt.Printf("❌ 未找到脚本: %s\n", rimeManager.GetLuaDirectory()+"/"+manager.LoggerLuaFile)
	}

	if configInstalled {
		fmt.Printf("✅ 找到脚本: %s\n", rimeManager.GetLuaDirectory()+"/"+manager.ConfigLuaFile)
	} else {
		fmt.Printf("❌ 未找到脚本: %s\n", rimeManager.GetLuaDirectory()+"/"+manager.ConfigLuaFile)
	}

	// Check schema configuration (matching Python logic)
	schemaPath := rimeManager.GetSchemaPath()
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		fmt.Printf("❌ 未找到输入方案文件: %s\n", schemaPath)
	} else {
		configured, err := rimeManager.CheckSchemaConfigured()
		if err != nil {
			fmt.Printf("❓ 无法读取输入方案文件。错误: %v\n", err)
		} else if configured {
			fmt.Printf("✅ 输入方案 '%s' 已为日志记录器正确配置。\n", manager.SchemaYamlFile)
		} else {
			fmt.Printf("❌ 输入方案 '%s' 尚未配置。\n", manager.SchemaYamlFile)
		}
	}

	// Check log file existence (matching Python logic)
	logFilePath, err := rimeManager.GetLogFilePath()
	if err != nil {
		fmt.Printf("❓ 无法确定日志文件路径。错误: %v\n", err)
	} else if rimeManager.LogFileExists() {
		fmt.Printf("✅ 找到日志文件: %s\n", logFilePath)
	} else {
		fmt.Printf("❌ 在 '%s' 未找到日志文件。请打字以生成日志。\n", logFilePath)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
