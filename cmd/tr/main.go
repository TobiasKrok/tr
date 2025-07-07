package main

import (
	"fmt"
	"os"

	"tr/internal/repl"
	"tr/internal/translator"

	"github.com/spf13/cobra"
)

var (
	version   = "1.0.0"
	direction string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "tr [text]",
	Short:   "Translate between English and Spanish",
	Long:    `TR is a command-line tool for translating between English and Spanish with interactive REPL mode and verb conjugations.`,
	Version: version,
	Args:    cobra.ArbitraryArgs,
	Run:     runTranslate,
}

func init() {
	rootCmd.Flags().StringVarP(&direction, "direction", "d", "", "Translation direction: es2en or en2es")

	// Add conjugate subcommand
	var conjugateCmd = &cobra.Command{
		Use:   "conjugate [verb]",
		Short: "Show conjugations for a Spanish verb",
		Long:  `Display conjugation tables for Spanish verbs with expandable tenses.`,
		Args:  cobra.ExactArgs(1),
		Run:   runConjugate,
	}

	rootCmd.AddCommand(conjugateCmd)
}

func runTranslate(cmd *cobra.Command, args []string) {
	// If no arguments provided, start interactive REPL mode
	if len(args) == 0 {
		fmt.Println("Starting interactive mode...")
		repl := repl.New()
		if err := repl.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting REPL: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Non-interactive mode: translate the provided text
	text := ""
	if len(args) == 1 {
		text = args[0]
	} else {
		// Join multiple arguments with spaces
		for i, arg := range args {
			if i > 0 {
				text += " "
			}
			text += arg
		}
	}

	// Determine translation direction
	fromLang, toLang := determineDirection(direction, text)

	// Create translator and perform translation
	t := translator.New()
	result, err := t.Translate(text, fromLang, toLang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Translation error: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayResult(result, fromLang, toLang)

	// If it's a Spanish verb, show conjugations
	if fromLang == "es" && result.IsVerb {
		conjugations, err := t.GetConjugations(text)
		if err == nil && len(conjugations) > 0 {
			fmt.Println()
			displayConjugations(conjugations)
		}
	}
}

func determineDirection(direction, text string) (from, to string) {
	switch direction {
	case "es2en":
		return "es", "en"
	case "en2es":
		return "en", "es"
	default:
		// Auto-detect based on text characteristics
		if isLikelySpanish(text) {
			return "es", "en"
		}
		return "en", "es"
	}
}

func isLikelySpanish(text string) bool {
	// Simple heuristic: check for Spanish-specific characters
	spanishChars := "ñáéíóúü¿¡"
	for _, char := range text {
		for _, sChar := range spanishChars {
			if char == sChar {
				return true
			}
		}
	}

	// Could add more sophisticated detection here
	// For now, default to assuming input is Spanish
	return true
}

func displayResult(result *translator.TranslationResult, fromLang, toLang string) {
	// Import the UI package functions
	translator.DisplayTranslation(result, fromLang, toLang)
}

func displayConjugations(conjugations map[string]map[string]string) {
	translator.DisplayConjugations(conjugations)
}

func runConjugate(cmd *cobra.Command, args []string) {
	verb := args[0]

	// Create translator and get conjugations
	t := translator.New()

	conjugations, err := t.GetConjugations(verb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting conjugations: %v\n", err)
		os.Exit(1)
	}

	if len(conjugations) == 0 {
		fmt.Printf("No conjugations found for verb: %s\n", verb)
		return
	}

	fmt.Printf("Verb Conjugations for: %s\n", verb)
	displayConjugations(conjugations)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
