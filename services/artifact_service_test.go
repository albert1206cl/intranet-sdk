package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hujia-team/intranet-sdk/client"
	"github.com/hujia-team/intranet-sdk/models"
)

func newArtifactTestService(t *testing.T, handler http.HandlerFunc) *artifactService {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	httpClient, err := client.NewHTTPClient(&client.Config{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("new http client: %v", err)
	}
	return NewArtifactService(httpClient).(*artifactService)
}

func decodeBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	defer r.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return payload
}

func TestListArtifacts(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := decodeBody(t, r)
		if payload["page"].(float64) != 1 || payload["pageSize"].(float64) != 20 {
			t.Fatalf("unexpected payload: %#v", payload)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"total":1,"data":[{"id":12,"name":"artifact-a"}]}}`))
	})

	result, err := service.ListArtifacts(&models.ArtifactListReq{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListArtifacts error: %v", err)
	}
	if result.Total != 1 || result.Data[0].Name == nil || *result.Data[0].Name != "artifact-a" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestGetArtifactByName(t *testing.T) {
	call := 0
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		call++
		switch r.URL.Path {
		case "/aiplorer/artifact/list":
			payload := decodeBody(t, r)
			if payload["name"].(string) != "artifact-a" {
				t.Fatalf("unexpected name payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"total":1,"data":[{"id":12,"name":"artifact-a"}]}}`))
		case "/aiplorer/artifact":
			payload := decodeBody(t, r)
			if payload["id"].(float64) != 12 {
				t.Fatalf("unexpected id payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","projectName":"proj-a"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	result, err := service.GetArtifactByName("artifact-a", nil)
	if err != nil {
		t.Fatalf("GetArtifactByName error: %v", err)
	}
	if call != 2 || result.ID == nil || *result.ID != 12 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestGetArtifactByCommitHashAndChildHashes(t *testing.T) {
	call := 0
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		call++
		switch r.URL.Path {
		case "/aiplorer/artifact/by-commit-hash":
			payload := decodeBody(t, r)
			if payload["commitHash"].(string) != "root-hash" {
				t.Fatalf("unexpected commitHash payload: %#v", payload)
			}
			if payload["type"].(string) != "pkg" {
				t.Fatalf("unexpected type payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"pkg","commitHash":"root-hash","dependencies":[{"id":21,"name":"child-a","type":"pkg","commitHash":"child-hash-a","parentId":12},{"id":22,"name":"child-b","type":"mcu","commitHash":"child-hash-b","parentId":21}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	artifact, err := service.GetArtifactByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	})
	if err != nil {
		t.Fatalf("GetArtifactByCommitHash error: %v", err)
	}
	if artifact.ID == nil || *artifact.ID != 12 {
		t.Fatalf("unexpected artifact: %#v", artifact)
	}

	childHashes, err := service.GetChildArtifactHashesByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	})
	if err != nil {
		t.Fatalf("GetChildArtifactHashesByCommitHash error: %v", err)
	}
	if len(childHashes.ChildHashes) != 2 {
		t.Fatalf("unexpected child hashes: %#v", childHashes)
	}
	if childHashes.ChildHashes[0].CommitHash == nil || *childHashes.ChildHashes[0].CommitHash != "child-hash-a" {
		t.Fatalf("unexpected first child hash: %#v", childHashes.ChildHashes[0])
	}
	if call != 2 {
		t.Fatalf("unexpected call count: %d", call)
	}
}

func TestGetArtifactByCommitHashAllowsExplicitEmptyPlatform(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/by-commit-hash" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := decodeBody(t, r)
		if payload["commitHash"].(string) != "root-hash" {
			t.Fatalf("unexpected commitHash payload: %#v", payload)
		}
		platform, ok := payload["platform"]
		if !ok {
			t.Fatalf("expected platform field to be present, payload: %#v", payload)
		}
		if platform.(string) != "" {
			t.Fatalf("expected empty platform payload, got %#v", platform)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"app","commitHash":"root-hash","platform":""}}`))
	})

	emptyPlatform := ""
	artifact, err := service.GetArtifactByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "app",
		Platform:     &emptyPlatform,
	})
	if err != nil {
		t.Fatalf("GetArtifactByCommitHash error: %v", err)
	}
	if artifact.ID == nil || *artifact.ID != 12 {
		t.Fatalf("unexpected artifact: %#v", artifact)
	}
}

func TestCheckExistsByCommitHashAllowsExplicitEmptyPlatform(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/by-commit-hash" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := decodeBody(t, r)
		if payload["commitHash"].(string) != "root-hash" {
			t.Fatalf("unexpected commitHash payload: %#v", payload)
		}
		platform, ok := payload["platform"]
		if !ok {
			t.Fatalf("expected platform field to be present, payload: %#v", payload)
		}
		if platform.(string) != "" {
			t.Fatalf("expected empty platform payload, got %#v", platform)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"app","commitHash":"root-hash","platform":"","fullPath":"repo/path/artifact.tar.gz","fileHash":"abc123"}}`))
	})

	emptyPlatform := ""
	exists, err := service.CheckExistsByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "app",
		Platform:     &emptyPlatform,
	})
	if err != nil {
		t.Fatalf("CheckExistsByCommitHash error: %v", err)
	}
	if !exists {
		t.Fatal("expected artifact to exist")
	}
}

