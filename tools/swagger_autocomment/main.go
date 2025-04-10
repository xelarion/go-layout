// Package main provides a tool to generate intelligent Swagger comments for handler methods.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// Config holds configuration options for the Swagger comment generator.
type Config struct {
	HandlerDir     string   // Directory containing handler files
	OutputDir      string   // Output directory for generated files
	SecurityScheme string   // Security scheme for Swagger docs
	RouterFile     string   // Router file to extract routes from
	TypesPaths     []string // Paths to search for type definitions
	HandlerPattern string   // Pattern to match handler files
	Verbose        bool     // Whether to output verbose logs
	Concurrency    int      // Number of concurrent workers
	ApiPrefix      string   // API prefix for paths
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	// Use CPU core count for concurrency, with minimum of 2
	cpuCount := runtime.NumCPU()
	if cpuCount < 2 {
		cpuCount = 2
	}

	return &Config{
		HandlerDir:     "./internal/api/http/web/handler",
		SecurityScheme: "BearerAuth",
		RouterFile:     "./internal/api/http/web/router.go",
		TypesPaths:     []string{"./internal/api/http/web/types/*.go", "./pkg/types/*.go"},
		HandlerPattern: "*_handler.go",
		Concurrency:    cpuCount,
		ApiPrefix:      "",
		Verbose:        true,
	}
}

// SetupFlags sets up command line flags for the configuration
func (cfg *Config) SetupFlags() {
	// Save original usage function
	originalUsage := flag.Usage

	// Set up flag variables with more descriptive names
	flag.StringVar(&cfg.HandlerDir, "handlers", cfg.HandlerDir, "Directory containing handler files")
	flag.StringVar(&cfg.OutputDir, "output", cfg.OutputDir, "Output directory (default: same as handler directory)")
	flag.StringVar(&cfg.SecurityScheme, "security", cfg.SecurityScheme, "Security scheme name")
	flag.StringVar(&cfg.RouterFile, "router", cfg.RouterFile, "Router file path")
	flag.StringVar(&cfg.HandlerPattern, "handler-pattern", cfg.HandlerPattern, "Pattern to match handler files")
	flag.StringVar(&cfg.ApiPrefix, "api-prefix", cfg.ApiPrefix, "API prefix for paths")
	flag.IntVar(&cfg.Concurrency, "concurrency", cfg.Concurrency, "Number of concurrent workers")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Enable verbose output")

	// Add quiet flag that sets verbose to false when used
	silent := flag.Bool("silent", false, "Silent mode (no output except errors)")

	// Allow comma-separated list of types paths
	var typesPaths string
	defaultTypesPaths := strings.Join(cfg.TypesPaths, ",")
	flag.StringVar(&typesPaths, "types", defaultTypesPaths, "Comma-separated list of glob patterns for type definition files")

	// Add help flags
	help := flag.Bool("help", false, "Print detailed usage information")
	h := flag.Bool("h", false, "Print detailed usage information")

	// Override default usage to include examples and detailed explanation
	flag.Usage = func() {
		// First, show the standard Go flag help
		originalUsage()

		// Then, add our custom examples and additional information
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Automatically generates intelligent Swagger comments for handler methods.")
		fmt.Println("  Extracts parameter types from request structs when available.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Use default configuration")
		fmt.Println("  go run tools/swagger_autocomment/main.go")
		fmt.Println()
		fmt.Println("  # Process custom directory")
		fmt.Println("  go run tools/swagger_autocomment/main.go -handlers ./internal/api/http/wx-api/handler")
		fmt.Println()
		fmt.Println("  # Process custom API with custom type directories")
		fmt.Println("  go run tools/swagger_autocomment/main.go \\")
		fmt.Println("    -handlers ./internal/api/http/wx-api/handler \\")
		fmt.Println("    -router ./internal/api/http/wx-api/router.go \\")
		fmt.Println("    -types \"./internal/api/http/wx-api/types/*.go,./pkg/types/*.go\" \\")
		fmt.Println("    -api-prefix \"/wx-api\" \\")
		fmt.Println("    -concurrency 8")
		fmt.Println()
		fmt.Println("  # Run in silent mode (minimal output)")
		fmt.Println("  go run tools/swagger_autocomment/main.go -silent")
		fmt.Println()
		fmt.Println("Makefile Integration:")
		fmt.Println("  # Add to your Makefile:")
		fmt.Println("  swagger-comment:")
		fmt.Println("  \tgo run tools/swagger_autocomment/main.go $(ARGS)")
		fmt.Println()
		fmt.Println("  swagger-wx-comment:")
		fmt.Println("  \tgo run tools/swagger_autocomment/main.go \\")
		fmt.Println("  \t  -handlers ./internal/api/http/wx-api/handler \\")
		fmt.Println("  \t  -router ./internal/api/http/wx-api/router.go \\")
		fmt.Println("  \t  -types \"./internal/api/http/wx-api/types/*.go,./pkg/types/*.go\" \\")
		fmt.Println("  \t  -api-prefix \"/wx-api\" $(ARGS)")
		fmt.Println()
		fmt.Println("  # To run in silent mode:")
		fmt.Println("  make swagger-comment ARGS=\"-silent\"")
	}

	// Parse the flags
	flag.Parse()

	// Check if help was requested
	if *help || *h {
		// Call the now-enhanced usage function
		flag.Usage()
		os.Exit(0)
	}

	// Process types paths after flags are parsed
	if typesPaths != "" {
		cfg.TypesPaths = strings.Split(typesPaths, ",")
	}

	// Use handler directory as output directory if not specified
	if cfg.OutputDir == "" {
		cfg.OutputDir = cfg.HandlerDir
	}

	// If quiet mode is set, it overrides verbose
	if *silent {
		cfg.Verbose = false
	}
}

