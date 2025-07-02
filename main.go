package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// ASCII Art Banner
const banner = `
` + "\033[36m" + `██████╗ ██████╗ ██╗   ██╗████████╗ █████╗ ██╗         ██████╗ ██╗   ██╗███████╗████████╗███████╗██████╗` + "\033[0m" + `
` + "\033[36m" + `██╔══██╗██╔══██╗██║   ██║╚══██╔══╝██╔══██╗██║         ██╔══██╗██║   ██║██╔════╝╚══██╔══╝██╔════╝██╔══██╗` + "\033[0m" + `
` + "\033[35m" + `██████╔╝██████╔╝██║   ██║   ██║   ███████║██║         ██████╔╝██║   ██║███████╗   ██║   █████╗  ██████╔╝` + "\033[0m" + `
` + "\033[35m" + `██╔══██╗██╔══██╗██║   ██║   ██║   ██╔══██║██║         ██╔══██╗██║   ██║╚════██║   ██║   ██╔══╝  ██╔══██╗` + "\033[0m" + `
` + "\033[31m" + `██████╔╝██║  ██║╚██████╔╝   ██║   ██║  ██║███████╗    ██████╔╝╚██████╔╝███████║   ██║   ███████╗██║  ██║` + "\033[0m" + `
` + "\033[31m" + `╚═════╝ ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚══════╝    ╚═════╝  ╚═════╝ ╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝` + "\033[0m" + `
                                                                                                          
` + "\033[33m" + `                           🚀 High-Performance HTTP Load Testing Tool 🚀` + "\033[0m" + `
`

// Config holds the configuration for load testing
type Config struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	Concurrent  int               `json:"concurrent"`
	Requests    int               `json:"requests"`
	Duration    time.Duration     `json:"duration"`
	Timeout     time.Duration     `json:"timeout"`
	InsecureTLS bool              `json:"insecure_tls"`
}

// Result holds the result of a single request
type Result struct {
	StatusCode   int
	ResponseTime time.Duration
	ContentSize  int64
	Error        error
	Timestamp    time.Time
}

// Stats holds aggregated statistics
type Stats struct {
	TotalRequests   int
	SuccessfulReqs  int
	FailedReqs      int
	TotalTime       time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	AvgResponseTime time.Duration
	ResponseTimes   []time.Duration
	StatusCodes     map[int]int
	TotalBytes      int64
	RequestsPerSec  float64
	Percentiles     map[int]time.Duration
}

// LoadTester represents the load testing tool
type LoadTester struct {
	config     Config
	httpClient *http.Client
	results    []Result
	mu         sync.Mutex
}

// Global variables for command flags
var (
	url        string
	method     string
	headers    string
	body       string
	concurrent int
	requests   int
	timeout    time.Duration
	insecure   bool
	output     string
	noBanner   bool
	version    string = "dev"
)

// NewLoadTester creates a new load tester instance
func NewLoadTester(config Config) *LoadTester {
	transport := &http.Transport{
		MaxIdleConns:        config.Concurrent * 2,
		MaxIdleConnsPerHost: config.Concurrent,
		IdleConnTimeout:     30 * time.Second,
	}

	if config.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &LoadTester{
		config:     config,
		httpClient: client,
		results:    make([]Result, 0),
	}
}

// makeRequest performs a single HTTP request
func (lt *LoadTester) makeRequest() Result {
	start := time.Now()

	var bodyReader io.Reader
	if lt.config.Body != "" {
		bodyReader = strings.NewReader(lt.config.Body)
	}

	req, err := http.NewRequest(lt.config.Method, lt.config.URL, bodyReader)
	if err != nil {
		return Result{Error: err, ResponseTime: time.Since(start), Timestamp: time.Now()}
	}

	// Add headers
	for key, value := range lt.config.Headers {
		req.Header.Set(key, value)
	}

	// Set default User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Go Brutal/1.0")
	}

	resp, err := lt.httpClient.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return Result{Error: err, ResponseTime: responseTime, Timestamp: time.Now()}
	}
	defer resp.Body.Close()

	// Read response body to get content size
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{
			StatusCode:   resp.StatusCode,
			ResponseTime: responseTime,
			Error:        err,
			Timestamp:    time.Now(),
		}
	}

	return Result{
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		ContentSize:  int64(len(bodyBytes)),
		Timestamp:    time.Now(),
	}
}

