# Contributing to EdgeX MessageBus Client

Thank you for your interest in contributing to the EdgeX MessageBus Client! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- EdgeX Foundry development environment (optional, for testing)

### Setting up the Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/edgex-messagebus-client.git
   cd edgex-messagebus-client
   ```

3. Add the original repository as upstream:
   ```bash
   git remote add upstream https://github.com/clint456/edgex-messagebus-client.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Run tests to ensure everything works:
   ```bash
   go test ./...
   ```

## 📝 How to Contribute

### Reporting Issues

Before creating an issue, please:

1. Check if the issue already exists in the [Issues](https://github.com/clint456/edgex-messagebus-client/issues) section
2. Use the issue templates when available
3. Provide as much detail as possible:
   - Go version
   - Operating system
   - EdgeX version (if applicable)
   - Steps to reproduce
   - Expected vs actual behavior
   - Code samples or logs

### Submitting Changes

1. **Create a new branch** for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-description
   ```

2. **Make your changes** following the coding standards below

3. **Add or update tests** for your changes

4. **Run tests** to ensure nothing is broken:
   ```bash
   go test ./...
   go vet ./...
   ```

5. **Update documentation** if needed (README.md, code comments, etc.)

6. **Commit your changes** with a clear commit message:
   ```bash
   git commit -m "feat: add support for custom message headers"
   # or
   git commit -m "fix: resolve connection timeout issue"
   ```

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request** on GitHub

## 📋 Coding Standards

### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use `go vet` to check for common errors
- Follow Go naming conventions
- Add comments for exported functions, types, and constants
- Keep functions focused and reasonably sized
- Use meaningful variable and function names

### Documentation

- All exported functions must have documentation comments
- Use examples in documentation when helpful
- Update README.md for significant changes
- Include inline comments for complex logic

### Testing

- Write unit tests for new functionality
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error conditions

### Commit Messages

Use conventional commit format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions/changes
- `refactor:` for code refactoring
- `chore:` for maintenance tasks

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestClientConnect
```

### Test Requirements

- All new code should have corresponding tests
- Tests should be deterministic and not rely on external services
- Use mocks for EdgeX dependencies when possible
- Include both positive and negative test cases

## 📚 Documentation

### Code Documentation

- Use godoc format for function and type documentation
- Include examples in documentation comments when helpful
- Document parameters, return values, and any side effects

### README Updates

When making significant changes, update:
- Feature lists
- API documentation
- Examples
- Installation instructions

## 🔄 Pull Request Process

1. **Ensure your PR**:
   - Has a clear title and description
   - References any related issues
   - Includes tests for new functionality
   - Updates documentation as needed
   - Passes all CI checks

2. **PR Review Process**:
   - Maintainers will review your PR
   - Address any feedback or requested changes
   - Once approved, your PR will be merged

3. **After Merge**:
   - Delete your feature branch
   - Pull the latest changes from upstream
   - Thank you for your contribution! 🎉

## 📞 Getting Help

If you need help or have questions:

1. Check the [documentation](README.md)
2. Look through existing [issues](https://github.com/clint456/edgex-messagebus-client/issues)
3. Create a new issue with the "question" label
4. Join EdgeX community discussions

## 📄 License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.

## 🙏 Recognition

Contributors will be recognized in the project documentation and release notes. Thank you for helping make this project better!
