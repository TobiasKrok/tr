package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"tr/internal/config"
	"tr/internal/translator"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// REPL represents the interactive Read-Eval-Print Loop
type REPL struct {
	translator translator.Translator
	direction  string // "es2en" or "en2es"
	running    bool
	config     *config.Config
}

// New creates a new REPL instance
func New() *REPL {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	return &REPL{
		translator: translator.New(),
		direction:  cfg.DefaultDirection,
		running:    false,
		config:     cfg,
	}
}

// Start begins the interactive REPL session
func (r *REPL) Start() error {
	r.running = true

	// Setup signal handling for graceful shutdown
	r.setupSignalHandling()

	// Display welcome message
	r.displayWelcome()

	// Setup terminal for raw input to capture key combinations
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err == nil {
		defer term.Restore(int(os.Stdin.Fd()), oldState)
		return r.runRawMode()
	}

	// Fallback to line-by-line input if raw mode fails
	return r.runLineMode()
}

// setupSignalHandling sets up graceful shutdown on Ctrl+C
func (r *REPL) setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		r.shutdown()
	}()
}

// runRawMode runs the REPL with raw terminal input for key combinations
func (r *REPL) runRawMode() error {
	var input strings.Builder
	buf := make([]byte, 1)

	for r.running {
		r.displayPrompt()
		input.Reset()

	innerLoop:
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}

			char := buf[0]

			switch char {
			case 3: // Ctrl+C
				r.shutdown()
				return nil
			case 20: // Ctrl+T
				r.toggleDirection()
				fmt.Print("\r\033[K") // Clear line
				break innerLoop       // Break inner loop to show new prompt
			case 13, 10: // Enter (CR or LF)
				fmt.Println() // Move to next line
				text := strings.TrimSpace(input.String())
				if text != "" {
					r.processInput(text)
				}
				break innerLoop // Break inner loop to show new prompt
			case 127, 8: // Backspace or Delete
				if input.Len() > 0 {
					// Remove last character from input
					inputStr := input.String()
					input.Reset()
					input.WriteString(inputStr[:len(inputStr)-1])
					// Update display
					fmt.Print("\b \b")
				}
			default:
				if char >= 32 && char <= 126 || char >= 128 { // Printable characters
					input.WriteByte(char)
					fmt.Print(string(char))
				}
			}
		}
	}

	return nil
}

