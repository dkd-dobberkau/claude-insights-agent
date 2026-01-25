package watcher

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dkd/claude-insights-agent/internal/client"
	"github.com/dkd/claude-insights-agent/internal/config"
	"github.com/dkd/claude-insights-agent/internal/filter"
	"github.com/dkd/claude-insights-agent/internal/parser"
)

// State tracks which sessions and plans have been synced
type State struct {
	SyncedSessions map[string]time.Time `json:"synced_sessions"`
	SyncedPlans    map[string]time.Time `json:"synced_plans"`
	LastSync       time.Time            `json:"last_sync"`
}

// Watcher monitors Claude logs and syncs to server
type Watcher struct {
	cfg       *config.Config
	client    *client.Client
	filter    *filter.Filter
	state     *State
	statePath string
	logsPath  string
	logger    *log.Logger
	stopCh    chan struct{}
}

// New creates a new Watcher
func New(cfg *config.Config, logger *log.Logger) *Watcher {
	return &Watcher{
		cfg:       cfg,
		client:    client.New(cfg.Server.URL, cfg.Server.APIKey),
		filter:    filter.New(&cfg.Sharing),
		statePath: config.StatePath(),
		logsPath:  config.ClaudeLogsPath(),
		logger:    logger,
		stopCh:    make(chan struct{}),
	}
}

// Start begins watching for new sessions
func (w *Watcher) Start() error {
	// Load state
	if err := w.loadState(); err != nil {
		w.logger.Printf("Warning: could not load state: %v", err)
		w.state = &State{
			SyncedSessions: make(map[string]time.Time),
			SyncedPlans:    make(map[string]time.Time),
		}
	}

	// Initial sync
	w.logger.Println("Starting initial sync...")
	if err := w.sync(); err != nil {
		w.logger.Printf("Initial sync error: %v", err)
	}

	// Start periodic sync
	ticker := time.NewTicker(time.Duration(w.cfg.Sync.Interval) * time.Second)
	defer ticker.Stop()

	w.logger.Printf("Watching %s (interval: %ds)", w.logsPath, w.cfg.Sync.Interval)

	for {
		select {
		case <-ticker.C:
			if err := w.sync(); err != nil {
				w.logger.Printf("Sync error: %v", err)
			}
		case <-w.stopCh:
			w.logger.Println("Watcher stopped")
			return nil
		}
	}
}

// Stop halts the watcher
func (w *Watcher) Stop() {
	close(w.stopCh)
}

// SyncOnce performs a single sync operation
func (w *Watcher) SyncOnce() error {
	if err := w.loadState(); err != nil {
		w.state = &State{
			SyncedSessions: make(map[string]time.Time),
			SyncedPlans:    make(map[string]time.Time),
		}
	}
	return w.sync()
}

// sync finds and uploads new sessions
func (w *Watcher) sync() error {
	// Find all JSONL session files
	projectsDir := filepath.Join(w.logsPath, "projects")
	files, err := w.findSessions(projectsDir)
	if err != nil {
		return err
	}

	// Filter to new sessions only
	var newFiles []string
	for _, f := range files {
		sessionID := filepath.Base(strings.TrimSuffix(f, ".jsonl"))
		if _, synced := w.state.SyncedSessions[sessionID]; !synced {
			newFiles = append(newFiles, f)
		}
	}

	if len(newFiles) == 0 {
		w.logger.Println("No new sessions to sync")
	} else {
		w.logger.Printf("Found %d new sessions", len(newFiles))

		// Parse and upload sessions
		var toUpload []*parser.Session
		for _, f := range newFiles {
			session, err := parser.ParseJSONL(f)
			if err != nil {
				w.logger.Printf("Error parsing %s: %v", f, err)
				continue
			}

			// Apply privacy filter
			filtered := w.filter.Apply(session)
			if filtered == nil {
				w.logger.Printf("Session %s excluded by filter", session.ID)
				// Mark as synced anyway to avoid re-processing
				w.state.SyncedSessions[session.ID] = time.Now()
				continue
			}

			toUpload = append(toUpload, filtered)
		}

		if len(toUpload) > 0 {
			// Upload in batches of 10
			batchSize := 10
			for i := 0; i < len(toUpload); i += batchSize {
				end := i + batchSize
				if end > len(toUpload) {
					end = len(toUpload)
				}
				batch := toUpload[i:end]

				var uploadErr error
				for attempt := 1; attempt <= w.cfg.Sync.RetryAttempts; attempt++ {
					responses, err := w.client.UploadBatch(batch)
					if err == nil {
						for j, resp := range responses {
							w.state.SyncedSessions[batch[j].ID] = time.Now()
							if len(resp.Warnings) > 0 {
								w.logger.Printf("Session %s: warnings: %v", resp.SessionID, resp.Warnings)
							}
						}
						w.logger.Printf("Uploaded %d sessions", len(batch))
						uploadErr = nil
						break
					}
					uploadErr = err
					w.logger.Printf("Upload attempt %d failed: %v", attempt, err)
					time.Sleep(time.Duration(attempt*2) * time.Second)
				}

				if uploadErr != nil {
					w.logger.Printf("Failed to upload batch after %d attempts", w.cfg.Sync.RetryAttempts)
				}
			}
		}
	}

	// Sync plans
	if err := w.syncPlans(); err != nil {
		w.logger.Printf("Plan sync error: %v", err)
	}

	w.state.LastSync = time.Now()
	return w.saveState()
}

