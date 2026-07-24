package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const serverName = "diagram-mcp"

// vsCodeConfigPath returns the path to VS Code's global (user) MCP server
// config file, which lists servers under a "servers" key.
func vsCodeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% is not set")
		}
		return filepath.Join(appData, "Code", "User", "mcp.json"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "mcp.json"), nil
	default:
		return filepath.Join(home, ".config", "Code", "User", "mcp.json"), nil
	}
}

// claudeConfigPath returns the path to Claude Desktop's config file, which
// lists servers under a "mcpServers" key.
func claudeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("%%APPDATA%% is not set")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	default:
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	}
}

func registerVSCode(binPath string) (string, error) {
	path, err := vsCodeConfigPath()
	if err != nil {
		return "", err
	}
	entry := map[string]any{
		"type":    "stdio",
		"command": binPath,
		"args":    []string{},
	}
	return path, mergeJSONServerEntry(path, "servers", entry)
}

func registerClaude(binPath string) (string, error) {
	path, err := claudeConfigPath()
	if err != nil {
		return "", err
	}
	entry := map[string]any{
		"command": binPath,
		"args":    []string{},
	}
	return path, mergeJSONServerEntry(path, "mcpServers", entry)
}

// mergeJSONServerEntry reads the JSON document at path (if any), sets
// doc[topKey][serverName] = entry, and writes it back, preserving any other
// keys/servers already present.
func mergeJSONServerEntry(path, topKey string, entry map[string]any) error {
	doc := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if len(data) > 0 {
			if err := json.Unmarshal(data, &doc); err != nil {
				return fmt.Errorf("parsing existing %s: %w", path, err)
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	servers, ok := doc[topKey].(map[string]any)
	if !ok {
		servers = map[string]any{}
	}
	servers[serverName] = entry
	doc[topKey] = servers

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// unregisterVSCode removes the diagram-mcp entry from VS Code's MCP config.
func unregisterVSCode() error {
	path, err := vsCodeConfigPath()
	if err != nil {
		return err
	}
	return removeJSONServerEntry(path, "servers")
}

// unregisterClaude removes the diagram-mcp entry from Claude Desktop's config.
func unregisterClaude() error {
	path, err := claudeConfigPath()
	if err != nil {
		return err
	}
	return removeJSONServerEntry(path, "mcpServers")
}

// removeJSONServerEntry reads the JSON document at path, removes
// doc[topKey][serverName], and writes it back.
func removeJSONServerEntry(path, topKey string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // file doesn't exist, nothing to clean
		}
		return err
	}

	doc := map[string]any{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
	}

	servers, ok := doc[topKey].(map[string]any)
	if !ok {
		return nil // no servers key, nothing to clean
	}
	delete(servers, serverName)
	doc[topKey] = servers

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}
