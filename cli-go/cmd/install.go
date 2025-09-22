package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"rime-wanxiang-logger-go/internal/assets"
	"rime-wanxiang-logger-go/internal/manager"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the logger scripts into the Rime user directory.",
	Long: `This command detects the Rime user directory, copies the necessary Lua scripts
(input_habit_logger.lua and input_habit_logger_config.lua) into it,
and modifies the Rime schema to enable the logger with interactive preset selection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("--- 开始安装日志记录器 ---")

		// 1. Initialize RimeManager
		rimeManager, err := manager.NewRimeManager()
		if err != nil {
			fmt.Printf("❌ 错误: 未找到 Rime 用户目录，无法安装。\n")
			return fmt.Errorf("failed to initialize Rime manager: %w", err)
		}

		fmt.Printf("找到 Rime 目录: %s\n", rimeManager.UserDirectory)

		// 2. Interactive preset selection (matching Python version)
		presetMap := map[string]string{
			"✅ 普通模式 (Normal) - 推荐，用于计算输入法预测准确率":      "normal",
			"👩‍💻 词库贡献者模式 (Developer) - 用于调试，关注非首选上屏": "developer",
			"🔬 高级模式 (Advanced) - 记录几乎所有信息，用于深度分析":    "advanced",
			"⚙️ 自定义 (Custom) - (需要手动修改配置文件)":         "custom",
		}

		var presetChoices []string
		for key := range presetMap {
			presetChoices = append(presetChoices, key)
		}

		prompt := promptui.Select{
			Label: "请选择一个日志记录预设模式",
			Items: presetChoices,
		}

		_, selectedChoice, err := prompt.Run()
		if err != nil {
			fmt.Println("Installation cancelled.")
			return nil
		}

		selectedPreset := presetMap[selectedChoice]

		// Handle custom preset selection
		if selectedPreset == "custom" {
			fmt.Printf("您选择了 '自定义' 模式。\n")
			fmt.Println("请先使用其他模式安装，然后手动编辑以下文件以符合您的需求:")
			fmt.Printf("%s\n", filepath.Join(rimeManager.GetLuaDirectory(), "input_habit_logger_config.lua"))
			return nil
		}

		fmt.Printf("已选择预设: %s\n", selectedPreset)

		// 3. 步骤 1: 复制 Lua 脚本
		fmt.Println("\n--> 步骤 1: 复制 Lua 脚本...")

		luaDir := rimeManager.GetLuaDirectory()
		if err := os.MkdirAll(luaDir, 0755); err != nil {
			return fmt.Errorf("failed to create lua directory at %s: %w", luaDir, err)
		}

		// Copy logger script
		loggerScriptPath := filepath.Join(luaDir, "input_habit_logger.lua")
		if err := os.WriteFile(loggerScriptPath, assets.LoggerScript, 0644); err != nil {
			return fmt.Errorf("failed to write logger script: %w", err)
		}
		fmt.Printf("    [+] 已安装: %s\n", loggerScriptPath)

		// Copy and modify config script with selected preset
		configScriptPath := filepath.Join(luaDir, "input_habit_logger_config.lua")

		// Read the original config content
		configContent := string(assets.ConfigScript)

		// Replace the preset choice (matching Python logic)
		presetRegex := regexp.MustCompile(`local\s+preset_choice\s*=\s*".*"`)
		newConfigContent := presetRegex.ReplaceAllString(configContent, fmt.Sprintf(`local preset_choice = "%s"`, selectedPreset))

		if err := os.WriteFile(configScriptPath, []byte(newConfigContent), 0644); err != nil {
			return fmt.Errorf("failed to write config script: %w", err)
		}
		fmt.Printf("    [+] 已安装配置文件，预设为 '%s': %s\n", selectedPreset, configScriptPath)

		// 4. 步骤 2: 修改输入方案文件
		fmt.Println("\n--> 步骤 2: 修改输入方案文件...")
		if err := modifySchemaForInstall(rimeManager); err != nil {
			return fmt.Errorf("failed to modify schema file: %w", err)
		}

		// 5. Installation complete
		fmt.Println("\n--- ✅ 安装成功！ ---")
		fmt.Println("\n重要提示: 您必须立即 '重新部署' Rime才能使更改生效。")

		return nil
	},
}

// modifySchemaForInstall handles the schema file modification (matching Python _modify_schema_for_install)
func modifySchemaForInstall(rimeManager *manager.RimeManager) error {
	schemaPath := rimeManager.GetSchemaPath()

	// Check if schema file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		fmt.Printf("    ❌ 错误: 未找到 '%s'。\n", schemaPath)
		fmt.Println("    请确保已安装 '万象' 输入方案并已至少部署过一次。")
		return fmt.Errorf("schema file not found: %s", schemaPath)
	}

	// Read the schema file
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Printf("    ❌ 发生意外错误: %v\n", err)
		return err
	}

	lines := strings.Split(string(content), "\n")
	lineToAdd := strings.TrimSpace(manager.SchemaLineToAdd)

	// Check if already configured
	for _, line := range lines {
		if strings.Contains(line, lineToAdd) {
			fmt.Println("    [*] 输入方案已配置，无需更改。")
			return nil
		}
	}

	// Find punctuator line (matching Python logic)
	punctuatorIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "punctuator") && strings.TrimSpace(line)[0] == '-' {
			punctuatorIndex = i
			break
		}
	}

	if punctuatorIndex == -1 {
		fmt.Println("    ❌ 错误: 在输入方案中未找到 'punctuator' 入口。")
		return fmt.Errorf("could not find 'punctuator' entry in schema")
	}

	// Get indentation from punctuator line
	punctuatorLine := lines[punctuatorIndex]
	indentation := ""
	for _, char := range punctuatorLine {
		if char == ' ' || char == '\t' {
			indentation += string(char)
		} else {
			break
		}
	}

	// Create backup (matching Python logic)
	backupPath := schemaPath + ".bak"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		fmt.Printf("    ❌ 发生意外错误: %v\n", err)
		return err
	}
	fmt.Printf("    [*] 已将原始输入方案备份至: %s\n", backupPath)

	// Insert the line with proper indentation
	lineWithIndent := indentation + lineToAdd

	// Create new lines slice with the inserted line
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:punctuatorIndex+1]...)
	newLines = append(newLines, lineWithIndent)
	newLines = append(newLines, lines[punctuatorIndex+1:]...)

	// Write the modified content back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(schemaPath, []byte(newContent), 0644); err != nil {
		fmt.Printf("    ❌ 发生意外错误: %v\n", err)
		return err
	}

	fmt.Printf("    [+] 已成功配置 '%s'。\n", manager.SchemaYamlFile)
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