// syncPlans finds and uploads new plans
func (w *Watcher) syncPlans() error {
	plansDir := filepath.Join(w.logsPath, "plans")

	// Check if plans directory exists
	if _, err := os.Stat(plansDir); os.IsNotExist(err) {
		return nil // No plans directory, nothing to sync
	}

	files, err := w.findPlans(plansDir)
	if err != nil {
		return err
	}

	// Filter to new/updated plans
	var newFiles []string
	for _, f := range files {
		name := filepath.Base(strings.TrimSuffix(f, ".md"))
		info, err := os.Stat(f)
		if err != nil {
			continue
		}

		// Check if plan is new or modified since last sync
		lastSynced, synced := w.state.SyncedPlans[name]
		if !synced || info.ModTime().After(lastSynced) {
			newFiles = append(newFiles, f)
		}
	}

	if len(newFiles) == 0 {
		return nil
	}

	w.logger.Printf("Found %d new/updated plans", len(newFiles))

	// Parse and upload plans
	var toUpload []*parser.Plan
	for _, f := range newFiles {
		plan, err := parser.ParsePlan(f)
		if err != nil {
			w.logger.Printf("Error parsing plan %s: %v", f, err)
			continue
		}
		toUpload = append(toUpload, plan)
	}

	if len(toUpload) == 0 {
		return nil
	}

	// Upload in batches of 10
	batchSize := 10
	for i := 0; i < len(toUpload); i += batchSize {
		end := i + batchSize
		if end > len(toUpload) {
			end = len(toUpload)
		}
		batch := toUpload[i:end]

		var uploadErr error
		for attempt := 1; attempt <= w.cfg.Sync.RetryAttempts; attempt++ {
			responses, err := w.client.UploadPlanBatch(batch)
			if err == nil {
				for j, resp := range responses {
					w.state.SyncedPlans[batch[j].Name] = time.Now()
					if len(resp.Warnings) > 0 {
						w.logger.Printf("Plan %s: warnings: %v", resp.Name, resp.Warnings)
					}
				}
				w.logger.Printf("Uploaded %d plans", len(batch))
				uploadErr = nil
				break
			}
			uploadErr = err
			w.logger.Printf("Plan upload attempt %d failed: %v", attempt, err)
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		if uploadErr != nil {
			w.logger.Printf("Failed to upload plan batch after %d attempts", w.cfg.Sync.RetryAttempts)
		}
	}

	return nil
}

// findPlans finds all markdown plan files in the plans directory
func (w *Watcher) findPlans(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't access
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// findSessions finds all JSONL session files in the projects directory
func (w *Watcher) findSessions(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't access
		}
		if !d.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// loadState loads sync state from disk
func (w *Watcher) loadState() error {
	data, err := os.ReadFile(w.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.state = &State{
				SyncedSessions: make(map[string]time.Time),
				SyncedPlans:    make(map[string]time.Time),
			}
			return nil
		}
		return err
	}

	w.state = &State{
		SyncedSessions: make(map[string]time.Time),
		SyncedPlans:    make(map[string]time.Time),
	}
	return json.Unmarshal(data, w.state)
}

// saveState persists sync state to disk
func (w *Watcher) saveState() error {
	// Ensure directory exists
	dir := filepath.Dir(w.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(w.state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(w.statePath, data, 0600)
}

// GetStats returns sync statistics
func (w *Watcher) GetStats() Stats {
	if err := w.loadState(); err != nil {
		return Stats{}
	}

	return Stats{
		TotalSynced:      len(w.state.SyncedSessions),
		TotalPlansSynced: len(w.state.SyncedPlans),
		LastSync:         w.state.LastSync,
	}
}

// Stats contains watcher statistics
type Stats struct {
	TotalSynced      int       `json:"total_synced"`
	TotalPlansSynced int       `json:"total_plans_synced"`
	LastSync         time.Time `json:"last_sync"`
}
