# Commit Convention

This document describes the commit message format used in this project.

## Commit Types

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Style changes (formatting, missing semi colons, etc)
- `refactor`: Code refactoring without behavior changes
- `perf`: Performance improvements
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks, dependency updates
- `ci`: CI/CD related changes
- `build`: Build system changes
- `revert`: Reverting a previous commit

## Scopes

- `core`: Core application logic
- `config`: Configuration system
- `downloader`: Download system
- `logging`: Logging system
- `repository`: Repository management
- `cmd`: CLI commands

## Format

```
type(scope): description

[optional body]

[optional footer]
```

Example:
```
docs(config): update configuration examples in README

- Add new example for custom repository configuration
- Update outdated JDK paths
```
