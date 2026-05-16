package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	gitignore "github.com/sabhiram/go-gitignore"
)

var (
	ignoreDirs  = []string{".git", "node_modules", "vendor", ".idea", ".vscode", "dist", "build", "__pycache__", ".next", ".cache"}
	ignoreExts  = []string{".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".exe", ".dll", ".so", ".dylib", ".wasm", ".zip", ".tar.gz", ".7z", ".lock", "-lock.json"}
	maxFileSize = int64(500 * 1024) // 500KB

	sigRegexes = map[string]*regexp.Regexp{
		".go":   regexp.MustCompile(`^(func|type|interface|struct)\s+(\w+)`),
		".py":   regexp.MustCompile(`^(class|def)\s+(\w+)`),
		".ts":   regexp.MustCompile(`^(export\s+)?(class|interface|function|const|let|var)\s+(\w+)`),
		".js":   regexp.MustCompile(`^(export\s+)?(class|function|const|let|var)\s+(\w+)`),
		".java": regexp.MustCompile(`^(public|protected|private|static|\s) +[\w\<\>\[\]]+\s+(\w+)`),
		".rs":   regexp.MustCompile(`^(pub\s+)?(fn|struct|enum|type|trait|impl)\s+(\w+)`),
	}
)

type Config struct {
	MaxDepth    int
	MaxTokens   int
	TotalTokens int
	Lines       bool // prefix each line with its line number
	Verbose     bool // include characters/tokens/mode in file tags
}

func main() {
	config := Config{
		MaxDepth:  -1,
		MaxTokens: 1000000,
	}

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]
	path := "."

	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--depth":
			if i+1 >= len(args) {
				break
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: --depth requires an integer, got %q\n", args[i+1])
				return
			}
			config.MaxDepth = v
			i++
		case "--max-tokens":
			if i+1 >= len(args) {
				break
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: --max-tokens requires an integer, got %q\n", args[i+1])
				return
			}
			config.MaxTokens = v
			i++
		case "--lines":
			config.Lines = true
		case "--verbose":
			config.Verbose = true
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	if len(filteredArgs) > 0 {
		if !strings.HasPrefix(filteredArgs[0], "-") {
			path = filteredArgs[0]
			filteredArgs = filteredArgs[1:]
		}
	}

	switch command {
	case "bundle":
		runBundle(path, false, &config)
	case "skeleton":
		runBundle(path, true, &config)
	case "index":
		runIndex(path)
	case "tree":
		runTree(path, false, &config)
	case "map":
		runTree(path, true, &config)
	case "search":
		query := ""
		if len(filteredArgs) > 0 {
			query = filteredArgs[0]
		} else if path != "." {
			query = path
			path = "."
		}
		if query == "" {
			fmt.Println("Usage: codecontext search <query> [path]")
			return
		}
		runSearch(query, path, &config)
	case "extract":
		target := path
		rng := ""
		if len(filteredArgs) > 0 {
			rng = filteredArgs[0]
		}
		if target == "." || target == "" {
			fmt.Println("Usage: codecontext extract <file> [range]")
			fmt.Println("Example: codecontext extract main.go 10:20")
			return
		}
		runExtract(target, rng, &config)
	case "help", "-h", "--help":
		printUsage()
	default:
		if info, err := os.Stat(command); err == nil && info.IsDir() {
			runBundle(command, false, &config)
		} else {
			printUsage()
		}
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "codecontext - Context-Efficient Repository Packing for AI Agents\n\n")
	fmt.Fprintf(os.Stderr, "USAGE:\n  codecontext <command> [path] [options]\n\n")
	fmt.Fprintf(os.Stderr, "COMMANDS:\n")
	fmt.Fprintf(os.Stderr, "  index     Generate a compact symbol map of the project. Use this FIRST to find symbols.\n")
	fmt.Fprintf(os.Stderr, "  skeleton  Pack only function/type signatures + doc-comments into XML. Use to see APIs.\n")
	fmt.Fprintf(os.Stderr, "  bundle    Pack full file contents into XML with line numbers. Use when you need the code.\n")
	fmt.Fprintf(os.Stderr, "  tree      Show directory structure. Use to understand project layout.\n")
	fmt.Fprintf(os.Stderr, "  map       Show tree view with the first 3 lines of each file. Good for quick context.\n")
	fmt.Fprintf(os.Stderr, "  search    Find files containing a query and return them as full XML blocks.\n")
	fmt.Fprintf(os.Stderr, "  extract   Extract a specific line range from a file. (e.g., extract main.go 10:20)\n\n")
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  --depth <n>      Limit recursion depth for tree/map commands.\n")
	fmt.Fprintf(os.Stderr, "  --max-tokens <n> Safety limit to truncate output if it exceeds N tokens (default 1M).\n")
	fmt.Fprintf(os.Stderr, "  --lines          Prefix each line with its line number (default: off).\n")
	fmt.Fprintf(os.Stderr, "  --verbose        Include character count, token estimate, and mode in file tags (default: off).\n\n")
	fmt.Fprintf(os.Stderr, "AGENT GUIDANCE:\n")
	fmt.Fprintf(os.Stderr, "  1. Start with 'index' or 'tree' to get a high-level overview without wasting tokens.\n")
	fmt.Fprintf(os.Stderr, "  2. Use 'skeleton' on specific directories to understand their interface/contract.\n")
	fmt.Fprintf(os.Stderr, "  3. Use 'bundle' ONLY on the specific files or folders you need to refactor or debug.\n")
	fmt.Fprintf(os.Stderr, "  4. Files are wrapped in compact <f p=\"...\"> tags. Use --verbose for full metadata.\n")
}