// RouteInfo stores information about an API route
type RouteInfo struct {
	Path      string // API path with parameter placeholders
	Method    string // HTTP method (lowercase)
	Handler   string // Handler function name
	IsSecured bool   // Whether the route requires authentication
}

// Parameter represents a parameter for a handler function
type Parameter struct {
	Name     string // Parameter name
	Type     string // Parameter type
	Location string // Parameter location (path, query, body)
	Required bool   // Whether the parameter is required
}

// FileStats holds statistics for processed files
type FileStats struct {
	TotalMethods     int
	HandlerMethods   int
	AlreadyCommented int
	NewlyCommented   int
}

// URIParameter stores information about URI parameters from request structs
type URIParameter struct {
	Name     string
	Type     string
	Required bool
}

var (
	// Regular expressions - compiled once for performance
	pathParamRegex   = regexp.MustCompile(`\{([^}]+)\}`)
	camelCaseRegex   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	apiRouteRegex    = regexp.MustCompile(`(api|authorized)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)", ([a-zA-Z0-9]+)\.([a-zA-Z0-9]+)\)`)
	routerParamRegex = regexp.MustCompile(`:([^/]+)`)

	// Slice of common prefixes for path determination
	pathPrefixes = []string{"get-", "list-", "create-", "update-", "delete-", "handle-"}

	// Common patterns for public endpoints
	publicPatterns = []string{"Login", "Register", "Captcha", "Public"}
)

// FileCache provides caching for parsed files to avoid repeated parsing
type FileCache struct {
	parsedFiles map[string]*ast.File
	mutex       sync.RWMutex
}

// NewFileCache creates a new file cache
func NewFileCache() *FileCache {
	return &FileCache{
		parsedFiles: make(map[string]*ast.File),
	}
}

// GetParsedFile gets a parsed file from cache or parses it if not cached
func (c *FileCache) GetParsedFile(filePath string) (*ast.File, error) {
	// Check cache first
	c.mutex.RLock()
	if file, ok := c.parsedFiles[filePath]; ok {
		c.mutex.RUnlock()
		return file, nil
	}
	c.mutex.RUnlock()

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}

	// Cache the result
	c.mutex.Lock()
	c.parsedFiles[filePath] = node
	c.mutex.Unlock()

	return node, nil
}

// TypeFinder provides an interface for finding and caching type information
type TypeFinder interface {
	// FindType finds a type by name and returns information about it
	FindType(typeName string) (map[string]URIParameter, error)
}

// CachedTypeFinder caches type information to avoid repeated parsing
type CachedTypeFinder struct {
	typePaths  []string
	cache      map[string]map[string]URIParameter
	fileCache  *FileCache
	cacheMutex sync.RWMutex
}

