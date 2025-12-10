package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RAMA Theme colors
var (
	BgBase       = lipgloss.Color("#2b2d42") // Space cadet blue
	Primary      = lipgloss.Color("#edf2f4") // Anti-flash white
	Accent       = lipgloss.Color("#ef233c") // Pantone red
	AccentDark   = lipgloss.Color("#d90429") // Fire engine red
	FgMuted      = lipgloss.Color("#8d99ae") // Cool gray
	ErrorColor   = lipgloss.Color("#ef233c")
	WarningColor = lipgloss.Color("#d90429")
)

// Styles
var (
	checkMark   = lipgloss.NewStyle().Foreground(Accent).SetString("[OK]")
	failMark    = lipgloss.NewStyle().Foreground(ErrorColor).SetString("[FAIL]")
	skipMark    = lipgloss.NewStyle().Foreground(WarningColor).SetString("[SKIP]")
	headerStyle = lipgloss.NewStyle().Foreground(Accent).Bold(true)
)

type installStep int

const (
	stepWelcome installStep = iota
	stepInstalling
	stepComplete
)

type taskStatus int

const (
	statusPending taskStatus = iota
	statusRunning
	statusComplete
	statusFailed
	statusSkipped
)

type installTask struct {
	name        string
	description string
	execute     func(*model) error
	optional    bool
	status      taskStatus
}

type model struct {
	step             installStep
	tasks            []installTask
	currentTaskIndex int
	width            int
	height           int
	spinner          spinner.Model
	errors           []string
	uninstallMode    bool
	selectedOption   int // 0 = Install, 1 = Uninstall
	installPath      string
	sourcePath       string
	user             string
	serviceName      string
}

type taskCompleteMsg struct {
	index   int
	success bool
	error   string
}

const (
	defaultSourcePath  = "/home/nomadx/searxng-custom"
	defaultInstallPath = "/opt/searxng-rama"
	defaultUser        = "searxng"
	defaultServiceName = "searxng-rama"
)

func newModel() model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(FgMuted)
	s.Spinner = spinner.Dot

	installTasks := []installTask{
		{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
		{name: "Validate source", description: "Validating SearXNG source", execute: validateSource, status: statusPending},
		{name: "Create install directory", description: "Creating installation directory", execute: createInstallDir, status: statusPending},
		{name: "Copy SearXNG files", description: "Copying SearXNG files", execute: copySearxngFiles, status: statusPending},
		{name: "Setup Python venv", description: "Creating venv and installing dependencies", execute: installPythonDeps, status: statusPending},
		{name: "Setup configuration", description: "Setting up configuration", execute: setupConfiguration, status: statusPending},
		{name: "Set permissions", description: "Setting permissions", execute: setPermissions, status: statusPending},
		{name: "Create systemd service", description: "Creating systemd service", execute: createSystemdService, status: statusPending},
		{name: "Enable and start service", description: "Enabling and starting RAMA SearXNG service", execute: enableAndStartService, status: statusPending},
	}

	m := model{
		step:             stepWelcome,
		tasks:            installTasks,
		currentTaskIndex: -1,
		spinner:          s,
		errors:           []string{},
		installPath:      defaultInstallPath,
		sourcePath:       defaultSourcePath,
		user:             defaultUser,
		serviceName:      defaultServiceName,
	}

	return m
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step != stepInstalling {
				return m, tea.Quit
			}
		case "up", "k":
			if m.step == stepWelcome && m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.step == stepWelcome && m.selectedOption < 1 {
				m.selectedOption++
			}
		case "enter":
			if m.step == stepWelcome {
				m.uninstallMode = (m.selectedOption == 1)

				if m.uninstallMode {
					m.tasks = []installTask{
						{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
						{name: "Stop service", description: "Stopping SearXNG service", execute: stopService, status: statusPending},
						{name: "Disable service", description: "Disabling systemd service", execute: disableService, status: statusPending},
						{name: "Remove service file", description: "Removing service file", execute: removeServiceFile, status: statusPending},
						{name: "Remove installation", description: "Removing installation files", execute: removeInstallation, status: statusPending},
					}
				}

				m.step = stepInstalling
				m.currentTaskIndex = 0
				m.tasks[0].status = statusRunning
				return m, tea.Batch(
					m.spinner.Tick,
					executeTask(0, &m),
				)
			} else if m.step == stepComplete {
				return m, tea.Quit
			}
		}

	case taskCompleteMsg:
		if msg.success {
			m.tasks[msg.index].status = statusComplete
		} else {
			if m.tasks[msg.index].optional {
				m.tasks[msg.index].status = statusSkipped
				m.errors = append(m.errors, fmt.Sprintf("%s (skipped): %s", m.tasks[msg.index].name, msg.error))
			} else {
				m.tasks[msg.index].status = statusFailed
				m.errors = append(m.errors, fmt.Sprintf("%s: %s", m.tasks[msg.index].name, msg.error))
				m.step = stepComplete
				return m, nil
			}
		}

		m.currentTaskIndex++
		if m.currentTaskIndex >= len(m.tasks) {
			m.step = stepComplete
			return m, nil
		}

		m.tasks[m.currentTaskIndex].status = statusRunning
		return m, executeTask(m.currentTaskIndex, &m)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// SearXNG ASCII Header
	asciiArt := ` ████████ ██████████   ██████   ████████   ████  ████ ████    ████   ████████
████       ████       ████  ████ ████  ████ ████  ████ ██████  ████ ████
██████████ ████████   ████  ████ ████  ████   ██████   ████████████ ████  ████
      ████ ████       ██████████ ████████   ████  ████ ████  ██████ ████  ████
████████   ██████████ ████  ████ ████  ████ ████  ████ ████    ████   ████████`

	content.WriteString(headerStyle.Render(asciiArt))
	content.WriteString("\n\n")

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true).
		Align(lipgloss.Center)
	title := "SearXNG Installer : RAMA Edition"
	if m.uninstallMode {
		title = "SearXNG Uninstaller : RAMA Edition"
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Main content based on step
	var mainContent string
	switch m.step {
	case stepWelcome:
		mainContent = m.renderWelcome()
	case stepInstalling:
		mainContent = m.renderInstalling()
	case stepComplete:
		mainContent = m.renderComplete()
	}

	// Wrap in border
	mainStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Accent).
		Width(m.width - 4)
	content.WriteString(mainStyle.Render(mainContent))
	content.WriteString("\n")

	// Help text
	helpText := m.getHelpText()
	if helpText != "" {
		helpStyle := lipgloss.NewStyle().
			Foreground(FgMuted).
			Italic(true).
			Align(lipgloss.Center)
		content.WriteString("\n" + helpStyle.Render(helpText))
	}

	return content.String()
}

