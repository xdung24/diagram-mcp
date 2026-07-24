// Package installer implements `diagram-mcp install`: it copies the running
// binary to a user-writable location, ensures that location is on PATH, and
// registers the server (stdio transport) with locally installed MCP clients
// (VS Code, Claude Desktop). It supports Windows, Linux, and macOS.
package installer

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

// Options controls what Run does.
type Options struct {
	// Dir overrides the auto-detected install directory.
	Dir string
	// SkipPath disables adding the install directory to PATH.
	SkipPath bool
	// SkipVSCode disables registering the server with VS Code.
	SkipVSCode bool
	// SkipClaude disables registering the server with Claude Desktop.
	SkipClaude bool
}

// Result reports what Run actually did.
type Result struct {
	BinPath          string
	Dir              string
	Copied           bool
	AddedToPath      bool
	VSCodeConfigPath string
	ClaudeConfigPath string
	Warnings         []string
}

const binBaseName = "diagram-mcp"

// RunInstall implements `diagram-mcp install`: it copies this binary to a
// user-writable location on (or added to) PATH, and registers it as a
// stdio MCP server with VS Code and Claude Desktop.
func RunInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	dir := fs.String("dir", "", "install to this directory instead of an auto-detected one")
	noPath := fs.Bool("no-path", false, "don't add the install directory to PATH")
	noVSCode := fs.Bool("no-vscode", false, "don't register the server with VS Code")
	noClaude := fs.Bool("no-claude", false, "don't register the server with Claude Desktop")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("diagram-mcp install: %v", err)
	}

	res, err := Run(Options{
		Dir:        *dir,
		SkipPath:   *noPath,
		SkipVSCode: *noVSCode,
		SkipClaude: *noClaude,
	})
	if err != nil {
		log.Fatalf("diagram-mcp install failed: %v", err)
	}

	fmt.Printf("Installed diagram-mcp to %s\n", res.BinPath)
	if res.Copied {
		fmt.Printf("Updated file at %s\n", res.BinPath)
	}
	if *noPath {
		fmt.Println("Skipped PATH setup (--no-path)")
	} else if res.AddedToPath {
		fmt.Printf("Added %s to PATH (restart your terminal/shell to pick it up)\n", res.Dir)
	} else {
		fmt.Printf("%s is already on PATH\n", res.Dir)
	}
	if !*noVSCode {
		fmt.Println("Registered diagram-mcp with VS Code (global stdio MCP server)")
		if res.VSCodeConfigPath != "" {
			fmt.Printf("Updated VS Code MCP config at %s\n", res.VSCodeConfigPath)
		}
	}
	if !*noClaude {
		fmt.Println("Registered diagram-mcp with Claude Desktop (stdio MCP server)")
		if res.ClaudeConfigPath != "" {
			fmt.Printf("Updated Claude Desktop MCP config at %s\n", res.ClaudeConfigPath)
		}
	}
	for _, w := range res.Warnings {
		fmt.Println("warning:", w)
	}
}

// RunUninstall implements `diagram-mcp uninstall`: it removes the installed binary
// and unregisters the server from VS Code and Claude Desktop.
func RunUninstall(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	noBinary := fs.Bool("no-binary", false, "don't remove the installed binary")
	noVSCode := fs.Bool("no-vscode", false, "don't unregister from VS Code")
	noClaude := fs.Bool("no-claude", false, "don't unregister from Claude Desktop")
	if err := fs.Parse(args); err != nil {
		log.Fatalf("diagram-mcp uninstall: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("diagram-mcp uninstall: locating home directory: %v", err)
	}

	dir := filepath.Join(defaultInstallDir(home), "mcp-server")
	binName := binBaseName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	destPath := filepath.Join(dir, binName)

	// Remove binary if requested
	if !*noBinary {
		// Stop any running copy before removing
		if err := stopRunningCopy(destPath); err != nil {
			log.Printf("warning: could not stop running process: %v", err)
		}
		if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: could not remove binary at %s: %v", destPath, err)
		} else if err == nil {
			fmt.Printf("Removed %s\n", destPath)
		}
	}

	// Unregister from VS Code
	if !*noVSCode {
		if err := unregisterVSCode(); err != nil {
			log.Printf("warning: could not unregister from VS Code: %v", err)
		} else {
			fmt.Println("Unregistered diagram-mcp from VS Code")
		}
	}

	// Unregister from Claude Desktop
	if !*noClaude {
		if err := unregisterClaude(); err != nil {
			log.Printf("warning: could not unregister from Claude Desktop: %v", err)
		} else {
			fmt.Println("Unregistered diagram-mcp from Claude Desktop")
		}
	}

	fmt.Println("Uninstall completed")
}

// Run performs the install and overwrites any existing installed copy.
func Run(opts Options) (*Result, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locating running executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("locating home directory: %w", err)
	}

	dir := opts.Dir
	if dir == "" {
		dir = filepath.Join(defaultInstallDir(home), "mcp-server")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating install directory %s: %w", dir, err)
	}

	binName := binBaseName
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	destPath := filepath.Join(dir, binName)
	res := &Result{BinPath: destPath, Dir: dir}

	if filepath.Clean(destPath) != filepath.Clean(exePath) {
		if err := stopRunningCopy(destPath); err != nil {
			return nil, err
		}
		if err := copyFile(exePath, destPath); err != nil {
			return nil, fmt.Errorf("copying binary to %s: %w", destPath, err)
		}
		res.Copied = true
	}

	applyInstallActions(res, opts, dir, destPath)
	return res, nil
}

func applyInstallActions(res *Result, opts Options, dir, destPath string) {
	if !opts.SkipPath {
		if isDirInPath(dir) {
			res.AddedToPath = false
		} else if err := addToPath(dir); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("could not add %s to PATH: %v", dir, err))
		} else {
			res.AddedToPath = true
		}
	}

	if !opts.SkipVSCode {
		if path, err := registerVSCode(destPath); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("could not register with VS Code: %v", err))
		} else {
			res.VSCodeConfigPath = path
		}
	}
	if !opts.SkipClaude {
		if path, err := registerClaude(destPath); err != nil {
			res.Warnings = append(res.Warnings, fmt.Sprintf("could not register with Claude Desktop: %v", err))
		} else {
			res.ClaudeConfigPath = path
		}
	}
}

// defaultInstallDir returns a per-user, no-admin-required install directory
// used when no existing PATH entry is suitable.
func defaultInstallDir(home string) string {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, "Programs")
	default: // linux, darwin
		return filepath.Join(home, ".local", "bin")
	}
}

func isDirInPath(dir string) bool {
	target := cleanForCompare(dir)
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if cleanForCompare(p) == target {
			return true
		}
	}
	return false
}

func cleanForCompare(p string) string {
	p = filepath.Clean(p)
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}
	return p
}

func stopRunningCopy(destPath string) error {
	cleanDest := filepath.Clean(destPath)
	currentPID := os.Getpid()
	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("listing running processes: %w", err)
	}
	for _, proc := range procs {
		pid := int(proc.Pid)
		if pid == currentPID {
			continue
		}
		exePath, err := proc.Exe()
		if err != nil || exePath == "" {
			continue
		}
		if filepath.Clean(exePath) != cleanDest {
			continue
		}
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("stopping running server pid %d: %w", pid, err)
		}
	}
	return nil
}

// copyFile copies src to dst (via a temp file + rename), replacing any
// existing file at dst, and marks dst executable.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
