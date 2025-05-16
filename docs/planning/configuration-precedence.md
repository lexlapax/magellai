# Configuration Resolution and Precedence

This document outlines the configuration precedence rules and implementation approach for Magellai.

## Core Configuration Values

These are the primary configuration values that affect LLM behavior:

1. **Provider** - Which LLM provider to use (e.g., anthropic, openai, gemini)
   - Unique key: `provider<string>`

2. **Provider URL** - Custom endpoint URL for the provider
   - Unique key: `provider_url<string>`

3. **Provider Timeout** - Timeout value for provider requests
   - Unique key: `provider_timeout<float>`

4. **Model** - Which model to use for generation
   - Unique key: `model<string>`

5. **Temperature** - Controls randomness of responses
   - Unique key: `temperature<float>`

6. **Stream** - Whether to stream responses token by token
   - Unique key: `stream<bool>`

## System-Level Configuration

These settings affect system behavior rather than LLM functionality:

- `log.level` - Logging level (unique key: `log_level<string>`)
- `log.format` - Log format (text/JSON) (unique key: `log_format<string>`)

## Multi-Layered Precedence Model

Configuration in Magellai has two distinct precedence hierarchies:

### 1. Internal Configuration File Precedence

Within the config file itself, settings at more specific levels override more general settings:

1. **Provider-specific defaults** (lowest precedence)
   - `providers.anthropic.model`
   - `providers.anthropic.temperature`

2. **Global defaults**
   - `default.provider`
   - `default.temperature`

3. **Command-specific settings**
   - `commands.ask.provider`
   - `commands.ask.model` 

4. **Profile overrides**
   - `profiles.creative.default.temperature`
   - `profiles.creative.commands.ask.model`

### 2. External Input Precedence (across sources)

Sources of configuration are applied in this order:

1. **Built-in defaults**
2. **System configuration** (`/etc/magellai/config.yaml`) 
3. **User configuration** (`~/.config/magellai/config.yaml`)
4. **Project configuration** (`./.magellai.yaml` or searched upwards)
5. **Profile overrides** (from the selected profile)
6. **Environment variables** (`MAGELLAI_*`)
7. **Command-line flags** (highest precedence for initial settings)
8. **REPL commands** (highest precedence, runtime only)

## Implementation Approach

### Configuration Resolution Method

We will implement a `GetResolvedValue` method that follows the precedence rules:

```go
// GetResolvedValue returns the value for a key following the precedence rules
func (m *Manager) GetResolvedValue(key string, command string) interface{} {
    value := interface{}(nil)
    
    // Find default value (lowest precedence)
    defaultKey := fmt.Sprintf("default.%s", strings.TrimPrefix(key, "commands."+command+"."))
    value = m.koanf.Get(defaultKey)
    
    // Find provider-specific value if applicable
    providerKey := ""
    if strings.HasSuffix(key, ".model") || strings.HasSuffix(key, ".temperature") {
        providerVal := m.GetString("default.provider")
        if providerVal != "" {
            providerKey = fmt.Sprintf("providers.%s.%s", 
                providerVal, strings.TrimPrefix(key, "commands."+command+"."))
            if m.koanf.Exists(providerKey) {
                value = m.koanf.Get(providerKey)
            }
        }
    }
    
    // Command-specific value overrides defaults
    commandKey := fmt.Sprintf("commands.%s.%s", command, 
        strings.TrimPrefix(key, "commands."+command+"."))
    if m.koanf.Exists(commandKey) {
        value = m.koanf.Get(commandKey)
    }
    
    // Check for value in the key itself (may come from profile or CLI)
    if m.koanf.Exists(key) {
        value = m.koanf.Get(key)
    }
    
    return value
}
```

### Profile Implementation

Profiles will be implemented with more explicit structure:

```yaml
profiles:
  creative:
    # Global overrides
    default:
      temperature: 1.0
      provider: openai
    
    # Provider overrides
    providers:
      openai:
        model: o3
    
    # Command overrides (takes precedence over the above)
    commands:
      ask:
        provider: openai
        model: o3
```

When loading a profile, we'll flatten this structure internally to individual key-value pairs:

```go
// LoadProfile loads and applies a profile
func (m *Manager) LoadProfile(profileName string) error {
    profilesKey := "profiles"
    profileKey := fmt.Sprintf("%s.%s", profilesKey, profileName)
    
    if !m.koanf.Exists(profileKey) {
        return fmt.Errorf("profile '%s' not found", profileName)
    }
    
    // Get all profile settings as a map
    profileSettings := m.koanf.Cut(profileKey).Raw()
    
    // Log what we're applying for clarity
    log.Debug("Applying profile settings", "profile", profileName, "settings", profileSettings)
    
    // Apply each setting
    for key, value := range profileSettings.(map[string]interface{}) {
        m.koanf.Set(key, value)
    }
    
    // Set active profile
    m.activeProfile = profileName
    return nil
}
```

### Configuration Inspection

We'll add tools to inspect the current configuration state:

```go
// InspectConfig returns the effective config values for a command
func (m *Manager) InspectConfig(command string) map[string]interface{} {
    result := make(map[string]interface{})
    
    // Get all relevant keys for this command
    keys := []string{
        "provider", "model", "temperature", "stream",
    }
    
    for _, k := range keys {
        commandKey := fmt.Sprintf("commands.%s.%s", command, k)
        value := m.GetResolvedValue(commandKey, command)
        source := m.determineValueSource(commandKey, command)
        
        result[k] = map[string]interface{}{
            "value": value,
            "source": source,
        }
    }
    
    return result
}
```

## Implementation Plan

1. **Update Configuration Manager**
   - Add the `GetResolvedValue` method for accurate resolution
   - Modify profile loading to use a more explicit approach
   - Add inspection methods for debugging

2. **Update Configuration File Format**
   - Support the more explicit profile structure
   - Maintain backward compatibility with dot notation

3. **Add Configuration Inspection Commands**
   - `magellai config inspect [command]` to show effective values
   - `magellai config profile show [profile]` to preview profile changes

4. **Update Documentation**
   - Document the precedence rules clearly
   - Provide examples of correct profile usage

## Alias Handling

- **Alias definition**: Aliases are defined in the configuration file and loaded during startup
- **Alias resolution**: Aliases are resolved at runtime against the current configuration

## Expected User Workflow

The expected user interaction follows this pattern:

1. **Default behavior**: Uses settings from configuration files
2. **Command-specific behavior**: Uses command-specific settings when available
3. **Profile activation**: When a profile is selected, its settings override defaults and command settings
4. **Flag overrides**: Command-line flags override all previous settings
5. **REPL commands**: In interactive mode, REPL commands override settings for the current session only

This model provides intuitive behavior while maintaining precise control over configuration values.