// NewCachedTypeFinder creates a new type finder with caching
func NewCachedTypeFinder(typePaths []string, fileCache *FileCache) *CachedTypeFinder {
	return &CachedTypeFinder{
		typePaths: typePaths,
		cache:     make(map[string]map[string]URIParameter),
		fileCache: fileCache,
	}
}

// FindType finds a type by name and caches the result
func (tf *CachedTypeFinder) FindType(typeName string) (map[string]URIParameter, error) {
	// Check cache first
	tf.cacheMutex.RLock()
	params, found := tf.cache[typeName]
	tf.cacheMutex.RUnlock()

	if found {
		return params, nil
	}

	// Not in cache, find it
	params, err := tf.extractURIParametersFromType(typeName)
	if err != nil {
		return params, err
	}

	// Cache the result
	tf.cacheMutex.Lock()
	tf.cache[typeName] = params
	tf.cacheMutex.Unlock()

	return params, nil
}

// extractURIParametersFromType extracts URI parameters from request struct types
func (tf *CachedTypeFinder) extractURIParametersFromType(typeName string) (map[string]URIParameter, error) {
	uriParams := make(map[string]URIParameter)

	// Find files with types package in the codebase
	var typesFiles []string
	for _, pattern := range tf.typePaths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		typesFiles = append(typesFiles, matches...)
	}

	// Get the struct name without package prefix (e.g., "UserReq" from "types.UserReq")
	parts := strings.Split(typeName, ".")
	if len(parts) != 2 {
		return uriParams, nil
	}
	structName := parts[1]

	for _, filePath := range typesFiles {
		// Get the parsed file from cache
		node, err := tf.fileCache.GetParsedFile(filePath)
		if err != nil {
			continue
		}

		// Find the struct declaration
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != structName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Examine each field for URI tag
				for _, field := range structType.Fields.List {
					if field.Tag == nil {
						continue
					}

					tag := field.Tag.Value

					// Extract URI parameter name from the tag
					uriMatch := regexp.MustCompile(`uri:"([^"]+)"`).FindStringSubmatch(tag)
					if len(uriMatch) < 2 {
						continue
					}

					uriParamName := uriMatch[1]
					required := regexp.MustCompile(`binding:"[^"]*required[^"]*"`).MatchString(tag)

					// Determine field type
					paramType := "string" // default
					if ident, ok := field.Type.(*ast.Ident); ok {
						switch ident.Name {
						case "int", "int32", "int64", "uint", "uint32", "uint64":
							paramType = "integer"
						case "float32", "float64":
							paramType = "number"
						case "bool":
							paramType = "boolean"
						}
					}

					// Store parameter info
					uriParams[uriParamName] = URIParameter{
						Name:     uriParamName,
						Type:     paramType,
						Required: required,
					}
				}
			}
		}
	}

	return uriParams, nil
}

// SwaggerGenerator handles the generation of Swagger comments
type SwaggerGenerator struct {
	config     *Config
	routes     map[string]RouteInfo
	fileCache  *FileCache
	typeFinder TypeFinder
}

// NewSwaggerGenerator creates a new Swagger generator
func NewSwaggerGenerator(config *Config) (*SwaggerGenerator, error) {
	// Create file cache
	fileCache := NewFileCache()

	// Extract routes from router file
	routes, err := extractRoutes(config.RouterFile)
	if err != nil && config.Verbose {
		fmt.Printf("Warning: Could not extract routes from router file: %v\n", err)
		fmt.Println("Will use function name analysis to determine routes.")
	}

	// Create type finder
	typeFinder := NewCachedTypeFinder(config.TypesPaths, fileCache)

	return &SwaggerGenerator{
		config:     config,
		routes:     routes,
		fileCache:  fileCache,
		typeFinder: typeFinder,
	}, nil
}