func getIgnoreFilter(root string) *gitignore.GitIgnore {
	gi, err := gitignore.CompileIgnoreFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return nil
	}
	return gi
}

func shouldIgnore(path string, info fs.FileInfo, gi *gitignore.GitIgnore) bool {
	if gi != nil && gi.MatchesPath(path) {
		return true
	}
	name := info.Name()
	if info.IsDir() {
		for _, d := range ignoreDirs {
			if name == d {
				return true
			}
		}
	} else {
		if info.Size() > maxFileSize {
			return true
		}
		ext := strings.ToLower(filepath.Ext(name))
		for _, e := range ignoreExts {
			if ext == e || strings.HasSuffix(name, e) {
				return true
			}
		}
	}
	return false
}

func runBundle(root string, signaturesOnly bool, config *Config) {
	gi := getIgnoreFilter(root)
	var mu sync.Mutex
	var wg sync.WaitGroup
	files := make(chan string, 100)

	// Worker pool
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range files {
				mu.Lock()
				if config.TotalTokens > config.MaxTokens {
					mu.Unlock()
					continue
				}
				mu.Unlock()

				tokensUsed := printFileXML(path, signaturesOnly, config, &mu)

				mu.Lock()
				config.TotalTokens += tokensUsed
				mu.Unlock()
			}
		}()
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		mu.Lock()
		if config.TotalTokens > config.MaxTokens {
			mu.Unlock()
			return filepath.SkipAll
		}
		mu.Unlock()

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if shouldIgnore(path, info, gi) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		files <- path
		return nil
	})
	close(files)
	wg.Wait()

	if config.TotalTokens > config.MaxTokens {
		fmt.Fprintf(os.Stderr, "\n[WARNING] Token limit reached. Output truncated.\n")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func runIndex(root string) {
	fmt.Printf("<idx p=\"%s\">\n", xmlEscape(root))
	gi := getIgnoreFilter(root)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldIgnore(path, info, gi) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldIgnore(path, info, gi) {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		sigRegex, ok := sigRegexes[ext]
		if !ok {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if sigRegex.MatchString(line) {
				fmt.Printf("[%s:%d] %s\n", path, lineNum, strings.TrimSpace(line))
			}
			lineNum++
		}
		return nil
	})
	fmt.Println("</idx>")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

var xmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"'", "&apos;",
)

func xmlEscape(s string) string {
	return xmlReplacer.Replace(s)
}

func printFileXML(path string, signaturesOnly bool, config *Config, mu *sync.Mutex) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	ext := strings.ToLower(filepath.Ext(path))
	sigRegex := sigRegexes[ext]
	scanner := bufio.NewScanner(f)
	var output strings.Builder
	var lastComment strings.Builder
	lineNum := 1
	charCount := 0
	inBody := false
	braceCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "\"\"\"")

		if signaturesOnly {
			if inBody {
				output.WriteString(formatLine(line, lineNum, config.Lines))
				braceCount += strings.Count(line, "{")
				braceCount -= strings.Count(line, "}")
				if braceCount <= 0 {
					inBody = false
				}
			} else if isComment {
				lastComment.WriteString(formatLine(line, lineNum, config.Lines))
			} else if sigRegex != nil && sigRegex.MatchString(line) {
				if lastComment.Len() > 0 {
					output.WriteString(lastComment.String())
				}
				output.WriteString(formatLine(line, lineNum, config.Lines))
				lastComment.Reset()

				// If it's a struct or interface (mostly Go specific here), capture the body
				if strings.Contains(line, "struct") || strings.Contains(line, "interface") {
					braceCount = strings.Count(line, "{") - strings.Count(line, "}")
					if braceCount > 0 {
						inBody = true
					}
				}
			} else if trimmed != "" {
				lastComment.Reset()
			}
		} else {
			output.WriteString(formatLine(line, lineNum, config.Lines))
		}
		lineNum++
		charCount += len(line) + 1
	}
	tokens := charCount / 4
	if output.Len() > 0 {
		if mu != nil {
			mu.Lock()
			defer mu.Unlock()
		}
		if config.Verbose {
			fmt.Printf("<f p=\"%s\" m=\"%s\" chars=\"%d\" tokens=\"~%d\">\n",
				xmlEscape(path), map[bool]string{true: "skeleton", false: "full"}[signaturesOnly], charCount, tokens)
		} else if signaturesOnly {
			fmt.Printf("<f p=\"%s\" m=\"skeleton\">\n", xmlEscape(path))
		} else {
			fmt.Printf("<f p=\"%s\">\n", xmlEscape(path))
		}
		fmt.Print(output.String())
		fmt.Println("</f>")
	}
	return tokens
}

