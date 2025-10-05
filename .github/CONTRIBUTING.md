# Contributing to MailCatch

Thank you for your interest in contributing to MailCatch! We welcome contributions from the community.

## How to Contribute

### Reporting Issues

Before creating a new issue, please:

1. Check if the issue already exists in our [issue tracker](../../issues)
2. Make sure you're using the latest version
3. Provide detailed information about your environment and the problem

When creating an issue, please include:

- Operating system and version
- Go version
- Steps to reproduce the issue
- Expected vs actual behavior
- Relevant log output or error messages

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch from `main`: `git checkout -b feature/your-feature-name`
3. Make your changes with clear, descriptive commit messages
4. Add tests for new functionality
5. Ensure all tests pass: `make test`
6. Run linters: `make lint`
7. Update documentation if needed
8. Submit a pull request with a clear description of your changes

### Development Setup

1. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/mailcatch.git
   cd mailcatch
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   make build
   ```

4. Run tests:
   ```bash
   make test
   ```

### Code Style

- Follow Go best practices and idioms
- Use `gofmt` for formatting
- Write clear, self-documenting code
- Add comments for complex logic
- Keep functions focused and small

### Commit Messages

Use clear and descriptive commit messages:

```
feat: add SMTP authentication support

- Implement PLAIN and LOGIN mechanisms
- Add configuration options for auth
- Update documentation

Closes #123
```

Format: `type: description`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Testing

- Write unit tests for new functionality
- Ensure existing tests still pass
- Test both success and error cases
- Include integration tests where appropriate

### Documentation

- Update README.md if adding new features
- Add inline comments for complex code
- Update configuration documentation
- Include examples where helpful

## Project Structure

```
mailcatch/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── models/         # Data models
│   ├── smtp/           # SMTP server implementation
│   ├── storage/        # Storage backends
│   └── web/            # Web interface
├── scripts/            # Build and deployment scripts
└── web/static/         # Static web assets
```

## Getting Help

If you need help or have questions:

- Check existing [issues](../../issues) and [discussions](../../discussions)
- Join our community chat (if available)
- Contact maintainers directly

## Code of Conduct

Please be respectful and considerate in all interactions. We want to maintain a welcoming environment for all contributors.

## License

By contributing to MailCatch, you agree that your contributions will be licensed under the same license as the project (MIT License).

Thank you for contributing! 🚀