// Run runs the swagger generation process
func (sg *SwaggerGenerator) Run() (FileStats, error) {
	totalStats := FileStats{}

	// Print configuration if verbose
	if sg.config.Verbose {
		printConfig(sg.config, len(sg.routes))
	}

	// Find all handler files
	handlerFiles, err := findHandlerFiles(sg.config.HandlerDir, sg.config.HandlerPattern)
	if err != nil {
		return totalStats, fmt.Errorf("error finding handler files: %v", err)
	}

	if len(handlerFiles) == 0 {
		return totalStats, fmt.Errorf("no handler files found in %s matching pattern %s",
			sg.config.HandlerDir, sg.config.HandlerPattern)
	}

	if sg.config.Verbose {
		fmt.Printf("Found %d handler files to process\n", len(handlerFiles))
	}

	// Create and start worker pool
	workerPool := NewWorkerPool(sg.config, sg.routes, sg.fileCache, sg.typeFinder)
	workerPool.Start()

	// Add files to work queue
	for _, path := range handlerFiles {
		workerPool.AddFile(path)
	}

	// Close worker pool and process results
	workerPool.Close()
	totalStats = workerPool.ProcessResults()

	return totalStats, nil
}

// WorkerPool manages a pool of workers for processing files
type WorkerPool struct {
	concurrency int
	workChan    chan string
	resultChan  chan fileResult
	wg          *sync.WaitGroup
	config      *Config
	routes      map[string]RouteInfo
	fileCache   *FileCache
	typeFinder  TypeFinder
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config *Config, routes map[string]RouteInfo, fileCache *FileCache, typeFinder TypeFinder) *WorkerPool {
	return &WorkerPool{
		concurrency: config.Concurrency,
		workChan:    make(chan string, config.Concurrency*2),
		resultChan:  make(chan fileResult, config.Concurrency*2),
		wg:          &sync.WaitGroup{},
		config:      config,
		routes:      routes,
		fileCache:   fileCache,
		typeFinder:  typeFinder,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	// Start workers
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker processes files from the work channel
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for filePath := range wp.workChan {
		stats, err := processFile(filePath, *wp.config, wp.routes, wp.fileCache, wp.typeFinder)
		wp.resultChan <- fileResult{path: filePath, stats: stats, err: err}
	}
}

// AddFile adds a file to the work queue
func (wp *WorkerPool) AddFile(filePath string) {
	wp.workChan <- filePath
}

// Close closes the work channel and waits for all workers to finish
func (wp *WorkerPool) Close() {
	close(wp.workChan)
	wp.wg.Wait()
	close(wp.resultChan)
}

// ProcessResults processes results from the result channel and returns total stats
func (wp *WorkerPool) ProcessResults() FileStats {
	totalStats := FileStats{}
	successCount := 0
	errorCount := 0

	if wp.config.Verbose {
		fmt.Printf("Processing files with %d workers...\n", wp.concurrency)
		fmt.Println()
	}

	for result := range wp.resultChan {
		if result.err != nil {
			errorCount++
			fmt.Printf("Error processing %s: %v\n", result.path, result.err)
			continue
		}

		successCount++
		totalStats.TotalMethods += result.stats.TotalMethods
		totalStats.HandlerMethods += result.stats.HandlerMethods
		totalStats.AlreadyCommented += result.stats.AlreadyCommented
		totalStats.NewlyCommented += result.stats.NewlyCommented

		if wp.config.Verbose {
			printFileStats(result.path, result.stats)
			fmt.Printf("Progress: %d files processed (%d successful, %d failed)\n",
				successCount+errorCount, successCount, errorCount)
			fmt.Println()
		}
	}

	if errorCount > 0 {
		fmt.Printf("Completed with %d errors and %d successful files\n", errorCount, successCount)
	}

	return totalStats
}

func main() {
	// Initialize default configuration
	config := NewDefaultConfig()

	// Parse command line flags
	config.SetupFlags()

	// Validate handler directory
	if !dirExists(config.HandlerDir) {
		fmt.Printf("Error: Directory %s does not exist!\n", config.HandlerDir)
		fmt.Println("Current working directory:", getCwd())
		fmt.Println("Please provide a valid directory using the -handlers flag")
		os.Exit(1)
	}

	// Create Swagger generator
	generator, err := NewSwaggerGenerator(config)
	if err != nil {
		fmt.Printf("Error creating Swagger generator: %v\n", err)
		os.Exit(1)
	}

	// Run the generator
	totalStats, err := generator.Run()
	if err != nil {
		fmt.Printf("Error running Swagger generator: %v\n", err)
		os.Exit(1)
	}

	// Calculate overall percentage
	var totalCoverage float64
	if totalStats.HandlerMethods > 0 {
		totalCoverage = float64(totalStats.AlreadyCommented+totalStats.NewlyCommented) / float64(totalStats.HandlerMethods) * 100
	}

	fmt.Println("======================================")
	fmt.Println("Swagger Comment Generation Complete!")
	fmt.Println("======================================")
	fmt.Printf("Total methods examined:  %d\n", totalStats.TotalMethods)
	fmt.Printf("Handler methods found:   %d\n", totalStats.HandlerMethods)
	fmt.Printf("Already commented:       %d\n", totalStats.AlreadyCommented)
	fmt.Printf("New comments added:      %d\n", totalStats.NewlyCommented)
	fmt.Printf("Total documentation:     %d/%d (%.1f%%)\n",
		totalStats.AlreadyCommented+totalStats.NewlyCommented,
		totalStats.HandlerMethods,
		totalCoverage)
}

// fileResult holds the result of processing a file
type fileResult struct {
	path  string
	stats FileStats
	err   error
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// getCwd returns the current working directory
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}

// printConfig prints the configuration
func printConfig(config *Config, routesCount int) {
	fmt.Println("Configuration:")
	fmt.Printf("  Handler directory: %s\n", config.HandlerDir)
	fmt.Printf("  Output directory: %s\n", config.OutputDir)
	fmt.Printf("  Router file: %s\n", config.RouterFile)
	fmt.Printf("  Handler pattern: %s\n", config.HandlerPattern)
	fmt.Printf("  API prefix: %s\n", config.ApiPrefix)
	fmt.Printf("  Security scheme: %s\n", config.SecurityScheme)
	fmt.Printf("  Concurrency: %d\n", config.Concurrency)
	fmt.Printf("  Type paths: %s\n", strings.Join(config.TypesPaths, ", "))
	fmt.Printf("  Extracted %d routes from router file\n", routesCount)
	fmt.Println()
}

// findHandlerFiles finds all handler files in the given directory
func findHandlerFiles(dirPath string, pattern string) ([]string, error) {
	var handlerFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if file matches pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if !info.IsDir() && matched {
			handlerFiles = append(handlerFiles, path)
		}
		return nil
	})

	return handlerFiles, err
}

