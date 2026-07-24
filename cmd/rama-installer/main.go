package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
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
	stepThemeSelect
	stepInstalling
	stepComplete
)

const (
	defaultThemeIndex = 0
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

type theme struct {
	id          string
	name        string
	description string
	path        string
}

var availableThemes = []theme{
	{
		id:          "rama",
		name:        "RAMA (Dark)",
		description: "Original dark theme with space cadet blue and RAMA red accents",
		path:        "rama",
	},
	{
		id:          "google-light",
		name:        "Google (Light)",
		description: "Google-inspired light theme with refined minimalism",
		path:        "google",
	},
	{
		id:          "google-dark",
		name:        "Google (Dark)",
		description: "Google-inspired dark theme with refined minimalism",
		path:        "google",
	},
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
	switchThemeMode  bool
	selectedOption   int // 0 = Install, 1 = Switch Theme, 2 = Uninstall
	installPath      string
	sourcePath       string
	user             string
	serviceName      string
	selectedTheme    int
}

type taskCompleteMsg struct {
	index   int
	success bool
	error   string
}

const (
	defaultSourcePath  = "/opt/searxng-rama"
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
		{name: "Apply theme", description: "Applying selected theme", execute: applyTheme, status: statusPending},
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
		// Source/install paths default to /opt/searxng-rama (AUR + switch/uninstall
		// modes), but the cross-distro install.sh overrides RAMA_SOURCE_PATH to point
		// at the freshly-built SearXNG checkout it copies from.
		installPath:   getenvDefault("RAMA_INSTALL_PATH", defaultInstallPath),
		sourcePath:    getenvDefault("RAMA_SOURCE_PATH", defaultSourcePath),
		user:          defaultUser,
		serviceName:   defaultServiceName,
		selectedTheme: defaultThemeIndex,
	}

	return m
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
			} else if m.step == stepThemeSelect && m.selectedTheme > 0 {
				m.selectedTheme--
			}
		case "down", "j":
			if m.step == stepWelcome && m.selectedOption < 2 {
				m.selectedOption++
			} else if m.step == stepThemeSelect && m.selectedTheme < len(availableThemes)-1 {
				m.selectedTheme++
			}
		case "enter":
			if m.step == stepWelcome {
				if m.selectedOption == 2 {
					// Uninstall mode
					m.uninstallMode = true
					m.tasks = []installTask{
						{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
						{name: "Stop service", description: "Stopping SearXNG service", execute: stopService, status: statusPending},
						{name: "Disable service", description: "Disabling systemd service", execute: disableService, status: statusPending},
						{name: "Remove service file", description: "Removing service file", execute: removeServiceFile, status: statusPending},
						{name: "Remove installation", description: "Removing installation files", execute: removeInstallation, status: statusPending},
					}
					m.step = stepInstalling
					m.currentTaskIndex = 0
					m.tasks[0].status = statusRunning
					return m, tea.Batch(
						m.spinner.Tick,
						executeTask(0, &m),
					)
				} else if m.selectedOption == 1 {
					// Theme switching mode - minimal tasks
					m.switchThemeMode = true
					m.step = stepThemeSelect
					return m, nil
				} else {
					// Regular install
					m.switchThemeMode = false
					m.step = stepThemeSelect
					return m, nil
				}
			} else if m.step == stepThemeSelect {
				// Create appropriate task list based on mode
				if m.switchThemeMode {
					// Minimal tasks for theme switching
					m.tasks = []installTask{
						{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
						{name: "Stop service", description: "Stopping SearXNG service", execute: stopService, status: statusPending},
						{name: "Apply theme", description: "Applying selected theme", execute: applyTheme, status: statusPending},
						{name: "Start service", description: "Starting SearXNG service", execute: enableAndStartService, status: statusPending},
					}
				} else {
					// Full install tasks (already set in newModel)
					m.tasks = []installTask{
						{name: "Check privileges", description: "Checking root access", execute: checkPrivileges, status: statusPending},
						{name: "Validate source", description: "Validating SearXNG source", execute: validateSource, status: statusPending},
						{name: "Create install directory", description: "Creating installation directory", execute: createInstallDir, status: statusPending},
						{name: "Copy SearXNG files", description: "Copying SearXNG files", execute: copySearxngFiles, status: statusPending},
						{name: "Setup Python venv", description: "Creating venv and installing dependencies", execute: installPythonDeps, status: statusPending},
						{name: "Apply theme", description: "Applying selected theme", execute: applyTheme, status: statusPending},
						{name: "Setup configuration", description: "Setting up configuration", execute: setupConfiguration, status: statusPending},
						{name: "Set permissions", description: "Setting permissions", execute: setPermissions, status: statusPending},
						{name: "Create systemd service", description: "Creating systemd service", execute: createSystemdService, status: statusPending},
						{name: "Enable and start service", description: "Enabling and starting RAMA SearXNG service", execute: enableAndStartService, status: statusPending},
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
		case "esc":
			if m.step == stepThemeSelect {
				m.step = stepWelcome
				return m, nil
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

	// SearXNG ASCII Header - render as single block to avoid lipgloss padding issues
	asciiArt := `  ████████ ██████████   ██████   ████████   ████  ████ ████    ████   ████████
████       ████       ████  ████ ████  ████ ████  ████ ██████  ████ ████
██████████ ████████   ████  ████ ████  ████   ██████   ████████████ ████  ████
      ████ ████       ██████████ ████████   ████  ████ ████  ██████ ████  ████
████████   ██████████ ████  ████ ████  ████ ████  ████ ████    ████   ████████`

	// Apply styling to entire ASCII block at once
	styledASCII := headerStyle.Render(asciiArt)

	// Center the ASCII art
	centeredASCII := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(styledASCII)

	content.WriteString(centeredASCII)
	content.WriteString("\n\n")

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true).
		Align(lipgloss.Center)
	title := "SearXNG Installer : RAMA Edition"
	if m.step == stepThemeSelect {
		title = "SearXNG Installer : Select Theme"
	} else if m.switchThemeMode {
		title = "SearXNG Theme Switcher : RAMA Edition"
	} else if m.uninstallMode {
		title = "SearXNG Uninstaller : RAMA Edition"
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Main content based on step
	var mainContent string
	switch m.step {
	case stepWelcome:
		mainContent = m.renderWelcome()
	case stepThemeSelect:
		mainContent = m.renderThemeSelect()
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

	// Wrap everything in background
	bgStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#000000")).
		Foreground(Primary).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Top)

	return bgStyle.Render(content.String())
}

func (m model) renderWelcome() string {
	var b strings.Builder

	b.WriteString("Select an option:\n\n")

	// Install option
	installPrefix := "  "
	if m.selectedOption == 0 {
		installPrefix = lipgloss.NewStyle().Foreground(Accent).Render("▸ ")
	}
	b.WriteString(installPrefix + "Install SearXNG (RAMA Edition)\n\n")

	// Switch theme option
	themePrefix := "  "
	if m.selectedOption == 1 {
		themePrefix = lipgloss.NewStyle().Foreground(Accent).Render("▸ ")
	}
	b.WriteString(themePrefix + "Switch Theme\n\n")

	// Uninstall option
	uninstallPrefix := "  "
	if m.selectedOption == 2 {
		uninstallPrefix = lipgloss.NewStyle().Foreground(Accent).Render("▸ ")
	}
	b.WriteString(uninstallPrefix + "Uninstall SearXNG (RAMA Edition)\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("Requires root privileges"))

	return b.String()
}

func (m model) renderThemeSelect() string {
	var b strings.Builder

	b.WriteString("Select a theme:\n\n")

	for i, theme := range availableThemes {
		prefix := "  "
		if m.selectedTheme == i {
			prefix = lipgloss.NewStyle().Foreground(Accent).Render("▸ ")
		}

		b.WriteString(prefix + theme.name + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(FgMuted).Render("    "+theme.description) + "\n")

		if i < len(availableThemes)-1 {
			b.WriteString("\n")
		}
	}

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

	selectedTheme := availableThemes[m.selectedTheme]

	if m.switchThemeMode {
		return fmt.Sprintf(`Theme switched successfully!

New theme: %s
Service restarted automatically.

Access RAMA Search at http://localhost:8855

Press Enter to exit`,
			selectedTheme.name)
	}

	return fmt.Sprintf(`Installation complete!

Installation directory: %s
Configuration file: %s/searx/settings.yml
Selected theme: %s

Systemd service: %s
  Enable and start: sudo systemctl enable --now %s
  Check status:     sudo systemctl status %s
  View logs:        sudo journalctl -u %s -f

Access RAMA Search at http://localhost:8855

Press Enter to exit`,
		m.installPath,
		m.installPath,
		selectedTheme.name,
		m.serviceName,
		m.serviceName,
		m.serviceName,
		m.serviceName)
}

func (m model) getHelpText() string {
	switch m.step {
	case stepWelcome:
		return "↑/↓: Navigate  •  Enter: Continue  •  Ctrl+C: Quit"
	case stepThemeSelect:
		if m.switchThemeMode {
			return "↑/↓: Navigate  •  Enter: Select theme & apply  •  Esc: Back  •  Ctrl+C: Quit"
		}
		return "↑/↓: Navigate  •  Enter: Select theme & continue  •  Esc: Back  •  Ctrl+C: Quit"
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

	// Validate pre-built theme bundles exist (warnings only, not failures).
	// These are produced by the build; without them, theme-switch can't work.
	for _, theme := range availableThemes {
		bundle := filepath.Join(m.sourcePath, "searx", "static", "themes", "simple", fmt.Sprintf("sxng-ltr.%s.min.css", theme.id))
		if !fileExists(bundle) {
			fmt.Fprintf(os.Stderr, "[WARNING] theme bundle not found: %s (switching to %q will fail)\n", bundle, theme.id)
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
	// Note: .git is intentionally NOT copied — it bloats the install and would
	// expose repository history if the install dir were ever web-served.
	files := []string{"Makefile", "manage", "requirements.txt", "requirements-dev.txt", "setup.py", "babel.cfg"}

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

// applyTheme switches the served theme by swapping in the pre-built CSS bundle
// for the selected variant. SearXNG serves precompiled CSS (sxng-<side>.min.css)
// — there is no runtime LESS compilation — so the package/installer builds one
// bundle per variant (sxng-<side>.<variant>.min.css) and switching is just a
// file copy + service restart. No recompilation needed.
func applyTheme(m *model) error {
	variant := availableThemes[m.selectedTheme].id
	cssDir := filepath.Join(m.installPath, "searx", "static", "themes", "simple")

	for _, side := range []string{"ltr", "rtl"} {
		src := filepath.Join(cssDir, fmt.Sprintf("sxng-%s.%s.min.css", side, variant))
		dst := filepath.Join(cssDir, fmt.Sprintf("sxng-%s.min.css", side))
		if !fileExists(src) {
			return fmt.Errorf("pre-built theme bundle not found: %s (was the install built with theme variants?)", src)
		}
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("apply theme %s (%s): %w", variant, side, err)
		}
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Applied theme variant: %s\n", variant)
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

	// Install dependencies in venv, bounded by a timeout so a hung network
	// doesn't leave the installer spinning forever.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	pipBin := filepath.Join(venvPath, "bin", "pip")
	pipCmd := exec.CommandContext(ctx, pipBin, "install", "-r", requirementsPath)
	pipCmd.Dir = m.installPath
	if output, err := pipCmd.CombinedOutput(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("pip install timed out after 15m (check network connectivity)")
		}
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

	// Backup original to a single stable path (avoids accumulating .bak.<ts> files on reinstall)
	backupPath := settingsPath + ".bak"
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
	u, err := user.Lookup(m.user)
	if err != nil {
		// Only "user does not exist" is a legitimate skip (e.g. bare-metal install
		// before the service user is created). Any other lookup error is real.
		if _, ok := err.(user.UnknownUserError); ok {
			fmt.Fprintf(os.Stderr, "[WARNING] user %q not found; skipping ownership change\n", m.user)
			return nil
		}
		return fmt.Errorf("lookup user %s: %w", m.user, err)
	}

	chownCmd := exec.Command("chown", "-R", u.Uid+":"+u.Gid, m.installPath)
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
		copyFile(servicePath, servicePath+".bak")
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
		if err := os.RemoveAll(m.installPath); err != nil {
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
