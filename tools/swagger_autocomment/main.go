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
	"strings"
	"sync"
)

// Config holds configuration options for the Swagger comment generator.
type Config struct {
	HandlerDir     string // Directory containing handler files
	OutputDir      string // Output directory for generated files
	SecurityScheme string // Security scheme for Swagger docs
	RouterFile     string // Router file to extract routes from
	Verbose        bool   // Whether to output verbose logs
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

func main() {
	// Initialize default configuration
	config := Config{
		HandlerDir:     "./internal/api/http/web/handler",
		SecurityScheme: "BearerAuth",
		RouterFile:     "./internal/api/http/web/router.go",
		Verbose:        true,
	}

	// Parse command line flags
	flag.StringVar(&config.HandlerDir, "dir", config.HandlerDir, "Directory containing handler files")
	flag.StringVar(&config.OutputDir, "out", "", "Output directory (default: same as handler directory)")
	flag.StringVar(&config.SecurityScheme, "security", config.SecurityScheme, "Security scheme name")
	flag.StringVar(&config.RouterFile, "router", config.RouterFile, "Router file path")
	flag.BoolVar(&config.Verbose, "v", config.Verbose, "Verbose output")

	help := flag.Bool("help", false, "Print usage information")
	h := flag.Bool("h", false, "Print usage information")

	flag.Parse()

	if *help || *h {
		printUsage()
		return
	}

	// Use handler directory as output directory if not specified
	if config.OutputDir == "" {
		config.OutputDir = config.HandlerDir
	}

	// Validate handler directory
	if !dirExists(config.HandlerDir) {
		fmt.Printf("Error: Directory %s does not exist!\n", config.HandlerDir)
		fmt.Println("Current working directory:", getCwd())
		fmt.Println("Please provide a valid directory using the -dir flag")
		os.Exit(1)
	}

	// Extract routes from router file
	routes, err := extractRoutes(config.RouterFile)
	if err != nil && config.Verbose {
		fmt.Printf("Warning: Could not extract routes from router file: %v\n", err)
		fmt.Println("Will use function name analysis to determine routes.")
	}

	// Print configuration if verbose
	if config.Verbose {
		printConfig(config, len(routes))
	}

	// Find all handler files
	handlerFiles, err := findHandlerFiles(config.HandlerDir)
	if err != nil {
		fmt.Printf("Error finding handler files: %v\n", err)
		os.Exit(1)
	}

	if len(handlerFiles) == 0 {
		fmt.Println("No handler files found in", config.HandlerDir)
		fmt.Println("Make sure the directory contains *_handler.go files")
		os.Exit(1)
	}

	if config.Verbose {
		fmt.Printf("Found %d handler files to process\n", len(handlerFiles))
	}

	// Process files
	var wg sync.WaitGroup
	results := make(chan fileResult, len(handlerFiles))

	// Process each handler file - limited concurrency
	semaphore := make(chan struct{}, 4) // Limit to 4 concurrent file operations
	for _, path := range handlerFiles {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(path string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if config.Verbose {
				fmt.Printf("Processing file: %s\n", path)
			}

			stats, err := processFile(path, config, routes)
			results <- fileResult{path: path, stats: stats, err: err}
		}(path)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and process results
	totalStats := FileStats{}
	for result := range results {
		if result.err != nil {
			fmt.Printf("Error processing %s: %v\n", result.path, result.err)
			continue
		}

		totalStats.TotalMethods += result.stats.TotalMethods
		totalStats.HandlerMethods += result.stats.HandlerMethods
		totalStats.AlreadyCommented += result.stats.AlreadyCommented
		totalStats.NewlyCommented += result.stats.NewlyCommented

		if config.Verbose {
			printFileStats(result.path, result.stats)
		}
	}

	fmt.Println("Swagger comment generation complete!")
	fmt.Printf("Total methods: %d, Handler methods: %d, Already commented: %d, Newly commented: %d\n",
		totalStats.TotalMethods, totalStats.HandlerMethods,
		totalStats.AlreadyCommented, totalStats.NewlyCommented)
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
func printConfig(config Config, routesCount int) {
	fmt.Printf("Processing handlers in %s\n", config.HandlerDir)
	fmt.Printf("Output directory: %s\n", config.OutputDir)
	fmt.Printf("Security Scheme: %s\n", config.SecurityScheme)
	fmt.Printf("Router file: %s\n", config.RouterFile)
	fmt.Printf("Verbose mode: %v\n", config.Verbose)
	fmt.Printf("Extracted %d routes from router file\n", routesCount)
}

// printUsage prints usage information
func printUsage() {
	fmt.Println("Usage: go run tools/swagger_autocomment/main.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -dir <path>       Directory containing handler files")
	fmt.Println("                    Default: ./internal/api/http/web/handler")
	fmt.Println("  -out <path>       Output directory")
	fmt.Println("                    Default: same as -dir")
	fmt.Println("  -security <name>  Security scheme name")
	fmt.Println("                    Default: BearerAuth")
	fmt.Println("  -router <path>    Router file path")
	fmt.Println("                    Default: ./internal/api/http/web/router.go")
	fmt.Println("  -v                Verbose output")
	fmt.Println("  -help, -h         Print this help message")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  go run tools/swagger_autocomment/main.go -dir ./internal/api/http/web/handler")
}

// findHandlerFiles finds all handler files in the given directory
func findHandlerFiles(dirPath string) ([]string, error) {
	var handlerFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "_handler.go") {
			handlerFiles = append(handlerFiles, path)
		}
		return nil
	})

	return handlerFiles, err
}

