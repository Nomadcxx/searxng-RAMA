package main

import (
    "archive/zip"
    "bytes"
    "flag"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

// This installer copies RAMA assets into a SearXNG checkout and optionally patches settings.yml
// to set the default theme to RAMA.
//
// Usage:
//   rama-installer -searxng /path/to/searxng [ -theme-name rama ] [ -set-default-theme ]
//
// It copies:
//   theme/rama/definitions.less -> <searxng>/client/simple/src/less/themes/rama/definitions.less (creates directory)
//   brand/rama.svg -> <searxng>/client/simple/src/brand/rama.svg
//
// If -set-default-theme is provided, it will attempt to set settings.yml "theme" to the given theme name.

var (
    searxngPath   = flag.String("searxng", "", "Path to SearXNG repo")
    themeName     = flag.String("theme-name", "rama", "Theme name to register/use")
    setDefault    = flag.Bool("set-default-theme", false, "Set default theme in settings.yml")
    settingsPath  = flag.String("settings", "", "Optional explicit path to settings.yml (defaults to <searxng>/searx/settings.yml)")
)

func main() {
    flag.Parse()
    if *searxngPath == "" {
        exitErr("-searxng path is required")
    }
    root, err := filepath.Abs(*searxngPath)
    if err != nil {
        exitErr("resolve searxng path: %v", err)
    }
    // Copy theme
    srcTheme := mustAbs("theme/rama/definitions.less")
    dstThemeDir := filepath.Join(root, "client", "simple", "src", "less", "themes", *themeName)
    dstTheme := filepath.Join(dstThemeDir, "definitions.less")
    if err := copyFile(srcTheme, dstTheme); err != nil {
        exitErr("copy theme: %v", err)
    }

    // Copy brand logo
    srcLogo := mustAbs("brand/rama.svg")
    dstLogoDir := filepath.Join(root, "client", "simple", "src", "brand")
    dstLogo := filepath.Join(dstLogoDir, fmt.Sprintf("%s.svg", *themeName))
    if err := copyFile(srcLogo, dstLogo); err != nil {
        exitErr("copy logo: %v", err)
    }

    if *setDefault {
        sp := *settingsPath
        if sp == "" {
            sp = filepath.Join(root, "searx", "settings.yml")
        }
        if err := patchSettings(sp, *themeName); err != nil {
            exitErr("patch settings.yml: %v", err)
        }
        fmt.Printf("Patched default theme to '%s' in %s\n", *themeName, sp)
    }

    fmt.Println("Installed RAMA assets successfully.")
    fmt.Printf("Theme: %s\n", dstTheme)
    fmt.Printf("Logo : %s\n", dstLogo)
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
    return nil
}

func patchSettings(path, theme string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read %s: %w", path, err)
    }
    // naive patch: if a line starts with "theme:" replace, else prepend near top
    lines := strings.Split(string(data), "\n")
    patched := false
    for i, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "theme:") {
            // preserve indentation of existing line
            prefix := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
            lines[i] = fmt.Sprintf("%stheme: %s", prefix, theme)
            patched = true
            break
        }
    }
    if !patched {
        // insert near top after possible comments/shebang; put after first non-comment
        insertAt := len(lines)
        for i, line := range lines {
            t := strings.TrimSpace(line)
            if t != "" && !strings.HasPrefix(t, "#") {
                insertAt = i
                break
            }
        }
        // build new slice
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

func exitErr(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}

// dummy dependency to avoid unused imports; keeping in case we extend with zip packaging
var _ = zip.FileHeader{}
var _ = bytes.Buffer{}
var _ = fs.FileInfo(nil)
var _ = io.Reader(nil)
