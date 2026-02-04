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
func (s *CandidateService) Add(candidate Candidate) error {
	if candidate.ID == "" {
		candidate.ID = GenerateID()
	}
	if candidate.Status == "" {
		candidate.Status = CandidateStatusPending
	}
	if candidate.DiscoveredAt.IsZero() {
		candidate.DiscoveredAt = time.Now()
	}

	// Check for duplicates in queue
	existing, err := s.store.FindByID(candidate.ID)
	if err != nil {
		return fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("candidate with ID %s already exists", candidate.ID)
	}

	// Check if previously rejected (FR-012)
	externalID := getExternalID(candidate)
	if externalID != "" {
		rejected, err := s.store.IsRejected(externalID, candidate.Type)
		if err != nil {
			return fmt.Errorf("check rejected: %w", err)
		}
		if rejected {
			return fmt.Errorf("candidate was previously rejected (external ID: %s)", externalID)
		}
	}

	return s.store.Append(candidate)
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
	for _, candidate := range candidates {
		if statusFilter != nil && candidate.Status != *statusFilter {
			continue
		}
		if typeFilter != nil && candidate.Type != *typeFilter {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered, nil
}

// Get returns a candidate by ID.
func (s *CandidateService) Get(id string) (*Candidate, error) {
	return s.store.FindByID(id)
}

// Approve approves a candidate and triggers appropriate actions.
func (s *CandidateService) Approve(id string, reviewedBy string, notes string) error {
	candidate, err := s.store.FindByID(id)
	if err != nil {
		return err
	}
	if candidate == nil {
		return fmt.Errorf("candidate not found: %s", id)
	}
	if candidate.Status != CandidateStatusPending {
		return fmt.Errorf("candidate %s is not pending (status: %s)", id, candidate.Status)
	}

	now := time.Now()
	candidate.Status = CandidateStatusApproved
	candidate.ReviewedAt = &now
	candidate.ReviewedBy = &reviewedBy
	candidate.ReviewNotes = &notes

	// Update the store
	if err := s.store.Update(*candidate); err != nil {
		return err
	}

	// Trigger bipartite add for papers and repos
	if err := triggerBipartiteAdd(*candidate); err != nil {
		// Approval succeeded but bipartite integration failed - propagate error
		// The candidate status is already saved, so a retry will find it already approved
		return fmt.Errorf("candidate %s approved but bipartite integration failed: %w", id, err)
	}

	return nil
}

// Reject rejects a candidate.
func (s *CandidateService) Reject(id string, reviewedBy string, reason string) error {
	candidate, err := s.store.FindByID(id)
	if err != nil {
		return err
	}
	if candidate == nil {
		return fmt.Errorf("candidate not found: %s", id)
	}
	if candidate.Status != CandidateStatusPending {
		return fmt.Errorf("candidate %s is not pending (status: %s)", id, candidate.Status)
	}

	now := time.Now()
	candidate.Status = CandidateStatusRejected
	candidate.ReviewedAt = &now
	candidate.ReviewedBy = &reviewedBy
	candidate.RejectionReason = &reason

	// Update the store
	if err := s.store.Update(*candidate); err != nil {
		return err
	}

	// Also add to rejected.jsonl for re-discovery prevention (FR-012)
	return s.store.AppendRejected(*candidate)
}

// Stats returns queue statistics.
func (s *CandidateService) Stats() (*QueueStats, error) {
	return s.store.GetStats()
}

// getExternalID returns the external identifier for duplicate checking.
func getExternalID(candidate Candidate) string {
	switch candidate.Type {
	case CandidateTypePaper:
		if candidate.PaperData != nil {
			return candidate.PaperData.S2ID
		}
	case CandidateTypeRepo:
		if candidate.RepoData != nil {
			return candidate.RepoData.URL
		}
	case CandidateTypeCodeLocation:
		if candidate.CodeLocationData != nil {
			return candidate.CodeLocationData.PermalinkURL
		}
	}
	return ""
}

// triggerBipartiteAdd adds approved candidates to bipartite.
func triggerBipartiteAdd(candidate Candidate) error {
	// Check if bip CLI is available
	if _, err := exec.LookPath("bip"); err != nil {
		return fmt.Errorf("bip CLI not found")
	}

	switch candidate.Type {
	case CandidateTypePaper:
		if candidate.PaperData != nil {
			cmd := exec.Command("bip", "s2", "add", candidate.PaperData.S2ID)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("bip s2 add failed: %s", stderr.String())
			}
		}
	case CandidateTypeRepo:
		if candidate.RepoData != nil {
			cmd := exec.Command("bip", "repo", "add", candidate.RepoData.URL)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("bip repo add failed: %s", stderr.String())
			}
		}
	}
	return nil
}