// printFileStats prints statistics for a processed file
func printFileStats(filePath string, stats FileStats) {
	fmt.Printf("File stats for %s:\n", filePath)
	fmt.Printf("  Total methods: %d\n", stats.TotalMethods)
	fmt.Printf("  Handler methods: %d\n", stats.HandlerMethods)
	fmt.Printf("  Already commented: %d\n", stats.AlreadyCommented)
	fmt.Printf("  Newly commented: %d\n", stats.NewlyCommented)
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
func processFile(filePath string, config Config, routes map[string]RouteInfo) (FileStats, error) {
	stats := FileStats{}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return stats, fmt.Errorf("parse error: %v", err)
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
		comment := generateSwaggerComment(funcDecl, config, routes)
		if config.Verbose {
			fmt.Printf("  Generated comment for %s\n", handlerName)
		}

		// Update the file with the new comment
		if err := updateFileWithComment(filePath, fset, funcDecl, comment); err != nil {
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
func generateSwaggerComment(funcDecl *ast.FuncDecl, config Config, routes map[string]RouteInfo) string {
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
	addParametersToComment(&comment, handlerName, path, method)

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
func addParametersToComment(comment *strings.Builder, handlerName, path, method string) {
	// Try to extract URI parameters from request struct
	reqTypeName := "types." + handlerName + "Req"
	uriParams, _ := extractURIParametersFromType(reqTypeName)

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
func updateFileWithComment(filePath string, fset *token.FileSet, funcDecl *ast.FuncDecl, comment string) error {
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

// extractURIParametersFromType extracts URI parameters from request struct types
func extractURIParametersFromType(typeName string) (map[string]URIParameter, error) {
	uriParams := make(map[string]URIParameter)

	// Find files with types package in the codebase
	typesFiles, err := findTypesFiles()
	if err != nil {
		return uriParams, err
	}

	for _, filePath := range typesFiles {
		// Parse the file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Get the struct name without package prefix (e.g., "UserReq" from "types.UserReq")
		parts := strings.Split(typeName, ".")
		if len(parts) != 2 {
			continue
		}
		structName := parts[1]

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

// findTypesFiles finds files that might contain type definitions
func findTypesFiles() ([]string, error) {
	var typesFiles []string

	// Common patterns for types package files
	patterns := []string{
		"./internal/api/http/web/types/*.go",
		"./pkg/types/*.go",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		typesFiles = append(typesFiles, matches...)
	}

	return typesFiles, nil
}
