# Agent Development Guide - SearXNG RAMA Edition

## Repository Overview

This is a Go-based installer and custom theme packaging for SearXNG. The project:
- **Primary language**: Go (installer) + Python (SearXNG runtime)
- **Purpose**: Builds a TUI installer for a customized SearXNG instance with RAMA theme
- **Main component**: `cmd/rama-installer/main.go` - Terminal UI installer
- **Packaging**: Arch Linux AUR package (PKGBUILD)

## Build & Development Commands

### Core Build Commands
```bash
# Build the RAMA installer
go build -o rama-installer ./cmd/rama-installer/

# One-line installer (full workflow)
curl -fsSL https://raw.githubusercontent.com/Nomadcxx/searxng-RAMA/main/install.sh | sudo bash

# Manual install process
1. Build: go build -o rama-installer ./cmd/rama-installer/
2. Run: sudo ./rama-installer
```

### Go Dependencies
```bash
# Install dependencies
go mod download

# Clean dependencies
go mod tidy

# Vendor dependencies (if needed)
go mod vendor
```

### Testing Commands
**Note**: This repository doesn't contain Go test files. Testing is primarily through:
- Manual installer testing
- AUR package testing via `makepkg`
- Systematic installer flow verification

### Packaging & Release
```bash
# Build Arch Linux package
makepkg -si

# Clean package build
makepkg --cleanbuild

# Check package validity
namcap PKGBUILD
```

## Code Style & Conventions

### Go Code Style
Based on the existing `cmd/rama-installer/main.go`:

#### Import Organization
```go
// Standard library imports first, then external packages
import (
    "fmt"
    "os"
    "os/exec"
    // ... other stdlib

    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)
```

#### Naming Conventions
- **Exported identifiers**: PascalCase (`type InstallTask struct`, `func NewModel()`)
- **Private identifiers**: camelCase (`var installTasks []installTask`, `func executeTask()`)
- **Constants**: UPPER_SNAKE_CASE (`const DefaultInstallPath = "/opt/searxng-rama"`)
- **Error variables**: prefix with `Err` (`var ErrInvalidPath = errors.New("invalid path")`)

#### Error Handling
```go
// Standard Go error checking
if err != nil {
    return fmt.Errorf("create install directory: %w", err)  // Wrap with context
}

// For optional steps that can fail
if m.tasks[index].optional {
    m.tasks[index].status = statusSkipped
}
```

#### Comments & Documentation
- **Package comments**: Brief description at top of file
- **Function comments**: Use complete sentences starting with function name
- **Complex logic**: Add explanatory comments
- **Public API**: Full godoc documentation

#### Type Definitions
```go
// Use clear, descriptive type names
type installTask struct {
    name        string
    description string
    execute     func(*model) error
    optional    bool
    status      taskStatus
}

// Constants for state management
const (
    stepWelcome installStep = iota
    stepInstalling
    stepComplete
)
```

#### Variable Declarations
```go
// Group related variables
var (
    BgBase       = lipgloss.Color("#2b2d42")
    Primary      = lipgloss.Color("#edf2f4")
    Accent       = lipgloss.Color("#ef233c")
)

// Function-local variables with clear names
installPath := defaultInstallPath
user := defaultUser
```

### Shell Script Style (`install.sh`, `scripts/*.sh`)
```bash
#!/bin/bash
# Script header with purpose and usage

set -e  # Exit on error

# Functions use snake_case
check_privileges() {
    if [ "$EUID" -ne 0 ]; then
        echo "Error: This script must be run as root"
        exit 1
    fi
}

# Clear variable naming
INSTALL_PATH="/opt/searxng-rama"
SOURCE_PATH="/home/nomadx/searxng-custom"

# Use double quotes for variable expansion
echo "Installing to $INSTALL_PATH"
```

### Python Code (SearXNG dependencies)
This project packages SearXNG Python code but doesn't directly contain Python source. Style follows:
- SearXNG's existing Python conventions
- PEP 8 when modifying or adding Python code

