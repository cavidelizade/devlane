package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

func issueIDSet(t *testing.T, rr *httptest.ResponseRecorder) map[string]bool {
	t.Helper()
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	var rows []struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &rows))
	set := make(map[string]bool, len(rows))
	for _, r := range rows {
		set[r.ID] = true
	}
	return set
}

// Drafts must stay out of the active project issue list but show up in the
// draft list, and accepting a draft (is_draft=false) moves it into the active
// list. Covers #123's acceptance criteria.
func TestIssue_DraftsExcludedFromActiveList(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)

	active := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	draft := testutil.CreateIssue(t, ts.DB, w.Project.ID, w.Workspace.ID, w.User.ID)
	require.NoError(t, ts.DB.Model(&model.Issue{}).
		Where("id = ?", draft.ID).Update("is_draft", true).Error)

	listURL := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/issues/"
	draftsURL := "/api/workspaces/" + w.Workspace.Slug + "/draft-issues/"

	// Active list excludes the draft.
	active1 := issueIDSet(t, ts.GET(listURL, w.Session))
	require.True(t, active1[active.ID.String()], "non-draft should be in the active list")
	require.False(t, active1[draft.ID.String()], "draft must not appear in the active list")

	// Draft list includes the draft and not the active issue.
	drafts := issueIDSet(t, ts.GET(draftsURL, w.Session))
	require.True(t, drafts[draft.ID.String()], "draft should appear in the draft list")
	require.False(t, drafts[active.ID.String()], "non-draft should not be in the draft list")

	// Accepting the draft moves it into the active list.
	patchURL := listURL + draft.ID.String() + "/"
	require.Equal(t, http.StatusOK,
		ts.PATCH(patchURL, map[string]any{"is_draft": false}, w.Session).Code)
	active2 := issueIDSet(t, ts.GET(listURL, w.Session))
	require.True(t, active2[draft.ID.String()], "accepted draft should now appear in the active list")
}