// Run executes the load test
func (lt *LoadTester) Run(progressCallback func(completed, total int)) *Stats {
	startTime := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, lt.config.Concurrent)

	completed := 0
	progressMu := sync.Mutex{}

	for i := 0; i < lt.config.Requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			result := lt.makeRequest()

			lt.mu.Lock()
			lt.results = append(lt.results, result)
			lt.mu.Unlock()

			progressMu.Lock()
			completed++
			currentCompleted := completed
			progressMu.Unlock()

			if progressCallback != nil {
				progressCallback(currentCompleted, lt.config.Requests)
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	return lt.calculateStats(totalTime)
}

// calculateStats computes statistics from results
func (lt *LoadTester) calculateStats(totalTime time.Duration) *Stats {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	stats := &Stats{
		TotalRequests: len(lt.results),
		StatusCodes:   make(map[int]int),
		Percentiles:   make(map[int]time.Duration),
		TotalTime:     totalTime,
	}

	var responseTimes []time.Duration
	var totalBytes int64

	for _, result := range lt.results {
		if result.Error == nil {
			stats.SuccessfulReqs++
			totalBytes += result.ContentSize
		} else {
			stats.FailedReqs++
		}

		responseTimes = append(responseTimes, result.ResponseTime)
		stats.StatusCodes[result.StatusCode]++
	}

	stats.TotalBytes = totalBytes
	stats.ResponseTimes = responseTimes

	if len(responseTimes) > 0 {
		sort.Slice(responseTimes, func(i, j int) bool {
			return responseTimes[i] < responseTimes[j]
		})

		stats.MinResponseTime = responseTimes[0]
		stats.MaxResponseTime = responseTimes[len(responseTimes)-1]

		var total time.Duration
		for _, rt := range responseTimes {
			total += rt
		}
		stats.AvgResponseTime = total / time.Duration(len(responseTimes))

		// Calculate percentiles
		percentiles := []int{50, 95, 99}
		for _, p := range percentiles {
			index := int(math.Ceil(float64(p)/100*float64(len(responseTimes)))) - 1
			if index < 0 {
				index = 0
			}
			if index >= len(responseTimes) {
				index = len(responseTimes) - 1
			}
			stats.Percentiles[p] = responseTimes[index]
		}
	}

	if totalTime.Seconds() > 0 {
		stats.RequestsPerSec = float64(stats.TotalRequests) / totalTime.Seconds()
	}

	return stats
}

