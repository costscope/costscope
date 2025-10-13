package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/costscope/costscope/internal/core/compliance"
)

// ComplianceCommands groups compliance CLI commands
type ComplianceCommands struct{}

func NewComplianceCommands() *ComplianceCommands { return &ComplianceCommands{} }

// BuildComplianceCommand creates the `compliance` root command
func (c *ComplianceCommands) BuildComplianceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compliance",
		Short: "Compliance checks and reporting",
	}
	cmd.AddCommand(c.buildCheckCommand())
	return cmd
}

func (c *ComplianceCommands) buildCheckCommand() *cobra.Command {
	chk := &cobra.Command{
		Use:   "check",
		Short: "Run compliance checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			stdStr, _ := cmd.Flags().GetString("standard")
			var std compliance.Standard
			switch stdStr {
			case string(compliance.SOC2):
				std = compliance.SOC2
			case string(compliance.GDPR):
				std = compliance.GDPR
			case string(compliance.HIPAA):
				std = compliance.HIPAA
			default:
				std = compliance.SOC2
			}
			res, err := compliance.NewChecker().Check(std)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(res)
		},
	}
	chk.Flags().String("standard", string(compliance.SOC2), "Compliance standard (SOC2, GDPR, HIPAA)")
	return chk
}
