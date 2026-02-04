package queue

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultQueuePath is the default path for the candidate queue.
const DefaultQueuePath = ".candidates/queue.jsonl"

// DefaultRejectedPath is the default path for rejected candidates.
const DefaultRejectedPath = ".candidates/rejected.jsonl"

// Store provides JSONL storage operations for candidates.
type Store struct {
	queuePath    string
	rejectedPath string
}

// NewStore creates a new Store with the given paths.
// If paths are empty, defaults are used.
func NewStore(queuePath, rejectedPath string) *Store {
	if queuePath == "" {
		queuePath = DefaultQueuePath
	}
	if rejectedPath == "" {
		rejectedPath = DefaultRejectedPath
	}
	return &Store{
		queuePath:    queuePath,
		rejectedPath: rejectedPath,
	}
}

// ensureDir creates the directory for a file path if it doesn't exist.
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

// ReadAll reads all candidates from the queue file.
func (s *Store) ReadAll() ([]Candidate, error) {
	return readJSONL[Candidate](s.queuePath)
}

// ReadRejected reads all rejected candidates.
func (s *Store) ReadRejected() ([]Candidate, error) {
	return readJSONL[Candidate](s.rejectedPath)
}

// Append appends a candidate to the queue file.
func (s *Store) Append(c Candidate) error {
	return appendJSONL(s.queuePath, c)
}

// AppendRejected appends a candidate to the rejected file.
func (s *Store) AppendRejected(c Candidate) error {
	return appendJSONL(s.rejectedPath, c)
}

// WriteAll writes all candidates to the queue file, overwriting existing content.
func (s *Store) WriteAll(candidates []Candidate) error {
	return writeJSONL(s.queuePath, candidates)
}

// FindByID finds a candidate by ID in the queue.
func (s *Store) FindByID(id string) (*Candidate, error) {
	candidates, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, c := range candidates {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

// Update updates a candidate in the queue.
func (s *Store) Update(c Candidate) error {
	candidates, err := s.ReadAll()
	if err != nil {
		return err
	}

	found := false
	for i, existing := range candidates {
		if existing.ID == c.ID {
			candidates[i] = c
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("candidate not found: %s", c.ID)
	}

	return s.WriteAll(candidates)
}

// IsRejected checks if a candidate with the given external ID has been previously rejected.
// This is used to prevent re-discovery of rejected items per FR-012.
func (s *Store) IsRejected(externalID string, candidateType CandidateType) (bool, error) {
	rejected, err := s.ReadRejected()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	for _, c := range rejected {
		if c.Type != candidateType {
			continue
		}
		// Check type-specific external IDs
		switch candidateType {
		case CandidateTypePaper:
			if c.PaperData != nil && c.PaperData.S2ID == externalID {
				return true, nil
			}
		case CandidateTypeRepo:
			if c.RepoData != nil && c.RepoData.URL == externalID {
				return true, nil
			}
		case CandidateTypeCodeLocation:
			if c.CodeLocationData != nil && c.CodeLocationData.PermalinkURL == externalID {
				return true, nil
			}
		}
	}
	return false, nil
}

// GetStats returns statistics about the queue.
func (s *Store) GetStats() (*QueueStats, error) {
	candidates, err := s.ReadAll()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &QueueStats{ByType: make(map[CandidateType]int)}, nil
		}
		return nil, err
	}

	stats := &QueueStats{
		ByType: make(map[CandidateType]int),
	}

	for _, c := range candidates {
		stats.Total++
		stats.ByType[c.Type]++
		switch c.Status {
		case CandidateStatusPending:
			stats.Pending++
		case CandidateStatusApproved:
			stats.Approved++
		case CandidateStatusRejected:
			stats.Rejected++
		}
	}

	return stats, nil
}

// readJSONL reads a JSONL file and returns a slice of items.
func readJSONL[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []T{}, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var items []T
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			return nil, fmt.Errorf("parse line %d in %s: %w", lineNum, path, err)
		}
		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return items, nil
}

// appendJSONL appends an item to a JSONL file.
func appendJSONL[T any](path string, item T) error {
	if err := ensureDir(path); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write to %s: %w", path, err)
	}

	return nil
}

// writeJSONL writes a slice of items to a JSONL file, overwriting existing content.
func writeJSONL[T any](path string, items []T) error {
	if err := ensureDir(path); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()

	for i, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshal item %d: %w", i, err)
		}
		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write item %d to %s: %w", i, path, err)
		}
	}

	return nil
}
