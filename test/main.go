package main

// This file can be used to quickly test the tr tool during development

import (
	"fmt"
	"log"
	"tr/internal/translator"
)

func main() {
	// Quick test of the translation functionality
	t := translator.New()
	
	// Test Spanish to English
	result, err := t.Translate("hola", "es", "en")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Translation: %s -> %s\n", result.OriginalText, result.Translation)
	
	// Test verb conjugation
	if result.IsVerb {
		conjugations, err := t.GetConjugations("hola")
		if err == nil {
			fmt.Printf("Found %d conjugation sets\n", len(conjugations))
		}
	}
	
	// Test a verb
	verbResult, err := t.Translate("caminar", "es", "en")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Verb translation: %s -> %s (IsVerb: %v)\n", 
		verbResult.OriginalText, verbResult.Translation, verbResult.IsVerb)
	
	if verbResult.IsVerb {
		conjugations, err := t.GetConjugations("caminar")
		if err == nil && len(conjugations) > 0 {
			fmt.Println("Conjugations found:")
			for tense, persons := range conjugations {
				fmt.Printf("  %s:\n", tense)
				for person, form := range persons {
					fmt.Printf("    %s: %s\n", person, form)
				}
			}
		}
	}
}
