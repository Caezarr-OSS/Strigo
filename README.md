# Strigo - SDK & JDK Version Manager

Strigo is a lightweight and efficient CLI tool designed for managing JDKs and SDKs locally. It allows users to easily install, remove, and switch between multiple versions of Java Development Kits (JDKs) and other SDKs.

![Strigo Logo](assets/img/strigo.jpeg)

## Table of Contents
- [Installation](#installation)
- [Features](#features)
- [Configuration](#configuration)
- [Command Reference](#command-reference)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Architecture](#architecture)

## Installation

### Prerequisites
- Go 1.21 or higher
- A Unix-like operating system (Linux or macOS)
- Access to a Nexus repository (for JDK downloads)

### From Source
```bash
# Clone the repository
git clone https://github.com/your-username/strigo.git
cd strigo

# Build using Task (recommended)
task build

# Or build directly with Go
go build -o bin/strigo
```

### From Releases
Download the appropriate binary for your system from the [releases page](https://github.com/Caezarr-OSS/Strigo/releases).

## Features

- **Multiple JDK Distributions**: Supports all kind of distribution like Temurin, Corretto, Zulu ... 
- **Customizable Configuration**: Uses `strigo.toml` for repository definitions
- **Flexible Shell Configuration**: Supports `.bashrc` and `.zshrc`
- **Nexus Repository Integration**: Fetches JDKs from a Nexus repository
- **Advanced Logging**: Multi-level logging with file and console output
- **Environment Management**: Flexible handling of environment variables
- **Cross-Platform**: Supports Linux and macOS (both amd64 and arm64)
- **Shell Completion**: Built-in completion support for bash, zsh, fish, and powershell

## Configuration

The configuration file (`strigo.toml`) is organized into several sections:

- `[general]`: Basic settings including log level, installation directories, and shell configuration
  - `log_level`: Logging verbosity ("debug", "info", "error")
  - `sdk_install_dir`: Base directory for all SDK installations
  - `cache_dir`: Directory for downloaded artifacts
  - `log_path`: Log file location (empty for console output)
  - `shell_config_path`: Shell RC file to modify for environment variables

- `[registries]`: Defines where Strigo will look for SDKs
  - Currently supports Nexus repositories
  - `{repository}` in the URL is replaced with values from `[sdk_repositories]`

- `[sdk_type]`: Defines supported SDK types and their installation directories
  - Each SDK type needs a unique identifier and install directory
  - Installation paths are relative to `sdk_install_dir`
  - Example: JDKs will be installed in `sdk_install_dir/jdks`

- `[sdk_repositories]`: Maps SDK distributions to their sources
  - Links to a registry defined in `[registries]`
  - Specifies repository name and path within that registry
  - Must reference a type from `[sdk_type]`

Here's a complete example:

```toml
[general]
log_level = "info"
sdk_install_dir = "/home/user/.sdks"
cache_dir = "/home/user/.cache/strigo"
log_path = "/home/user/.logs/strigo"
shell_config_path = "~/.bashrc"  # Shell configuration file path (e.g. ~/.bashrc, ~/.zshrc)

[registries]
nexus = { 
    type = "nexus", 
    api_url = "http://localhost:8081/service/rest/v1/assets?repository={repository}"
}

[sdk_type]
jdk = {
    type = "jdk",
    install_dir = "jdks"     # JDKs will be installed in sdk_install_dir/jdks
}
node = {
    type = "node",
    install_dir = "nodes"    # Node.js will be installed in sdk_install_dir/nodes
}

[sdk_repositories]
temurin = { registry = "nexus", repository = "raw", path = "jdk/adoptium/temurin" }
corretto = { registry = "nexus", repository = "raw", path = "jdk/amazon/corretto" }
zulu = { registry = "nexus", repository = "raw", path = "jdk/azul/zulu" }
```

After installation, your directory structure will look like this:
```
~/.sdks/
├── jdks/
│   ├── temurin-17.0.8/
│   └── corretto-11.0.19/
└── nodes/
    └── node-18.16.0/

~/.cache/strigo/
└── downloads/
    └── (temporary download files)
```

## Command Reference

### Core Commands
- `strigo available [type]`: List available SDK versions from repositories
  - `type`: SDK type (jdk, node)
  - Example: `strigo available jdk`

- `strigo install <type> <version>`: Install a specific SDK version
  - `type`: SDK type (jdk, node)
  - `version`: Version to install (e.g., "17.0.8", "18.16.0")
  - Example: `strigo install jdk 17.0.8`

- `strigo use <type> <version>`: Switch to a specific SDK version
  - `type`: SDK type (jdk, node)
  - `version`: Version to use
  - `--set-env`: Automatically configure environment variables
  - Example: `strigo use jdk 17.0.8 --set-env`

- `strigo list`: List installed SDK versions
  - Example: `strigo list jdk`

- `strigo remove <type> <version>`: Remove an installed SDK version
  - `type`: SDK type (jdk, node)
  - `version`: Version to remove
  - Example: `strigo remove jdk 17.0.8`

- `strigo clean`: Remove invalid environment configurations
  - Example: `strigo clean`

### Utility Commands
- `strigo completion [shell]`: Generate shell completion scripts
  - `shell`: Target shell (bash, zsh, fish, powershell)
  - Example: `strigo completion bash`

- `strigo help [command]`: Get help about any command
  - Example: `strigo help install`

### Global Flags
- `--config <path>`: Specify a custom configuration file path
  - Default: `~/.config/strigo/strigo.toml`
  - Example: `strigo --config /custom/path/strigo.toml install jdk 17.0.8`
  - Use this when you want to use a different configuration file than the default

- `--help, -h`: Show help information for any command
  - Example: `strigo install --help`

## Environment Variables

Strigo manages environment variables for different SDK types:

### Java Environment
- `JAVA_HOME`: Points to the selected JDK installation
- `PATH`: Updated to include `$JAVA_HOME/bin`

### Node.js Environment
- `NODE_HOME`: Points to the selected Node.js installation
- `PATH`: Updated to include `$NODE_HOME/bin`
- `NPM_CONFIG_PREFIX`: User-specific global npm installation directory

### Environment Management
- Use `--set-env` to automatically configure variables
- Without `--set-env`, variables are displayed for manual configuration
- Use `clean` command to remove environment variables for specific SDK type
- Environment changes are shell-specific and require shell restart to take effect

### Shell RC Files
Strigo detects and modifies the appropriate RC file based on your shell:
- Bash: `~/.bashrc`
- Zsh: `~/.zshrc`
- Fish: `~/.config/fish/config.fish`

### Manual Configuration
If automatic configuration is disabled, Strigo will output the necessary commands:
```bash
# For JDK
export JAVA_HOME=/path/to/jdk
export PATH=$JAVA_HOME/bin:$PATH

# For Node.js
export NODE_HOME=/path/to/node
export PATH=$NODE_HOME/bin:$PATH
export NPM_CONFIG_PREFIX=/path/to/npm/global
```

## Troubleshooting

### Common Issues

1. **Invalid JAVA_HOME**
   ```bash
   strigo clean  # Removes invalid configuration
   ```

2. **Download Failures**
   - Check Nexus repository connectivity
   - Verify repository structure
   - Check available disk space

3. **Permission Issues**
   - Ensure write permissions in installation directory
   - Check shell configuration file permissions

## Development

### Project Structure
```
strigo/
├── cmd/          # Command implementations
├── config/       # Configuration handling
├── downloader/   # Download management
├── logging/      # Logging system
└── repository/   # Repository interactions
```

### Adding New Features
1. Follow Go best practices
2. Add tests for new functionality
3. Update documentation
4. Submit a pull request


## Architecture

### Component Diagram
```
+-------------+     +--------------+     +----------------+
|    CLI      |---->| Config       |---->| Repository     |
+-------------+     +--------------+     +----------------+
       |                  |                      |
       v                  v                      v
+-------------+     +--------------+     +----------------+
| Commands    |     | Downloader   |     | Installation   |
+-------------+     +--------------+     +----------------+
```

### Data Flow
1. Command parsing
2. Configuration loading
3. Repository interaction
4. Download management
5. Installation handling
6. Environment configuration

---

## Nexus Repository Structure

The Nexus repository must follow this directory structure:

```
/raw
└── jdk
    ├── adoptium
    │   └── temurin
    │       ├── 11
    │       │   ├── jdk-11.0.24_8
    │       │   ├── jdk-11.0.26_4
    │       ├── 21
    │       │   ├── jdk-21.0.6_7
    │       │       └── OpenJDK21U-jdk_x64_linux_hotspot_21.0.6_7.tar.gz
    │       ├── 8
    │       │   ├── jdk-8u442b06
    │       │       └── OpenJDK8U-jdk_x64_linux_hotspot_8u442b06.tar.gz
    ├── amazon
    │   └── corretto
    │       ├── 11
    │       │   ├── jdk-11.0.26.4.1
    │       │       └── amazon-corretto-11.0.26.4.1-linux-x64.tar.gz
    │       ├── 8
    │       │   ├── jdk-8.442.06.1
    │       │       └── amazon-corretto-8.442.06.1-linux-x64.tar.gz
    ├── azul
    │   └── zulu
    │       ├── 11
    │       │   ├── jdk-11.0.26
    │       │       └── zulu11.78.15-ca-jdk11.0.26-linux_x64.tar.gz
└── node
    └── v22
        └── node-v22.13.1-linux-x64.tar.xz
```

---

## Command Usage

### Checking Available JDKs

#### Command:
```sh
strigo available jdk temurin
```

#### Output:
```
🔹 Available versions:
  - 11:
    ✅ OpenJDK11U-jdk_x64_linux_hotspot_11.0.24_8.tar.gz
    ✅ OpenJDK11U-jdk_x64_linux_hotspot_11.0.26_4.tar.gz
  - 21:
    ✅ OpenJDK21U-jdk_x64_linux_hotspot_21.0.6_7.tar.gz
  - 8:
    ✅ OpenJDK8U-jdk_x64_linux_hotspot_8u442b06.tar.gz
```

### Checking Available Corretto JDKs

#### Command:
```sh
strigo available jdk corretto
```

#### Output:
```
🔹 Available versions:
  - 11:
    ✅ amazon-corretto-11.0.26.4.1-linux-x64.tar.gz
  - 8:
    ✅ amazon-corretto-8.442.06.1-linux-x64.tar.gz
```

### Checking Available Corretto JDKs

#### Command:
```sh
strigo available jdk corretto
```


### Checking specific version (temurin 11)

#### Command:
```sh
strigo available jdk temurin 11
```

#### Output:
```
🔹 Available versions:
  - 11:
    ✅ OpenJDK11U-jdk_x64_linux_hotspot_11.0.24_8.tar.gz
    ✅ OpenJDK11U-jdk_x64_linux_hotspot_11.0.26_4.tar.gz
```



---

## Logging System

Strigo supports a **multi-level logging system**, configurable in `strigo.toml`. The available log levels are:

- `debug`: Logs everything, including detailed debugging information.
- `info`: Logs general execution details and warnings.
- `error`: Logs only critical failures.

Logs are stored in the directory specified in `log_path` under `strigo.toml`.