## Quality Assurance

### Pre-commit Hooks
```bash
# Install pre-commit
pre-commit install

# Run all hooks
pre-commit run --all-files

# Run specific hook
pre-commit run check-ai-content
pre-commit run check-secrets
```

### Linting & Formatting
**Go**:
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...
```

**Shell**:
```bash
# Use shellcheck for bash scripts
shellcheck install.sh scripts/*.sh
```

### Security Checks
- **Secrets**: `scripts/check-secrets.sh` checks for hardcoded credentials
- **AI content**: `scripts/check-ai-content.sh` identifies AI-generated documentation
- **Default secrets**: Always use placeholder secrets (`ultrasecretkey`) replaced at install time

## Project Structure Conventions

### File Organization
```
/
├── cmd/rama-installer/     # Go installer source
├── theme/rama/             # RAMA theme files (Less/CSS)
├── assets/                 # Branding assets (logos, icons)
├── brand/                  # Screenshots and branding
├── scripts/                # Utility and verification scripts
├── docs/                   # Documentation
└── .pre-commit-config.yaml # Quality hooks
```

### Naming Files
- **Go files**: `camelCase.go` for main files, `snake_case_test.go` for tests (if added)
- **Scripts**: `kebab-case.sh` or `snake_case.sh`
- **Configuration**: `kebab-case.yml` or `PascalCase.config`
- **Theme files**: Follow Less CSS conventions

## Development Workflow

### Adding Features
1. **Understand scope**: This is an installer, not the SearXNG application itself
2. **Modify installer**: Add tasks to `cmd/rama-installer/main.go`
3. **Test flow**: Build and test installer manually
4. **Update packaging**: Modify PKGBUILD if installation process changes
5. **Verify pre-commit**: Run hooks before committing

### Debugging
```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o rama-installer ./cmd/rama-installer/

# Run with strace for system calls
strace -f -o installer.log ./rama-installer
```

### Performance
- Installer should complete within 1-2 minutes
- Optimize file copying and Python venv creation
- Use concurrent operations where safe

## Agent-Specific Guidelines

### For AI Coding Agents
1. **Preserve installer flow**: The TUI has specific state transitions (welcome → installing → complete)
2. **Keep tasks atomic**: Each `installTask` should do one specific thing
3. **Error handling**: Always provide clear error messages for debugging
4. **User experience**: Installer should be informative but not overwhelming

### When Modifying
- **Theme changes**: Update `theme/rama/definitions.less` and rebuild via PKGBUILD
- **Installer logic**: Add new `installTask` entries with clear `execute` functions
- **Configuration**: Defaults are in `settings.yml` template, modified at install time

### Security Considerations
- Never commit actual secrets (use placeholders)
- Validate user input in installer
- Ensure proper permissions (use `searxng` user)
- Generate secure random keys at installation

## CI/CD & Automation

This project currently uses:
- **Pre-commit hooks**: Local quality gates
- **Manual testing**: AUR package validation
- **User testing**: Community feedback via GitHub issues

Potential future additions:
- GitHub Actions for automated builds
- Integration tests for installer flow
- Package testing in clean environments

---

## Quick Reference

### Essential Commands
```bash
# Build and test
go build -o rama-installer ./cmd/rama-installer/
./rama-installer

# Quality checks
pre-commit run --all-files
shellcheck install.sh

# Packaging
makepkg -si
```

### Code Patterns to Follow
- **Error handling**: Always wrap errors with context
- **Task design**: Each installer task should be independent and idempotent
- **User feedback**: Clear progress indicators in TUI
## **Disclaimer**

This is an installer project. The actual SearXNG Python code is a submodule/dependency. Changes to SearXNG itself should be made upstream or through theme customization only.
```

### Maintainer Notes
Last updated: $(date +%Y-%m-%d)
Repository: https://github.com/Nomadcxx/searxng-RAMA
Primary language: Go (installer) + Python (runtime)