package cmd

import (
	"fmt"
	"os"

	"rime-wanxiang-logger-go/internal/analyzer"
	"rime-wanxiang-logger-go/internal/manager"
	"rime-wanxiang-logger-go/internal/ui"

	"github.com/spf13/cobra"
)

var exportMissesCmd = &cobra.Command{
	Use:   "export-misses",
	Short: "Export misprediction report to a CSV file.",
	Long: `This command processes the log data and filters for entries where the
selected candidate was not the first one predicted by Rime. It then generates
a CSV report of these mispredictions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Section("导出预测错误报告")

		// Initialize RimeManager to get proper log file path
		rimeManager, err := manager.NewRimeManager()
		if err != nil {
			return fmt.Errorf("could not initialize Rime manager: %w", err)
		}

		// Get the actual log file path (parses Lua config)
		logFilePath, err := rimeManager.GetLogFilePath()
		if err != nil {
			return fmt.Errorf("failed to determine log file path: %w", err)
		}

		// Check if log file exists
		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			return fmt.Errorf("❌ 未找到日志文件: %s\n请确保记录器已安装并生成了数据", logFilePath)
		}

		// Get output path from flags or use default
		outFilePath, _ := cmd.Flags().GetString("output")

		ui.Infof("正在读取日志文件: %s", logFilePath)

		// Read and parse the log file
		events, err := analyzer.ReadLogFile(logFilePath)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}

		if len(events) == 0 {
			ui.Warnf("日志文件中未找到 'text_committed' 事件。")
			return nil
		}

		// Count mispredictions first
		var missCount int
		for _, event := range events {
			if event.SelectedCandidateRank != nil && *event.SelectedCandidateRank > 0 {
				missCount++
			}
		}

		if missCount == 0 {
			ui.Successf("太好了！未发现预测错误。您的输入法表现完美！")
			return nil
		}

		ui.Infof("在 %d 次总提交中发现 %d 次预测错误", len(events), missCount)
		ui.Infof("正在导出到: %s", outFilePath)

		// Export mispredictions to CSV
		if err := analyzer.ExportMisses(events, outFilePath); err != nil {
			return fmt.Errorf("failed to export mispredictions: %w", err)
		}

		ui.Successf("成功导出 %d 条预测错误记录到 '%s'", missCount, outFilePath)
		ui.Infof("您可以打开此 CSV 文件来查看具体的预测失误案例。")
		ui.Infof("最常见的错误会显示在文件顶部。")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportMissesCmd)

	// Add flags for customizing output path
	exportMissesCmd.Flags().StringP("output", "o", "mispredictions.csv", "输出 CSV 文件的路径")
}
