package queue

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// CandidateService provides candidate management operations.
type CandidateService struct {
	store *Store
}

// NewCandidateService creates a new CandidateService.
func NewCandidateService(store *Store) *CandidateService {
	return &CandidateService{store: store}
}

// GenerateID generates a new candidate ID in c-YYYYMMDDHHMM format.
func GenerateID() string {
	return fmt.Sprintf("c-%s", time.Now().Format("200601021504"))
}

// Add adds a new candidate to the queue.
func (s *CandidateService) Add(c Candidate) error {
	if c.ID == "" {
		c.ID = GenerateID()
	}
	if c.Status == "" {
		c.Status = CandidateStatusPending
	}
	if c.DiscoveredAt.IsZero() {
		c.DiscoveredAt = time.Now()
	}

	// Check for duplicates in queue
	existing, err := s.store.FindByID(c.ID)
	if err != nil {
		return fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("candidate with ID %s already exists", c.ID)
	}

	// Check if previously rejected (FR-012)
	externalID := getExternalID(c)
	if externalID != "" {
		rejected, err := s.store.IsRejected(externalID, c.Type)
		if err != nil {
			return fmt.Errorf("check rejected: %w", err)
		}
		if rejected {
			return fmt.Errorf("candidate was previously rejected (external ID: %s)", externalID)
		}
	}

	return s.store.Append(c)
}

// List returns all candidates, optionally filtered.
func (s *CandidateService) List(statusFilter *CandidateStatus, typeFilter *CandidateType) ([]Candidate, error) {
	candidates, err := s.store.ReadAll()
	if err != nil {
		return nil, err
	}

	if statusFilter == nil && typeFilter == nil {
		return candidates, nil
	}

	var filtered []Candidate
	for _, c := range candidates {
		if statusFilter != nil && c.Status != *statusFilter {
			continue
		}
		if typeFilter != nil && c.Type != *typeFilter {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered, nil
}

// Get returns a candidate by ID.
func (s *CandidateService) Get(id string) (*Candidate, error) {
	return s.store.FindByID(id)
}

// Approve approves a candidate and triggers appropriate actions.
func (s *CandidateService) Approve(id string, reviewedBy string, notes string) error {
	c, err := s.store.FindByID(id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("candidate not found: %s", id)
	}
	if c.Status != CandidateStatusPending {
		return fmt.Errorf("candidate %s is not pending (status: %s)", id, c.Status)
	}

	now := time.Now()
	c.Status = CandidateStatusApproved
	c.ReviewedAt = &now
	c.ReviewedBy = &reviewedBy
	c.ReviewNotes = &notes

	// Update the store
	if err := s.store.Update(*c); err != nil {
		return err
	}

	// Trigger bipartite add for papers and repos
	if err := triggerBipartiteAdd(*c); err != nil {
		// Log warning but don't fail the approval
		fmt.Printf("Warning: failed to add to bipartite: %v\n", err)
	}

	return nil
}

// Reject rejects a candidate.
func (s *CandidateService) Reject(id string, reviewedBy string, reason string) error {
	c, err := s.store.FindByID(id)
	if err != nil {
		return err
	}
	if c == nil {
		return fmt.Errorf("candidate not found: %s", id)
	}
	if c.Status != CandidateStatusPending {
		return fmt.Errorf("candidate %s is not pending (status: %s)", id, c.Status)
	}

	now := time.Now()
	c.Status = CandidateStatusRejected
	c.ReviewedAt = &now
	c.ReviewedBy = &reviewedBy
	c.RejectionReason = &reason

	// Update the store
	if err := s.store.Update(*c); err != nil {
		return err
	}

	// Also add to rejected.jsonl for re-discovery prevention (FR-012)
	return s.store.AppendRejected(*c)
}

// Stats returns queue statistics.
func (s *CandidateService) Stats() (*QueueStats, error) {
	return s.store.GetStats()
}

// getExternalID returns the external identifier for duplicate checking.
func getExternalID(c Candidate) string {
	switch c.Type {
	case CandidateTypePaper:
		if c.PaperData != nil {
			return c.PaperData.S2ID
		}
	case CandidateTypeRepo:
		if c.RepoData != nil {
			return c.RepoData.URL
		}
	case CandidateTypeCodeLocation:
		if c.CodeLocationData != nil {
			return c.CodeLocationData.PermalinkURL
		}
	}
	return ""
}

// triggerBipartiteAdd adds approved candidates to bipartite.
func triggerBipartiteAdd(c Candidate) error {
	// Check if bip CLI is available
	if _, err := exec.LookPath("bip"); err != nil {
		return fmt.Errorf("bip CLI not found")
	}

	switch c.Type {
	case CandidateTypePaper:
		if c.PaperData != nil {
			cmd := exec.Command("bip", "s2", "add", c.PaperData.S2ID)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("bip s2 add failed: %s", stderr.String())
			}
		}
	case CandidateTypeRepo:
		if c.RepoData != nil {
			cmd := exec.Command("bip", "repo", "add", c.RepoData.URL)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("bip repo add failed: %s", stderr.String())
			}
		}
	}
	return nil
}