// printFileStats prints statistics for a processed file
func printFileStats(filePath string, stats FileStats) {
	filename := filepath.Base(filePath)
	fmt.Printf("File: %s\n", filename)
	fmt.Printf("  Total methods examined: %d\n", stats.TotalMethods)
	fmt.Printf("  Handler methods found: %d\n", stats.HandlerMethods)
	fmt.Printf("  Already commented: %d\n", stats.AlreadyCommented)
	fmt.Printf("  New comments added: %d\n", stats.NewlyCommented)

	// Calculate percentage
	var commentedPercent float64
	if stats.HandlerMethods > 0 {
		commentedPercent = float64(stats.AlreadyCommented+stats.NewlyCommented) / float64(stats.HandlerMethods) * 100
	}

	fmt.Printf("  Total documentation coverage: %.1f%%\n", commentedPercent)
}

// extractRoutes extracts routes from the router file
func extractRoutes(routerFile string) (map[string]RouteInfo, error) {
	routes := make(map[string]RouteInfo)

	// Read the router file
	content, err := os.ReadFile(routerFile)
	if err != nil {
		return routes, fmt.Errorf("error reading router file: %v", err)
	}

	// Extract all route registrations
	matches := apiRouteRegex.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		group := match[1]   // api or authorized
		method := match[2]  // HTTP method
		path := match[3]    // Path
		handler := match[5] // Handler function name

		// Convert path params from :id to {id}
		convertedPath := routerParamRegex.ReplaceAllString(path, "{$1}")

		routes[handler] = RouteInfo{
			Path:      convertedPath,
			Method:    strings.ToLower(method),
			Handler:   handler,
			IsSecured: group == "authorized",
		}
	}

	return routes, nil
}

