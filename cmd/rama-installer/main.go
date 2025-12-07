package main

import (
    "errors"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// Installer copies RAMA assets into a SearXNG checkout and can patch settings.yml.
// If -searxng is omitted, it will try to auto-detect common install locations and
// perform a shallow search. Requires sudo/root (writes into system paths).

var (
    flagSearxng   string
    flagTheme     string
    flagSettings  string
    flagSetDefault bool
    flagVerbose   bool
    flagDryRun    bool
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "rama-installer",
        Short: "Install RAMA theme assets into SearXNG",
        RunE: func(cmd *cobra.Command, args []string) error {
            if os.Geteuid() != 0 {
                return errors.New("rama-installer must be run as root (sudo)")
            }
            viper.AutomaticEnv()
            viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

            searxRoot, err := detectSearxngRoot(flagSearxng, flagVerbose)
            if err != nil {
                return err
            }
            logf("Using SearXNG root: %s", searxRoot)

            if flagDryRun {
                logf("Dry run complete (no files written)")
                return nil
            }

            // Copy theme
            srcTheme := mustAbs("theme/rama/definitions.less")
            dstThemeDir := filepath.Join(searxRoot, "client", "simple", "src", "less", "themes", flagTheme)
            dstTheme := filepath.Join(dstThemeDir, "definitions.less")
            if err := copyFileWithBackup(srcTheme, dstTheme); err != nil {
                return fmt.Errorf("copy theme: %w", err)
            }
            logf("Copied theme -> %s", dstTheme)

            // Copy brand logo
            srcLogo := mustAbs("brand/rama.svg")
            dstLogoDir := filepath.Join(searxRoot, "client", "simple", "src", "brand")
            dstLogo := filepath.Join(dstLogoDir, fmt.Sprintf("%s.svg", flagTheme))
            if err := copyFileWithBackup(srcLogo, dstLogo); err != nil {
                return fmt.Errorf("copy logo: %w", err)
            }
            logf("Copied logo -> %s", dstLogo)

            if flagSetDefault {
                sp := flagSettings
                if sp == "" {
                    sp = filepath.Join(searxRoot, "searx", "settings.yml")
                }
                if err := patchSettingsWithBackup(sp, flagTheme); err != nil {
                    return fmt.Errorf("patch settings.yml: %w", err)
                }
                logf("Patched default theme to '%s' in %s", flagTheme, sp)
            }

            printHeader()
    fmt.Println("Installed RAMA assets successfully.")
            fmt.Printf("Theme: %s\n", dstTheme)
            fmt.Printf("Logo : %s\n", dstLogo)
            return nil
        },
    }

    rootCmd.Flags().StringVar(&flagSearxng, "searxng", "", "Path to SearXNG repo (optional; auto-detect if omitted)")
    rootCmd.Flags().StringVar(&flagTheme, "theme-name", "rama", "Theme name to register/use")
    rootCmd.Flags().StringVar(&flagSettings, "settings", "", "Explicit path to settings.yml (default: <searxng>/searx/settings.yml)")
    rootCmd.Flags().BoolVar(&flagSetDefault, "set-default-theme", false, "Set default theme in settings.yml")
    rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Verbose logging")
    rootCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Detect and validate without writing files")

    // Bind env overrides for convenience
    _ = viper.BindPFlag("searxng", rootCmd.Flags().Lookup("searxng"))
    _ = viper.BindPFlag("theme-name", rootCmd.Flags().Lookup("theme-name"))
    _ = viper.BindPFlag("settings", rootCmd.Flags().Lookup("settings"))

    flagSearxng = viper.GetString("searxng")
    flagTheme = viper.GetString("theme-name")
    flagSettings = viper.GetString("settings")

    if err := rootCmd.Execute(); err != nil {
        exitErr("%v", err)
    }
}

