// Package queue implements candidate queue management for the scribe CLI.
package queue

import "time"

// CandidateType represents the type of a candidate.
type CandidateType string

const (
	CandidateTypePaper        CandidateType = "paper"
	CandidateTypeConcept      CandidateType = "concept"
	CandidateTypeRepo         CandidateType = "repo"
	CandidateTypeCodeLocation CandidateType = "code-location"
)

// CandidateStatus represents the review status of a candidate.
type CandidateStatus string

const (
	CandidateStatusPending  CandidateStatus = "pending"
	CandidateStatusApproved CandidateStatus = "approved"
	CandidateStatusRejected CandidateStatus = "rejected"
)

// Candidate represents a discovered item awaiting human review.
type Candidate struct {
	ID               string          `json:"id"`
	Type             CandidateType   `json:"type"`
	Status           CandidateStatus `json:"status"`
	DiscoveredAt     time.Time       `json:"discovered_at"`
	DiscoveredBy     string          `json:"discovered_by"`
	DiscoveryContext string          `json:"discovery_context"`

	// Type-specific data (one of these based on type)
	PaperData        *PaperData        `json:"paper_data,omitempty"`
	ConceptData      *ConceptData      `json:"concept_data,omitempty"`
	RepoData         *RepoData         `json:"repo_data,omitempty"`
	CodeLocationData *CodeLocationData `json:"code_location_data,omitempty"`

	// Review metadata (populated on approve/reject)
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy      *string    `json:"reviewed_by,omitempty"`
	ReviewNotes     *string    `json:"review_notes,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
}

// PaperData contains paper-specific candidate data.
type PaperData struct {
	S2ID           string   `json:"s2_id"`
	Title          string   `json:"title"`
	Authors        []string `json:"authors"`
	Year           int      `json:"year"`
	RelevanceNotes string   `json:"relevance_notes"`
}

// ConceptData contains concept-specific candidate data.
type ConceptData struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	RelatedPapers []string `json:"related_papers"` // S2 IDs
	RelatedRepos  []string `json:"related_repos"`  // Repo URLs
}

// RepoData contains repository-specific candidate data.
type RepoData struct {
	URL            string `json:"url"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	RelevanceNotes string `json:"relevance_notes"`
}

// CodeLocationData contains code location-specific candidate data.
type CodeLocationData struct {
	RepoURL            string  `json:"repo_url"`
	FilePath           string  `json:"file_path"`
	StartLine          int     `json:"start_line"`
	EndLine            int     `json:"end_line"`
	CommitSHA          string  `json:"commit_sha"`
	PermalinkURL       string  `json:"permalink_url"`
	FunctionName       *string `json:"function_name,omitempty"`
	Description        string  `json:"description"`
	SurroundingContext string  `json:"surrounding_context"` // ~10 lines around the location
}

// QueueStats contains statistics about the candidate queue.
type QueueStats struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`

	ByType map[CandidateType]int `json:"by_type"`
}
