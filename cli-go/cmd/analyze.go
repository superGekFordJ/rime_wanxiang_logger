package cmd

import (
	"fmt"

	"rime-wanxiang-logger-go/internal/analyzer"
	"rime-wanxiang-logger-go/internal/manager"
	"rime-wanxiang-logger-go/internal/ui"

	"github.com/spf13/cobra"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze the collected log data.",
	Long: `This command reads the JSONL log file, calculates various metrics such as
prediction accuracy (first choice hit rate, top-3 hit rate), and displays the results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Section("输入习惯分析")

		// Initialize RimeManager to get the correct log file path
		rimeManager, err := manager.NewRimeManager()
		if err != nil {
			return fmt.Errorf("failed to initialize Rime manager: %w", err)
		}

		// Get the actual log file path by parsing the config
		logFilePath, err := rimeManager.GetLogFilePath()
		if err != nil {
			return fmt.Errorf("failed to determine log file path: %w", err)
		}

		// Check if log file exists
		if !rimeManager.LogFileExists() {
			ui.Errorf("未找到日志文件: %s", logFilePath)
			return nil
		}

		ui.Infof("正在分析日志文件: %s", logFilePath)

		// Read and parse the log file
		events, err := analyzer.ReadLogFile(logFilePath)
		if err != nil {
			return fmt.Errorf("分析过程中发生错误: %w", err)
		}

		if len(events) == 0 {
			ui.Warnf("日志文件中未找到 'text_committed' 事件。")
			return nil
		}

		// Perform comprehensive analysis matching Python version
		results := analyzer.PerformAnalysis(events)

		// Display prediction accuracy metrics
		ui.Subsection("预测准确度指标")

		if !results.HasValidSelections {
			ui.Warnf("未找到可供分析的有效候选词选择。")
		} else {
			ui.PrintKV([][2]string{
				{"总候选词选择数", fmt.Sprintf("%d", results.TotalSelections)},
				{"首选命中率", fmt.Sprintf("%.2f%%", results.FirstChoiceHitRate)},
				{"前三候选命中率", fmt.Sprintf("%.2f%%", results.Top3HitRate)},
				{"平均选择排名", fmt.Sprintf("%.2f", results.AverageRank)},
				{"综合预测得分", fmt.Sprintf("%.3f / 1.000", results.OverallAccuracyScore)},
			})
		}

		// Display general statistics
		ui.Subsection("常规统计")
		ui.PrintKV([][2]string{{"总上屏次数 (包括直接上屏)", fmt.Sprintf("%d", results.TotalCommits)}})
		if results.TotalCommits > 0 {
			ui.PrintKV([][2]string{{"直接上屏率 (非候选词)", fmt.Sprintf("%.2f%%", results.DirectInputRate)}})
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
