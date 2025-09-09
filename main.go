package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// --- TUI Styles and Model for 'pick' command ---
var (
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}


// --- Global Flags ---
var shouldStash bool

// --- Path and State Helpers ---

func getGitDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		stat, err := os.Stat(gitPath)
		if err == nil && stat.IsDir() {
			return gitPath, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("fatal: not a git repository (or any of the parent directories): .git")
		}
		dir = parentDir
	}
}

func getStackFilePath() (string, error) {
	gitDir, err := getGitDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gitDir, "git_branch_stack"), nil
}

func getRebaseStateFilePath() (string, error) {
	gitDir, err := getGitDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(gitDir, "git_stack_rebase_state"), nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// --- File I/O Helpers ---

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// --- Git Command & Stash Helpers ---

func runGitCommand(args ...string) error {
	fmt.Printf("+ git %s\n", strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stashPushIfDirty() (bool, error) {
	statusCmd := exec.Command("git", "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	if len(strings.TrimSpace(string(output))) == 0 {
		return false, nil // Working directory is clean
	}

	fmt.Println("--- Stashing local changes ---")
	err = runGitCommand("stash", "push", "-m", "git-stack auto-stash")
	if err != nil {
		return false, fmt.Errorf("failed to stash changes: %w", err)
	}
	return true, nil
}

func stashPop() error {
	fmt.Println("--- Applying stashed changes ---")
	return runGitCommand("stash", "pop")
}

// --- Cobra Command Definitions ---

var rootCmd = &cobra.Command{
	Use:   "git-stack",
	Short: "A utility to manage a stack of checked-out Git branches.",
	Long:  `git-stack provides a set of commands to manage a stack of Git branches, simplifying workflows with dependent branches.`,
	Run: func(cmd *cobra.Command, args []string) {
		listCmd.Run(cmd, args)
	},
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout <branch>",
	Short: "Pushes current branch, checks out <branch>.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchToCheckout := args[0]
		currentBranch, err := getCurrentBranch()
		if err != nil {
			return err
		}
		if currentBranch == branchToCheckout {
			fmt.Printf("Already on '%s'. Nothing to do.\n", currentBranch)
			return nil
		}

		var didStash bool
		if shouldStash {
			didStash, err = stashPushIfDirty()
			if err != nil {
				return err
			}
		}

		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		file, err := os.OpenFile(stackFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open stack file for writing: %w", err)
		}
		defer file.Close()
		fmt.Printf("Pushing '%s' onto the stack...\n", currentBranch)
		if _, err := fmt.Fprintln(file, currentBranch); err != nil {
			return fmt.Errorf("failed to write to stack file: %w", err)
		}

		fmt.Printf("Checking out '%s'...\n", branchToCheckout)
		err = runGitCommand(append([]string{"checkout"}, args...)...)
		if err != nil {
			return err
		}

		if didStash {
			return stashPop()
		}
		return nil
	},
}

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Pops the last branch from the stack and checks it out.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		lines, err := readLines(stackFile)
		if err != nil {
			return fmt.Errorf("could not read stack file: %w", err)
		}
		if len(lines) == 0 {
			fmt.Println("Stack is empty. Nothing to pop.")
			return nil
		}

		var didStash bool
		if shouldStash {
			didStash, err = stashPushIfDirty()
			if err != nil {
				return err
			}
		}

		branchToPop := lines[len(lines)-1]
		newStack := lines[:len(lines)-1]
		if err := os.WriteFile(stackFile, []byte(strings.Join(newStack, "\n")+"\n"), 0644); err != nil {
			return fmt.Errorf("could not update stack file: %w", err)
		}

		fmt.Printf("Popping '%s' from the stack and checking out...\n", branchToPop)
		err = runGitCommand("checkout", branchToPop)
		if err != nil {
			return err
		}

		if didStash {
			return stashPop()
		}
		return nil
	},
}

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Interactively pick a branch from the stack to checkout.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		lines, err := readLines(stackFile)
		if err != nil {
			return fmt.Errorf("could not read stack file: %w", err)
		}

		if len(lines) == 0 {
			fmt.Println("Stack is empty. Nothing to pick.")
			return nil
		}

		items := make([]list.Item, len(lines))
		for i, branch := range lines {
			items[i] = item(branch)
		}

		l := list.New(items, itemDelegate{}, 0, 0)
		l.Title = "Select a branch to checkout"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle

		m := model{list: l}
		p, err := tea.NewProgram(m).Run()
		if err != nil {
			return fmt.Errorf("error running picker: %w", err)
		}

		finalModel := p.(model)
		branchToCheckout := finalModel.choice

		if branchToCheckout == "" {
			fmt.Println("No branch selected.")
			return nil
		}

		currentBranch, err := getCurrentBranch()
		if err != nil {
			return err
		}
		if currentBranch == branchToCheckout {
			fmt.Printf("Already on '%s'.\n", currentBranch)
			return nil
		}

		var didStash bool
		if shouldStash {
			didStash, err = stashPushIfDirty()
			if err != nil {
				return err
			}
		}

		fmt.Printf("Checking out '%s'...\n", branchToCheckout)
		err = runGitCommand("checkout", branchToCheckout)
		if err != nil {
			return err
		}

		if didStash {
			return stashPop()
		}
		return nil
	},
}