func formatLine(line string, lineNum int, showLines bool) string {
	if showLines {
		return fmt.Sprintf("[%d] %s\n", lineNum, line)
	}
	return line + "\n"
}

func runExtract(path, rng string, config *Config) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	defer f.Close()

	start, end := 0, 0
	if rng != "" {
		parts := strings.Split(rng, ":")
		if len(parts) == 2 {
			start, _ = strconv.Atoi(parts[0])
			end, _ = strconv.Atoi(parts[1])
		} else {
			start, _ = strconv.Atoi(parts[0])
			end = start
		}
	}

	scanner := bufio.NewScanner(f)
	var output strings.Builder
	lineNum := 1
	for scanner.Scan() {
		if (start == 0 && end == 0) || (lineNum >= start && lineNum <= end) {
			output.WriteString(formatLine(scanner.Text(), lineNum, config.Lines))
		}
		lineNum++
	}

	if output.Len() > 0 {
		fmt.Printf("<f p=\"%s\" r=\"%s\">\n", xmlEscape(path), rng)
		fmt.Print(output.String())
		fmt.Println("</f>")
	}
}

func runSearch(query, root string, config *Config) {
	gi := getIgnoreFilter(root)
	var mu sync.Mutex
	var wg sync.WaitGroup
	files := make(chan string, 100)

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range files {
				mu.Lock()
				if config.TotalTokens > config.MaxTokens {
					mu.Unlock()
					continue
				}
				mu.Unlock()

				f, err := os.Open(path)
				if err != nil {
					continue
				}
				found := false
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), query) {
						found = true
						break
					}
				}
				f.Close()

				if found {
					tokensUsed := printFileXML(path, false, config, &mu)
					mu.Lock()
					config.TotalTokens += tokensUsed
					mu.Unlock()
				}
			}
		}()
	}

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		mu.Lock()
		if config.TotalTokens > config.MaxTokens {
			mu.Unlock()
			return filepath.SkipAll
		}
		mu.Unlock()

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if shouldIgnore(path, info, gi) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		files <- path
		return nil
	})
	close(files)
	wg.Wait()
}

func runTree(root string, showHeaders bool, config *Config) {
	fmt.Printf("%s/\n", root)
	gi := getIgnoreFilter(root)
	walkTree(root, "", 0, showHeaders, config, gi)
}

func walkTree(path string, indent string, depth int, showHeaders bool, config *Config, gi *gitignore.GitIgnore) {
	if config.MaxDepth != -1 && depth >= config.MaxDepth {
		return
	}
	entries, _ := os.ReadDir(path)
	var validEntries []fs.DirEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if shouldIgnore(filepath.Join(path, entry.Name()), info, gi) {
			continue
		}
		validEntries = append(validEntries, entry)
	}
	for i, entry := range validEntries {
		isLast := i == len(validEntries)-1
		prefix := "├── "
		if isLast {
			prefix = "└── "
		}
		fmt.Printf("%s%s%s\n", indent, prefix, entry.Name())
		if entry.IsDir() {
			newIndent := indent + "│   "
			if isLast {
				newIndent = indent + "    "
			}
			walkTree(filepath.Join(path, entry.Name()), newIndent, depth+1, showHeaders, config, gi)
		} else if showHeaders {
			printHeader(filepath.Join(path, entry.Name()), indent)
		}
	}
}

func printHeader(path string, indent string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() && count < 3 {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			fmt.Printf("%s    %s\n", indent, line)
			count++
		}
	}
}