func (m model) renderWelcome() string {
	var b strings.Builder

	b.WriteString("Select an option:\n\n")

	// Install option
	installPrefix := "  "
	if m.selectedOption == 0 {
		installPrefix = lipgloss.NewStyle().Foreground(Accent).Render("▸ ")
	}
	b.WriteString(installPrefix + "Install RAMA SearXNG\n")
	b.WriteString("    Deploys pre-built RAMA-themed SearXNG\n")
	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render(fmt.Sprintf("    Source: %s\n", m.sourcePath)))
	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render(fmt.Sprintf("    Target: %s\n", m.installPath)))
	b.WriteString("\n")

	// Uninstall option
	uninstallPrefix := "  "
	if m.selectedOption == 1 {
		uninstallPrefix = lipgloss.NewStyle().Foreground(Accent).Render("▸ ")
	}
	b.WriteString(uninstallPrefix + "Uninstall RAMA SearXNG\n")
	b.WriteString("    Removes RAMA SearXNG from your system\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("Requires root privileges"))

	return b.String()
}

func (m model) renderInstalling() string {
	var b strings.Builder

	for i, task := range m.tasks {
		var line string
		switch task.status {
		case statusPending:
			line = lipgloss.NewStyle().Foreground(FgMuted).Render("  " + task.name)
		case statusRunning:
			line = m.spinner.View() + " " + lipgloss.NewStyle().Foreground(Primary).Render(task.description)
		case statusComplete:
			line = checkMark.String() + " " + task.name
		case statusFailed:
			line = failMark.String() + " " + task.name
		case statusSkipped:
			line = skipMark.String() + " " + task.name
		}

		b.WriteString(line)
		if i < len(m.tasks)-1 {
			b.WriteString("\n")
		}
	}

	if len(m.errors) > 0 {
		b.WriteString("\n\n")
		for _, err := range m.errors {
			b.WriteString(lipgloss.NewStyle().Foreground(WarningColor).Render(err))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m model) renderComplete() string {
	hasCriticalFailure := false
	for _, task := range m.tasks {
		if task.status == statusFailed && !task.optional {
			hasCriticalFailure = true
			break
		}
	}

	if hasCriticalFailure {
		return lipgloss.NewStyle().Foreground(ErrorColor).Render(
			"Installation failed.\nCheck errors above.\n\nPress Enter to exit")
	}

	if m.uninstallMode {
		return `Uninstall complete.
RAMA SearXNG has been removed.

Press Enter to exit`
	}

	return fmt.Sprintf(`Installation complete!

Installation directory: %s
Configuration file: %s/searx/settings.yml

Systemd service: %s
  Enable and start: sudo systemctl enable --now %s
  Check status:     sudo systemctl status %s
  View logs:        sudo journalctl -u %s -f

Access RAMA Search at http://localhost:8855

Press Enter to exit`, m.installPath, m.installPath, m.serviceName, m.serviceName, m.serviceName, m.serviceName)
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome:
		return "↑/↓: Navigate  •  Enter: Continue  •  Ctrl+C: Quit"
	case stepComplete:
		return "Enter: Exit  •  Ctrl+C: Quit"
	default:
		return "Installing... Please wait"
	}
}

func executeTask(index int, m *model) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(200 * time.Millisecond)

		err := m.tasks[index].execute(m)

		if err != nil {
			fmt.Fprintf(os.Stderr, "\n[DEBUG] Task '%s' failed: %v\n", m.tasks[index].name, err)
			return taskCompleteMsg{
				index:   index,
				success: false,
				error:   err.Error(),
			}
		}

		return taskCompleteMsg{
			index:   index,
			success: true,
		}
	}
}