// SaveResultsToJSON saves results to a JSON file
func (lt *LoadTester) SaveResultsToJSON(filename string, stats *Stats) error {
	data := map[string]interface{}{
		"config":             lt.config,
		"stats":              stats,
		"individual_results": lt.results,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func printStats(stats *Stats) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful: %d (%.2f%%)\n", stats.SuccessfulReqs, float64(stats.SuccessfulReqs)/float64(stats.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.2f%%)\n", stats.FailedReqs, float64(stats.FailedReqs)/float64(stats.TotalRequests)*100)
	fmt.Printf("Total Time: %v\n", stats.TotalTime)
	fmt.Printf("Requests/sec: %.2f\n", stats.RequestsPerSec)
	fmt.Printf("Data Transfer: %.2f MB\n", float64(stats.TotalBytes)/(1024*1024))

	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("RESPONSE TIMES")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Min: %v\n", stats.MinResponseTime)
	fmt.Printf("Max: %v\n", stats.MaxResponseTime)
	fmt.Printf("Avg: %v\n", stats.AvgResponseTime)

	for p, time := range stats.Percentiles {
		fmt.Printf("%dth percentile: %v\n", p, time)
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("STATUS CODES")
	fmt.Println(strings.Repeat("-", 40))
	for code, count := range stats.StatusCodes {
		percentage := float64(count) / float64(stats.TotalRequests) * 100
		if code == 0 {
			fmt.Printf("Errors: %d (%.1f%%)\n", count, percentage)
		} else {
			fmt.Printf("%d: %d (%.1f%%)\n", code, count, percentage)
		}
	}
	fmt.Println(strings.Repeat("=", 60))
}

func printBanner() {
	if !noBanner {
		fmt.Print(banner)
	}
}

func runLoadTest(cmd *cobra.Command, args []string) error {

	if url == "" && len(args) == 0 {
		return fmt.Errorf("URL is required")
	}

	// Use URL from args if not provided via flag
	if url == "" && len(args) > 0 {
		url = args[0]
	}

	config := Config{
		URL:         url,
		Method:      strings.ToUpper(method),
		Concurrent:  concurrent,
		Requests:    requests,
		Timeout:     timeout,
		InsecureTLS: insecure,
		Headers:     make(map[string]string),
	}

	// Parse headers if provided
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &config.Headers); err != nil {
			return fmt.Errorf("error parsing headers: %v", err)
		}
	}

	if body != "" {
		config.Body = body
		// Set Content-Type if not provided and body is present
		if config.Headers["Content-Type"] == "" {
			config.Headers["Content-Type"] = "application/json"
		}
	}

	tester := NewLoadTester(config)

	// Print banner and configuration
	printBanner()
	fmt.Printf("Starting load test...\n")
	fmt.Printf("URL: %s\n", config.URL)
	fmt.Printf("Method: %s\n", config.Method)
	fmt.Printf("Concurrent users: %d\n", config.Concurrent)
	fmt.Printf("Total requests: %d\n", config.Requests)
	fmt.Printf("Timeout: %v\n", config.Timeout)
	fmt.Println(strings.Repeat("-", 50))

	// Run the load test with progress callback
	stats := tester.Run(func(completed, total int) {
		percent := float64(completed) / float64(total) * 100
		fmt.Printf("\rProgress: %d/%d (%.1f%%)", completed, total, percent)
	})

	fmt.Printf("\rCompleted: %d/%d (100.0%%)\n", config.Requests, config.Requests)
	printStats(stats)

	// Save results to JSON if output file specified
	if output != "" {
		if err := tester.SaveResultsToJSON(output, stats); err != nil {
			log.Printf("Error saving results to JSON: %v", err)
		} else {
			fmt.Printf("Results saved to: %s\n", output)
		}
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "brutal [URL]",
		Short: "Brutal - A powerful HTTP load testing tool",
		Long: `Brutal is a blazingly fast HTTP load testing tool with comprehensive analytics.
It provides detailed statistics, percentile analysis, and supports various HTTP methods.`,
		Example: `  brutal https://api.example.com
  brutal https://api.example.com -n 1000 -c 50
  brutal https://api.example.com -method POST -body '{"test": "data"}'`,
		RunE: runLoadTest,
		Args: cobra.MaximumNArgs(1),
	}

	// Add flags
	rootCmd.Flags().StringVarP(&url, "url", "u", "", "Target URL to test")
	rootCmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP method")
	rootCmd.Flags().StringVarP(&headers, "headers", "H", "", "Headers in JSON format")
	rootCmd.Flags().StringVarP(&body, "body", "d", "", "Request body")
	rootCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent requests")
	rootCmd.Flags().IntVarP(&requests, "requests", "n", 100, "Total number of requests")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Request timeout")
	rootCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Skip TLS certificate verification")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output file for JSON results")

	// Add persistent flags
	rootCmd.PersistentFlags().BoolVarP(&noBanner, "no-banner", "", false, "Disable ASCII art banner")

	// Add version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Brutal",
		Run: func(cmd *cobra.Command, args []string) {
			if !noBanner {
				printBanner()
			}
			fmt.Printf("Brutal Load Tester v%s\n", version)
			fmt.Printf("Built with Go %s\n", "1.23+")
			fmt.Printf("Powered by Cobra CLI Framework\n")
			fmt.Println()
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Add completion command
	var completionCmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(brutal completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ brutal completion bash > /etc/bash_completion.d/brutal
  # macOS:
  $ brutal completion bash > /usr/local/etc/bash_completion.d/brutal

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ brutal completion zsh > "${fpath[1]}/_brutal"

  # You will need to start a new shell for this setup to take effect.

fish:
  $ brutal completion fish | source

  # To load completions for each session, execute once:
  $ brutal completion fish > ~/.config/fish/completions/brutal.fish

PowerShell:
  PS> brutal completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> brutal completion powershell > brutal.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
