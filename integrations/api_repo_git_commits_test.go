// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"testing"

	"code.gitea.io/gitea/models"
	api "code.gitea.io/gitea/modules/structs"

	"github.com/stretchr/testify/assert"
)

func TestAPIReposGitCommits(t *testing.T) {
	defer prepareTestEnv(t)()
	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session)

	//check invalid requests for GetCommitsBySHA
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/git/commits/master?token="+token, user.Name)
	session.MakeRequest(t, req, http.StatusUnprocessableEntity)

	req = NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/git/commits/12345?token="+token, user.Name)
	session.MakeRequest(t, req, http.StatusNotFound)

	//check invalid requests for GetCommitsByRef
	req = NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/commits/..?token="+token, user.Name)
	session.MakeRequest(t, req, http.StatusUnprocessableEntity)

	req = NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/commits/branch-not-exist?token="+token, user.Name)
	session.MakeRequest(t, req, http.StatusNotFound)

	for _, ref := range [...]string{
		"master", // Branch
		"v1.1",   // Tag
		"65f1",   // short sha
		"65f1bf27bc3bf70f64657658635e66094edbcb4d", // full sha
	} {
		req = NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/commits/%s?token="+token, user.Name, ref)
		resp := session.MakeRequest(t, req, http.StatusOK)
		commitByRef := new(api.Commit)
		DecodeJSON(t, resp, commitByRef)
		assert.Len(t, commitByRef.SHA, 40)
		assert.EqualValues(t, commitByRef.SHA, commitByRef.RepoCommit.Tree.SHA)
		req = NewRequestf(t, "GET", "/api/v1/repos/%s/repo1/git/commits/%s?token="+token, user.Name, commitByRef.SHA)
		resp = session.MakeRequest(t, req, http.StatusOK)
		commitBySHA := new(api.Commit)
		DecodeJSON(t, resp, commitBySHA)

		assert.EqualValues(t, commitByRef.SHA, commitBySHA.SHA)
		assert.EqualValues(t, commitByRef.HTMLURL, commitBySHA.HTMLURL)
		assert.EqualValues(t, commitByRef.RepoCommit.Message, commitBySHA.RepoCommit.Message)
	}
}

func TestAPIReposGitCommitList(t *testing.T) {
	defer prepareTestEnv(t)()
	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session)

	// Test getting commits (Page 1)
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/repo16/commits?token="+token, user.Name)
	resp := session.MakeRequest(t, req, http.StatusOK)

	var apiData []api.Commit
	DecodeJSON(t, resp, &apiData)

	assert.Equal(t, 3, len(apiData))
	assert.Equal(t, "69554a64c1e6030f051e5c3f94bfbd773cd6a324", apiData[0].CommitMeta.SHA)
	assert.Equal(t, "27566bd5738fc8b4e3fef3c5e72cce608537bd95", apiData[1].CommitMeta.SHA)
	assert.Equal(t, "5099b81332712fe655e34e8dd63574f503f61811", apiData[2].CommitMeta.SHA)
}

func TestAPIReposGitCommitListPage2Empty(t *testing.T) {
	defer prepareTestEnv(t)()
	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session)

	// Test getting commits (Page=2)
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/repo16/commits?token="+token+"&page=2", user.Name)
	resp := session.MakeRequest(t, req, http.StatusOK)

	var apiData []api.Commit
	DecodeJSON(t, resp, &apiData)

	assert.Equal(t, 0, len(apiData))
}

func TestAPIReposGitCommitListDifferentBranch(t *testing.T) {
	defer prepareTestEnv(t)()
	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session)

	// Test getting commits (Page=1, Branch=good-sign)
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/repo16/commits?token="+token+"&sha=good-sign", user.Name)
	resp := session.MakeRequest(t, req, http.StatusOK)

	var apiData []api.Commit
	DecodeJSON(t, resp, &apiData)

	assert.Equal(t, 1, len(apiData))
	assert.Equal(t, "f27c2b2b03dcab38beaf89b0ab4ff61f6de63441", apiData[0].CommitMeta.SHA)
}