// Task execution functions

func checkPrivileges(m *model) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("root privileges required - run with sudo")
	}
	return nil
}

func validateSource(m *model) error {
	required := []string{
		filepath.Join(m.sourcePath, "searx"),
		filepath.Join(m.sourcePath, "searx", "static"),
		filepath.Join(m.sourcePath, "searx", "templates"),
	}

	for _, dir := range required {
		if !dirExists(dir) {
			return fmt.Errorf("missing required directory: %s", dir)
		}
	}

	return nil
}

func createInstallDir(m *model) error {
	if err := os.MkdirAll(m.installPath, 0o755); err != nil {
		return fmt.Errorf("create install directory: %w", err)
	}
	return nil
}

func copySearxngFiles(m *model) error {
	dirs := []string{"searx", "dockerfiles", "docs", "utils"}
	files := []string{"Makefile", "manage", "requirements.txt", "requirements-dev.txt", "setup.py", "babel.cfg", ".git"}

	for _, dir := range dirs {
		srcDir := filepath.Join(m.sourcePath, dir)
		dstDir := filepath.Join(m.installPath, dir)

		if !dirExists(srcDir) {
			continue
		}

		if err := copyDir(srcDir, dstDir); err != nil {
			return fmt.Errorf("copy %s: %w", dir, err)
		}
	}

	for _, file := range files {
		srcFile := filepath.Join(m.sourcePath, file)
		dstFile := filepath.Join(m.installPath, file)

		if !fileExists(srcFile) && !dirExists(srcFile) {
			continue
		}

		if dirExists(srcFile) {
			if err := copyDir(srcFile, dstFile); err != nil {
				return fmt.Errorf("copy %s: %w", file, err)
			}
		} else {
			if err := copyFile(srcFile, dstFile); err != nil {
				return fmt.Errorf("copy %s: %w", file, err)
			}
		}
	}

	return nil
}

func installPythonDeps(m *model) error {
	requirementsPath := filepath.Join(m.installPath, "requirements.txt")
	if !fileExists(requirementsPath) {
		return fmt.Errorf("requirements.txt not found")
	}

	// Check if python3 is available
	pythonBin := "python3"
	if _, err := exec.LookPath(pythonBin); err != nil {
		pythonBin = "python"
		if _, err := exec.LookPath(pythonBin); err != nil {
			return fmt.Errorf("python not found - install python3")
		}
	}

	venvPath := filepath.Join(m.installPath, "venv")

	// Create virtual environment
	venvCmd := exec.Command(pythonBin, "-m", "venv", venvPath)
	venvCmd.Dir = m.installPath
	if output, err := venvCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("venv creation failed: %s", string(output))
	}

	// Install dependencies in venv
	pipBin := filepath.Join(venvPath, "bin", "pip")
	pipCmd := exec.Command(pipBin, "install", "-r", requirementsPath)
	pipCmd.Dir = m.installPath
	if output, err := pipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pip install failed: %s", string(output))
	}

	return nil
}

