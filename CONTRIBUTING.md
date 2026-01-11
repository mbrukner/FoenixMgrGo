# Contributing to FoenixMgr

Thank you for your interest in contributing to FoenixMgr! This document provides guidelines and information for contributors.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/foenixmgr.git
   cd foenixmgr
   ```
3. **Create a branch** for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Environment

### Prerequisites

- **Go 1.23 or later** - [Download](https://go.dev/dl/)
- **Git** - Version control
- **Make** (optional) - For build automation
- **Foenix hardware** (optional) - For testing, though not required for all contributions

### Building

```bash
# Build the binary
go build -o foenixmgr .

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o foenixmgr-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o foenixmgr-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o foenixmgr-windows-amd64.exe .
```

## Code Guidelines

### Go Style

Follow standard Go conventions:

- Use `gofmt` to format code (automatic in most editors)
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

### Project Structure

```
foenixmgr/
├── main.go              # Entry point
├── cmd/                 # CLI commands (Cobra framework)
│   ├── root.go         # Base command, global flags
│   ├── upload.go       # Upload-related commands
│   ├── flash.go        # Flash programming commands
│   └── ...
├── pkg/
│   ├── config/         # Configuration loading
│   ├── connection/     # Serial/TCP connection abstraction
│   ├── protocol/       # Debug port protocol implementation
│   ├── loader/         # File format parsers (HEX, SREC, etc.)
│   └── util/           # Utilities (hex dump, labels, CRC, etc.)
└── test/               # Test data files
```

### Adding a New Command

1. Create a new file in `cmd/` (e.g., `cmd/mycommand.go`)
2. Define the command using Cobra:
   ```go
   var myCmd = &cobra.Command{
       Use:   "mycommand [args]",
       Short: "Brief description",
       Long:  `Detailed description...`,
       Args:  cobra.ExactArgs(1),
       RunE: func(cmd *cobra.Command, args []string) error {
           return myCommandFunc(args[0])
       },
   }
   ```
3. Register in `init()`:
   ```go
   func init() {
       rootCmd.AddCommand(myCmd)
   }
   ```
4. Follow the connection pattern:
   - Validate flags
   - Create connection
   - Enter debug mode (if needed)
   - Perform operation
   - Exit debug mode
   - Close connection

### Adding a File Loader

1. Create a new file in `pkg/loader/` implementing the `Loader` interface
2. Embed `BaseLoader` for common functionality
3. Implement `Open()`, `Process()` methods
4. Use `SetHandler()` callback for parsed data
5. Add tests in `*_test.go`
6. Add upload command in `cmd/upload.go`

### Protocol Extensions

New protocol commands go in `pkg/protocol/`:

1. Add constant in `commands.go`:
   ```go
   CMDNewCommand = 0x??
   ```
2. Add method to `DebugPort` type:
   ```go
   func (dp *DebugPort) NewCommand(params) error {
       data, err := dp.transfer(CMDNewCommand, address, data, readLength)
       // ... handle response
       return err
   }
   ```
3. Update `protocol.go` documentation

## Testing

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests where appropriate
- Test edge cases and error conditions
- Aim for high coverage on critical paths

Example test:
```go
func TestParseHexAddress(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected uint32
        wantErr  bool
    }{
        {"Simple hex", "1234", 0x1234, false},
        {"With 0x prefix", "0x1234", 0x1234, false},
        {"Invalid", "GHIJ", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParseHexAddress(tt.input)
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if result != tt.expected {
                    t.Errorf("got 0x%X, want 0x%X", result, tt.expected)
                }
            }
        })
    }
}
```

### Hardware Testing

If you have access to Foenix hardware:

1. Test with your specific model (F256jr, A2560, etc.)
2. Test all modified commands
3. Verify output matches expectations
4. Document hardware configuration in PR

If you don't have hardware:
- Focus on unit tests
- Test parsing and protocol logic
- Reviewers with hardware will test functionality

## Documentation

### Code Comments

- Add GoDoc comments for all exported functions/types
- Explain non-obvious logic
- Document protocol details
- Include usage examples in comments

### User Documentation

Update relevant documentation when making user-facing changes:

- `README.md` - For new features or commands
- `CLAUDE.md` - For architecture changes
- `--help` text - For command changes

## Commit Guidelines

### Commit Messages

Use clear, descriptive commit messages:

```
Add flash sector programming command

- Implement sector-based flash programming
- Add CSV bulk programming support
- Include safety confirmations
- Update documentation

Fixes #123
```

Format:
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed description (wrap at 72 chars)
- Reference issues: `Fixes #123` or `Closes #456`

### Commits

- Keep commits focused on a single change
- Test before committing
- Avoid committing debugging code or temporary files

## Pull Request Process

1. **Update tests** - Add/update tests for your changes
2. **Update documentation** - Update README, help text, etc.
3. **Run tests** - Ensure all tests pass: `go test ./...`
4. **Format code** - Run `go fmt ./...`
5. **Build** - Ensure clean build: `go build -o foenixmgr .`
6. **Commit** - Follow commit guidelines
7. **Push** - Push to your fork
8. **Create PR** - Open pull request with description

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Tested on hardware (specify model)
- [ ] All tests pass

## Checklist
- [ ] Code follows project style
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No breaking changes (or documented)

## Hardware Testing
Tested on: [F256jr/F256k/A2560/etc.]
Commands tested: [list]
```

## Bug Reports

When reporting bugs, include:

1. **FoenixMgr version** - `./foenixmgr --version` (or git commit)
2. **Go version** - `go version`
3. **Operating system** - Linux/macOS/Windows version
4. **Hardware model** - F256jr, A2560, etc.
5. **Configuration** - Relevant `foenixmgr.ini` settings
6. **Command used** - Exact command line
7. **Expected behavior** - What should happen
8. **Actual behavior** - What actually happened
9. **Error output** - Full error messages
10. **Steps to reproduce** - Minimal reproduction steps

## Feature Requests

For feature requests:

1. Check existing issues first
2. Describe the use case
3. Explain the desired behavior
4. Consider implementation complexity
5. Note hardware requirements (if any)

## Code Review

All contributions go through code review:

- Be responsive to feedback
- Make requested changes
- Ask questions if unclear
- Be patient - reviews may take time

## Community

- Be respectful and constructive
- Help other contributors
- Share hardware testing results
- Document lessons learned

## License

By contributing, you agree to release your contributions under the MIT License, matching the project license.

## Questions?

- Open an issue for questions
- Check `CLAUDE.md` for architecture details
- Review existing code for examples

Thank you for contributing to FoenixMgr! 🚀
