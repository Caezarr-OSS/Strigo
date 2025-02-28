# Strigo - SDK & JDK Version Manager

Strigo is a lightweight and efficient CLI tool designed for managing JDKs and SDKs locally. It allows users to easily install, remove, and switch between multiple versions of Java Development Kits (JDKs) and other SDKs.

## Features

- **Multiple JDK Distributions Support**: Supports Temurin, Corretto, and Zulu distributions.
- **Customizable Configuration**: Uses a `strigo.toml` file for repository definitions.
- **Nexus Repository Integration**: Fetches JDKs from a Nexus repository with a predefined structure.
- **Logging System**: Multi-level logging (`DEBUG`, `INFO`, `ERROR`) with file and console output.
- **Portable and Self-Contained**: No dependencies other than the configured Nexus repository.

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

---

## Configuration File (`strigo.toml`)

The configuration file should be structured as follows:

```toml
[general]
log_level = "info"
sdk_install_dir = "/home/user/.sdks"
cache_dir = "/home/user/.cache/strigo"
log_path = "/home/user/.logs/strigo"

[registries]
nexus = { 
    type = "nexus", 
    api_url = "http://localhost:8081/service/rest/v1/assets?repository={repository}"
}

[sdk_repositories]
temurin = { registry = "nexus", repository = "raw", path = "jdk/adoptium/temurin" }
corretto = { registry = "nexus", repository = "raw", path = "jdk/amazon/corretto" }
zulu = { registry = "nexus", repository = "raw", path = "jdk/azul/zulu" }
```




---

