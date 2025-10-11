package commands

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"local/costscope/internal/core/logging"
	"local/costscope/internal/core/security"
)

// SecurityCommands groups security-related CLI commands
type SecurityCommands struct {
	logger  *logging.Logger
	rbac    *security.RBACService
	scanner *security.Scanner
	auditor *security.Auditor
}

// NewSecurityCommands constructs SecurityCommands
func NewSecurityCommands(logger *logging.Logger) *SecurityCommands {
	store := security.NewFileRBACStore("")
	_ = store.Load()
	return &SecurityCommands{
		logger:  logger,
		rbac:    security.NewRBACService(store, logger),
		scanner: security.NewScanner(logger),
		auditor: security.NewAuditor(""),
	}
}

// BuildSecurityCommand builds the top-level `security` command
func (s *SecurityCommands) BuildSecurityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Security management (RBAC, scanning, audit)",
	}

	cmd.AddCommand(s.buildRBACCommand())
	cmd.AddCommand(s.buildScanCommand())

	return cmd
}

func (s *SecurityCommands) buildRBACCommand() *cobra.Command {
	rbacCmd := &cobra.Command{
		Use:   "rbac",
		Short: "RBAC management",
	}

	// costscope security rbac create-role
	createRoleCmd := &cobra.Command{
		Use:   "create-role",
		Short: "Create a new role",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("description")
			perms, _ := cmd.Flags().GetStringArray("permission")
			var permissions []security.Permission
			for _, p := range perms {
				// expected format resource:action
				parts := strings.SplitN(p, ":", 2)
				if len(parts) == 2 {
					permissions = append(permissions, security.Permission{Resource: parts[0], Action: parts[1]})
				}
			}
			role, err := s.rbac.CreateRole(name, desc, permissions)
			if err != nil {
				return err
			}
			_ = s.auditor.Write(security.AuditEvent{Actor: "cli", Action: "create_role", Resource: name, Result: "success", Timestamp: role.CreatedAt})
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(role)
		},
	}
	createRoleCmd.Flags().String("name", "", "Role name")
	_ = createRoleCmd.MarkFlagRequired("name")
	createRoleCmd.Flags().String("description", "", "Role description")
	createRoleCmd.Flags().StringArray("permission", []string{}, "Permission in resource:action format (repeatable)")

	rbacCmd.AddCommand(createRoleCmd)
	return rbacCmd
}

func (s *SecurityCommands) buildScanCommand() *cobra.Command {
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Security scanning",
	}

	vulnsCmd := &cobra.Command{
		Use:   "vulnerabilities",
		Short: "Scan for vulnerabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")
			res, err := s.scanner.ScanVulnerabilities(target)
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(res)
		},
	}
	vulnsCmd.Flags().String("target", ".", "Target path or component to scan")

	scanCmd.AddCommand(vulnsCmd)
	return scanCmd
}
