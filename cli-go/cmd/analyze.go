package cmd

import (
	"fmt"

	"rime-wanxiang-logger-go/internal/analyzer"
	"rime-wanxiang-logger-go/internal/manager"

	"github.com/spf13/cobra"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze the collected log data.",
	Long: `This command reads the JSONL log file, calculates various metrics such as
prediction accuracy (first choice hit rate, top-3 hit rate), and displays the results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("--- 输入习惯分析 ---")

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
			fmt.Printf("❌ 未找到日志文件: %s\n", logFilePath)
			return nil
		}

		fmt.Printf("📊 正在分析日志文件: %s\n", logFilePath)

		// Read and parse the log file
		events, err := analyzer.ReadLogFile(logFilePath)
		if err != nil {
			return fmt.Errorf("❌ 分析过程中发生错误: %w", err)
		}

		if len(events) == 0 {
			fmt.Println("日志文件中未找到 'text_committed' 事件。")
			return nil
		}

		// Perform comprehensive analysis matching Python version
		results := analyzer.PerformAnalysis(events)

		// Display prediction accuracy metrics
		fmt.Println("\n## 预测准确度指标")

		if !results.HasValidSelections {
			fmt.Println("未找到可供分析的有效候选词选择。")
		} else {
			fmt.Printf("  - 总候选词选择数: %d\n", results.TotalSelections)
			fmt.Printf("  - 首选命中率:      %.2f%%\n", results.FirstChoiceHitRate)
			fmt.Printf("  - 前三候选命中率:   %.2f%%\n", results.Top3HitRate)
			fmt.Printf("  - 平均选择排名:     %.2f\n", results.AverageRank)
			fmt.Printf("  - 综合预测得分:   %.3f / 1.000\n", results.OverallAccuracyScore)
		}

		// Display general statistics
		fmt.Println("\n## 常规统计")
		fmt.Printf("  - 总上屏次数 (包括直接上屏): %d\n", results.TotalCommits)
		if results.TotalCommits > 0 {
			fmt.Printf("  - 直接上屏率 (非候选词): %.2f%%\n", results.DirectInputRate)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