func setupConfiguration(m *model) error {
	settingsPath := filepath.Join(m.installPath, "searx", "settings.yml")

	// Read the default settings file
	defaultSettings, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("read default settings: %w", err)
	}

	// Backup original
	backupPath := settingsPath + ".bak." + fmt.Sprint(time.Now().Unix())
	if err := os.WriteFile(backupPath, defaultSettings, 0o644); err == nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] Backed up original settings to: %s\n", backupPath)
	}

	// Generate secret key
	secretKey := generateSecretKey()

	// Modify settings: update secret_key, port, and bind_address
	settingsStr := string(defaultSettings)
	settingsStr = strings.Replace(settingsStr, `secret_key: "ultrasecretkey"`, `secret_key: "`+secretKey+`"`, 1)
	settingsStr = strings.Replace(settingsStr, "port: 8888", "port: 8855", 1)
	settingsStr = strings.Replace(settingsStr, `bind_address: "127.0.0.1"`, `bind_address: "0.0.0.0"`, 1)

	return os.WriteFile(settingsPath, []byte(settingsStr), 0o644)
}

func generateSecretKey() string {
	cmd := exec.Command("openssl", "rand", "-hex", "32")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("CHANGE_THIS_SECRET_KEY_%d", time.Now().Unix())
	}
	return strings.TrimSpace(string(output))
}

func setPermissions(m *model) error {
	cmd := exec.Command("id", "-u", m.user)
	uidBytes, err := cmd.Output()
	if err != nil {
		return nil // User doesn't exist, skip
	}

	uid := strings.TrimSpace(string(uidBytes))

	cmd = exec.Command("id", "-g", m.user)
	gidBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("get group id for %s: %w", m.user, err)
	}

	gid := strings.TrimSpace(string(gidBytes))

	chownCmd := exec.Command("chown", "-R", uid+":"+gid, m.installPath)
	if err := chownCmd.Run(); err != nil {
		return fmt.Errorf("chown: %w", err)
	}

	return nil
}

func createSystemdService(m *model) error {
	venvPython := filepath.Join(m.installPath, "venv", "bin", "python")
	serviceContent := fmt.Sprintf(`[Unit]
Description=RAMA SearXNG
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
Environment="SEARXNG_SETTINGS_PATH=%s/searx/settings.yml"
ExecStart=%s -m searx.webapp
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, m.user, m.installPath, m.installPath, venvPython)

	servicePath := filepath.Join("/etc/systemd/system", m.serviceName+".service")

	if fileExists(servicePath) {
		backupPath := servicePath + ".bak." + fmt.Sprint(time.Now().Unix())
		copyFile(servicePath, backupPath)
	}

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0o644); err != nil {
		return fmt.Errorf("write service file: %w", err)
	}

	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemd daemon-reload: %w", err)
	}

	return nil
}

func enableAndStartService(m *model) error {
	// Enable the service
	enableCmd := exec.Command("systemctl", "enable", m.serviceName+".service")
	if err := enableCmd.Run(); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	// Start the service
	startCmd := exec.Command("systemctl", "start", m.serviceName+".service")
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("start service: %w", err)
	}

	return nil
}

// Uninstall functions

func stopService(m *model) error {
	cmd := exec.Command("systemctl", "stop", m.serviceName+".service")
	cmd.Run() // Ignore errors if service not running
	return nil
}

func disableService(m *model) error {
	cmd := exec.Command("systemctl", "disable", m.serviceName+".service")
	cmd.Run() // Ignore errors if service not enabled
	return nil
}

func removeServiceFile(m *model) error {
	servicePath := filepath.Join("/etc/systemd/system", m.serviceName+".service")
	if fileExists(servicePath) {
		if err := os.Remove(servicePath); err != nil {
			return fmt.Errorf("remove service file: %w", err)
		}
	}

	cmd := exec.Command("systemctl", "daemon-reload")
	cmd.Run()
	return nil
}

func removeInstallation(m *model) error {
	if dirExists(m.installPath) {
		cmd := exec.Command("rm", "-rf", m.installPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("remove installation: %w", err)
		}
	}
	return nil
}

// Utility functions

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(src, dst)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
