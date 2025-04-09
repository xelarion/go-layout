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
	HandlerDir     string
	OutputDir      string
	SecurityScheme string
	RouterFile     string
	Verbose        bool
}

// RouteInfo stores information about an API route
type RouteInfo struct {
	Path      string
	Method    string
	Handler   string
	IsSecured bool
}

// Parameter represents a parameter extracted from a handler function.
type Parameter struct {
	Name     string
	Type     string
	Location string // path, query, body
	Required bool
}

func main() {
	config := Config{
		HandlerDir:     "./internal/api/http/web/handler",
		SecurityScheme: "BearerAuth",
		RouterFile:     "./internal/api/http/web/router.go",
		Verbose:        true,
	}

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

	if config.OutputDir == "" {
		config.OutputDir = config.HandlerDir
	}

	// 检查目录是否存在
	if _, err := os.Stat(config.HandlerDir); os.IsNotExist(err) {
		fmt.Printf("Error: Directory %s does not exist!\n", config.HandlerDir)
		fmt.Println("Current working directory:", getCwd())
		fmt.Println("Please provide a valid directory using the -dir flag")
		os.Exit(1)
	}

	// Extract routes from router file
	routes, err := extractRoutes(config.RouterFile)
	if err != nil {
		fmt.Printf("Warning: Could not extract routes from router file: %v\n", err)
		fmt.Println("Will use function name analysis to determine routes.")
	}

	fmt.Printf("Processing handlers in %s\n", config.HandlerDir)
	fmt.Printf("Output directory: %s\n", config.OutputDir)
	fmt.Printf("Security Scheme: %s\n", config.SecurityScheme)
	fmt.Printf("Router file: %s\n", config.RouterFile)
	fmt.Printf("Verbose mode: %v\n", config.Verbose)
	fmt.Printf("Extracted %d routes from router file\n", len(routes))

	// Process all handler files
	handlerFiles := []string{}
	err = filepath.Walk(config.HandlerDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "_handler.go") {
			handlerFiles = append(handlerFiles, path)
			if config.Verbose {
				fmt.Printf("Found handler file: %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(handlerFiles) == 0 {
		fmt.Printf("No handler files found in %s\n", config.HandlerDir)
		fmt.Println("Make sure the directory contains *_handler.go files")
		os.Exit(1)
	}

	fmt.Printf("Found %d handler files to process\n", len(handlerFiles))

	// 处理每个文件
	for _, path := range handlerFiles {
		fmt.Printf("Processing file: %s\n", path)
		err = processFile(path, config, routes)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
		}
	}

	fmt.Println("Swagger comment generation complete!")
}

// 获取当前工作目录
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}

// Print usage information
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

// Extract routes from router file
func extractRoutes(routerFile string) (map[string]RouteInfo, error) {
	routes := make(map[string]RouteInfo)

	// Read the router file
	content, err := os.ReadFile(routerFile)
	if err != nil {
		return routes, fmt.Errorf("error reading router file: %v", err)
	}

	// Parse the content to extract routes
	routerContent := string(content)

	// Extract all route registrations
	// Pattern: api.METHOD("/path", handler.HandlerFunc)
	// or authorized.METHOD("/path", handler.HandlerFunc)
	apiRegex := regexp.MustCompile(`(api|authorized)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)", ([a-zA-Z0-9]+)\.([a-zA-Z0-9]+)\)`)
	matches := apiRegex.FindAllStringSubmatch(routerContent, -1)

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		group := match[1]   // api or authorized
		method := match[2]  // HTTP method
		path := match[3]    // Path
		handler := match[5] // Handler function name

		// 转换路径参数格式从 :id 到 {id}
		convertedPath := convertPathParams(path)

		routes[handler] = RouteInfo{
			Path:      convertedPath,
			Method:    strings.ToLower(method),
			Handler:   handler,
			IsSecured: group == "authorized",
		}
	}

	return routes, nil
}

// 将路径参数从 :id 格式转换为 {id} 格式
func convertPathParams(path string) string {
	paramRegex := regexp.MustCompile(`:([^/]+)`)
	return paramRegex.ReplaceAllString(path, "{$1}")
}

