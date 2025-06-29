package translator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

// TranslationResult represents the result of a translation
type TranslationResult struct {
	OriginalText string   `json:"original_text"`
	Translation  string   `json:"translation"`
	IsVerb       bool     `json:"is_verb"`
	Definitions  []string `json:"definitions"`
	Examples     []string `json:"examples"`
}

// Translator interface defines the contract for translation services
type Translator interface {
	Translate(text, from, to string) (*TranslationResult, error)
	GetConjugations(verb string) (map[string]map[string]string, error)
}

// translator is the main translator implementation
type translator struct {
	client    *http.Client
	cache     map[string]map[string]map[string]string
	cacheMux  sync.RWMutex
	cacheFile string
}

// New creates a new translator instance
func New() Translator {
	homeDir, _ := os.UserHomeDir()
	cacheFile := filepath.Join(homeDir, ".config", "tr", "conjugations-cache.json")

	t := &translator{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		cache:     make(map[string]map[string]map[string]string),
		cacheFile: cacheFile,
	}

	// Load cached conjugations
	t.loadCache()

	return t
}

// Translate translates text from one language to another using MyMemory API
func (t *translator) Translate(text, from, to string) (*TranslationResult, error) {
	// Clean and prepare the text
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	// Build the API URL for MyMemory (free translation service)
	baseURL := "https://api.mymemory.translated.net/get"
	params := url.Values{}
	params.Add("q", text)
	params.Add("langpair", fmt.Sprintf("%s|%s", from, to))

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make the HTTP request
	resp, err := t.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make translation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("translation service returned status %d", resp.StatusCode)
	}

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
		ResponseStatus int `json:"responseStatus"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse translation response: %w", err)
	}

	if response.ResponseStatus != 200 {
		return nil, fmt.Errorf("translation failed with status %d", response.ResponseStatus)
	}

	// Check if the word is likely a verb (simple heuristic)
	isVerb := from == "es" && isLikelySpanishVerb(text)

	return &TranslationResult{
		OriginalText: text,
		Translation:  response.ResponseData.TranslatedText,
		IsVerb:       isVerb,
		Definitions:  []string{response.ResponseData.TranslatedText},
		Examples:     []string{},
	}, nil
}

// GetConjugations retrieves verb conjugations for Spanish verbs using SpanishDict
func (t *translator) GetConjugations(verb string) (map[string]map[string]string, error) {
	verb = strings.ToLower(strings.TrimSpace(verb))

	// Check cache for verbs
	if cached := t.getCachedConjugations(verb); cached != nil {
		return cached, nil
	}

	// Try to get conjugations from SpanishDict
	var conjugations map[string]map[string]string
	var err error
	conjugations, err = t.getConjugationsFromSpanishDict(verb)

	// Generate rule-based as backup for regular verbs only
	ruleBasedConjugations := t.generateRuleBasedConjugations(verb)

	// If HTML parsing failed completely, use rule-based
	if err != nil || len(conjugations) == 0 {
		conjugations = ruleBasedConjugations
	} else {
		// HTML parsing succeeded partially - fill in missing entries with rule-based
		for tense, ruleConjugations := range ruleBasedConjugations {
			if htmlConjugations, exists := conjugations[tense]; exists {
				// Tense exists in HTML data, fill in missing persons
				for person, ruleConjugation := range ruleConjugations {
					if _, personExists := htmlConjugations[person]; !personExists || htmlConjugations[person] == "" || htmlConjugations[person] == "-" {
						htmlConjugations[person] = ruleConjugation
					}
				}
			} else {
				// Tense doesn't exist in HTML data, add the entire rule-based tense
				conjugations[tense] = ruleConjugations
			}
		}
	}

	// Cache the results if we got any
	if len(conjugations) > 0 {
		t.cacheConjugations(verb, conjugations)
	}

	return conjugations, nil
}

// isLikelySpanishVerb checks if a word is likely a Spanish verb
func isLikelySpanishVerb(word string) bool {
	word = strings.ToLower(word)
	return strings.HasSuffix(word, "ar") ||
		strings.HasSuffix(word, "er") ||
		strings.HasSuffix(word, "ir")
}

// getVerbStem extracts the stem from a Spanish verb
func getVerbStem(verb string) string {
	if len(verb) < 3 {
		return verb
	}
	return verb[:len(verb)-2]
}

// conjugateArVerb conjugates regular -ar verbs
func conjugateArVerb(stem string) map[string]map[string]string {
	return map[string]map[string]string{
		"present": {
			"yo":       stem + "o",
			"tú":       stem + "as",
			"él/ella":  stem + "a",
			"nosotros": stem + "amos",
			"vosotros": stem + "áis",
			"ellos":    stem + "an",
		},
		"preterite": {
			"yo":       stem + "é",
			"tú":       stem + "aste",
			"él/ella":  stem + "ó",
			"nosotros": stem + "amos",
			"vosotros": stem + "asteis",
			"ellos":    stem + "aron",
		},
	}
}

// DisplayTranslation displays translation results in a formatted table
func DisplayTranslation(result *TranslationResult, fromLang, toLang string) {
	// Create color objects for text only (no background colors)
	headerColor := color.New(color.FgCyan, color.Bold)

	// Create and configure the table with a simple style
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)

	// Set headers based on language direction
	var fromHeader, toHeader string
	switch fromLang {
	case "es":
		fromHeader = "Spanish"
	case "en":
		fromHeader = "English"
	default:
		fromHeader = strings.ToUpper(fromLang)
	}

	switch toLang {
	case "es":
		toHeader = "Spanish"
	case "en":
		toHeader = "English"
	default:
		toHeader = strings.ToUpper(toLang)
	}

	t.AppendHeader(table.Row{
		headerColor.Sprint(fromHeader),
		headerColor.Sprint(toHeader),
	})

	t.AppendRow(table.Row{
		result.OriginalText,
		result.Translation,
	})

	fmt.Println(t.Render())
}

// DisplayConjugations displays verb conjugations in a formatted table
func DisplayConjugations(conjugations map[string]map[string]string) {
	if len(conjugations) == 0 {
		return
	}

	// Create color objects for text only (no background colors)
	headerColor := color.New(color.FgGreen, color.Bold)
	personColor := color.New(color.FgYellow)

	fmt.Println("\n" + headerColor.Sprint("Verb Conjugations:"))

	// Create and configure the table with simple style
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)

	// Add headers
	headers := []interface{}{headerColor.Sprint("Person")}
	tenses := []string{}
	for tense := range conjugations {
		tenseTitle := FormatTenseName(tense)
		headers = append(headers, headerColor.Sprint(tenseTitle))
		tenses = append(tenses, tense)
	}
	t.AppendHeader(table.Row(headers))

	// Add rows for each person
	persons := []string{"yo", "tú", "él/ella", "nosotros", "vosotros", "ellos"}
	for _, person := range persons {
		row := []interface{}{personColor.Sprint(person)}
		for _, tense := range tenses {
			if conjugation, exists := conjugations[tense][person]; exists {
				row = append(row, conjugation)
			} else {
				row = append(row, "-")
			}
		}
		t.AppendRow(table.Row(row))
	}

	fmt.Println(t.Render())
}

// DisplayConjugationsExpandable displays verb conjugations with expandable options
func DisplayConjugationsExpandable(conjugations map[string]map[string]string, defaultTenses []string, showAll bool) {
	if len(conjugations) == 0 {
		return
	}

	// Create color objects
	headerColor := color.New(color.FgGreen, color.Bold)
	personColor := color.New(color.FgYellow)
	verbColor := color.New(color.FgWhite)
	infoColor := color.New(color.FgCyan)

	fmt.Println("\n" + headerColor.Sprint("Verb Conjugations:"))

	// Determine which tenses to show
	tensesToShow := defaultTenses
	if showAll {
		tensesToShow = []string{}
		for tense := range conjugations {
			tensesToShow = append(tensesToShow, tense)
		}
	}

	// Filter tenses that actually exist in the conjugations
	availableTenses := []string{}
	for _, tense := range tensesToShow {
		if _, exists := conjugations[tense]; exists {
			availableTenses = append(availableTenses, tense)
		}
	}

	if len(availableTenses) == 0 {
		fmt.Println(infoColor.Sprint("No conjugations available for the specified tenses."))
		return
	}

	// Create and configure the table with simple styling
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)

	// Add headers
	headers := []interface{}{headerColor.Sprint("Person")}
	for _, tense := range availableTenses {
		tenseTitle := FormatTenseName(tense)
		headers = append(headers, headerColor.Sprint(tenseTitle))
	}
	t.AppendHeader(table.Row(headers))

	// Add rows for each person
	persons := []string{"yo", "tú", "él/ella", "nosotros", "vosotros", "ellos"}
	for _, person := range persons {
		row := []interface{}{personColor.Sprint(person)}
		for _, tense := range availableTenses {
			if conjugation, exists := conjugations[tense][person]; exists {
				row = append(row, verbColor.Sprint(conjugation))
			} else {
				row = append(row, verbColor.Sprint("-"))
			}
		}
		t.AppendRow(table.Row(row))
	}

	fmt.Println(t.Render())

	// Show expansion hint if not showing all tenses
	if !showAll && len(conjugations) > len(availableTenses) {
		hiddenCount := len(conjugations) - len(availableTenses)
		fmt.Printf("\n%s\n",
			infoColor.Sprintf("💡 %d more tenses available. Type 'expand %s' to see all conjugations.",
				hiddenCount, GetLastTranslatedVerb()))
	}
}

// FormatTenseName converts internal tense names to display names
func FormatTenseName(tense string) string {
	switch tense {
	case "present":
		return "Present"
	case "preterite":
		return "Preterite"
	case "conditional":
		return "Conditional"
	case "present_subjunctive":
		return "Pres. Subj."
	case "imperfect":
		return "Imperfect"
	case "future":
		return "Future"
	case "imperfect_subjunctive":
		return "Imp. Subj."
	case "present_perfect":
		return "Pres. Perfect"
	case "pluperfect":
		return "Pluperfect"
	case "future_perfect":
		return "Fut. Perfect"
	case "conditional_perfect":
		return "Cond. Perfect"
	case "present_perfect_subjunctive":
		return "Pres. Perf. Subj."
	default:
		// Simple title case without deprecated strings.Title
		words := strings.Split(strings.ReplaceAll(tense, "_", " "), " ")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		return strings.Join(words, " ")
	}
}

// Global variable to track last translated verb for expansion
var lastTranslatedVerb string

func GetLastTranslatedVerb() string {
	return lastTranslatedVerb
}

func SetLastTranslatedVerb(verb string) {
	lastTranslatedVerb = verb
}

// Cache management methods

// loadCache loads cached conjugations from file
func (t *translator) loadCache() {
	t.cacheMux.Lock()
	defer t.cacheMux.Unlock()

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(t.cacheFile), 0755); err != nil {
		return // Silently fail, caching is optional
	}

	data, err := os.ReadFile(t.cacheFile)
	if err != nil {
		return // Cache file doesn't exist or can't be read
	}

	var cache map[string]map[string]map[string]string
	if err := json.Unmarshal(data, &cache); err == nil {
		t.cache = cache
	}
}

// saveCache saves current cache to file
func (t *translator) saveCache() {
	t.cacheMux.RLock()
	defer t.cacheMux.RUnlock()

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(t.cacheFile), 0755); err != nil {
		return // Silently fail
	}

	data, err := json.MarshalIndent(t.cache, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(t.cacheFile, data, 0644)
}

// getCachedConjugations retrieves conjugations from cache
func (t *translator) getCachedConjugations(verb string) map[string]map[string]string {
	t.cacheMux.RLock()
	defer t.cacheMux.RUnlock()

	if conjugations, exists := t.cache[verb]; exists {
		return conjugations
	}
	return nil
}

// cacheConjugations stores conjugations in cache
func (t *translator) cacheConjugations(verb string, conjugations map[string]map[string]string) {
	t.cacheMux.Lock()
	defer t.cacheMux.Unlock()

	t.cache[verb] = conjugations

	// Save cache asynchronously
	go t.saveCache()
}

// cleanConjugation removes HTML tags and entities from conjugation text
func (t *translator) cleanConjugation(text string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")

	// Replace common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&#8203;", "") // Zero-width space

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

// getConjugationsFromSpanishDict fetches conjugations from SpanishDict using web scraping
func (t *translator) getConjugationsFromSpanishDict(verb string) (map[string]map[string]string, error) {
	// Build the SpanishDict URL
	baseURL := fmt.Sprintf("https://www.spanishdict.com/conjugate/%s", url.QueryEscape(verb))

	// Make the HTTP request
	resp, err := t.client.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch conjugations from SpanishDict: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SpanishDict returned status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read SpanishDict response: %w", err)
	}

	// Parse the HTML to extract conjugations
	return t.parseSpanishDictHTML(string(body), verb)
}

// parseSpanishDictHTML extracts conjugation data from SpanishDict HTML using goquery
func (t *translator) parseSpanishDictHTML(html, verb string) (map[string]map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	conjugations := make(map[string]map[string]string)

	// First, try to get the first SpanishDict table which usually contains the main conjugations
	firstSpanishDictTable := doc.Find("table.sTe03NLF").First()
	if firstSpanishDictTable.Length() > 0 {
		extractedConjugations := t.extractFromSpanishDictTable(firstSpanishDictTable, verb)
		if len(extractedConjugations) > 0 {
			conjugations = extractedConjugations
			return conjugations, nil
		}
	}

	// If the first table didn't work, try fallback methods
	if len(conjugations) == 0 {
		// Try to extract from any table with 6 rows (typical for conjugation)
		doc.Find("table").Each(func(i int, table *goquery.Selection) {
			rows := table.Find("tr")
			if rows.Length() == 6 || rows.Length() == 7 { // 6 persons + optional header
				extractedConjugations := t.extractConjugationsFromTable(table, verb)
				if len(extractedConjugations) > 0 {
					conjugations = extractedConjugations
					return
				}
			}
		})
	}

	return conjugations, nil
}

// extractConjugationsFromTable extracts conjugations from a goquery table selection
func (t *translator) extractConjugationsFromTable(table *goquery.Selection, verb string) map[string]map[string]string {
	conjugations := make(map[string]map[string]string)

	// First try SpanishDict specific structure
	if table.HasClass("sTe03NLF") {
		return t.extractFromSpanishDictTable(table, verb)
	}

	// Fallback to generic table parsing
	persons := []string{"yo", "tú", "él/ella", "nosotros", "vosotros", "ellos"}

	rows := table.Find("tr")

	// Determine if first row is header
	headerOffset := 0
	firstRow := rows.First()
	if firstRow.Find("th").Length() > 0 ||
		strings.Contains(strings.ToLower(firstRow.Text()), "presente") ||
		strings.Contains(strings.ToLower(firstRow.Text()), "preterite") {
		headerOffset = 1
	}

	// Extract conjugations from each row
	rows.Each(func(rowIndex int, row *goquery.Selection) {
		personIndex := rowIndex - headerOffset
		if personIndex < 0 || personIndex >= len(persons) {
			return
		}

		person := persons[personIndex]
		cells := row.Find("td")

		// Process each cell as a potential tense
		cells.Each(func(cellIndex int, cell *goquery.Selection) {
			conjugation := strings.TrimSpace(cell.Text())
			conjugation = t.cleanConjugation(conjugation)

			if conjugation != "" && conjugation != "-" && t.isValidConjugation(conjugation, verb) {
				var tense string
				switch cellIndex {
				case 0:
					tense = "present"
				case 1:
					tense = "preterite"
				case 2:
					tense = "imperfect"
				case 3:
					tense = "conditional"
				case 4:
					tense = "future"
				default:
					return // Skip extra columns
				}

				if conjugations[tense] == nil {
					conjugations[tense] = make(map[string]string)
				}
				conjugations[tense][person] = conjugation
			}
		})
	})

	return conjugations
}

// extractFromSpanishDictTable extracts conjugations from SpanishDict's specific table structure
func (t *translator) extractFromSpanishDictTable(table *goquery.Selection, verb string) map[string]map[string]string {
	conjugations := make(map[string]map[string]string)

	// Get the header row to determine tense order
	headerRow := table.Find("tr").First()
	tenseNames := []string{}

	// Check header text for tense order
	headerText := strings.TrimSpace(headerRow.Text())

	// SpanishDict first table usually has: Present, Preterite, Imperfect, Conditional, Future
	if strings.Contains(headerText, "PresentPreteriteImperfectConditionalFuture") {
		tenseNames = []string{"present", "preterite", "imperfect", "conditional", "future"}
	} else {
		// Fallback to parsing individual headers
		headerRow.Find("th").Each(func(i int, th *goquery.Selection) {
			if i == 0 {
				return // Skip first empty header
			}
			tenseText := strings.ToLower(strings.TrimSpace(th.Text()))

			var tense string
			switch {
			case strings.Contains(tenseText, "present"):
				tense = "present"
			case strings.Contains(tenseText, "preterite"):
				tense = "preterite"
			case strings.Contains(tenseText, "imperfect"):
				tense = "imperfect"
			case strings.Contains(tenseText, "conditional"):
				tense = "conditional"
			case strings.Contains(tenseText, "future"):
				tense = "future"
			default:
				tense = fmt.Sprintf("tense_%d", i)
			}
			tenseNames = append(tenseNames, tense)
		})
	}

	// Process each row (skip header)
	table.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
		if rowIndex == 0 {
			return // Skip header row
		}

		rowText := strings.TrimSpace(row.Text())

		// Extract pronoun and conjugations from the concatenated text
		var pronoun string

		if strings.HasPrefix(rowText, "yo") {
			pronoun = "yo"
		} else if strings.HasPrefix(rowText, "tú") {
			pronoun = "tú"
		} else if strings.HasPrefix(rowText, "él/ella/Ud.") {
			pronoun = "él/ella"
		} else if strings.HasPrefix(rowText, "nosotros") {
			pronoun = "nosotros"
		} else if strings.HasPrefix(rowText, "vosotros") {
			pronoun = "vosotros"
		} else if strings.HasPrefix(rowText, "ellos/ellas/Uds.") {
			pronoun = "ellos"
		} else {
			return // Skip unknown row format
		}

		// Parse each cell as a different tense
		cells := row.Find("td")
		if cells.Length() == len(tenseNames) {
			cells.Each(func(cellIndex int, cell *goquery.Selection) {
				if cellIndex >= len(tenseNames) {
					return
				}

				tense := tenseNames[cellIndex]
				conjugation := strings.TrimSpace(cell.Text())

				// Clean up the conjugation
				conjugation = t.cleanConjugation(conjugation)

				if conjugation != "" && conjugation != "-" && t.isValidConjugation(conjugation, verb) {
					if conjugations[tense] == nil {
						conjugations[tense] = make(map[string]string)
					}
					conjugations[tense][pronoun] = conjugation
				}
			})
		}
	})

	return conjugations
}

// isValidConjugation checks if a string looks like a valid conjugation
func (t *translator) isValidConjugation(text, verb string) bool {
	if text == "" || len(text) < 2 || len(text) > 20 {
		return false
	}

	// Remove common non-conjugation patterns
	if strings.Contains(text, "http") || strings.Contains(text, "@") || strings.Contains(text, "#") {
		return false
	}

	// Check if it contains only letters and common Spanish characters
	validChars := regexp.MustCompile(`^[a-záéíóúüñ]+$`)
	if !validChars.MatchString(strings.ToLower(text)) {
		return false
	}

	// For irregular verbs, be more lenient
	// For regular verbs, check if it relates to the verb stem
	stem := getVerbStem(verb)

	// Allow conjugations that either contain the stem or are reasonable length
	if strings.Contains(strings.ToLower(text), strings.ToLower(stem)) ||
		(len(text) >= 2 && len(text) <= 10) {
		return true
	}

	return false
}

// generateRuleBasedConjugations creates conjugations using basic grammatical rules
func (t *translator) generateRuleBasedConjugations(verb string) map[string]map[string]string {
	conjugations := make(map[string]map[string]string)

	if !strings.HasSuffix(verb, "ar") && !strings.HasSuffix(verb, "er") && !strings.HasSuffix(verb, "ir") {
		return conjugations // Not a verb
	}

	stem := getVerbStem(verb)
	ending := verb[len(stem):]

	// Generate present tense
	present := make(map[string]string)
	switch ending {
	case "ar":
		present["yo"] = stem + "o"
		present["tú"] = stem + "as"
		present["él/ella"] = stem + "a"
		present["nosotros"] = stem + "amos"
		present["vosotros"] = stem + "áis"
		present["ellos"] = stem + "an"
	case "er":
		present["yo"] = stem + "o"
		present["tú"] = stem + "es"
		present["él/ella"] = stem + "e"
		present["nosotros"] = stem + "emos"
		present["vosotros"] = stem + "éis"
		present["ellos"] = stem + "en"
	case "ir":
		present["yo"] = stem + "o"
		present["tú"] = stem + "es"
		present["él/ella"] = stem + "e"
		present["nosotros"] = stem + "imos"
		present["vosotros"] = stem + "ís"
		present["ellos"] = stem + "en"
	}
	conjugations["present"] = present

	// Generate preterite tense
	preterite := make(map[string]string)
	switch ending {
	case "ar":
		preterite["yo"] = stem + "é"
		preterite["tú"] = stem + "aste"
		preterite["él/ella"] = stem + "ó"
		preterite["nosotros"] = stem + "amos"
		preterite["vosotros"] = stem + "asteis"
		preterite["ellos"] = stem + "aron"
	case "er", "ir":
		preterite["yo"] = stem + "í"
		preterite["tú"] = stem + "iste"
		preterite["él/ella"] = stem + "ió"
		preterite["nosotros"] = stem + "imos"
		preterite["vosotros"] = stem + "isteis"
		preterite["ellos"] = stem + "ieron"
	}
	conjugations["preterite"] = preterite

	// Generate future tense (same for all verbs)
	future := make(map[string]string)
	future["yo"] = verb + "é"
	future["tú"] = verb + "ás"
	future["él/ella"] = verb + "á"
	future["nosotros"] = verb + "emos"
	future["vosotros"] = verb + "éis"
	future["ellos"] = verb + "án"
	conjugations["future"] = future

	// Generate conditional tense (same for all verbs)
	conditional := make(map[string]string)
	conditional["yo"] = verb + "ía"
	conditional["tú"] = verb + "ías"
	conditional["él/ella"] = verb + "ía"
	conditional["nosotros"] = verb + "íamos"
	conditional["vosotros"] = verb + "íais"
	conditional["ellos"] = verb + "ían"
	conjugations["conditional"] = conditional

	return conjugations
}
