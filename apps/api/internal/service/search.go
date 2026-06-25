package service

import (
	"context"
	"strings"

	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/google/uuid"
)

// Per-group hit cap for a search response.
const searchPerGroupLimit = 25

// SearchService runs cross-entity search after authorizing the workspace member.
type SearchService struct {
	ss *store.SearchStore
	ws *store.WorkspaceStore
}

func NewSearchService(ss *store.SearchStore, ws *store.WorkspaceStore) *SearchService {
	return &SearchService{ss: ss, ws: ws}
}

// Search resolves the workspace, verifies the caller is a member, then queries.
// A blank query short-circuits to empty results without touching entity tables.
// When projectID is non-nil, results are scoped to that project.
func (s *SearchService) Search(ctx context.Context, workspaceSlug, query string, projectID *uuid.UUID, userID uuid.UUID) (store.SearchResults, error) {
	wrk, err := s.ws.GetBySlug(ctx, workspaceSlug)
	if err != nil {
		return store.EmptyResults(), ErrProjectForbidden
	}
	ok, _ := s.ws.IsMember(ctx, wrk.ID, userID)
	if !ok {
		return store.EmptyResults(), ErrProjectForbidden
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return store.EmptyResults(), nil
	}
	return s.ss.Search(ctx, wrk.ID, q, projectID, searchPerGroupLimit)
}
