# TR - English-Spanish Translation CLI Tool

A powerful command-line tool for translating between English and Spanish with an interactive REPL mode, verb conjugations, and beautiful terminal output.

## Features

- 🔄 **Interactive REPL Mode**: Enter an interactive prompt for continuous translation
- 🌐 **Bidirectional Translation**: Spanish ↔ English translation support
- 📚 **Verb Conjugation**: Get detailed Spanish verb conjugations
- ⌨️ **Keyboard Shortcuts**: Toggle translation direction with `Ctrl+T`
- 🎨 **Beautiful Terminal UI**: Syntax highlighting and formatted tables
- 🚀 **Non-interactive Mode**: One-off translations from command line
- 📝 **UTF-8 Support**: Full Unicode support for proper character display

## Installation

### Prerequisites

- Go 1.21 or later
- Internet connection (for translation API calls)

### Build from Source

```bash
# Clone the repository
git clone <your-repo-url>
cd tr

# Download dependencies
go mod download

# Build the executable
go build -o tr ./cmd/tr

# Optional: Install globally
go install ./cmd/tr
```

### Cross-platform Builds

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o tr.exe ./cmd/tr

# macOS
GOOS=darwin GOARCH=amd64 go build -o tr-macos ./cmd/tr

# Linux
GOOS=linux GOARCH=amd64 go build -o tr-linux ./cmd/tr
```

## Usage

### Interactive REPL Mode

Launch the tool without arguments to enter interactive mode:

```bash
./tr
```

In interactive mode:
- Type any Spanish or English word/phrase and press Enter
- Use `Ctrl+T` to toggle translation direction (ES→EN or EN→ES)
- Use `Ctrl+C` or type `exit` to quit
- Type `help` for available commands

Example session:
```
TR - English-Spanish Translator
Current direction: Spanish → English
Type 'help' for commands, Ctrl+T to toggle direction, Ctrl+C to exit

> hola
┌─────────────┬─────────────────────┐
│ Spanish     │ English             │
├─────────────┼─────────────────────┤
│ hola        │ hello, hi           │
└─────────────┴─────────────────────┘

> caminar
┌─────────────┬─────────────────────┐
│ Spanish     │ English             │
├─────────────┼─────────────────────┤
│ caminar     │ to walk             │
└─────────────┴─────────────────────┘

Verb Conjugations:
┌─────────────┬─────────────┬─────────────┐
│ Person      │ Present     │ Past        │
├─────────────┼─────────────┼─────────────┤
│ yo          │ camino      │ caminé      │
│ tú          │ caminas     │ caminaste   │
│ él/ella     │ camina      │ caminó      │
│ nosotros    │ caminamos   │ caminamos   │
│ vosotros    │ camináis    │ caminasteis │
│ ellos       │ caminan     │ caminaron   │
└─────────────┴─────────────┴─────────────┘
```

### Non-interactive Mode

For one-off translations:

```bash
# Spanish to English
./tr -d es2en "hola mundo"
./tr --direction es2en caminar

# English to Spanish
./tr -d en2es "good morning"
./tr --direction en2es "how are you"

# Default direction (Spanish to English)
./tr hola
```

### Command Line Options

- `-d, --direction`: Set translation direction (`es2en` or `en2es`)
- `-h, --help`: Show help information
- `-v, --version`: Show version information

## Configuration

### API Keys

Currently, the tool uses free translation services. For production use, you may want to configure API keys:

1. Create a `.env` file in the project directory:
```env
GOOGLE_TRANSLATE_API_KEY=your_key_here
SPANISHDICT_API_KEY=your_key_here
```

2. The application will automatically detect and use these keys if available.

### Supported Translation Services

The tool supports multiple translation backends:
- Google Translate API (with API key)
- MyMemory Translation API (free, no key required)
- SpanishDict API (for verb conjugations)

## Development

### Project Structure

```
tr/
├── cmd/tr/              # Main application entry point
│   └── main.go
├── internal/
│   ├── translator/      # Translation logic
│   │   ├── translator.go
│   │   ├── google.go
│   │   └── mymemory.go
│   ├── conjugator/      # Verb conjugation logic
│   │   └── conjugator.go
│   ├── repl/           # Interactive REPL logic
│   │   └── repl.go
│   └── ui/             # Terminal UI components
│       ├── table.go
│       └── colors.go
├── go.mod
├── go.sum
└── README.md
```

### Adding New Translation Services

To add a new translation service:

1. Create a new file in `internal/translator/`
2. Implement the `Translator` interface:
```go
type Translator interface {
    Translate(text, from, to string) (*TranslationResult, error)
}
```
3. Register the translator in the factory function

### Adding New Languages

To support additional languages:

1. Update the language codes in `internal/translator/translator.go`
2. Add language-specific conjugation rules if needed
3. Update the CLI flags and help text

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Troubleshooting

### Common Issues

**Translation not working:**
- Check your internet connection
- Verify API keys if using premium services
- Try switching translation backends

**Terminal display issues:**
- Ensure your terminal supports UTF-8
- Try running with `--no-color` flag if colors cause problems
- Update your terminal emulator if tables don't display correctly

**Build errors:**
- Ensure Go 1.21+ is installed
- Run `go mod download` to fetch dependencies
- Check that your GOPATH is set correctly

### Getting Help

- Open an issue on GitHub for bugs or feature requests
- Check the documentation for detailed API information
- Join our community Discord for real-time help

## Roadmap

- [ ] Add more language pairs (French, German, Italian)
- [ ] Implement offline dictionary support
- [ ] Add pronunciation guides
- [ ] Voice input/output support
- [ ] Web interface
- [ ] Mobile app companion

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [go-pretty](https://github.com/jedib0t/go-pretty) for beautiful tables
- [fatih/color](https://github.com/fatih/color) for terminal colors
- Translation services and dictionaries used
