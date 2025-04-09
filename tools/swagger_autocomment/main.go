// Package main provides a tool to generate intelligent Swagger comments for handler methods.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

var (
	// Regular expressions
	pathParamRegex   = regexp.MustCompile(`\{([^}]+)\}`)
	camelCaseRegex   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	apiRouteRegex    = regexp.MustCompile(`(api|authorized)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)", ([a-zA-Z0-9]+)\.([a-zA-Z0-9]+)\)`)
	routerParamRegex = regexp.MustCompile(`:([^/]+)`)
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
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(handlerFiles) == 0 {
		fmt.Printf("No handler files found in %s\n", config.HandlerDir)
		fmt.Println("Make sure the directory contains *_handler.go files")
		os.Exit(1)
	}

	if config.Verbose {
		fmt.Printf("Found %d handler files to process\n", len(handlerFiles))
	}

	// Process each handler file
	for _, path := range handlerFiles {
		if config.Verbose {
			fmt.Printf("Processing file: %s\n", path)
		}

		stats, err := processFile(path, config, routes)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
			continue
		}

		if config.Verbose {
			printFileStats(path, stats)
		}
	}

	fmt.Println("Swagger comment generation complete!")
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
	handlerFiles := []string{}

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

	// Get output path
	outputPath := filepath.Join(config.OutputDir, filepath.Base(filePath))
	if config.Verbose {
		fmt.Printf("Output will be written to: %s\n", outputPath)
	}

	// Collect all handler methods
	handlerFuncs := []*ast.FuncDecl{}

	// Iterate through all declarations in the file
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			stats.TotalMethods++

			// Check if it's a handler method
			if isHandlerMethod(funcDecl, config.Verbose) {
				stats.HandlerMethods++
				handlerName := funcDecl.Name.Name

				// Skip if already has swagger comment
				if hasSwaggerComment(funcDecl) {
					stats.AlreadyCommented++
					if config.Verbose {
						fmt.Printf("  Already has swagger comment: %s\n", handlerName)
					}
					continue
				}

				// Add to processing list
				handlerFuncs = append(handlerFuncs, funcDecl)
			}
		}
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

