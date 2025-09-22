package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"rime-wanxiang-logger-go/internal/manager"

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
		fmt.Println("Starting Rime logger uninstallation...")

		rimeManager, err := manager.NewRimeManager()
		if err != nil {
			return fmt.Errorf("could not initialize Rime manager: %w", err)
		}
		fmt.Printf("✅ Detected Rime user directory at: %s\n", rimeManager.UserDirectory)

		loggerScriptPath := filepath.Join(rimeManager.GetLuaDirectory(), "input_habit_logger.lua")
		configScriptPath := filepath.Join(rimeManager.GetLuaDirectory(), "input_habit_logger_config.lua")

		// Check if files exist before trying to delete
		if _, err := os.Stat(loggerScriptPath); os.IsNotExist(err) {
			fmt.Println("Logger script not found. Nothing to uninstall.")
			return nil
		}

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("This will delete the logger scripts from %s. Are you sure", rimeManager.GetLuaDirectory()),
			IsConfirm: true,
		}

		if _, err := prompt.Run(); err != nil {
			fmt.Println("Uninstallation cancelled.")
			return nil
		}

		// Delete the scripts
		if err := os.Remove(loggerScriptPath); err != nil {
			return fmt.Errorf("failed to delete 'input_habit_logger.lua': %w", err)
		}
		fmt.Println("✅ Deleted 'input_habit_logger.lua'")

		if err := os.Remove(configScriptPath); err != nil {
			return fmt.Errorf("failed to delete 'input_habit_logger_config.lua': %w", err)
		}
		fmt.Println("✅ Deleted 'input_habit_logger_config.lua'")

		fmt.Println("\n📝 Please remember to manually remove the following line from your .schema.yaml file:")
		fmt.Println("  - lua_processor@input_habit_logger")
		fmt.Println("\nThen, redeploy Rime to finalize the uninstallation.")

		fmt.Println("\n🎉 Uninstallation complete!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
