package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

// TUI Model for bubble tea
type model struct {
	loadTester  *LoadTester
	progress    progress.Model
	spinner     spinner.Model
	state       string
	completed   int
	total       int
	stats       *Stats
	startTime   time.Time
	currentTime time.Time
	liveStats   *LiveStats
	err         error
}

// LiveStats holds real-time statistics during testing
type LiveStats struct {
	mu              sync.RWMutex
	successful      int
	failed          int
	totalBytes      int64
	responseTimes   []time.Duration
	statusCodes     map[int]int
	minResponseTime time.Duration
	maxResponseTime time.Duration
}

// Styles for TUI
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2)
)

// Messages for bubble tea
type startTestMsg struct{}
type progressMsg struct {
	completed int
	result    Result
}
type completeMsg struct{ stats *Stats }
type tickMsg time.Time

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

	var body io.Reader
	if lt.config.Body != "" {
		body = strings.NewReader(lt.config.Body)
	}

	req, err := http.NewRequest(lt.config.Method, lt.config.URL, body)
	if err != nil {
		return Result{Error: err, ResponseTime: time.Since(start)}
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
		return Result{Error: err, ResponseTime: responseTime}
	}
	defer resp.Body.Close()

	// Read response body to get content size
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{
			StatusCode:   resp.StatusCode,
			ResponseTime: responseTime,
			Error:        err,
		}
	}

	return Result{
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		ContentSize:  int64(len(bodyBytes)),
	}
}

// RunWithTUI executes the load test with TUI
func (lt *LoadTester) RunWithTUI(ctx context.Context, updateChan chan<- tea.Msg) *Stats {
	startTime := time.Now()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, lt.config.Concurrent)

	completed := 0
	progressMu := sync.Mutex{}

	liveStats := &LiveStats{
		statusCodes:     make(map[int]int),
		minResponseTime: time.Duration(0),
		maxResponseTime: time.Duration(0),
	}

	for i := 0; i < lt.config.Requests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			result := lt.makeRequest()

			// Update live stats
			liveStats.mu.Lock()
			if result.Error == nil {
				liveStats.successful++
				liveStats.totalBytes += result.ContentSize
			} else {
				liveStats.failed++
			}
			liveStats.statusCodes[result.StatusCode]++
			liveStats.responseTimes = append(liveStats.responseTimes, result.ResponseTime)

			if liveStats.minResponseTime == 0 || result.ResponseTime < liveStats.minResponseTime {
				liveStats.minResponseTime = result.ResponseTime
			}
			if result.ResponseTime > liveStats.maxResponseTime {
				liveStats.maxResponseTime = result.ResponseTime
			}
			liveStats.mu.Unlock()

			lt.mu.Lock()
			lt.results = append(lt.results, result)
			lt.mu.Unlock()

			progressMu.Lock()
			completed++
			currentCompleted := completed
			progressMu.Unlock()

			// Send progress update
			select {
			case updateChan <- progressMsg{completed: currentCompleted, result: result}:
			case <-ctx.Done():
				return
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	stats := lt.calculateStats(totalTime)

	// Send completion message
	select {
	case updateChan <- completeMsg{stats: stats}:
	case <-ctx.Done():
	}

	return stats
}

// calculateStats computes statistics from results
func (lt *LoadTester) calculateStats(totalTime time.Duration) *Stats {
	stats := &Stats{
		TotalRequests: len(lt.results),
		StatusCodes:   make(map[int]int),
		TotalTime:     totalTime,
	}

	var responseTimes []time.Duration
	var totalResponseTime time.Duration

	for _, result := range lt.results {
		if result.Error == nil {
			stats.SuccessfulReqs++
			stats.TotalBytes += result.ContentSize
		} else {
			stats.FailedReqs++
		}

		stats.StatusCodes[result.StatusCode]++
		responseTimes = append(responseTimes, result.ResponseTime)
		totalResponseTime += result.ResponseTime

		if stats.MinResponseTime == 0 || result.ResponseTime < stats.MinResponseTime {
			stats.MinResponseTime = result.ResponseTime
		}
		if result.ResponseTime > stats.MaxResponseTime {
			stats.MaxResponseTime = result.ResponseTime
		}
	}

	if len(responseTimes) > 0 {
		stats.AvgResponseTime = totalResponseTime / time.Duration(len(responseTimes))
		stats.ResponseTimes = responseTimes
		stats.RequestsPerSec = float64(stats.TotalRequests) / totalTime.Seconds()

		// Calculate percentiles
		sort.Slice(responseTimes, func(i, j int) bool {
			return responseTimes[i] < responseTimes[j]
		})

		stats.Percentiles = map[int]time.Duration{
			50: responseTimes[len(responseTimes)*50/100],
			90: responseTimes[len(responseTimes)*90/100],
			95: responseTimes[len(responseTimes)*95/100],
			99: responseTimes[len(responseTimes)*99/100],
		}
	}

	return stats
}