// isHandlerMethod checks if a function declaration is a handler method
func isHandlerMethod(funcDecl *ast.FuncDecl, verbose bool) bool {
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
	if !hasContextParameter(funcDecl) {
		return false
	}

	// This is a handler method
	if verbose {
		fmt.Printf("Found handler method: %s with receiver %s\n", funcDecl.Name.Name, receiverType)
	}

	return true
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
	method := ""
	path := ""
	isSecured := false

	if exists {
		method = routeInfo.Method
		path = routeInfo.Path
		isSecured = routeInfo.IsSecured
	} else {
		// Fallback to function name analysis
		method = strings.ToLower(determineHTTPMethod(handlerName))
		path = determinePath(handlerName)
	}

	// Build the comment
	var comment strings.Builder
	comment.WriteString(fmt.Sprintf("// %s godoc\n", handlerName))
	comment.WriteString(fmt.Sprintf("// @Summary %s\n", generateSummary(handlerName)))
	comment.WriteString(fmt.Sprintf("// @Description %s\n", generateDescription(handlerName)))
	comment.WriteString("// @Tags " + determineTagFromHandler(receiverType) + "\n")
	comment.WriteString("// @Accept json\n")
	comment.WriteString("// @Produce json\n")

	// Add parameters
	params := determineParameters(handlerName, path, method)
	for _, param := range params {
		required := "false"
		if param.Required {
			required = "true"
		}

		comment.WriteString(fmt.Sprintf("// @Param %s %s %s %s \"%s\"\n",
			param.Name, param.Location, param.Type, required, param.Name))
	}

	// Add response type - special case for List operations
	statusCode := determineStatusCode(method)
	if strings.HasPrefix(handlerName, "List") {
		// List operations typically have paginated responses
		comment.WriteString(fmt.Sprintf("// @Success %d {object} types.Response{data=types.%sResp} \"Success\"\n", statusCode, handlerName))
	} else {
		comment.WriteString(fmt.Sprintf("// @Success %d {object} types.Response{data=types.%sResp} \"Success\"\n", statusCode, handlerName))
	}

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

// updateFileWithComment updates a file with a new Swagger comment
func updateFileWithComment(filePath string, fset *token.FileSet, funcDecl *ast.FuncDecl, comment string) error {
	// Get function position
	startPos := funcDecl.Pos()
	startOffset := fset.Position(startPos).Offset

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Determine where to insert comment
	docStartOffset := startOffset
	if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
		docPos := funcDecl.Doc.Pos()
		docStartOffset = fset.Position(docPos).Offset
	}

	// Create new content with comment
	newContent := string(content[:docStartOffset]) + comment + string(content[startOffset:])

	// Write back to file
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// determineHTTPMethod determines the HTTP method based on the function name
func determineHTTPMethod(handlerName string) string {
	if strings.HasPrefix(handlerName, "Create") || strings.HasPrefix(handlerName, "Add") {
		return "post"
	} else if strings.HasPrefix(handlerName, "Get") || strings.HasPrefix(handlerName, "List") || strings.HasPrefix(handlerName, "Find") {
		return "get"
	} else if strings.HasPrefix(handlerName, "Update") || strings.HasPrefix(handlerName, "Modify") {
		return "put"
	} else if strings.HasPrefix(handlerName, "Patch") || strings.HasPrefix(handlerName, "Partial") {
		return "patch"
	} else if strings.HasPrefix(handlerName, "Delete") || strings.HasPrefix(handlerName, "Remove") {
		return "delete"
	}

	return "get" // Default
}

// determinePath determines the API path based on the function name
func determinePath(handlerName string) string {
	// Convert camel case to kebab case
	path := camelCaseRegex.ReplaceAllString(handlerName, "$1-$2")
	path = strings.ToLower(path)

	// Remove common prefixes
	prefixes := []string{"get-", "list-", "create-", "update-", "delete-", "handle-"}
	for _, prefix := range prefixes {
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
	prefix := ""
	if strings.HasPrefix(handlerName, "Create") {
		prefix = "Creates a new "
	} else if strings.HasPrefix(handlerName, "Get") && !strings.HasPrefix(handlerName, "List") {
		prefix = "Retrieves a single "
	} else if strings.HasPrefix(handlerName, "List") {
		prefix = "Retrieves a list of "
	} else if strings.HasPrefix(handlerName, "Update") {
		prefix = "Updates an existing "
	} else if strings.HasPrefix(handlerName, "Delete") {
		prefix = "Deletes an existing "
	}

	if prefix != "" {
		resourceName := strings.TrimPrefix(desc, strings.Split(desc, " ")[0]+" ")
		return prefix + resourceName
	}

	return desc
}

// determineTagFromHandler determines the Swagger tag from a handler name
func determineTagFromHandler(handlerName string) string {
	// Remove Handler suffix and convert to lowercase
	return strings.ToLower(strings.TrimSuffix(handlerName, "Handler"))
}

// determineParameters determines the parameters for a handler method
func determineParameters(handlerName string, path string, method string) []Parameter {
	params := []Parameter{}

	// Add path parameters
	pathParams := pathParamRegex.FindAllStringSubmatch(path, -1)
	for _, match := range pathParams {
		if len(match) >= 2 {
			params = append(params, Parameter{
				Name:     match[1],
				Type:     "string",
				Location: "path",
				Required: true,
			})
		}
	}

	// Add request parameter
	reqType := "types." + handlerName + "Req"
	location := "body"
	required := true

	if method == "get" {
		location = "query"
		required = false
	}

	params = append(params, Parameter{
		Name:     "req",
		Type:     reqType,
		Location: location,
		Required: required,
	})

	return params
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
	// Common patterns for public endpoints
	publicPatterns := []string{"Login", "Register", "Captcha", "Public"}

	for _, pattern := range publicPatterns {
		if strings.Contains(handlerName, pattern) {
			return false
		}
	}

	return true
}
