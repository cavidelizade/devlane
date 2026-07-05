package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// reloadIssue reads a fresh copy — GORM's First does not reset a reused struct's
// NULL fields, so each assertion needs its own value.
func reloadIssue(t *testing.T, ts *testutil.TestServer, id uuid.UUID) model.Issue {
	t.Helper()
	var got model.Issue
	require.NoError(t, ts.DB.First(&got, "id = ?", id).Error)
	return got
}

func TestIssue_ClearDates(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	issue := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/issues/" + issue.ID.String() + "/"

	// Set both dates.
	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{
		"start_date": "2026-03-15", "target_date": "2026-03-20",
	}, w.Session).Code)
	set := reloadIssue(t, ts, issue.ID)
	require.NotNil(t, set.StartDate)
	require.NotNil(t, set.TargetDate)

	// Clearing with null persists as NULL.
	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{
		"start_date": nil, "target_date": nil,
	}, w.Session).Code)
	cleared := reloadIssue(t, ts, issue.ID)
	require.Nil(t, cleared.StartDate, "start_date should clear to null")
	require.Nil(t, cleared.TargetDate, "target_date should clear to null")

	// Omitting the field leaves it unchanged.
	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{"start_date": "2026-04-01"}, w.Session).Code)
	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{"name": "Renamed"}, w.Session).Code)
	kept := reloadIssue(t, ts, issue.ID)
	require.NotNil(t, kept.StartDate, "omitting start_date must not clear it")
}

func TestCycle_ClearDates(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	cy := testutil.CreateCycle(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/cycles/" + cy.ID.String() + "/"

	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{
		"start_date": "2026-03-15", "end_date": "2026-03-20",
	}, w.Session).Code)
	var set model.Cycle
	require.NoError(t, ts.DB.First(&set, "id = ?", cy.ID).Error)
	require.NotNil(t, set.StartDate)
	require.NotNil(t, set.EndDate)

	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{
		"start_date": nil, "end_date": nil,
	}, w.Session).Code)
	var cleared model.Cycle
	require.NoError(t, ts.DB.First(&cleared, "id = ?", cy.ID).Error)
	require.Nil(t, cleared.StartDate, "cycle start_date should clear")
	require.Nil(t, cleared.EndDate, "cycle end_date should clear")
}

func TestModule_ClearDates(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	mod := testutil.CreateModule(t, ts.DB, w.Project.ID, w.Workspace.ID)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() +
		"/modules/" + mod.ID.String() + "/"

	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{
		"start_date": "2026-03-15", "target_date": "2026-03-20",
	}, w.Session).Code)
	var set model.Module
	require.NoError(t, ts.DB.First(&set, "id = ?", mod.ID).Error)
	require.NotNil(t, set.StartDate)
	require.NotNil(t, set.TargetDate)

	require.Equal(t, http.StatusOK, ts.PATCH(base, map[string]any{
		"start_date": nil, "target_date": nil,
	}, w.Session).Code)
	var cleared model.Module
	require.NoError(t, ts.DB.First(&cleared, "id = ?", mod.ID).Error)
	require.Nil(t, cleared.StartDate, "module start_date should clear")
	require.Nil(t, cleared.TargetDate, "module target_date should clear")
}