func TestCheckExistsByCommitHashReturnsFalseForVirtualArtifact(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/by-commit-hash" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"app","commitHash":"root-hash"}}`))
	})

	exists, err := service.CheckExistsByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "app",
	})
	if err != nil {
		t.Fatalf("CheckExistsByCommitHash error: %v", err)
	}
	if exists {
		t.Fatal("expected virtual artifact without fullPath to be treated as non-existent")
	}
}

func TestCheckExistsByCommitHashReturnsFalseWithoutFileHash(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/by-commit-hash" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"app","commitHash":"root-hash","fullPath":"repo/path/artifact.tar.gz"}}`))
	})

	exists, err := service.CheckExistsByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "app",
	})
	if err != nil {
		t.Fatalf("CheckExistsByCommitHash error: %v", err)
	}
	if exists {
		t.Fatal("expected artifact without fileHash to be treated as non-existent")
	}
}

func TestCheckExistsByCommitHashReturnsFalseWhenArtifactMissing(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/by-commit-hash" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":500,"msg":"artifact not found by commit hash: root-hash"}`))
	})

	exists, err := service.CheckExistsByCommitHash("root-hash", &models.ArtifactLookupOptions{
		ArtifactType: "app",
	})
	if err != nil {
		t.Fatalf("CheckExistsByCommitHash error: %v", err)
	}
	if exists {
		t.Fatal("expected missing artifact to return false")
	}
}

func TestCheckExistsPrepareDownloadAndVersionMetadata(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/artifact":
			payload := decodeBody(t, r)
			if payload["id"].(float64) != 12 {
				t.Fatalf("unexpected artifact id payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"pkg","commitHash":"root-hash","projectName":"proj-a","fileHash":"abc123","fullPath":"repo/path/artifact.zip"}}`))
		case "/aiplorer/artifact/by-commit-hash":
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"pkg","commitHash":"root-hash","projectName":"proj-a","fileHash":"abc123","fullPath":"repo/path/artifact.zip"}}`))
		case "/aiplorer/jfrog/token":
			_, _ = w.Write([]byte(`{"code":0,"data":{"token_id":"tid","access_token":"token","expires_in":3600,"token_type":"Bearer","scope":"scope","url":"https://jfrog.example.com"}}`))
		case "/aiplorer/artifact/download-url":
			_, _ = w.Write([]byte(`{"code":0,"data":{"downloadUrl":"https://jfrog.example.com/file","expireTime":"2026-03-18 12:00:00","fileName":"artifact.zip","filePath":"repo/path/artifact.zip"}}`))
		case "/aiplorer/artifact/version-metadata":
			_, _ = w.Write([]byte(`{"code":0,"data":{"artifactId":12,"commitHash":"root-hash","name":"artifact-a","metadataPath":"repo/path/version.json","metadataFileName":"version.json","rawContent":"{\"version\":\"1.2.3\"}"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	service.downloadArtifact = func(token *models.JfrogTokenInfo, filePath, targetDir string) error {
		if token.AccessToken != "token" || filePath != "repo/path/artifact.zip" {
			t.Fatalf("unexpected download args: %#v %s %s", token, filePath, targetDir)
		}
		return nil
	}

	exists, err := service.CheckExistsByCommitHash("root-hash", &models.ArtifactLookupOptions{ArtifactType: "pkg"})
	if err != nil || !exists {
		t.Fatalf("CheckExistsByCommitHash = %v, %v", exists, err)
	}

	plan, err := service.PrepareDownloadByCommitHash("root-hash", &models.ArtifactLookupOptions{ArtifactType: "pkg"}, "downloads/")
	if err != nil {
		t.Fatalf("PrepareDownloadByCommitHash error: %v", err)
	}
	if plan.TargetPath != "downloads/artifact.zip" {
		t.Fatalf("unexpected target path: %#v", plan)
	}

	plan, err = service.PrepareDownloadByArtifactID(12, "downloads/")
	if err != nil {
		t.Fatalf("PrepareDownloadByArtifactID error: %v", err)
	}
	if plan.TargetPath != "downloads/artifact.zip" {
		t.Fatalf("unexpected artifact-id target path: %#v", plan)
	}

	plan, err = service.DownloadByCommitHash("root-hash", &models.ArtifactLookupOptions{ArtifactType: "pkg"}, "downloads/")
	if err != nil {
		t.Fatalf("DownloadByCommitHash error: %v", err)
	}
	if plan.DownloadURL == nil || plan.DownloadURL.FileName != "artifact.zip" {
		t.Fatalf("unexpected download plan: %#v", plan)
	}

	plan, err = service.DownloadByArtifactID(12, "downloads/")
	if err != nil {
		t.Fatalf("DownloadByArtifactID error: %v", err)
	}
	if plan.DownloadURL == nil || plan.DownloadURL.FileName != "artifact.zip" {
		t.Fatalf("unexpected artifact-id download plan: %#v", plan)
	}

	metadata, err := service.GetVersionMetadataByCommitHash("root-hash", &models.ArtifactLookupOptions{ArtifactType: "pkg"})
	if err != nil {
		t.Fatalf("GetVersionMetadataByCommitHash error: %v", err)
	}
	if metadata.Parsed["version"].(string) != "1.2.3" {
		t.Fatalf("unexpected parsed metadata: %#v", metadata.Parsed)
	}
}

func TestDownloadByCommitHashSkipsExistingFileWithMatchingHash(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/artifact/by-commit-hash":
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","type":"pkg","commitHash":"root-hash","projectName":"proj-a","fileHash":"5d41402abc4b2a76b9719d911017c592","fullPath":"repo/path/artifact.zip"}}`))
		case "/aiplorer/jfrog/token":
			_, _ = w.Write([]byte(`{"code":0,"data":{"token_id":"tid","access_token":"token","expires_in":3600,"token_type":"Bearer","scope":"scope","url":"https://jfrog.example.com"}}`))
		case "/aiplorer/artifact/download-url":
			_, _ = w.Write([]byte(`{"code":0,"data":{"downloadUrl":"https://jfrog.example.com/file","expireTime":"2026-03-18 12:00:00","fileName":"artifact.zip","filePath":"repo/path/artifact.zip"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	called := false
	service.downloadArtifact = func(token *models.JfrogTokenInfo, filePath, targetDir string) error {
		called = true
		return nil
	}

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "artifact.zip")
	if err := os.WriteFile(targetPath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	plan, err := service.DownloadByCommitHash("root-hash", &models.ArtifactLookupOptions{ArtifactType: "pkg"}, tempDir)
	if err != nil {
		t.Fatalf("DownloadByCommitHash error: %v", err)
	}
	if called {
		t.Fatal("expected downloadArtifact not to be called")
	}
	if !plan.SkippedExisting {
		t.Fatalf("expected skipped existing plan: %#v", plan)
	}
}

func TestGetArtifactTagSchemaJSONAndUpdateTags(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/artifact/tag-schema":
			_, _ = w.Write([]byte(`{"code":0,"data":{"version":"0.2.0","content":"{\"version\":\"0.2.0\",\"type\":\"object\"}"}}`))
		case "/aiplorer/artifact/update":
			payload := decodeBody(t, r)
			if payload["id"].(float64) != 12 {
				t.Fatalf("unexpected update payload: %#v", payload)
			}
			if !strings.Contains(payload["tags"].(string), `"schema_version":"0.2.0"`) {
				t.Fatalf("unexpected tags payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"msg":"updated"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	schema, err := service.GetArtifactTagSchemaJSON("0.2.0")
	if err != nil {
		t.Fatalf("GetArtifactTagSchemaJSON error: %v", err)
	}
	if schema["version"].(string) != "0.2.0" {
		t.Fatalf("unexpected schema: %#v", schema)
	}

	result, err := service.UpdateArtifactTags(12, map[string]any{
		"schema_version": "0.2.0",
		"decision":       "pass",
	}, "")
	if err != nil {
		t.Fatalf("UpdateArtifactTags error: %v", err)
	}
	if result.Msg != "updated" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestGetArtifactChangelogAndFiles(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/artifact/commit-diff-v2":
			payload := decodeBody(t, r)
			if payload["artifactIdA"].(float64) != 10 || payload["artifactIdB"].(float64) != 12 {
				t.Fatalf("unexpected changelog payload: %#v", payload)
			}
			if payload["page"].(float64) != 2 || payload["pageSize"].(float64) != 50 {
				t.Fatalf("unexpected pagination payload: %#v", payload)
			}
			if payload["repositoryId"].(float64) != 7 {
				t.Fatalf("unexpected repository payload: %#v", payload)
			}
			mode, ok := payload["compareMode"].(string)
			if !ok || mode != "distant" {
				t.Fatalf("unexpected compare mode payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"olderArtifactId":10,"newerArtifactId":12,"changedRepoCount":1,"unavailableRepoCount":1,"unavailableRepos":[{"repositoryId":9,"reason":"missing compare cache"}],"total":1,"data":[{"id":33,"gitlabProjectId":"100","repositoryId":7,"repositoryName":"repo-a","commitId":"abcdef","commitShortId":"abcdef0","title":"fix bug","authorName":"Alice","committedAt":1770000000,"webUrl":"https://git.example.com/commit/abcdef"}]}}`))
		case "/aiplorer/artifact/commit-diff-files":
			payload := decodeBody(t, r)
			if payload["gitCommitId"].(float64) != 33 {
				t.Fatalf("unexpected files payload: %#v", payload)
			}
			if payload["page"].(float64) != 1 || payload["pageSize"].(float64) != 20 {
				t.Fatalf("unexpected files pagination payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"total":1,"data":[{"id":44,"gitCommitId":33,"oldPath":"old.go","newPath":"new.go","newFile":false,"renamedFile":true,"deletedFile":false}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	page, pageSize, repositoryID := uint64(2), uint64(50), uint64(7)
	changelog, err := service.GetArtifactChangelog(10, 12, &models.ArtifactChangelogOptions{
		Page:         &page,
		PageSize:     &pageSize,
		RepositoryID: &repositoryID,
		CompareMode:  "distant",
	})
	if err != nil {
		t.Fatalf("GetArtifactChangelog error: %v", err)
	}
	if changelog.Total != 1 || len(changelog.Data) != 1 {
		t.Fatalf("unexpected changelog: %#v", changelog)
	}
	if changelog.Data[0].CommitID == nil || *changelog.Data[0].CommitID != "abcdef" {
		t.Fatalf("unexpected changelog commit: %#v", changelog.Data[0])
	}
	if changelog.UnavailableRepos[0].Reason == nil || *changelog.UnavailableRepos[0].Reason != "missing compare cache" {
		t.Fatalf("unexpected unavailable repos: %#v", changelog.UnavailableRepos)
	}

	filesPage, filesPageSize := uint64(1), uint64(20)
	files, err := service.GetArtifactChangelogFiles(33, &models.ArtifactChangelogFilesOptions{
		Page:     &filesPage,
		PageSize: &filesPageSize,
	})
	if err != nil {
		t.Fatalf("GetArtifactChangelogFiles error: %v", err)
	}
	if files.Total != 1 || len(files.Data) != 1 {
		t.Fatalf("unexpected files: %#v", files)
	}
	if files.Data[0].NewPath == nil || *files.Data[0].NewPath != "new.go" {
		t.Fatalf("unexpected file info: %#v", files.Data[0])
	}
}

func TestGetJfrogTokenAndDownloadURL(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/jfrog/token":
			_, _ = w.Write([]byte(`{"code":0,"data":{"token_id":"tid","access_token":"token","expires_in":3600,"token_type":"Bearer","scope":"scope","url":"https://jfrog.example.com"}}`))
		case "/aiplorer/artifact/download-url":
			payload := decodeBody(t, r)
			if payload["downloadType"].(string) != "artifact" {
				t.Fatalf("unexpected download payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"downloadUrl":"https://jfrog.example.com/file","expireTime":"2026-03-18 12:00:00","fileName":"artifact.zip","filePath":"repo/path/artifact.zip"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	token, err := service.GetJfrogToken("proj-a")
	if err != nil {
		t.Fatalf("GetJfrogToken error: %v", err)
	}
	if token.AccessToken != "token" {
		t.Fatalf("unexpected token: %#v", token)
	}

	download, err := service.GetArtifactDownloadURL(12, "")
	if err != nil {
		t.Fatalf("GetArtifactDownloadURL error: %v", err)
	}
	if download.FileName != "artifact.zip" {
		t.Fatalf("unexpected download: %#v", download)
	}
}

func TestDownloadByNameUsesResolvedArtifactID(t *testing.T) {
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/artifact/list":
			_, _ = w.Write([]byte(`{"code":0,"data":{"total":1,"data":[{"id":12,"name":"artifact-a"}]}}`))
		case "/aiplorer/artifact":
			payload := decodeBody(t, r)
			if payload["id"].(float64) != 12 {
				t.Fatalf("unexpected artifact id payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"name":"artifact-a","projectName":"proj-a","fileHash":"abc123","fullPath":"repo/path/artifact.zip"}}`))
		case "/aiplorer/jfrog/token":
			_, _ = w.Write([]byte(`{"code":0,"data":{"token_id":"tid","access_token":"token","expires_in":3600,"token_type":"Bearer","scope":"scope","url":"https://jfrog.example.com"}}`))
		case "/aiplorer/artifact/download-url":
			payload := decodeBody(t, r)
			if payload["artifactId"].(float64) != 12 {
				t.Fatalf("unexpected download payload: %#v", payload)
			}
			_, _ = w.Write([]byte(`{"code":0,"data":{"downloadUrl":"https://jfrog.example.com/file","expireTime":"2026-03-18 12:00:00","fileName":"artifact.zip","filePath":"repo/path/artifact.zip"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	service.downloadArtifact = func(token *models.JfrogTokenInfo, filePath, targetDir string) error {
		if token.AccessToken != "token" || filePath != "repo/path/artifact.zip" {
			t.Fatalf("unexpected download args: %#v %s %s", token, filePath, targetDir)
		}
		return nil
	}

	plan, err := service.DownloadByName("artifact-a", nil, "downloads/")
	if err != nil {
		t.Fatalf("DownloadByName error: %v", err)
	}
	if plan.DownloadURL == nil || plan.DownloadURL.FileName != "artifact.zip" {
		t.Fatalf("unexpected download plan: %#v", plan)
	}
}

func TestGetArtifactCommitDiffCompareMode(t *testing.T) {
	diffResponse := []byte(`{"code":0,"data":{"olderArtifactId":10,"newerArtifactId":12,"changedRepoCount":1,"repoDiffs":[{"repositoryId":7,"repositoryName":"repo-a","commits":[{"id":33,"commitHash":"abcdef"}]}]}}`)

	t.Run("default omits compare mode", func(t *testing.T) {
		service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/aiplorer/artifact/commit-diff" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			payload := decodeBody(t, r)
			if payload["artifactIdA"].(float64) != 10 || payload["artifactIdB"].(float64) != 12 {
				t.Fatalf("unexpected payload: %#v", payload)
			}
			if _, ok := payload["compareMode"]; ok {
				t.Fatalf("unexpected compare mode payload: %#v", payload)
			}
			_, _ = w.Write(diffResponse)
		})

		diff, err := service.GetArtifactCommitDiff(10, 12)
		if err != nil {
			t.Fatalf("GetArtifactCommitDiff error: %v", err)
		}
		if diff.ChangedRepoCount != 1 || len(diff.RepoDiffs) != 1 {
			t.Fatalf("unexpected diff: %#v", diff)
		}
		if diff.RepoDiffs[0].Commits[0].CommitHash == nil || *diff.RepoDiffs[0].Commits[0].CommitHash != "abcdef" {
			t.Fatalf("unexpected commits: %#v", diff.RepoDiffs[0].Commits)
		}
	})

	t.Run("distant passes compare mode", func(t *testing.T) {
		service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
			payload := decodeBody(t, r)
			mode, ok := payload["compareMode"].(string)
			if !ok || mode != "distant" {
				t.Fatalf("unexpected compare mode payload: %#v", payload)
			}
			_, _ = w.Write(diffResponse)
		})

		if _, err := service.GetArtifactCommitDiff(10, 12, &models.ArtifactCommitDiffOptions{CompareMode: "distant"}); err != nil {
			t.Fatalf("GetArtifactCommitDiff error: %v", err)
		}
	})

	t.Run("empty compare mode stays omitted", func(t *testing.T) {
		service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
			payload := decodeBody(t, r)
			if _, ok := payload["compareMode"]; ok {
				t.Fatalf("unexpected compare mode payload: %#v", payload)
			}
			_, _ = w.Write(diffResponse)
		})

		if _, err := service.GetArtifactCommitDiff(10, 12, nil, &models.ArtifactCommitDiffOptions{}); err != nil {
			t.Fatalf("GetArtifactCommitDiff error: %v", err)
		}
	})
}

const flatTagSchemaContent = `{"schema_version":"0.3.0","label_mode":"flat","vocab":{"dim":{"pass_rate":"ratio","state":"bool"}},"buckets":{"ratio":["0.8-0.9","0.9-1.0"]},"schema":{}}`
const legacyTagSchemaContent = `{"schema_version":"0.2.0","schema":{}}`

func tagSchemaHandler(t *testing.T, version, content string, extra func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aiplorer/artifact/tag-schema":
			_, _ = w.Write([]byte(`{"code":0,"data":{"version":"` + version + `","content":` + mustJSON(content) + `}}`))
		case "/aiplorer/artifact":
			_, _ = w.Write([]byte(`{"code":0,"data":{"id":12,"tagSchemaVersion":"` + version + `","tags":` + mustJSON(content) + `}}`))
		default:
			if extra != nil {
				extra(w, r)
				return
			}
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}
}

func TestGetArtifactLabelModeDetectsFlatAndLegacy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		version  string
		content  string
		wantMode string
		wantFlat bool
	}{
		{"flat schema", "0.3.0", flatTagSchemaContent, models.LabelModeFlat, true},
		{"legacy schema without label_mode", "0.2.0", legacyTagSchemaContent, models.LabelModeLegacy, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			service := newArtifactTestService(t, tagSchemaHandler(t, tc.version, tc.content, nil))

			mode, err := service.GetArtifactLabelMode(tc.version)
			if err != nil {
				t.Fatalf("GetArtifactLabelMode error: %v", err)
			}
			if mode != tc.wantMode {
				t.Fatalf("label mode = %q, want %q", mode, tc.wantMode)
			}

			flat, err := service.UsesFlatLabels(tc.version)
			if err != nil {
				t.Fatalf("UsesFlatLabels error: %v", err)
			}
			if flat != tc.wantFlat {
				t.Fatalf("UsesFlatLabels = %v, want %v", flat, tc.wantFlat)
			}
		})
	}
}

func TestParseArtifactTagsChecksSchemaVersionFromDocument(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, tagSchemaHandler(t, "0.3.0", flatTagSchemaContent, nil))

	// The schema document carries "schema_version"; a payload that matches it is
	// accepted even though the document has no "version" field.
	if _, err := service.ParseArtifactTags(
		`{"schema_version":"0.3.0","labels":{"gate.L2":"PASS"}}`,
		models.ArtifactTagSchemaInfo{Version: "0.3.0", Content: flatTagSchemaContent},
	); err != nil {
		t.Fatalf("ParseArtifactTags rejected matching version: %v", err)
	}

	if _, err := service.ParseArtifactTags(
		`{"schema_version":"9.9.9","labels":{}}`,
		models.ArtifactTagSchemaInfo{Version: "0.3.0", Content: flatTagSchemaContent},
	); err == nil {
		t.Fatal("ParseArtifactTags accepted a mismatched schema_version")
	}
}

func TestPatchArtifactTagsPostsMergePatchPayload(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/patch-tags" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := decodeBody(t, r)
		if payload["artifactId"].(float64) != 12 {
			t.Fatalf("unexpected artifactId: %#v", payload)
		}
		if payload["tagSchemaVersion"].(string) != "0.3.0" {
			t.Fatalf("unexpected tagSchemaVersion: %#v", payload)
		}
		var patch map[string]any
		if err := json.Unmarshal([]byte(payload["tagsPatch"].(string)), &patch); err != nil {
			t.Fatalf("tagsPatch is not JSON: %v", err)
		}
		if patch["gate.L2"] != "PASS" {
			t.Fatalf("unexpected tagsPatch: %#v", patch)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"update success"}`))
	})

	resp, err := service.PatchArtifactTags(12, map[string]any{"gate.L2": "PASS"}, "0.3.0")
	if err != nil {
		t.Fatalf("PatchArtifactTags error: %v", err)
	}
	if resp.Msg != "update success" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestPatchArtifactLabelsWrapsLabelsIntoSchemaPayload(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, tagSchemaHandler(t, "0.3.0", flatTagSchemaContent, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/patch-tags" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := decodeBody(t, r)
		var patch map[string]any
		if err := json.Unmarshal([]byte(payload["tagsPatch"].(string)), &patch); err != nil {
			t.Fatalf("tagsPatch is not JSON: %v", err)
		}
		if patch["schema_version"] != "0.3.0" {
			t.Fatalf("schema_version not injected: %#v", patch)
		}
		labels, ok := patch["labels"].(map[string]any)
		if !ok || labels["gate.L2"] != "PASS" {
			t.Fatalf("unexpected labels: %#v", patch)
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"update success"}`))
	}))

	if _, err := service.PatchArtifactLabels(12, map[string]any{"gate.L2": "PASS"}, "0.3.0"); err != nil {
		t.Fatalf("PatchArtifactLabels error: %v", err)
	}
}

func TestPatchArtifactLabelsRejectsLegacySchema(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, tagSchemaHandler(t, "0.2.0", legacyTagSchemaContent, nil))

	_, err := service.PatchArtifactLabels(12, map[string]any{"gate.L2": "PASS"}, "0.2.0")
	if err == nil || !strings.Contains(err.Error(), "does not declare label_mode") {
		t.Fatalf("expected legacy schema rejection, got %v", err)
	}
}

func TestPatchArtifactLabelsRequiresLabels(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})

	if _, err := service.PatchArtifactLabels(12, map[string]any{}, "0.3.0"); err == nil {
		t.Fatal("expected empty labels to be rejected")
	}
}

func TestListArtifactsByLabelsPostsTagFilter(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, tagSchemaHandler(t, "0.3.0", flatTagSchemaContent, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/aiplorer/artifact/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := decodeBody(t, r)
		if payload["tagSchemaVersion"].(string) != "0.3.0" {
			t.Fatalf("unexpected tagSchemaVersion: %#v", payload)
		}
		if payload["page"].(float64) != 2 || payload["pageSize"].(float64) != 50 {
			t.Fatalf("unexpected pagination: %#v", payload)
		}
		var tagFilter map[string]any
		if err := json.Unmarshal([]byte(payload["tagFilter"].(string)), &tagFilter); err != nil {
			t.Fatalf("tagFilter is not JSON: %v", err)
		}
		labels, ok := tagFilter["labels"].(map[string]any)
		if !ok || labels["gate.L2"] != "PASS" {
			t.Fatalf("unexpected tagFilter: %#v", tagFilter)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"total":1,"data":[{"id":12,"name":"artifact-a"}]}}`))
	}))

	result, err := service.ListArtifactsByLabels(map[string]any{"gate.L2": "PASS"}, "0.3.0", 2, 50)
	if err != nil {
		t.Fatalf("ListArtifactsByLabels error: %v", err)
	}
	if result.Total != 1 || len(result.Data) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestListArtifactsByLabelsRejectsLegacySchema(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, tagSchemaHandler(t, "0.2.0", legacyTagSchemaContent, nil))

	_, err := service.ListArtifactsByLabels(map[string]any{"gate.L2": "PASS"}, "0.2.0", 1, 20)
	if err == nil || !strings.Contains(err.Error(), "does not declare label_mode") {
		t.Fatalf("expected legacy schema rejection, got %v", err)
	}
}

func TestListArtifactsByLabelsValidatesArguments(t *testing.T) {
	t.Parallel()
	service := newArtifactTestService(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})

	if _, err := service.ListArtifactsByLabels(map[string]any{"gate.L2": "PASS"}, "", 1, 20); err == nil {
		t.Fatal("expected missing tagSchemaVersion to be rejected")
	}
	if _, err := service.ListArtifactsByLabels(map[string]any{}, "0.3.0", 1, 20); err == nil {
		t.Fatal("expected empty labels to be rejected")
	}
}
