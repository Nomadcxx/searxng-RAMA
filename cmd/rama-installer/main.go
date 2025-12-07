package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// Installer copies RAMA assets into a SearXNG checkout and can patch settings.yml.
// If -searxng is omitted, it will try to auto-detect common install locations.
//
// Assets copied:
//   theme/rama/definitions.less -> <searxng>/client/simple/src/less/themes/<theme>/definitions.less
//   brand/rama.svg              -> <searxng>/client/simple/src/brand/<theme>.svg
//
// Optional: patch settings.yml theme key when -set-default-theme is provided.

var (
    searxngPath  = flag.String("searxng", "", "Path to SearXNG repo (optional; auto-detect if omitted)")
    themeName    = flag.String("theme-name", "rama", "Theme name to register/use")
    setDefault   = flag.Bool("set-default-theme", false, "Set default theme in settings.yml")
    settingsPath = flag.String("settings", "", "Optional explicit path to settings.yml")
    verbose      = flag.Bool("v", false, "Verbose logging")
)

func main() {
    flag.Parse()

    root, err := detectSearxngRoot(*searxngPath, *verbose)
    if err != nil {
        exitErr("locate SearXNG: %v", err)
    }
    logf("Using SearXNG root: %s", root)

    // Copy theme
    srcTheme := mustAbs("theme/rama/definitions.less")
    dstThemeDir := filepath.Join(root, "client", "simple", "src", "less", "themes", *themeName)
    dstTheme := filepath.Join(dstThemeDir, "definitions.less")
    if err := copyFile(srcTheme, dstTheme); err != nil {
        exitErr("copy theme: %v", err)
    }
    logf("Copied theme -> %s", dstTheme)

    // Copy brand logo
    srcLogo := mustAbs("brand/rama.svg")
    dstLogoDir := filepath.Join(root, "client", "simple", "src", "brand")
    dstLogo := filepath.Join(dstLogoDir, fmt.Sprintf("%s.svg", *themeName))
    if err := copyFile(srcLogo, dstLogo); err != nil {
        exitErr("copy logo: %v", err)
    }
    logf("Copied logo -> %s", dstLogo)

    if *setDefault {
        sp := *settingsPath
        if sp == "" {
            sp = filepath.Join(root, "searx", "settings.yml")
        }
        if err := patchSettings(sp, *themeName); err != nil {
            exitErr("patch settings.yml: %v", err)
        }
        logf("Patched default theme to '%s' in %s", *themeName, sp)
    }

    fmt.Println("Installed RAMA assets successfully.")
    fmt.Printf("Theme: %s\n", dstTheme)
    fmt.Printf("Logo : %s\n", dstLogo)
}

func detectSearxngRoot(userPath string, verbose bool) (string, error) {
    candidates := []string{}
    if userPath != "" {
        candidates = append(candidates, userPath)
    }
    if env := os.Getenv("SEARXNG_PATH"); env != "" {
        candidates = append(candidates, env)
    }
    // Current and parent directories
    if cwd, err := os.Getwd(); err == nil {
        candidates = append(candidates, cwd, filepath.Dir(cwd))
    }
    home, _ := os.UserHomeDir()
    defaults := []string{
        filepath.Join(home, "searxng"),
        filepath.Join(home, "SearXNG"),
        "/opt/searxng", "/srv/searxng", "/usr/local/share/searxng", "/var/lib/searxng",
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
    return "", fmt.Errorf("could not find SearXNG; specify -searxng explicitly")
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

func copyFile(src, dst string) error {
    data, err := os.ReadFile(src)
    if err != nil {
        return fmt.Errorf("read %s: %w", src, err)
    }
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
        return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
    }
    if err := os.WriteFile(dst, data, 0o644); err != nil {
        return fmt.Errorf("write %s: %w", dst, err)
    }
    // validate write
    if info, err := os.Stat(dst); err != nil || info.Size() == 0 {
        return fmt.Errorf("validation failed for %s", dst)
    }
    return nil
}

func patchSettings(path, theme string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read %s: %w", path, err)
    }
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
    // backup
    _ = os.WriteFile(path+".bak", data, 0o644)
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