func detectSearxngRoot(userPath string, verbose bool) (string, error) {
    candidates := []string{}
    if userPath != "" {
        candidates = append(candidates, userPath)
    }
    if env := os.Getenv("SEARXNG_PATH"); env != "" {
        candidates = append(candidates, env)
    }
    if env := os.Getenv("SEARX_HOME"); env != "" {
        candidates = append(candidates, env)
    }
    if cwd, err := os.Getwd(); err == nil {
        candidates = append(candidates, cwd, filepath.Dir(cwd))
    }
    home, _ := os.UserHomeDir()
    defaults := []string{
        filepath.Join(home, "searxng"),
        filepath.Join(home, "SearXNG"),
        filepath.Join(home, ".local", "share", "searxng"),
        "/opt/searxng", "/opt/searx/searxng", "/srv/searxng", "/usr/local/share/searxng", "/var/lib/searxng", "/usr/share/searxng", "/etc/searxng",
    }
    candidates = append(candidates, defaults...)

    seen := map[string]bool{}
    for _, c := range candidates {
        if c == "" {
            continue
        }
        abs, err := filepath.Abs(c)
        if err != nil {
            continue
        }
        if seen[abs] {
            continue
        }
        seen[abs] = true
        if isSearxngRoot(abs) {
            if verbose {
                fmt.Printf("[detect] matched SearXNG at %s\n", abs)
            }
            return abs, nil
        } else if verbose {
            fmt.Printf("[detect] not SearXNG: %s\n", abs)
        }
    }

    // Fallback shallow search in likely bases
    searchBases := []string{home, "/opt", "/srv", "/usr/local/share", "/var/lib", "/usr/share", "/etc"}
    for _, base := range searchBases {
        if base == "" {
            continue
        }
        if verbose {
            fmt.Printf("[detect] scanning under %s (depth<=3)\n", base)
        }
        if found := shallowSearch(base, 3, verbose); found != "" {
            return found, nil
        }
    }

    return "", fmt.Errorf("could not find SearXNG; specify --searxng explicitly")
}

func shallowSearch(root string, maxDepth int, verbose bool) string {
    baseDepth := depth(root)
    found := ""
    filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return filepath.SkipDir
        }
        if d.IsDir() {
            if depth(path)-baseDepth > maxDepth {
                return filepath.SkipDir
            }
            name := strings.ToLower(d.Name())
            if strings.Contains(name, "searx") && isSearxngRoot(path) {
                found = path
                return errors.New("found")
            }
        }
        return nil
    })
    if found != "" && verbose {
        fmt.Printf("[detect] shallow search matched %s\n", found)
    }
    return found
}

func depth(path string) int {
    return len(strings.Split(filepath.Clean(path), string(os.PathSeparator)))
}

func isSearxngRoot(path string) bool {
    required := []string{
        filepath.Join(path, "client", "simple", "src", "less"),
        filepath.Join(path, "client", "simple", "src", "brand"),
    }
    for _, p := range required {
        if st, err := os.Stat(p); err != nil || !st.IsDir() {
            return false
        }
    }
    return true
}

func copyFileWithBackup(src, dst string) error {
    data, err := os.ReadFile(src)
    if err != nil {
        return fmt.Errorf("read %s: %w", src, err)
    }
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
        return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
    }
    if _, err := os.Stat(dst); err == nil {
        backupPath := fmt.Sprintf("%s.bak.%d", dst, time.Now().Unix())
        if err := os.Rename(dst, backupPath); err != nil {
            return fmt.Errorf("backup %s: %w", dst, err)
        }
    }
    mode := fs.FileMode(0o644)
    if srcInfo, err := os.Stat(src); err == nil {
        mode = srcInfo.Mode()
    }
    if err := os.WriteFile(dst, data, mode); err != nil {
        return fmt.Errorf("write %s: %w", dst, err)
    }
    if info, err := os.Stat(dst); err != nil || info.Size() == 0 {
        return fmt.Errorf("validation failed for %s", dst)
    }
    return nil
}

func patchSettingsWithBackup(path, theme string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read %s: %w", path, err)
    }
    backup := fmt.Sprintf("%s.bak.%d", path, time.Now().Unix())
    _ = os.WriteFile(backup, data, 0o644)

    lines := strings.Split(string(data), "\n")
    patched := false
    for i, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "theme:") {
            prefix := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
            lines[i] = fmt.Sprintf("%stheme: %s", prefix, theme)
            patched = true
            break
        }
    }
    if !patched {
        insertAt := len(lines)
        for i, line := range lines {
            t := strings.TrimSpace(line)
            if t != "" && !strings.HasPrefix(t, "#") {
                insertAt = i
                break
            }
        }
        newLines := []string{}
        newLines = append(newLines, lines[:insertAt]...)
        newLines = append(newLines, fmt.Sprintf("theme: %s", theme))
        newLines = append(newLines, lines[insertAt:]...)
        lines = newLines
    }
    out := strings.Join(lines, "\n")
    return os.WriteFile(path, []byte(out), 0o644)
}

func mustAbs(rel string) string {
    p, err := filepath.Abs(rel)
    if err != nil {
        exitErr("abs %s: %v", rel, err)
    }
    return p
}

func logf(format string, args ...interface{}) {
    fmt.Printf(format+"\n", args...)
}

func exitErr(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}

func printHeader() {
    path := mustAbs("assets/RAMA.txt")
    data, err := os.ReadFile(path)
    if err != nil {
        return
    }
    fmt.Println(string(data))
}
