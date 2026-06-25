package handler_test

import (
	"net/http"
	"testing"

	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestModule_Members(t *testing.T) {
	ts := testutil.NewTestServer(t)
	w := testutil.SeedWorld(t, ts.DB)
	base := "/api/workspaces/" + w.Workspace.Slug + "/projects/" + w.Project.ID.String() + "/modules/"
	uid := w.User.ID.String()

	// Create with a lead and one member.
	rr := ts.POST(base, map[string]any{
		"name":       "Sprint M",
		"lead_id":    uid,
		"member_ids": []string{uid},
	}, w.Session)
	require.Equal(t, http.StatusCreated, rr.Code, "body=%s", rr.Body.String())
	created := testutil.MustJSONMap(t, rr)
	id, _ := created["id"].(string)
	require.NotEmpty(t, id)
	require.Equal(t, uid, created["lead_id"])
	members, _ := created["member_ids"].([]any)
	require.Len(t, members, 1)
	require.Equal(t, uid, members[0])

	// Get returns the member.
	got := ts.GET(base+id+"/", w.Session)
	require.Equal(t, http.StatusOK, got.Code, "body=%s", got.Body.String())
	gm, _ := testutil.MustJSONMap(t, got)["member_ids"].([]any)
	require.Len(t, gm, 1)

	// Update clears members.
	upd := ts.PATCH(base+id+"/", map[string]any{"member_ids": []string{}}, w.Session)
	require.Equal(t, http.StatusOK, upd.Code, "body=%s", upd.Body.String())
	um, _ := testutil.MustJSONMap(t, upd)["member_ids"].([]any)
	require.Len(t, um, 0)
}
