package queue

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if len(id) != 14 { // "c-" + 12 digits
		t.Errorf("unexpected ID length: %d", len(id))
	}
	if id[0:2] != "c-" {
		t.Errorf("ID should start with 'c-': %s", id)
	}
}

func TestStore_ReadWrite(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	queuePath := filepath.Join(tmpDir, "queue.jsonl")
	rejectedPath := filepath.Join(tmpDir, "rejected.jsonl")

	store := NewStore(queuePath, rejectedPath)

	// Test empty read
	candidates, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on empty: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(candidates))
	}

	// Test append
	c := Candidate{
		ID:           "c-test001",
		Type:         CandidateTypePaper,
		Status:       CandidateStatusPending,
		DiscoveredAt: time.Now(),
		DiscoveredBy: "test",
		PaperData: &PaperData{
			S2ID:  "S2:abc123",
			Title: "Test Paper",
		},
	}

	if err := store.Append(c); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Test read after append
	candidates, err = store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].ID != "c-test001" {
		t.Errorf("unexpected ID: %s", candidates[0].ID)
	}
}

func TestStore_FindByID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))

	// Add some candidates
	for i := 0; i < 3; i++ {
		c := Candidate{
			ID:           GenerateID(),
			Type:         CandidateTypePaper,
			Status:       CandidateStatusPending,
			DiscoveredAt: time.Now(),
			DiscoveredBy: "test",
		}
		time.Sleep(time.Millisecond) // Ensure unique IDs
		if err := store.Append(c); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	candidates, _ := store.ReadAll()
	if len(candidates) == 0 {
		t.Fatal("no candidates to test")
	}

	// Find existing
	found, err := store.FindByID(candidates[0].ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil {
		t.Error("expected to find candidate")
	}

	// Find non-existing
	notFound, err := store.FindByID("nonexistent")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existing candidate")
	}
}

func TestStore_Update(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))

	c := Candidate{
		ID:           "c-update-test",
		Type:         CandidateTypePaper,
		Status:       CandidateStatusPending,
		DiscoveredAt: time.Now(),
		DiscoveredBy: "test",
	}
	if err := store.Append(c); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Update status
	c.Status = CandidateStatusApproved
	if err := store.Update(c); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Verify update
	updated, _ := store.FindByID("c-update-test")
	if updated.Status != CandidateStatusApproved {
		t.Errorf("expected approved status, got %s", updated.Status)
	}
}

func TestStore_GetStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))

	// Add candidates with different statuses and types
	candidates := []Candidate{
		{ID: "c-1", Type: CandidateTypePaper, Status: CandidateStatusPending},
		{ID: "c-2", Type: CandidateTypePaper, Status: CandidateStatusApproved},
		{ID: "c-3", Type: CandidateTypeRepo, Status: CandidateStatusPending},
		{ID: "c-4", Type: CandidateTypeCodeLocation, Status: CandidateStatusRejected},
	}

	for _, c := range candidates {
		c.DiscoveredAt = time.Now()
		c.DiscoveredBy = "test"
		if err := store.Append(c); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	if stats.Total != 4 {
		t.Errorf("expected 4 total, got %d", stats.Total)
	}
	if stats.Pending != 2 {
		t.Errorf("expected 2 pending, got %d", stats.Pending)
	}
	if stats.Approved != 1 {
		t.Errorf("expected 1 approved, got %d", stats.Approved)
	}
	if stats.Rejected != 1 {
		t.Errorf("expected 1 rejected, got %d", stats.Rejected)
	}
	if stats.ByType[CandidateTypePaper] != 2 {
		t.Errorf("expected 2 papers, got %d", stats.ByType[CandidateTypePaper])
	}
}

func TestCandidateService_AddAndList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))
	svc := NewCandidateService(store)

	// Add a candidate
	c := Candidate{
		Type:         CandidateTypePaper,
		DiscoveredBy: "test",
		PaperData: &PaperData{
			S2ID:  "S2:test123",
			Title: "Test Paper",
		},
	}

	if err := svc.Add(c); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// List all
	candidates, err := svc.List(nil, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(candidates))
	}

	// Auto-generated ID
	if candidates[0].ID == "" {
		t.Error("expected auto-generated ID")
	}

	// Auto-set status
	if candidates[0].Status != CandidateStatusPending {
		t.Errorf("expected pending status, got %s", candidates[0].Status)
	}
}