var peekCmd = &cobra.Command{
	Use:   "peek",
	Short: "Shows the branch at the top of the stack.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		lines, err := readLines(stackFile)
		if err != nil {
			return fmt.Errorf("could not read stack file: %w", err)
		}
		if len(lines) == 0 {
			fmt.Println("Stack is empty.")
			return nil
		}
		fmt.Printf("Top of stack: %s\n", lines[len(lines)-1])
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Displays all branches currently in the stack.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		lines, err := readLines(stackFile)
		if err != nil {
			return fmt.Errorf("could not read stack file: %w", err)
		}
		if len(lines) == 0 {
			fmt.Println("Stack is empty.")
			return nil
		}
		fmt.Println("Branch Stack (bottom to top):")
		for i, line := range lines {
			fmt.Printf("%d. %s\n", i+1, line)
		}
		return nil
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clears all branches from the stack.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		err = os.Remove(stackFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Stack is already empty.")
				return nil
			}
			return err
		}
		fmt.Println("Stack has been cleared.")
		return nil
	},
}

func navigateStack(direction int) error {
	stackFile, err := getStackFilePath()
	if err != nil {
		return err
	}
	lines, err := readLines(stackFile)
	if err != nil {
		return fmt.Errorf("could not read stack file: %w", err)
	}
	if len(lines) < 2 {
		return fmt.Errorf("not enough branches in stack to navigate")
	}

	currentBranch, err := getCurrentBranch()
	if err != nil {
		return err
	}
	currentIndex := -1
	for i, branch := range lines {
		if branch == currentBranch {
			currentIndex = i
			break
		}
	}
	if currentIndex == -1 {
		return fmt.Errorf("current branch '%s' not found in stack", currentBranch)
	}

	targetIndex := currentIndex + direction
	if targetIndex < 0 {
		fmt.Println("Already at the bottom of the stack.")
		return nil
	}
	if targetIndex >= len(lines) {
		fmt.Println("Already at the top of the stack.")
		return nil
	}

	var didStash bool
	if shouldStash {
		didStash, err = stashPushIfDirty()
		if err != nil {
			return err
		}
	}

	branchToCheckout := lines[targetIndex]
	action := "up"
	if direction < 0 {
		action = "down"
	}
	fmt.Printf("Moving %s to '%s'...\n", action, branchToCheckout)
	err = runGitCommand("checkout", branchToCheckout)
	if err != nil {
		return err
	}

	if didStash {
		return stashPop()
	}
	return nil
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Checks out the next branch up in the stack.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return navigateStack(1)
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Checks out the previous branch down in the stack.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return navigateStack(-1)
	},
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Rebases the entire stack starting from the bottom.",
	RunE: func(cmd *cobra.Command, args []string) error {
		isContinue, _ := cmd.Flags().GetBool("continue")
		if isContinue {
			return cmdRebaseContinue()
		}

		shouldPull, _ := cmd.Flags().GetBool("pull")
		rebaseStateFile, err := getRebaseStateFilePath()
		if err != nil {
			return err
		}
		_ = os.Remove(rebaseStateFile)

		var didStash bool
		if shouldStash {
			didStash, err = stashPushIfDirty()
			if err != nil {
				return err
			}
		}

		originalBranch, err := getCurrentBranch()
		if err != nil {
			return err
		}
		stackFile, err := getStackFilePath()
		if err != nil {
			return err
		}
		lines, err := readLines(stackFile)
		if err != nil {
			return err
		}
		if len(lines) < 2 {
			fmt.Println("Stack has fewer than two branches. Nothing to rebase.")
			return nil
		}

		err = executeRebaseLoop(lines, originalBranch, 0, shouldPull, didStash)
		if err != nil {
			return err
		}

		fmt.Printf("\n--- Stack rebase finished successfully. Returning to original branch '%s' ---\n", originalBranch)
		if err := runGitCommand("checkout", originalBranch); err != nil {
			return err
		}

		if didStash {
			return stashPop()
		}
		return nil
	},
}