// processFile processes a single file to add Swagger comments
func processFile(filePath string, config Config, routes map[string]RouteInfo, fileCache *FileCache, typeFinder TypeFinder) (FileStats, error) {
	stats := FileStats{}

	// Get parsed file from cache
	node, err := fileCache.GetParsedFile(filePath)
	if err != nil {
		return stats, fmt.Errorf("error getting parsed file: %v", err)
	}

	// Collect all handler methods that need comments
	handlerFuncs := collectHandlerFunctions(node, &stats, config.Verbose)

	// If no functions to update, return early
	if len(handlerFuncs) == 0 {
		return stats, nil
	}

	// Process functions in reverse order to avoid position changes
	for i := len(handlerFuncs) - 1; i >= 0; i-- {
		funcDecl := handlerFuncs[i]
		handlerName := funcDecl.Name.Name

		// Generate swagger comment
		comment := generateSwaggerComment(funcDecl, &config, routes, typeFinder)
		if config.Verbose {
			fmt.Printf("  Generated comment for %s\n", handlerName)
		}

		// Update the file with the new comment
		if err := updateFileWithComment(filePath, funcDecl, comment); err != nil {
			fmt.Printf("  Error updating comment for %s: %v\n", handlerName, err)
			continue
		}

		stats.NewlyCommented++
	}

	return stats, nil
}

// collectHandlerFunctions collects all handler functions that need Swagger comments
func collectHandlerFunctions(node *ast.File, stats *FileStats, verbose bool) []*ast.FuncDecl {
	var handlerFuncs []*ast.FuncDecl

	// Iterate through all declarations in the file
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		stats.TotalMethods++

		// Check if it's a handler method
		if !isHandlerMethod(funcDecl) {
			continue
		}

		stats.HandlerMethods++
		handlerName := funcDecl.Name.Name

		// Skip if already has swagger comment
		if hasSwaggerComment(funcDecl) {
			stats.AlreadyCommented++
			if verbose {
				fmt.Printf("  Already has swagger comment: %s\n", handlerName)
			}
			continue
		}

		// Add to processing list
		handlerFuncs = append(handlerFuncs, funcDecl)
	}

	return handlerFuncs
}

// isHandlerMethod checks if a function declaration is a handler method
func isHandlerMethod(funcDecl *ast.FuncDecl) bool {
	// Must have a receiver
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}

	// Receiver name must contain "Handler"
	receiverType := getReceiverType(funcDecl)
	if !strings.Contains(receiverType, "Handler") {
		return false
	}

	// Must have at least one parameter
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) < 1 {
		return false
	}

	// Must have a context parameter
	return hasContextParameter(funcDecl)
}

// getReceiverType extracts the receiver type name from a function
func getReceiverType(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return ""
	}

	// Handle pointer receiver (*UserHandler)
	if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
		if ident, ok := starExpr.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	// Handle value receiver (UserHandler)
	if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
		return ident.Name
	}

	return ""
}

// hasContextParameter checks if a function has a context-like parameter
func hasContextParameter(funcDecl *ast.FuncDecl) bool {
	for _, param := range funcDecl.Type.Params.List {
		// Look for *xxx.Context parameter
		if starExpr, ok := param.Type.(*ast.StarExpr); ok {
			if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name == "Context" {
					return true
				}
			}
		}
	}

	return false
}

// hasSwaggerComment checks if a function already has Swagger comments
func hasSwaggerComment(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Doc == nil {
		return false
	}

	for _, comment := range funcDecl.Doc.List {
		if strings.Contains(comment.Text, "@Router") {
			return true
		}
	}

	return false
}

