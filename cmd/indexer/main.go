package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultSphinxConfig = "/etc/sphinx/sphinx.conf"
	defaultIndexerBin   = "/usr/bin/indexer"
)

// IndexerConfig holds configuration for the indexer
type IndexerConfig struct {
	ConfigPath  string
	IndexerBin  string
	RotateIndex bool
	AllIndexes  bool
	IndexNames  []string
	Verbose     bool
	DeltaOnly   bool
}

func main() {
	config := parseFlags()
	
	if err := runIndexer(config); err != nil {
		fmt.Fprintf(os.Stderr, "Indexer failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Indexing completed successfully")
}

// parseFlags parses command-line flags
func parseFlags() *IndexerConfig {
	config := &IndexerConfig{}
	
	flag.StringVar(&config.ConfigPath, "config", defaultSphinxConfig, "Path to sphinx.conf")
	flag.StringVar(&config.IndexerBin, "indexer", defaultIndexerBin, "Path to indexer binary")
	flag.BoolVar(&config.RotateIndex, "rotate", true, "Rotate indexes seamlessly")
	flag.BoolVar(&config.AllIndexes, "all", false, "Index all indexes")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&config.DeltaOnly, "delta", false, "Only index delta (incremental update)")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [index_names...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOpenWan Sphinx Indexer Tool\n")
		fmt.Fprintf(os.Stderr, "Indexes media files for full-text search\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --all                    # Index all indexes\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --delta                  # Only delta index (incremental)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s main delta               # Index specific indexes\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config /path/sphinx.conf --all  # Custom config\n", os.Args[0])
	}
	
	flag.Parse()
	
	// Get index names from remaining arguments
	if !config.AllIndexes && !config.DeltaOnly {
		config.IndexNames = flag.Args()
		if len(config.IndexNames) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No index names specified. Use --all or specify index names.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	}
	
	return config
}

// runIndexer executes the indexer command
func runIndexer(config *IndexerConfig) error {
	// Check if indexer binary exists
	if _, err := os.Stat(config.IndexerBin); os.IsNotExist(err) {
		return fmt.Errorf("indexer binary not found at %s", config.IndexerBin)
	}
	
	// Check if config file exists
	if _, err := os.Stat(config.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("sphinx config not found at %s", config.ConfigPath)
	}
	
	// Build indexer command
	args := []string{
		"--config", config.ConfigPath,
	}
	
	if config.RotateIndex {
		args = append(args, "--rotate")
	}
	
	if config.Verbose {
		args = append(args, "--verbose")
	}
	
	// Determine which indexes to process
	var indexesToBuild []string
	if config.AllIndexes {
		args = append(args, "--all")
		indexesToBuild = []string{"all"}
	} else if config.DeltaOnly {
		indexesToBuild = []string{"delta"}
		args = append(args, "delta")
	} else {
		indexesToBuild = config.IndexNames
		args = append(args, config.IndexNames...)
	}
	
	// Print indexing info
	fmt.Printf("Starting indexing at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Config: %s\n", config.ConfigPath)
	fmt.Printf("Indexes: %s\n", strings.Join(indexesToBuild, ", "))
	fmt.Println(strings.Repeat("-", 60))
	
	// Execute indexer
	cmd := exec.Command(config.IndexerBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("indexer command failed: %w", err)
	}
	
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Indexing finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// If rotate was used, print merge reminder for delta
	if config.DeltaOnly && config.RotateIndex {
		fmt.Println("\nReminder: Delta index updated. Consider merging with main index periodically:")
		fmt.Printf("  %s --config %s --merge main delta --rotate\n", config.IndexerBin, config.ConfigPath)
	}
	
	return nil
}

// validateConfig validates the sphinx configuration file
func validateConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	
	config := string(data)
	
	// Check for required sections
	requiredSections := []string{"source", "index", "searchd"}
	for _, section := range requiredSections {
		if !strings.Contains(config, section) {
			return fmt.Errorf("missing required section: %s", section)
		}
	}
	
	return nil
}

// createDirectories creates necessary directories for Sphinx
func createDirectories(configPath string) error {
	// Parse config to find data paths
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	
	config := string(data)
	
	// Extract path directives
	lines := strings.Split(config, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "path") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
			}
		}
	}
	
	return nil
}

// getIndexList returns list of available indexes from config
func getIndexList(configPath string) ([]string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var indexes []string
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "index ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				indexName := parts[1]
				// Skip inheritance part if present
				if strings.Contains(indexName, ":") {
					indexName = strings.Split(indexName, ":")[0]
				}
				indexes = append(indexes, indexName)
			}
		}
	}
	
	return indexes, nil
}
