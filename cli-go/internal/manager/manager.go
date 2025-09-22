// Package manager provides core logic for interacting with the Rime input method directory.
package manager

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Constants for file names and configurations
const (
	LoggerLuaFile       = "input_habit_logger.lua"
	ConfigLuaFile       = "input_habit_logger_config.lua"
	SchemaYamlFile      = "wanxiang.schema.yaml"
	DefaultLogJsonlFile = "input_habit_log_structured.jsonl"
	SchemaLineToAdd     = "      - lua_processor@*input_habit_logger #输入习惯记录器 - 记录用户输入习惯"
)

// RimeManager handles all interactions with the Rime user directory.
type RimeManager struct {
	UserDirectory string
	LuaDirectory  string
	AssetsPath    string // Path to the internal assets (for reference)
}

// NewRimeManager creates a new RimeManager and automatically detects the user directory.
func NewRimeManager() (*RimeManager, error) {
	userDir, err := GetRimeUserDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to detect Rime user directory: %w", err)
	}

	info, err := os.Stat(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Rime user directory does not exist at '%s'", userDir)
		}
		return nil, fmt.Errorf("could not access Rime user directory at '%s': %w", userDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path '%s' is not a directory", userDir)
	}

	return &RimeManager{
		UserDirectory: userDir,
		LuaDirectory:  filepath.Join(userDir, "lua"),
	}, nil
}