func cmdRebaseContinue() error {
	rebaseStateFile, err := getRebaseStateFilePath()
	if err != nil {
		return err
	}
	content, err := os.ReadFile(rebaseStateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no rebase in progress")
		}
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(parts) != 3 {
		return fmt.Errorf("invalid rebase state file")
	}

	originalBranch := parts[0]
	lastSuccessfulIndex, _ := strconv.Atoi(parts[1])
	didStash, _ := strconv.ParseBool(parts[2])

	stackFile, err := getStackFilePath()
	if err != nil {
		return err
	}
	lines, err := readLines(stackFile)
	if err != nil {
		return err
	}

	fmt.Println("--- Continuing stack rebase ---")
	err = executeRebaseLoop(lines, originalBranch, lastSuccessfulIndex+1, false, didStash)
	if err != nil {
		return err
	}

	fmt.Printf("\n--- Stack rebase finished successfully. Returning to original branch '%s' ---\n", originalBranch)
	if err := runGitCommand("checkout", originalBranch); err != nil {
		return err
	}

	if didStash {
		return stashPop()
	}
	return nil
}

func executeRebaseLoop(stack []string, originalBranch string, startIndex int, shouldPull bool, didStash bool) error {
	rebaseStateFile, _ := getRebaseStateFilePath()

	if startIndex == 0 {
		baseBranch := stack[0]
		fmt.Printf("--- Checking out base branch '%s' ---\n", baseBranch)
		if err := runGitCommand("checkout", baseBranch); err != nil {
			return err
		}
		if shouldPull {
			fmt.Printf("--- Pulling latest changes for '%s' ---\n", baseBranch)
			if err := runGitCommand("pull"); err != nil {
				return err
			}
		}
		startIndex = 1
	}

	for i := startIndex; i < len(stack); i++ {
		branchToRebase := stack[i]
		rebaseOnto := stack[i-1]
		fmt.Printf("\n--- Checking out '%s' ---\n", branchToRebase)
		if err := runGitCommand("checkout", branchToRebase); err != nil {
			return err
		}
		fmt.Printf("--- Rebasing '%s' onto '%s' ---\n", branchToRebase, rebaseOnto)
		if err := runGitCommand("rebase", rebaseOnto); err != nil {
			state := fmt.Sprintf("%s\n%d\n%t", originalBranch, i-1, didStash)
			_ = os.WriteFile(rebaseStateFile, []byte(state), 0644)
			fmt.Fprintf(os.Stderr, "\nERROR: Rebase of '%s' failed.\n", branchToRebase)
			fmt.Fprintln(os.Stderr, "1. Resolve conflicts and run `git rebase --continue`.")
			fmt.Fprintln(os.Stderr, "2. Then run `git stack rebase --continue` to proceed with the stack.")
			return err
		}
	}

	_ = os.Remove(rebaseStateFile)
	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&shouldStash, "stash", true, "Auto-stash local changes before switching branches.")

	rebaseCmd.Flags().Bool("pull", true, "Pull latest changes on the base branch before rebasing.")
	rebaseCmd.Flags().Bool("continue", false, "Continue a stacked rebase after resolving conflicts.")

	rootCmd.AddCommand(checkoutCmd)
	rootCmd.AddCommand(popCmd)
	rootCmd.AddCommand(pickCmd)
	rootCmd.AddCommand(peekCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(rebaseCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