// Process a single file to add Swagger comments
func processFile(filePath string, config Config, routes map[string]RouteInfo) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	// Make a copy of the file path for output
	outputPath := filepath.Join(config.OutputDir, filepath.Base(filePath))
	if config.Verbose {
		fmt.Printf("Output will be written to: %s\n", outputPath)
	}

	// 计数器
	totalMethods := 0
	handlerMethods := 0
	alreadyCommented := 0
	newlyCommented := 0

	// 收集所有需要处理的方法
	handlerFuncs := []*ast.FuncDecl{}

	// Iterate through all declarations in the file
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			totalMethods++

			// 检查是否是处理函数
			if isHandlerMethod(funcDecl, config) {
				handlerMethods++
				handlerName := funcDecl.Name.Name

				// Skip if already has swagger comment
				if hasSwaggerComment(funcDecl) {
					alreadyCommented++
					if config.Verbose {
						fmt.Printf("  Already has swagger comment: %s\n", handlerName)
					}
					continue
				}

				// 添加到处理列表
				handlerFuncs = append(handlerFuncs, funcDecl)
			}
		}
	}

	// 从后向前处理函数，避免修改后的位置影响
	for i := len(handlerFuncs) - 1; i >= 0; i-- {
		funcDecl := handlerFuncs[i]
		handlerName := funcDecl.Name.Name

		// Generate swagger comment
		comment := generateSwaggerComment(funcDecl, config, routes)
		if config.Verbose {
			fmt.Printf("  Generated comment for %s\n", handlerName)
			if config.Verbose {
				fmt.Printf("  Comment content:\n%s\n", comment)
			}
		}

		// Update the file with the new comment
		updateFileWithComment(filePath, fset, funcDecl, comment)
		newlyCommented++
	}

	fmt.Printf("File stats for %s:\n", filePath)
	fmt.Printf("  Total methods: %d\n", totalMethods)
	fmt.Printf("  Handler methods: %d\n", handlerMethods)
	fmt.Printf("  Already commented: %d\n", alreadyCommented)
	fmt.Printf("  Newly commented: %d\n", newlyCommented)

	return nil
}

// Check if a function declaration is a handler method
func isHandlerMethod(funcDecl *ast.FuncDecl, config Config) bool {
	// Check if the function has a receiver
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}

	// Check if the receiver name has "Handler" in it
	receiverType := ""
	if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
		// For pointer receivers like *UserHandler
		if ident, ok := starExpr.X.(*ast.Ident); ok {
			receiverType = ident.Name
		}
	} else if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
		// For value receivers like UserHandler
		receiverType = ident.Name
	}

	isHandler := strings.Contains(receiverType, "Handler")
	if !isHandler {
		return false
	}

	// Check if at least one parameter exists
	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) < 1 {
		return false
	}

	// Check if one of the parameters is *gin.Context or similar context type
	hasContextParam := false
	for _, param := range funcDecl.Type.Params.List {
		// Check for *xxx.Context pattern
		if starExpr, ok := param.Type.(*ast.StarExpr); ok {
			if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name == "Context" {
					hasContextParam = true
					break
				}
			}
		}
	}

	if !hasContextParam {
		return false
	}

	// We've identified a handler method - it has a *Handler receiver and takes a context
	if config.Verbose {
		fmt.Printf("Found handler method: %s with receiver %s\n", funcDecl.Name.Name, receiverType)
	}
	return true
}

// Check if a function already has Swagger comments
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

// Generate a Swagger comment for a handler method
func generateSwaggerComment(funcDecl *ast.FuncDecl, config Config, routes map[string]RouteInfo) string {
	handlerName := funcDecl.Name.Name

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

	// 获取处理器名称（用于生成tag）
	receiverType := ""
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				receiverType = ident.Name
			}
		}
	}

	// Build the comment
	var comment strings.Builder

	// Summary and description
	comment.WriteString(fmt.Sprintf("// %s godoc\n", handlerName))
	comment.WriteString(fmt.Sprintf("// @Summary %s\n", generateSummary(handlerName)))
	comment.WriteString(fmt.Sprintf("// @Description %s\n", generateDescription(handlerName)))
	comment.WriteString("// @Tags " + determineTagFromHandler(receiverType) + "\n")
	comment.WriteString("// @Accept json\n")
	comment.WriteString("// @Produce json\n")

	// Parameters
	params := determineParameters(handlerName, path, method)
	for _, param := range params {
		required := "false"
		if param.Required {
			required = "true"
		}

		comment.WriteString(fmt.Sprintf("// @Param %s %s %s %s \"%s\"\n",
			param.Name, param.Location, param.Type, required, param.Name))
	}

	// Response type based on handler name
	responseType := "types." + handlerName + "Resp"
	statusCode := determineStatusCode(method)

	// Success response - special case for List operations which often have pagination
	if strings.HasPrefix(handlerName, "List") {
		// List operations typically return a response with results array and pagination info
		comment.WriteString(fmt.Sprintf("// @Success %d {object} types.Response{data=types.%sResp} \"Success\"\n", statusCode, handlerName))
	} else {
		comment.WriteString(fmt.Sprintf("// @Success %d {object} types.Response{data=%s} \"Success\"\n", statusCode, responseType))
	}

	// Error responses
	comment.WriteString("// @Failure 400 {object} types.Response \"Bad request\"\n")
	comment.WriteString("// @Failure 401 {object} types.Response \"Unauthorized\"\n")
	comment.WriteString("// @Failure 500 {object} types.Response \"Internal server error\"\n")

	// Security
	if isSecured || (!exists && isLikelySecured(handlerName)) {
		comment.WriteString(fmt.Sprintf("// @Security %s\n", config.SecurityScheme))
	}

	// Router
	comment.WriteString(fmt.Sprintf("// @Router %s [%s]\n", path, method))

	return comment.String()
}

