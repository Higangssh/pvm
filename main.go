package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{
		Use:     "pvm",
		Short:   "Python venv manager",
		Version: version,
	}
	root.AddCommand(
		listCmd(),
		scanCmd(),
		addCmd(),
		removeCmd(),
		aliasCmd(),
		runCmd(),
		execCmd(),
		shellCmd(),
		saveCmd(),
		doCmd(),
		uiCmd(),
	)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func mustConfig() *Config {
	c, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}
	return c
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered venvs",
		Run: func(cmd *cobra.Command, args []string) {
			c := mustConfig()
			if len(c.Venvs) == 0 {
				color.Yellow("No venvs registered.")
				fmt.Println("  Use `pvm scan <path>` or `pvm add <path>` to get started.")
				return
			}
			cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()
			gray := color.New(color.FgHiBlack).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			title := color.New(color.FgMagenta, color.Bold).SprintFunc()

			fmt.Println()
			fmt.Println(title("  🐍 Python Virtual Environments"))
			fmt.Println(gray("  ────────────────────────────────"))
			tw := tablewriter.NewWriter(os.Stdout)
			tw.Header("ALIAS", "PYTHON", "PATH", "COMMANDS")
			for _, v := range c.Venvs {
				cmds := make([]string, 0, len(v.Commands))
				for k := range v.Commands {
					cmds = append(cmds, k)
				}
				cmdStr := gray("-")
				if len(cmds) > 0 {
					cmdStr = green(strings.Join(cmds, ", "))
				}
				_ = tw.Append(cyan(v.Alias), yellow(pythonVersion(v.Path)), gray(v.Path), cmdStr)
			}
			_ = tw.Render()
			fmt.Printf("  %s %d venv(s)\n\n", gray("Total:"), len(c.Venvs))
		},
	}
}

func scanCmd() *cobra.Command {
	var depth int
	cmd := &cobra.Command{
		Use:   "scan <path>",
		Short: "Scan a directory for venvs and register them",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustConfig()
			found := scanDir(args[0], depth)
			added := 0
			for _, p := range found {
				if c.FindByPath(p) != nil {
					continue
				}
				alias := defaultAlias(p)
				orig := alias
				i := 2
				for existing, _ := c.Find(alias); existing != nil; existing, _ = c.Find(alias) {
					alias = fmt.Sprintf("%s%d", orig, i)
					i++
				}
				c.Venvs = append(c.Venvs, Venv{Alias: alias, Path: p})
				fmt.Printf("+ %s  %s\n", alias, p)
				added++
			}
			if added > 0 {
				if err := c.Save(); err != nil {
					fmt.Fprintln(os.Stderr, "save error:", err)
					os.Exit(1)
				}
			}
			fmt.Printf("Added %d venv(s).\n", added)
		},
	}
	cmd.Flags().IntVarP(&depth, "depth", "d", 4, "max search depth")
	return cmd
}

func addCmd() *cobra.Command {
	var alias string
	var yes bool
	cmd := &cobra.Command{
		Use:   "add <path>",
		Short: "Register a venv manually",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := filepath.Abs(args[0])
			if !isVenv(path) {
				fmt.Fprintf(os.Stderr, "not a venv: %s\n", path)
				os.Exit(1)
			}
			aliasName := alias
			if aliasName == "" {
				aliasName = defaultAlias(path)
			}
			if !yes {
				fmt.Printf("Add %s (%s)? [y/N]: ", aliasName, path)
				var ans string
				_, _ = fmt.Scanln(&ans)
				ans = strings.ToLower(strings.TrimSpace(ans))
				if ans != "y" && ans != "yes" {
					fmt.Println("Cancelled.")
					return
				}
			}
			c := mustConfig()
			if err := c.Add(Venv{Alias: aliasName, Path: path}); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err := c.Save(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("Added %s -> %s\n", aliasName, path)
		},
	}
	cmd.Flags().StringVarP(&alias, "alias", "a", "", "alias name")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "add without confirmation")
	return cmd
}

func removeCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "remove <alias>",
		Aliases: []string{"rm"},
		Short:   "Remove a venv from registry",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustConfig()
			v, i := c.Find(args[0])
			if i < 0 {
				fmt.Fprintf(os.Stderr, "alias not found: %s\n", args[0])
				os.Exit(1)
			}
			if !yes {
				fmt.Printf("Remove %s (%s)? [y/N]: ", v.Alias, v.Path)
				var ans string
				_, _ = fmt.Scanln(&ans)
				ans = strings.ToLower(strings.TrimSpace(ans))
				if ans != "y" && ans != "yes" {
					fmt.Println("Cancelled.")
					return
				}
			}
			c.Venvs = append(c.Venvs[:i], c.Venvs[i+1:]...)
			if err := c.Save(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("Removed %s\n", args[0])
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "remove without confirmation")
	return cmd
}

func aliasCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "alias <old> <new>",
		Short: "Rename a venv alias",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := mustConfig()
			v, _ := c.Find(args[0])
			if v == nil {
				fmt.Fprintf(os.Stderr, "alias not found: %s\n", args[0])
				os.Exit(1)
			}
			if existing, _ := c.Find(args[1]); existing != nil {
				fmt.Fprintf(os.Stderr, "alias already exists: %s\n", args[1])
				os.Exit(1)
			}
			v.Alias = args[1]
			if err := c.Save(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("%s -> %s\n", args[0], args[1])
		},
	}
}

func resolveVenv(alias string) *Venv {
	c := mustConfig()
	v, _ := c.Find(alias)
	if v == nil {
		fmt.Fprintf(os.Stderr, "alias not found: %s\n", alias)
		os.Exit(1)
	}
	return v
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "run <alias> [python-args...]",
		Short:              "Run the venv's python with args",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			v := resolveVenv(args[0])
			pythonArgs := args[1:]
			if len(pythonArgs) > 0 && pythonArgs[0] == "--" {
				pythonArgs = pythonArgs[1:]
			}
			c := exec.Command(pythonExe(v.Path), pythonArgs...)
			c.Env = activatedEnv(v.Path)
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					os.Exit(ee.ExitCode())
				}
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
}

func execCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "exec <alias> -- <cmd...>",
		Short:              "Run any command inside the venv (PATH injected)",
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			v := resolveVenv(args[0])
			rest := args[1:]
			if len(rest) > 0 && rest[0] == "--" {
				rest = rest[1:]
			}
			if len(rest) == 0 {
				fmt.Fprintln(os.Stderr, "no command given")
				os.Exit(1)
			}
			c, err := commandFromArgs(rest, v.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					os.Exit(ee.ExitCode())
				}
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
}

func shellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell <alias>",
		Short: "Open an interactive shell with the venv activated",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			v := resolveVenv(args[0])
			c, err := shellCommand(v.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
}

func saveCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "save <alias> <name> <cmd...>",
		Short:              "Save a custom command for a venv",
		Args:               cobra.MinimumNArgs(3),
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			c := mustConfig()
			v, _ := c.Find(args[0])
			if v == nil {
				fmt.Fprintf(os.Stderr, "alias not found: %s\n", args[0])
				os.Exit(1)
			}
			if v.Commands == nil {
				v.Commands = map[string]string{}
			}
			if v.CommandArgs == nil {
				v.CommandArgs = map[string][]string{}
			}
			joined := strings.Join(args[2:], " ")
			v.Commands[args[1]] = joined
			v.CommandArgs[args[1]] = append([]string(nil), args[2:]...)
			if err := c.Save(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("Saved %s.%s = %s\n", args[0], args[1], joined)
		},
	}
}

func doCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "do <alias> <name>",
		Short: "Run a saved custom command",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			v := resolveVenv(args[0])
			if len(v.CommandArgs[args[1]]) > 0 {
				parts := v.CommandArgs[args[1]]
				c, err := commandFromArgs(parts, v.Path)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
				if err := c.Run(); err != nil {
					if ee, ok := err.(*exec.ExitError); ok {
						os.Exit(ee.ExitCode())
					}
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return
			}
			cmdStr, ok := v.Commands[args[1]]
			if !ok {
				fmt.Fprintf(os.Stderr, "command not found: %s.%s\n", args[0], args[1])
				os.Exit(1)
			}
			c, err := commandFromString(cmdStr, v.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if c == nil {
				return
			}
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					os.Exit(ee.ExitCode())
				}
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
}