// GetRimeUserDirectory detects the default Rime user directory based on the OS.
// See: https://github.com/rime/home/blob/master/README.md#user-data-directory
func GetRimeUserDirectory() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\Rime
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New("%APPDATA% environment variable not set")
		}
		return filepath.Join(appData, "Rime"), nil
	case "darwin":
		// macOS: ~/Library/Rime
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}
		return filepath.Join(homeDir, "Library", "Rime"), nil
	case "linux":
		// Linux: check multiple possible locations
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get user home directory: %w", err)
		}

		// Check common locations in order of preference
		possiblePaths := []string{
			filepath.Join(homeDir, ".config", "rime"),
			filepath.Join(homeDir, ".config", "fcitx", "rime"),
			filepath.Join(homeDir, ".config", "fcitx5", "rime"),
			filepath.Join(homeDir, ".config", "ibus", "rime"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		// If none exist, default to the most common location
		return possiblePaths[0], nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// GetLuaDirectory returns the path to the 'lua' subdirectory within the Rime user directory.
func (m *RimeManager) GetLuaDirectory() string {
	return m.LuaDirectory
}

// GetLogFilePath gets the log file path by parsing the active preset in the Lua config file.
// This mimics the Python version's _get_log_file_path method.
func (m *RimeManager) GetLogFilePath() (string, error) {
	defaultPath := filepath.Join(m.UserDirectory, DefaultLogJsonlFile)
	configPath := filepath.Join(m.LuaDirectory, ConfigLuaFile)

	// If config file doesn't exist, return default path
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultPath, nil
	}

	// Read and parse the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not read config file, using default log path. Error: %v\n", err)
		return defaultPath, nil
	}

	contentStr := string(content)

	// 1. Find the active preset choice
	presetChoiceRegex := regexp.MustCompile(`local\s+preset_choice\s*=\s*"([^"]+)"`)
	presetMatch := presetChoiceRegex.FindStringSubmatch(contentStr)
	if len(presetMatch) < 2 {
		return defaultPath, nil
	}

	activePreset := presetMatch[1]

	// 2. Find the configuration block for the active preset
	// This regex looks for `preset_name = {` and captures everything until the matching `}`
	presetBlockRegex := regexp.MustCompile(fmt.Sprintf(`^\s*%s\s*=\s*\{([\s\S]*?)\n\s*\}`, regexp.QuoteMeta(activePreset)))
	presetBlockMatch := presetBlockRegex.FindStringSubmatch(contentStr)
	if len(presetBlockMatch) < 2 {
		return defaultPath, nil
	}

	presetContent := presetBlockMatch[1]

	// 3. Search for a non-commented log_file_path within that block
	logPathRegex := regexp.MustCompile(`^\s*log_file_path\s*=\s*"([^"]+)"`)
	logPathMatch := logPathRegex.FindStringSubmatch(presetContent)

	if len(logPathMatch) >= 2 {
		// Path must be un-escaped (e.g., "C:\\Users" -> "C:\Users")
		customPathStr := strings.ReplaceAll(logPathMatch[1], `\\`, `\`)
		if customPathStr != "" {
			fmt.Printf("Found active custom log path: %s\n", customPathStr)
			return customPathStr, nil
		}
	}

	return defaultPath, nil
}

// CheckScriptsInstalled checks if the Lua scripts are installed in the lua directory.
func (m *RimeManager) CheckScriptsInstalled() (bool, bool) {
	loggerPath := filepath.Join(m.LuaDirectory, LoggerLuaFile)
	configPath := filepath.Join(m.LuaDirectory, ConfigLuaFile)

	_, loggerErr := os.Stat(loggerPath)
	_, configErr := os.Stat(configPath)

	return loggerErr == nil, configErr == nil
}

// CheckSchemaConfigured checks if the schema file is configured with the logger.
func (m *RimeManager) CheckSchemaConfigured() (bool, error) {
	schemaPath := filepath.Join(m.UserDirectory, SchemaYamlFile)

	content, err := os.ReadFile(schemaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("schema file not found: %s", schemaPath)
		}
		return false, fmt.Errorf("could not read schema file: %w", err)
	}

	return strings.Contains(string(content), strings.TrimSpace(SchemaLineToAdd)), nil
}

// GetSchemaPath returns the path to the wanxiang.schema.yaml file.
func (m *RimeManager) GetSchemaPath() string {
	return filepath.Join(m.UserDirectory, SchemaYamlFile)
}

// BackupSchemaFile creates a backup of the schema file before modification.
func (m *RimeManager) BackupSchemaFile() error {
	schemaPath := m.GetSchemaPath()
	backupPath := schemaPath + ".bak"

	input, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	err = os.WriteFile(backupPath, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}

	fmt.Printf("Schema file backed up to: %s\n", backupPath)
	return nil
}

// ModifySchemaForInstall adds the lua_processor line to the schema file.
func (m *RimeManager) ModifySchemaForInstall() error {
	schemaPath := m.GetSchemaPath()

	// Check if file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return fmt.Errorf("schema file not found: %s. Please ensure the 'wanxiang' input method is installed and deployed at least once", schemaPath)
	}

	// Read the file
	file, err := os.Open(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to open schema file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Check if already configured
	lineToAdd := strings.TrimSpace(SchemaLineToAdd)
	for _, line := range lines {
		if strings.Contains(line, lineToAdd) {
			fmt.Println("Schema is already configured, no changes needed.")
			return nil
		}
	}

	// Find the punctuator line to insert after it
	punctuatorIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "punctuator") && strings.TrimSpace(line)[0] == '-' {
			punctuatorIndex = i
			break
		}
	}

	if punctuatorIndex == -1 {
		return errors.New("could not find 'punctuator' entry in schema file")
	}

	// Get the indentation from the punctuator line
	punctuatorLine := lines[punctuatorIndex]
	indentation := ""
	for _, char := range punctuatorLine {
		if char == ' ' || char == '\t' {
			indentation += string(char)
		} else {
			break
		}
	}

	// Create the new line with proper indentation
	lineWithIndent := indentation + lineToAdd

	// Insert the line after punctuator
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:punctuatorIndex+1]...)
	newLines = append(newLines, lineWithIndent)
	newLines = append(newLines, lines[punctuatorIndex+1:]...)

	// Write back to file
	content := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(schemaPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	fmt.Printf("Successfully configured '%s'.\n", SchemaYamlFile)
	return nil
}

// RevertSchemaForUninstall removes the lua_processor line from the schema file.
func (m *RimeManager) RevertSchemaForUninstall() error {
	schemaPath := m.GetSchemaPath()

	// Check if file exists
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		fmt.Println("Schema file not found, skipping.")
		return nil
	}

	// Read the file
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	lineToRemove := strings.TrimSpace(SchemaLineToAdd)

	// Filter out lines containing the logger configuration
	var filteredLines []string
	removed := false
	for _, line := range lines {
		if !strings.Contains(line, lineToRemove) {
			filteredLines = append(filteredLines, line)
		} else {
			removed = true
		}
	}

	if removed {
		// Write back to file
		newContent := strings.Join(filteredLines, "\n")
		if err := os.WriteFile(schemaPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write schema file: %w", err)
		}
		fmt.Printf("Successfully removed logger configuration from '%s'.\n", SchemaYamlFile)
	} else {
		fmt.Println("Logger configuration not found in schema file, no changes needed.")
	}

	return nil
}

// LogFileExists checks if the log file exists at the determined path.
func (m *RimeManager) LogFileExists() bool {
	logPath, err := m.GetLogFilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(logPath)
	return err == nil
}