func TestCandidateService_ApproveReject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))
	svc := NewCandidateService(store)

	// Add two candidates
	c1 := Candidate{
		ID:           "c-approve-test",
		Type:         CandidateTypePaper,
		Status:       CandidateStatusPending,
		DiscoveredAt: time.Now(),
		DiscoveredBy: "test",
	}
	c2 := Candidate{
		ID:           "c-reject-test",
		Type:         CandidateTypeRepo,
		Status:       CandidateStatusPending,
		DiscoveredAt: time.Now(),
		DiscoveredBy: "test",
		RepoData: &RepoData{
			URL: "https://github.com/test/repo",
		},
	}

	svc.Add(c1)
	svc.Add(c2)

	// Approve first
	if err := svc.Approve("c-approve-test", "reviewer", "looks good"); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	approved, _ := svc.Get("c-approve-test")
	if approved.Status != CandidateStatusApproved {
		t.Errorf("expected approved status, got %s", approved.Status)
	}
	if approved.ReviewedBy == nil || *approved.ReviewedBy != "reviewer" {
		t.Error("expected reviewer name")
	}

	// Reject second
	if err := svc.Reject("c-reject-test", "reviewer", "out of scope"); err != nil {
		t.Fatalf("Reject: %v", err)
	}

	rejected, _ := svc.Get("c-reject-test")
	if rejected.Status != CandidateStatusRejected {
		t.Errorf("expected rejected status, got %s", rejected.Status)
	}
	if rejected.RejectionReason == nil || *rejected.RejectionReason != "out of scope" {
		t.Error("expected rejection reason")
	}

	// Check rejected list
	rejectedList, _ := store.ReadRejected()
	if len(rejectedList) != 1 {
		t.Errorf("expected 1 in rejected list, got %d", len(rejectedList))
	}
}

func TestCandidateService_RejectPreventsReAdd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))
	svc := NewCandidateService(store)

	// Add and reject
	c1 := Candidate{
		ID:           "c-original",
		Type:         CandidateTypePaper,
		Status:       CandidateStatusPending,
		DiscoveredAt: time.Now(),
		DiscoveredBy: "test",
		PaperData: &PaperData{
			S2ID: "S2:unique123",
		},
	}
	svc.Add(c1)
	svc.Reject("c-original", "reviewer", "not relevant")

	// Try to add same paper again
	c2 := Candidate{
		ID:           "c-duplicate",
		Type:         CandidateTypePaper,
		Status:       CandidateStatusPending,
		DiscoveredAt: time.Now(),
		DiscoveredBy: "test",
		PaperData: &PaperData{
			S2ID: "S2:unique123", // Same S2 ID
		},
	}

	err = svc.Add(c2)
	if err == nil {
		t.Error("expected error when adding previously rejected candidate")
	}
}

func TestCandidateService_ListWithFilters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewStore(filepath.Join(tmpDir, "queue.jsonl"), filepath.Join(tmpDir, "rejected.jsonl"))
	svc := NewCandidateService(store)

	// Add various candidates
	candidates := []Candidate{
		{ID: "c-1", Type: CandidateTypePaper, Status: CandidateStatusPending, DiscoveredAt: time.Now(), DiscoveredBy: "test"},
		{ID: "c-2", Type: CandidateTypePaper, Status: CandidateStatusApproved, DiscoveredAt: time.Now(), DiscoveredBy: "test"},
		{ID: "c-3", Type: CandidateTypeRepo, Status: CandidateStatusPending, DiscoveredAt: time.Now(), DiscoveredBy: "test"},
	}
	for _, c := range candidates {
		store.Append(c)
	}

	// Filter by status
	pending := CandidateStatusPending
	filtered, _ := svc.List(&pending, nil)
	if len(filtered) != 2 {
		t.Errorf("expected 2 pending, got %d", len(filtered))
	}

	// Filter by type
	paper := CandidateTypePaper
	filtered, _ = svc.List(nil, &paper)
	if len(filtered) != 2 {
		t.Errorf("expected 2 papers, got %d", len(filtered))
	}

	// Filter by both
	filtered, _ = svc.List(&pending, &paper)
	if len(filtered) != 1 {
		t.Errorf("expected 1 pending paper, got %d", len(filtered))
	}
}