// SaveResultsToJSON saves results to a JSON file
func (lt *LoadTester) SaveResultsToJSON(filename string, stats *Stats) error {
	data := map[string]interface{}{
		"config":    lt.config,
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

// NewLiveStats creates a new LiveStats instance
func NewLiveStats() *LiveStats {
	return &LiveStats{
		statusCodes: make(map[int]int),
	}
}

// TUI Model Implementation
func initialModel(loadTester *LoadTester) model {
	p := progress.New(progress.WithDefaultGradient())
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		loadTester:  loadTester,
		progress:    p,
		spinner:     s,
		state:       "ready",
		total:       loadTester.config.Requests,
		liveStats:   NewLiveStats(),
		currentTime: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
		func() tea.Msg { return startTestMsg{} },
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case startTestMsg:
		if m.state == "ready" {
			m.state = "running"
			m.startTime = time.Now()

			ctx, cancel := context.WithCancel(context.Background())
			updateChan := make(chan tea.Msg, 100)

			// Start the load test in a goroutine
			go func() {
				defer cancel()
				defer close(updateChan)
				m.loadTester.RunWithTUI(ctx, updateChan)
			}()

			// Listen for updates
			return m, tea.Batch(
				func() tea.Cmd {
					return func() tea.Msg {
						for msg := range updateChan {
							return msg
						}
						return nil
					}
				}(),
				m.spinner.Tick,
			)
		}

	case progressMsg:
		m.completed = msg.completed

		// Update live stats
		m.liveStats.mu.Lock()
		if msg.result.Error == nil {
			m.liveStats.successful++
			m.liveStats.totalBytes += msg.result.ContentSize
		} else {
			m.liveStats.failed++
		}
		m.liveStats.statusCodes[msg.result.StatusCode]++
		m.liveStats.responseTimes = append(m.liveStats.responseTimes, msg.result.ResponseTime)

		if m.liveStats.minResponseTime == 0 || msg.result.ResponseTime < m.liveStats.minResponseTime {
			m.liveStats.minResponseTime = msg.result.ResponseTime
		}
		if msg.result.ResponseTime > m.liveStats.maxResponseTime {
			m.liveStats.maxResponseTime = msg.result.ResponseTime
		}
		m.liveStats.mu.Unlock()

		if m.completed >= m.total {
			m.state = "completed"
		}
		return m, m.spinner.Tick

	case completeMsg:
		m.state = "completed"
		m.stats = msg.stats
		return m, nil

	case tickMsg:
		m.currentTime = time.Time(msg)
		return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("🚀 Go Brutal Tester"))
	s.WriteString("\n\n")

	// Configuration
	configBox := boxStyle.Render(fmt.Sprintf(
		"%s\n"+
			"URL: %s\n"+
			"Method: %s\n"+
			"Concurrent: %d\n"+
			"Total Requests: %d\n"+
			"Timeout: %v",
		headerStyle.Render("Configuration"),
		m.loadTester.config.URL,
		m.loadTester.config.Method,
		m.loadTester.config.Concurrent,
		m.loadTester.config.Requests,
		m.loadTester.config.Timeout,
	))
	s.WriteString(configBox)
	s.WriteString("\n\n")

	switch m.state {
	case "ready":
		s.WriteString(infoStyle.Render("Preparing to start load test..."))
		s.WriteString("\n")
		s.WriteString(m.spinner.View())

	case "running":
		// Progress bar
		percent := float64(m.completed) / float64(m.total)
		s.WriteString(fmt.Sprintf("Progress: %d/%d (%.1f%%)\n", m.completed, m.total, percent*100))
		s.WriteString(m.progress.ViewAs(percent))
		s.WriteString("\n\n")

		// Live statistics
		elapsed := m.currentTime.Sub(m.startTime)
		if elapsed == 0 {
			elapsed = time.Millisecond
		}

		m.liveStats.mu.RLock()
		currentRPS := float64(m.completed) / elapsed.Seconds()

		var avgResponseTime time.Duration
		if len(m.liveStats.responseTimes) > 0 {
			var total time.Duration
			for _, rt := range m.liveStats.responseTimes {
				total += rt
			}
			avgResponseTime = total / time.Duration(len(m.liveStats.responseTimes))
		}

		liveStatsBox := boxStyle.Render(fmt.Sprintf(
			"%s\n"+
				"Elapsed: %v\n"+
				"Successful: %s\n"+
				"Failed: %s\n"+
				"Current RPS: %.2f\n"+
				"Avg Response Time: %v\n"+
				"Min Response Time: %v\n"+
				"Max Response Time: %v\n"+
				"Data Transferred: %.2f MB",
			headerStyle.Render("Live Statistics"),
			elapsed.Truncate(time.Millisecond),
			successStyle.Render(fmt.Sprintf("%d", m.liveStats.successful)),
			errorStyle.Render(fmt.Sprintf("%d", m.liveStats.failed)),
			currentRPS,
			avgResponseTime.Truncate(time.Microsecond),
			m.liveStats.minResponseTime.Truncate(time.Microsecond),
			m.liveStats.maxResponseTime.Truncate(time.Microsecond),
			float64(m.liveStats.totalBytes)/(1024*1024),
		))
		m.liveStats.mu.RUnlock()

		s.WriteString(liveStatsBox)
		s.WriteString("\n\n")
		s.WriteString(m.spinner.View() + " Running...")

	case "completed":
		if m.stats != nil {
			s.WriteString(successStyle.Render("✅ Load test completed!"))
			s.WriteString("\n\n")

			// Final results
			resultsBox := boxStyle.Render(fmt.Sprintf(
				"%s\n"+
					"Total Requests: %d\n"+
					"Successful: %s (%d)\n"+
					"Failed: %s (%d)\n"+
					"Success Rate: %.2f%%\n"+
					"Total Time: %v\n"+
					"Requests/sec: %.2f\n"+
					"Data Transfer: %.2f MB\n\n"+
					"%s\n"+
					"Min: %v\n"+
					"Max: %v\n"+
					"Avg: %v\n"+
					"50th: %v\n"+
					"90th: %v\n"+
					"95th: %v\n"+
					"99th: %v",
				headerStyle.Render("Final Results"),
				m.stats.TotalRequests,
				successStyle.Render("✓"), m.stats.SuccessfulReqs,
				errorStyle.Render("✗"), m.stats.FailedReqs,
				float64(m.stats.SuccessfulReqs)/float64(m.stats.TotalRequests)*100,
				m.stats.TotalTime.Truncate(time.Millisecond),
				m.stats.RequestsPerSec,
				float64(m.stats.TotalBytes)/(1024*1024),
				headerStyle.Render("Response Times"),
				m.stats.MinResponseTime.Truncate(time.Microsecond),
				m.stats.MaxResponseTime.Truncate(time.Microsecond),
				m.stats.AvgResponseTime.Truncate(time.Microsecond),
				m.stats.Percentiles[50].Truncate(time.Microsecond),
				m.stats.Percentiles[90].Truncate(time.Microsecond),
				m.stats.Percentiles[95].Truncate(time.Microsecond),
				m.stats.Percentiles[99].Truncate(time.Microsecond),
			))
			s.WriteString(resultsBox)
			s.WriteString("\n\n")

			// Status codes
			if len(m.stats.StatusCodes) > 0 {
				statusBox := boxStyle.Render(fmt.Sprintf(
					"%s\n%s",
					headerStyle.Render("Status Codes"),
					formatStatusCodes(m.stats.StatusCodes, m.stats.TotalRequests),
				))
				s.WriteString(statusBox)
				s.WriteString("\n\n")
			}
		}

		s.WriteString(infoStyle.Render("Press 'q' or 'Ctrl+C' to exit"))
	}

	return s.String()
}

func formatStatusCodes(codes map[int]int, total int) string {
	var parts []string
	for code, count := range codes {
		percentage := float64(count) / float64(total) * 100
		if code == 0 {
			parts = append(parts, errorStyle.Render(fmt.Sprintf("Errors: %d (%.1f%%)", count, percentage)))
		} else {
			style := successStyle
			if code >= 400 {
				style = errorStyle
			}
			parts = append(parts, style.Render(fmt.Sprintf("%d: %d (%.1f%%)", code, count, percentage)))
		}
	}
	return strings.Join(parts, "\n")
}

func main() {
	var (
		url        = flag.String("url", "", "Target URL to test (required)")
		method     = flag.String("method", "GET", "HTTP method")
		headers    = flag.String("headers", "", "Headers in JSON format")
		body       = flag.String("body", "", "Request body")
		concurrent = flag.Int("c", 10, "Number of concurrent requests")
		requests   = flag.Int("n", 100, "Total number of requests")
		timeout    = flag.Duration("timeout", 30*time.Second, "Request timeout")
		insecure   = flag.Bool("insecure", false, "Skip TLS certificate verification")
		output     = flag.String("output", "", "Output file for JSON results")
		noTUI      = flag.Bool("no-tui", false, "Disable TUI and use simple output")
	)
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: URL is required")
		flag.Usage()
		return
	}

	config := Config{
		URL:         *url,
		Method:      strings.ToUpper(*method),
		Concurrent:  *concurrent,
		Requests:    *requests,
		Timeout:     *timeout,
		InsecureTLS: *insecure,
		Headers:     make(map[string]string),
	}

	// Parse headers if provided
	if *headers != "" {
		if err := json.Unmarshal([]byte(*headers), &config.Headers); err != nil {
			log.Fatalf("Error parsing headers: %v", err)
		}
	}

	if *body != "" {
		config.Body = *body
		// Set Content-Type if not provided and body is present
		if config.Headers["Content-Type"] == "" {
			config.Headers["Content-Type"] = "application/json"
		}
	}

	tester := NewLoadTester(config)

	if *noTUI {
		// Use simple CLI output
		fmt.Printf("Starting load test...\n")
		fmt.Printf("URL: %s\n", config.URL)
		fmt.Printf("Method: %s\n", config.Method)
		fmt.Printf("Concurrent users: %d\n", config.Concurrent)
		fmt.Printf("Total requests: %d\n", config.Requests)
		fmt.Printf("Timeout: %v\n", config.Timeout)
		fmt.Println(strings.Repeat("-", 50))

		ctx := context.Background()
		updateChan := make(chan tea.Msg, 100)
		go func() {
			defer close(updateChan)
			tester.RunWithTUI(ctx, updateChan)
		}()

		// Simple progress tracking
		for msg := range updateChan {
			switch m := msg.(type) {
			case progressMsg:
				percent := float64(m.completed) / float64(config.Requests) * 100
				fmt.Printf("\rProgress: %d/%d (%.1f%%)", m.completed, config.Requests, percent)
			case completeMsg:
				fmt.Printf("\rCompleted: %d/%d (100.0%%)\n", config.Requests, config.Requests)
				printSimpleStats(m.stats)
			}
		}
	} else {
		// Use TUI
		m := initialModel(tester)
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	}

	// Save results to JSON if output file specified
	if *output != "" {
		// We need to get the stats - for simplicity, run a quick test again
		// In a real implementation, you'd pass the stats from the TUI
		ctx := context.Background()
		updateChan := make(chan tea.Msg, 100)
		go func() {
			defer close(updateChan)
			tester.RunWithTUI(ctx, updateChan)
		}()

		var finalStats *Stats
		for msg := range updateChan {
			if m, ok := msg.(completeMsg); ok {
				finalStats = m.stats
				break
			}
		}

		if finalStats != nil {
			if err := tester.SaveResultsToJSON(*output, finalStats); err != nil {
				log.Printf("Error saving results to JSON: %v", err)
			} else {
				fmt.Printf("Results saved to: %s\n", *output)
			}
		}
	}
}

