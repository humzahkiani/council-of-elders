package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/humzahkiani/council/internal/types"
)

// Storage handles session persistence to JSON files
type Storage struct {
	baseDir string
}

// New creates a new Storage instance and ensures the sessions directory exists
func New() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".council", "sessions")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &Storage{baseDir: baseDir}, nil
}

// Save persists a session to the default sessions directory
// Returns the file path where the session was saved
func (s *Storage) Save(session *types.Session) (string, error) {
	filename := s.generateFilename(session)
	path := filepath.Join(s.baseDir, filename)
	return path, s.SaveTo(session, path)
}

// SaveTo saves a session to a specific file path
func (s *Storage) SaveTo(session *types.Session, path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load reads a session from a file
func (s *Storage) Load(path string) (*types.Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session types.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// generateFilename creates a filename in the format: YYYY-MM-DD_HHMMSS_{uuid[:6]}.json
func (s *Storage) generateFilename(session *types.Session) string {
	timestamp := session.CreatedAt.Format("2006-01-02_150405")
	shortID := session.ID
	if len(shortID) > 6 {
		shortID = shortID[:6]
	}
	return fmt.Sprintf("%s_%s.json", timestamp, shortID)
}
