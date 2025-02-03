package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VersionInfo struct {
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

var validArchitectures = map[string][]string{
	"windows": {"x86", "x64", "arm64"},
	"darwin":  {"x64", "arm64"},
	"linux":   {"x86", "x64", "armv7l", "arm64"},
}

func getPort() string {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
		fmt.Println("Warning: .env file not found")
	}

	// Try to get PORT from environment (either from .env or system env)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	return port
}

func main() {
	http.HandleFunc("/patch/", handlePatchDownload)
	http.HandleFunc("/appVersion.info", handleVersionInfo)

	// Add root handler that returns 403 Forbidden
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.NotFound(w, r)
	})

	port := getPort()
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handlePatchDownload(w http.ResponseWriter, r *http.Request) {
	// Extract platform and timestamp from query parameters
	queryTimestamp := r.URL.Query().Get("timestamp")
	queryArch := r.URL.Query().Get("arch")

	// Get the requested file path
	filePath := strings.TrimPrefix(r.URL.Path, "/patch/")
	if filePath == "" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Validate file name format (yyyymmddhhmm-platform.zip)
	re := regexp.MustCompile(`^(\d{12})-([a-zA-Z]+)\.zip$`)
	matches := re.FindStringSubmatch(filePath)
	if matches == nil {
		http.Error(w, "Invalid file format", http.StatusBadRequest)
		return
	}

	timestamp := matches[1]
	platform := strings.ToLower(matches[2])

	// Validate platform
	validArchs, ok := validArchitectures[platform]
	if !ok {
		http.Error(w, "Invalid platform", http.StatusBadRequest)
		return
	}

	// Handle architecture
	targetArch := "x64" // default architecture
	if queryArch != "" {
		isValidArch := false
		for _, arch := range validArchs {
			if arch == queryArch {
				isValidArch = true
				targetArch = queryArch
				break
			}
		}
		if !isValidArch {
			http.Error(w, "Invalid architecture", http.StatusNotFound)
			return
		}
	}

	// Handle timestamp query
	if queryTimestamp != "" {
		requestedTime, err := strconv.ParseInt(queryTimestamp, 10, 64)
		if err != nil {
			http.Error(w, "Invalid timestamp", http.StatusBadRequest)
			return
		}

		// Find the latest version before the requested timestamp
		files, err := filepath.Glob("patches/" + platform + "/*-" + platform + ".zip")
		if err != nil {
			http.Error(w, "Error reading patches", http.StatusInternalServerError)
			return
		}

		var latestValidVersion string
		var latestTime int64

		for _, file := range files {
			base := filepath.Base(file)
			fileTimestamp := strings.Split(base, "-")[0]
			fileTime, err := time.ParseInLocation("200601021504", fileTimestamp, time.UTC)
			if err != nil {
				continue
			}

			fileUnix := fileTime.Unix()
			if fileUnix <= requestedTime && fileUnix > latestTime {
				latestTime = fileUnix
				latestValidVersion = fileTimestamp
			}
		}

		if latestValidVersion != "" {
			timestamp = latestValidVersion
		}
	}

	// Construct the final file path
	finalPath := filepath.Join("patches", platform, timestamp+"-"+platform+"-"+targetArch+".zip")

	// Check if file exists
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		http.Error(w, "Update file not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, finalPath)
}

func handleVersionInfo(w http.ResponseWriter, r *http.Request) {
	// Get the query parameters
	queryArch := r.URL.Query().Get("arch")
	platform := strings.ToLower(r.URL.Query().Get("platform"))

	// Set defaults if not specified
	if queryArch == "" {
		queryArch = "x64" // default to x64 if not specified
	}
	if platform == "" {
		http.Error(w, "Platform parameter is required", http.StatusBadRequest)
		return
	}

	// Validate the platform
	validArchs, ok := validArchitectures[platform]
	if !ok {
		http.Error(w, "Invalid platform", http.StatusBadRequest)
		return
	}

	// Validate the architecture for the platform
	isValidArch := false
	for _, arch := range validArchs {
		if arch == queryArch {
			isValidArch = true
			break
		}
	}

	if !isValidArch {
		http.Error(w, "Invalid architecture", http.StatusNotFound)
		return
	}

	// Find the latest version for the specific platform and architecture
	files, err := filepath.Glob("patches/" + platform + "/*-" + platform + "-" + queryArch + ".zip")
	if err != nil {
		http.Error(w, "Error reading patches", http.StatusInternalServerError)
		return
	}

	var latestTime time.Time
	var latestVersion string

	for _, file := range files {
		base := filepath.Base(file)
		timestamp := strings.Split(base, "-")[0]
		fileTime, err := time.ParseInLocation("200601021504", timestamp, time.UTC)
		if err != nil {
			continue
		}

		if fileTime.After(latestTime) {
			latestTime = fileTime
			latestVersion = timestamp
		}
	}

	if latestVersion == "" {
		http.Error(w, "No updates found", http.StatusNotFound)
		return
	}

	// Set plain text content type and write just the version string
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(latestVersion))
}