// generateSwaggerComment generates a Swagger comment for a handler method
func generateSwaggerComment(funcDecl *ast.FuncDecl, config *Config, routes map[string]RouteInfo, typeFinder TypeFinder) string {
	handlerName := funcDecl.Name.Name
	receiverType := getReceiverType(funcDecl)

	// Get route info from router file if available
	routeInfo, exists := routes[handlerName]

	// Determine HTTP method and path
	var method, path string
	var isSecured bool

	if exists {
		method = routeInfo.Method
		path = routeInfo.Path
		isSecured = routeInfo.IsSecured
	} else {
		// Fallback to function name analysis
		method = determineHTTPMethod(handlerName)
		path = determinePath(handlerName)
	}

	// Add API prefix if configured
	if config.ApiPrefix != "" && !strings.HasPrefix(path, config.ApiPrefix) {
		path = config.ApiPrefix + path
	}

	// Build the comment - use a pre-sized StringBuilder for better performance
	var comment strings.Builder
	comment.Grow(500) // Pre-allocate for typical comment size

	// Add basic info
	comment.WriteString(fmt.Sprintf("// %s godoc\n", handlerName))
	comment.WriteString(fmt.Sprintf("// @Summary %s\n", generateSummary(handlerName)))
	comment.WriteString(fmt.Sprintf("// @Description %s\n", generateDescription(handlerName)))
	comment.WriteString("// @Tags " + determineTagFromHandler(receiverType) + "\n")
	comment.WriteString("// @Accept json\n")
	comment.WriteString("// @Produce json\n")

	// Add parameters
	addParametersToComment(&comment, handlerName, path, method, typeFinder)

	// Add response type
	statusCode := determineStatusCode(method)
	comment.WriteString(fmt.Sprintf("// @Success %d {object} types.Response{data=types.%sResp} \"Success\"\n",
		statusCode, handlerName))

	// Add error responses
	comment.WriteString("// @Failure 400 {object} types.Response \"Bad request\"\n")
	comment.WriteString("// @Failure 401 {object} types.Response \"Unauthorized\"\n")
	comment.WriteString("// @Failure 500 {object} types.Response \"Internal server error\"\n")

	// Add security if applicable
	if isSecured || (!exists && isLikelySecured(handlerName)) {
		comment.WriteString(fmt.Sprintf("// @Security %s\n", config.SecurityScheme))
	}

	// Add router information
	comment.WriteString(fmt.Sprintf("// @Router %s [%s]\n", path, method))

	return comment.String()
}

// addParametersToComment adds parameter information to the comment
func addParametersToComment(comment *strings.Builder, handlerName, path, method string, typeFinder TypeFinder) {
	// Try to extract URI parameters from request struct
	reqTypeName := "types." + handlerName + "Req"
	uriParams, _ := typeFinder.FindType(reqTypeName)

	// Get path parameters
	pathParams := pathParamRegex.FindAllStringSubmatch(path, -1)
	for _, match := range pathParams {
		if len(match) >= 2 {
			paramName := match[1]
			paramType := "string" // default type
			required := "true"    // default required

			// Check if we have type info from the request struct
			if param, found := uriParams[paramName]; found {
				paramType = param.Type
				if param.Required {
					required = "true"
				} else {
					required = "false"
				}
			} else if paramName == "id" || strings.HasSuffix(paramName, "_id") {
				// Fallback: Check if parameter is "id" or ends with "_id"
				paramType = "integer"
			}

			comment.WriteString(fmt.Sprintf("// @Param %s path %s %s \"%s\"\n",
				paramName, paramType, required, paramName))
		}
	}

	// Add request parameter
	reqType := "types." + handlerName + "Req"
	location := "body"
	required := "true"

	if method == "get" {
		location = "query"
		required = "false"
	}

	comment.WriteString(fmt.Sprintf("// @Param req %s %s %s \"req\"\n",
		location, reqType, required))
}

// updateFileWithComment updates a file with a new Swagger comment
func updateFileWithComment(filePath string, funcDecl *ast.FuncDecl, comment string) error {
	// Create a new file set for this operation
	fset := token.NewFileSet()

	// Parse the file to get positions
	_, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	// Get function position
	startPos := funcDecl.Pos()
	startOffset := fset.Position(startPos).Offset

	// Create a temporary file
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "swagger-comment-*.go")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Open the source file
	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer src.Close()

	writer := bufio.NewWriter(tempFile)

	// Determine where to insert comment
	docStartOffset := startOffset
	if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
		docPos := funcDecl.Doc.Pos()
		docStartOffset = fset.Position(docPos).Offset
	}

	// Copy from start to doc position
	if _, err = io.CopyN(writer, src, int64(docStartOffset)); err != nil {
		return fmt.Errorf("error copying to temp file: %v", err)
	}

	// Write the new comment
	if _, err = writer.WriteString(comment); err != nil {
		return fmt.Errorf("error writing comment: %v", err)
	}

	// Skip the old doc if it exists
	if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
		if _, err = src.Seek(int64(startOffset), io.SeekStart); err != nil {
			return fmt.Errorf("error seeking in file: %v", err)
		}
	}

	// Copy the rest of the file
	if _, err = io.Copy(writer, src); err != nil {
		return fmt.Errorf("error copying rest of file: %v", err)
	}

	// Flush the writer
	if err = writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %v", err)
	}

	// Close the temp file
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	// Replace the original file with the temp file
	if err = os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("error replacing original file: %v", err)
	}

	return nil
}