// runLineMode runs the REPL with line-by-line input (fallback mode)
func (r *REPL) runLineMode() error {
	scanner := bufio.NewScanner(os.Stdin)

	for r.running {
		r.displayPrompt()

		if !scanner.Scan() {
			break
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		r.processInput(text)
	}

	return scanner.Err()
}

// displayWelcome shows the initial welcome message
func (r *REPL) displayWelcome() {
	titleColor := color.New(color.FgCyan, color.Bold)

	fmt.Println(titleColor.Sprint("TR - English-Spanish Translator"))
	fmt.Println("Type 'help' for commands, Ctrl+T to toggle direction, Ctrl+C to exit")
	fmt.Println()
}

// displayPrompt shows the current prompt with direction indicator
func (r *REPL) displayPrompt() {
	directionColor := color.New(color.FgGreen)
	promptColor := color.New(color.FgBlue, color.Bold)

	var directionText string
	switch r.direction {
	case "es2en":
		directionText = "Spanish → English"
	case "en2es":
		directionText = "English → Spanish"
	}

	fmt.Printf("%s\n%s ",
		directionColor.Sprintf("Current direction: %s", directionText),
		promptColor.Sprint(">"))
}

// toggleDirection switches between es2en and en2es
func (r *REPL) toggleDirection() {
	if r.direction == "es2en" {
		r.direction = "en2es"
	} else {
		r.direction = "es2en"
	}

	toggleColor := color.New(color.FgYellow, color.Bold)
	var directionText string
	switch r.direction {
	case "es2en":
		directionText = "Spanish → English"
	case "en2es":
		directionText = "English → Spanish"
	}

	fmt.Printf("\n%s\n\n", toggleColor.Sprintf("Switched to: %s", directionText))
}

// processInput handles user input and performs translation
func (r *REPL) processInput(input string) {
	input = strings.TrimSpace(input)

	// Handle expand command for conjugations
	if strings.HasPrefix(strings.ToLower(input), "expand ") {
		verb := strings.TrimSpace(input[7:]) // Remove "expand "
		r.expandConjugations(verb)
		return
	}

	switch strings.ToLower(input) {
	case "exit", "quit", "q":
		r.shutdown()
		return
	case "help", "h":
		r.showHelp()
		return
	case "toggle", "t":
		r.toggleDirection()
		return
	case "clear", "cls":
		r.clearScreen()
		return
	case "config":
		r.showConfig()
		return
	case "tenses":
		r.showAvailableTenses()
		return
	}

	// Perform translation
	fromLang, toLang := r.getLanguages()
	result, err := r.translator.Translate(input, fromLang, toLang)
	if err != nil {
		errorColor := color.New(color.FgRed)
		fmt.Printf("%s\n\n", errorColor.Sprintf("Translation error: %v", err))
		return
	}

	// Display translation result
	fmt.Println()
	translator.DisplayTranslation(result, fromLang, toLang)

	// Show conjugations if it's a Spanish verb
	if fromLang == "es" && result.IsVerb {
		translator.SetLastTranslatedVerb(input) // Store for expand command
		conjugations, err := r.translator.GetConjugations(input)
		if err == nil && len(conjugations) > 0 {
			translator.DisplayConjugationsExpandable(conjugations, r.config.DefaultTenses, r.config.ShowAllTenses)
		}
	}

	fmt.Println()
}

// getLanguages returns the from and to language codes based on current direction
func (r *REPL) getLanguages() (from, to string) {
	switch r.direction {
	case "es2en":
		return "es", "en"
	case "en2es":
		return "en", "es"
	default:
		return "es", "en"
	}
}

// showHelp displays available commands
func (r *REPL) showHelp() {
	helpColor := color.New(color.FgCyan, color.Bold)
	commandColor := color.New(color.FgYellow)

	fmt.Println()
	fmt.Println(helpColor.Sprint("Available Commands:"))
	fmt.Printf("  %s - Show this help message\n", commandColor.Sprint("help, h"))
	fmt.Printf("  %s - Toggle translation direction\n", commandColor.Sprint("toggle, t"))
	fmt.Printf("  %s - Clear the screen\n", commandColor.Sprint("clear, cls"))
	fmt.Printf("  %s - Exit the program\n", commandColor.Sprint("exit, quit, q"))
	fmt.Printf("  %s - Show current configuration\n", commandColor.Sprint("config"))
	fmt.Printf("  %s - Show available tenses\n", commandColor.Sprint("tenses"))
	fmt.Printf("  %s - Show all conjugations for a verb\n", commandColor.Sprint("expand [verb]"))
	fmt.Printf("  %s - Toggle direction (keyboard shortcut)\n", commandColor.Sprint("Ctrl+T"))
	fmt.Printf("  %s - Exit the program\n", commandColor.Sprint("Ctrl+C"))
	fmt.Println()
	fmt.Println("Simply type any word or phrase to translate it.")
	fmt.Println("For Spanish verbs, basic conjugations are shown automatically.")
	fmt.Println("Use 'expand' to see all available tenses and moods.")
	fmt.Println()
}

// clearScreen clears the terminal screen
func (r *REPL) clearScreen() {
	fmt.Print("\033[2J\033[H") // ANSI escape codes to clear screen and move cursor to top
	r.displayWelcome()
}

// shutdown gracefully stops the REPL
func (r *REPL) shutdown() {
	if !r.running {
		return
	}

	r.running = false
	farewellColor := color.New(color.FgGreen)
	fmt.Printf("\n%s\n", farewellColor.Sprint("¡Adiós! Goodbye!"))
	os.Exit(0)
}

// expandConjugations shows all conjugations for a specific verb
func (r *REPL) expandConjugations(verb string) {
	if verb == "" {
		verb = translator.GetLastTranslatedVerb()
		if verb == "" {
			errorColor := color.New(color.FgRed)
			fmt.Printf("%s\n\n", errorColor.Sprint("No verb to expand. Please translate a verb first."))
			return
		}
	}

	conjugations, err := r.translator.GetConjugations(verb)
	if err != nil {
		errorColor := color.New(color.FgRed)
		fmt.Printf("%s\n\n", errorColor.Sprintf("Error getting conjugations: %v", err))
		return
	}

	if len(conjugations) == 0 {
		infoColor := color.New(color.FgYellow)
		fmt.Printf("%s\n\n", infoColor.Sprintf("No conjugations found for '%s'", verb))
		return
	}

	// Show all available tenses
	fmt.Println()
	translator.DisplayConjugationsExpandable(conjugations, config.GetAvailableTenses(), true)
	fmt.Println()
}

// showConfig displays current configuration
func (r *REPL) showConfig() {
	configColor := color.New(color.FgCyan, color.Bold)
	keyColor := color.New(color.FgYellow)
	valueColor := color.New(color.FgWhite)

	fmt.Println()
	fmt.Println(configColor.Sprint("Current Configuration:"))
	fmt.Printf("  %s: %s\n", keyColor.Sprint("Default Direction"), valueColor.Sprint(r.config.DefaultDirection))
	fmt.Printf("  %s: %s\n", keyColor.Sprint("Default Tenses"), valueColor.Sprint(strings.Join(r.config.DefaultTenses, ", ")))
	fmt.Printf("  %s: %s\n", keyColor.Sprint("Show All Tenses"), valueColor.Sprint(r.config.ShowAllTenses))
	fmt.Println()
	fmt.Println("Configuration file location: ~/.config/tr/config.json")
	fmt.Println("Edit the file directly to change settings.")
	fmt.Println()
}

// showAvailableTenses displays all available tenses
func (r *REPL) showAvailableTenses() {
	tenseColor := color.New(color.FgGreen, color.Bold)
	listColor := color.New(color.FgWhite)

	fmt.Println()
	fmt.Println(tenseColor.Sprint("Available Tenses:"))

	allTenses := config.GetAvailableTenses()
	for i, tense := range allTenses {
		displayName := translator.FormatTenseName(tense)
		if contains(r.config.DefaultTenses, tense) {
			// Mark default tenses
			fmt.Printf("  %s %s (default)\n", listColor.Sprint(fmt.Sprintf("%2d.", i+1)),
				color.New(color.FgGreen).Sprint(displayName))
		} else {
			fmt.Printf("  %s %s\n", listColor.Sprint(fmt.Sprintf("%2d.", i+1)), listColor.Sprint(displayName))
		}
	}
	fmt.Println()
	fmt.Println("Default tenses are shown automatically. Use 'expand [verb]' to see all tenses.")
	fmt.Println()
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