// Update a file with a new comment
func updateFileWithComment(filePath string, fset *token.FileSet, funcDecl *ast.FuncDecl, comment string) {
	// 直接使用AST位置信息更精确地定位函数
	startPos := funcDecl.Pos()
	startOffset := fset.Position(startPos).Offset

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	// 检查函数是否已有注释，如果有，需要替换
	docStartOffset := startOffset
	if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
		docPos := funcDecl.Doc.Pos()
		docStartOffset = fset.Position(docPos).Offset
	}

	// 构建新内容：前面的内容 + 新注释 + 函数本身
	newContent := string(content[:docStartOffset]) + comment + string(content[startOffset:])

	// 写回文件
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", filePath, err)
	}
}

// Determine the HTTP method based on the function name
func determineHTTPMethod(handlerName string) string {
	if strings.HasPrefix(handlerName, "Create") || strings.HasPrefix(handlerName, "Add") {
		return "POST"
	} else if strings.HasPrefix(handlerName, "Get") || strings.HasPrefix(handlerName, "List") || strings.HasPrefix(handlerName, "Find") {
		return "GET"
	} else if strings.HasPrefix(handlerName, "Update") || strings.HasPrefix(handlerName, "Modify") {
		return "PUT"
	} else if strings.HasPrefix(handlerName, "Patch") || strings.HasPrefix(handlerName, "Partial") {
		return "PATCH"
	} else if strings.HasPrefix(handlerName, "Delete") || strings.HasPrefix(handlerName, "Remove") {
		return "DELETE"
	}

	return "GET" // Default
}

// Determine the API path based on the function name
func determinePath(handlerName string) string {
	// Convert camel case to kebab case for path
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	path := re.ReplaceAllString(handlerName, "$1-$2")
	path = strings.ToLower(path)

	// Remove common prefixes
	path = strings.TrimPrefix(path, "get-")
	path = strings.TrimPrefix(path, "list-")
	path = strings.TrimPrefix(path, "create-")
	path = strings.TrimPrefix(path, "update-")
	path = strings.TrimPrefix(path, "delete-")
	path = strings.TrimPrefix(path, "handle-")

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

// Generate a summary for a handler method
func generateSummary(handlerName string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	summary := re.ReplaceAllString(handlerName, "$1 $2")
	return summary
}

// Generate a description for a handler method
func generateDescription(handlerName string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	desc := re.ReplaceAllString(handlerName, "$1 $2")

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

// 从处理器名称确定tag
func determineTagFromHandler(handlerName string) string {
	// 移除Handler后缀
	tag := strings.TrimSuffix(handlerName, "Handler")
	return strings.ToLower(tag)
}

// Determine parameters based on handler name, path and HTTP method
func determineParameters(handlerName string, path string, method string) []Parameter {
	params := []Parameter{}

	// 1. 处理路径参数 - 保留这些参数，因为Swagger需要知道URL路径中的参数
	pathParamRegex := regexp.MustCompile(`\{([^}]+)\}`)
	pathParams := pathParamRegex.FindAllStringSubmatch(path, -1)

	for _, match := range pathParams {
		if len(match) >= 2 {
			paramName := match[1]
			params = append(params, Parameter{
				Name:     paramName,
				Type:     "string", // 默认为string类型
				Location: "path",
				Required: true, // 路径参数总是必需的
			})
		}
	}

	// 2. 添加请求结构体参数
	reqType := "types." + handlerName + "Req"

	// 根据HTTP方法确定参数位置
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

// Determine status code based on method
func determineStatusCode(method string) int {
	if method == "post" {
		return 201 // Created
	} else if method == "delete" {
		return 204 // No Content
	}
	return 200 // OK
}

// Determine if a handler is likely secured based on its name
func isLikelySecured(handlerName string) bool {
	// Most APIs except public ones like login, register, etc. are secured
	publicPatterns := []string{"Login", "Register", "Captcha", "Public"}

	for _, pattern := range publicPatterns {
		if strings.Contains(handlerName, pattern) {
			return false
		}
	}

	return true
}