// determineHTTPMethod determines the HTTP method based on the function name
func determineHTTPMethod(handlerName string) string {
	switch {
	case strings.HasPrefix(handlerName, "Create") || strings.HasPrefix(handlerName, "Add"):
		return "post"
	case strings.HasPrefix(handlerName, "Update") || strings.HasPrefix(handlerName, "Modify"):
		return "put"
	case strings.HasPrefix(handlerName, "Patch") || strings.HasPrefix(handlerName, "Partial"):
		return "patch"
	case strings.HasPrefix(handlerName, "Delete") || strings.HasPrefix(handlerName, "Remove"):
		return "delete"
	default:
		return "get" // Default to GET for Get/List/Find methods
	}
}

// determinePath determines the API path based on the function name
func determinePath(handlerName string) string {
	// Convert camel case to kebab case
	path := camelCaseRegex.ReplaceAllString(handlerName, "$1-$2")
	path = strings.ToLower(path)

	// Remove common prefixes
	for _, prefix := range pathPrefixes {
		path = strings.TrimPrefix(path, prefix)
	}

	// Add ID parameter for single-item operations
	if strings.Contains(handlerName, "ById") ||
		(strings.HasPrefix(handlerName, "Get") && !strings.HasPrefix(handlerName, "List")) ||
		strings.HasPrefix(handlerName, "Update") ||
		strings.HasPrefix(handlerName, "Delete") {

		// Extract resource name
		resourceName := path
		if strings.Contains(resourceName, "-by-id") {
			resourceName = strings.Split(resourceName, "-by-id")[0]
		}
		return "/" + resourceName + "/{id}"
	}

	// Special case for List operations
	if strings.HasPrefix(handlerName, "List") {
		resourceName := strings.TrimPrefix(path, "list-")
		return "/" + resourceName
	}

	return "/" + path
}

// generateSummary generates a summary for a handler method
func generateSummary(handlerName string) string {
	// Convert camel case to space-separated words
	return camelCaseRegex.ReplaceAllString(handlerName, "$1 $2")
}

// generateDescription generates a description for a handler method
func generateDescription(handlerName string) string {
	// Convert camel case to space-separated words
	desc := camelCaseRegex.ReplaceAllString(handlerName, "$1 $2")

	// Add appropriate prefix based on handler type
	var prefix string
	switch {
	case strings.HasPrefix(handlerName, "Create"):
		prefix = "Creates a new "
	case strings.HasPrefix(handlerName, "Get") && !strings.HasPrefix(handlerName, "List"):
		prefix = "Retrieves a single "
	case strings.HasPrefix(handlerName, "List"):
		prefix = "Retrieves a list of "
	case strings.HasPrefix(handlerName, "Update"):
		prefix = "Updates an existing "
	case strings.HasPrefix(handlerName, "Delete"):
		prefix = "Deletes an existing "
	default:
		return desc
	}

	resourceName := strings.TrimPrefix(desc, strings.Split(desc, " ")[0]+" ")
	return prefix + resourceName
}

// determineTagFromHandler determines the Swagger tag from a handler name
func determineTagFromHandler(handlerName string) string {
	// Remove Handler suffix and convert to lowercase
	return strings.ToLower(strings.TrimSuffix(handlerName, "Handler"))
}

// determineStatusCode determines the HTTP status code based on the method
func determineStatusCode(method string) int {
	switch method {
	case "post":
		return 201 // Created
	case "delete":
		return 204 // No Content
	default:
		return 200 // OK
	}
}

// isLikelySecured determines if a handler is likely secured based on its name
func isLikelySecured(handlerName string) bool {
	for _, pattern := range publicPatterns {
		if strings.Contains(handlerName, pattern) {
			return false
		}
	}

	return true
}