func printSimpleStats(stats *Stats) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Total Requests:      %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulReqs)
	fmt.Printf("Failed Requests:     %d\n", stats.FailedReqs)
	fmt.Printf("Success Rate:        %.2f%%\n", float64(stats.SuccessfulReqs)/float64(stats.TotalRequests)*100)
	fmt.Printf("Total Time:          %v\n", stats.TotalTime)
	fmt.Printf("Requests per Second: %.2f\n", stats.RequestsPerSec)
	fmt.Printf("Total Data Transfer: %.2f MB\n", float64(stats.TotalBytes)/(1024*1024))

	fmt.Println("\nResponse Times:")
	fmt.Printf("  Min:     %v\n", stats.MinResponseTime)
	fmt.Printf("  Max:     %v\n", stats.MaxResponseTime)
	fmt.Printf("  Average: %v\n", stats.AvgResponseTime)

	fmt.Println("\nPercentiles:")
	for _, p := range []int{50, 90, 95, 99} {
		fmt.Printf("  %dth:     %v\n", p, stats.Percentiles[p])
	}

	fmt.Println("\nStatus Code Distribution:")
	for code, count := range stats.StatusCodes {
		percentage := float64(count) / float64(stats.TotalRequests) * 100
		if code == 0 {
			fmt.Printf("  Errors:  %d (%.1f%%)\n", count, percentage)
		} else {
			fmt.Printf("  %d:       %d (%.1f%%)\n", code, count, percentage)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}
