package discover

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/config"
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize vulgar configuration",
		Long: `Initialize vulgar configuration by creating a config file at ~/.config/vulgar/config.toml.

This command will:
  - Create the config directory if it doesn't exist
  - Create a template config file with all available options
  - Show helpful tips for setting up credentials`,
		RunE: runInit,
	}

	cmd.Flags().Bool("force", false, "Overwrite existing config file")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	configDir := config.ConfigDir()
	configPath := config.ConfigPath()

	if config.Exists() && !force {
		cli.PrintWarning("Config file already exists: %s", configPath)
		fmt.Println(cli.Muted("Use --force to overwrite"))
		return nil
	}

	cli.PrintLoading("Creating config directory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		cli.PrintFailed()
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	cli.PrintDone()

	cli.PrintLoading("Creating config file")
	if err := os.WriteFile(configPath, []byte(config.DefaultConfigTemplate()), 0600); err != nil {
		cli.PrintFailed()
		return fmt.Errorf("failed to write config: %w", err)
	}
	cli.PrintDone()

	fmt.Println()
	cli.PrintSuccess("Configuration initialized!")
	fmt.Println()

	cli.PrintHeader("Config Location")
	fmt.Println("  " + cli.Code(configPath))

	cli.PrintHeader("Next Steps")
	fmt.Println("  Edit your config file to add credentials:")
	fmt.Println("     " + cli.Code("nano "+configPath))
	fmt.Println()
	fmt.Println("  You can reference environment variables in the config:")
	fmt.Println("     " + cli.Muted("api_key = \"${MY_SECRET_KEY}\""))

	cli.PrintHeader("Try Discovery Commands")
	fmt.Println("  " + cli.Code("vulgar gdrive list"))
	fmt.Println("  " + cli.Code("vulgar gdrive find \"report.pdf\""))
	fmt.Println("  " + cli.Code("vulgar gcalendar today"))

	return nil
}
