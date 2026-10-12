// Package models defines the data structures used in the MiniEye Intranet API.
package models

import "encoding/json"

// Label grammar modes declared by an artifact tag schema document.
//
// LabelModeFlat is the artifact-tag 0.3.0 grammar, whose payload is
// {"schema_version": ..., "labels": {...}} with dotted label keys.
// LabelModeLegacy is the original nested schema DSL used by 0.1.0/0.2.0;
// documents without a label_mode field keep using it.
const (
	LabelModeFlat   = "flat"
	LabelModeLegacy = "legacy"
)

// CommitInfo describes a git commit associated with an artifact.
type CommitInfo struct {
	ID             *uint64 `json:"id,omitempty"`
	CreatedAt      *int64  `json:"createdAt,omitempty"`
	UpdatedAt      *int64  `json:"updatedAt,omitempty"`
	RepositoryID   *uint64 `json:"repositoryId,omitempty"`
	RepositoryName *string `json:"repositoryName,omitempty"`
	RepositoryPath *string `json:"repositoryPath,omitempty"`
	CommitHash     *string `json:"commitHash,omitempty"`
	ShortHash      *string `json:"shortHash,omitempty"`
	Branch         *string `json:"branch,omitempty"`
	Author         *string `json:"author,omitempty"`
	AuthorEmail    *string `json:"authorEmail,omitempty"`
	Message        *string `json:"message,omitempty"`
	CommittedAt    *int64  `json:"committedAt,omitempty"`
	CommitTitle    *string `json:"commitTitle,omitempty"`
}

// ArtifactDependencyInfo describes a dependency or dependent artifact.
type ArtifactDependencyInfo struct {
	ID          *uint64      `json:"id,omitempty"`
	CreatedAt   *int64       `json:"createdAt,omitempty"`
	UpdatedAt   *int64       `json:"updatedAt,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Type        *string      `json:"type,omitempty"`
	Platform    *string      `json:"platform,omitempty"`
	IsVirtual   *bool        `json:"isVirtual,omitempty"`
	CommitHash  *string      `json:"commitHash,omitempty"`
	ModulePath  *string      `json:"modulePath,omitempty"`
	PipelineURL *string      `json:"pipelineUrl,omitempty"`
	ParentID    *uint64      `json:"parentId,omitempty"`
	Commits     []CommitInfo `json:"commits,omitempty"`
}

// ArtifactInfo describes artifact metadata and lineage information.
type ArtifactInfo struct {
	ID               *uint64                  `json:"id,omitempty"`
	CreatedAt        *int64                   `json:"createdAt,omitempty"`
	UpdatedAt        *int64                   `json:"updatedAt,omitempty"`
	FullPath         *string                  `json:"fullPath,omitempty"`
	MetadataPath     *string                  `json:"metadataPath,omitempty"`
	FileHash         *string                  `json:"fileHash,omitempty"`
	CommitHash       *string                  `json:"commitHash,omitempty"`
	Name             *string                  `json:"name,omitempty"`
	Type             *string                  `json:"type,omitempty"`
	Platform         *string                  `json:"platform,omitempty"`
	ProjectName      *string                  `json:"projectName,omitempty"`
	IsVirtual        *bool                    `json:"isVirtual,omitempty"`
	ModulePath       *string                  `json:"modulePath,omitempty"`
	SemanticVersion  *string                  `json:"semanticVersion,omitempty"`
	Tags             *string                  `json:"tags,omitempty"`
	TagSchemaVersion *string                  `json:"tagSchemaVersion,omitempty"`
	Extra            *string                  `json:"extra,omitempty"`
	PipelineID       *string                  `json:"pipelineId,omitempty"`
	PipelineURL      *string                  `json:"pipelineUrl,omitempty"`
	BuildDate        *int64                   `json:"buildDate,omitempty"`
	Commits          []CommitInfo             `json:"commits,omitempty"`
	Dependencies     []ArtifactDependencyInfo `json:"dependencies,omitempty"`
	Dependents       []ArtifactDependencyInfo `json:"dependents,omitempty"`
}

// ArtifactListReq lists artifacts with filters and pagination.
type ArtifactListReq struct {
	Page             uint64  `json:"page"`
	PageSize         uint64  `json:"pageSize"`
	CreatedAt        *int64  `json:"createdAt,omitempty"`
	UpdatedAt        *int64  `json:"updatedAt,omitempty"`
	FullPath         *string `json:"fullPath,omitempty"`
	FileHash         *string `json:"fileHash,omitempty"`
	CommitHash       *string `json:"commitHash,omitempty"`
	Name             *string `json:"name,omitempty"`
	ProjectName      *string `json:"projectName,omitempty"`
	Type             *string `json:"type,omitempty"`
	Platform         *string `json:"platform,omitempty"`
	IsVirtual        *bool   `json:"isVirtual,omitempty"`
	ModulePath       *string `json:"modulePath,omitempty"`
	SemanticVersion  *string `json:"semanticVersion,omitempty"`
	Tags             *string `json:"tags,omitempty"`
	TagSchemaVersion *string `json:"tagSchemaVersion,omitempty"`
	// TagFilter filters by tag content and requires TagSchemaVersion to be set.
	// For flat label schemas its shape is {"labels": {...}}.
	TagFilter       *string `json:"tagFilter,omitempty"`
	Extra           *string `json:"extra,omitempty"`
	PipelineID      *string `json:"pipelineId,omitempty"`
	PipelineURL     *string `json:"pipelineUrl,omitempty"`
	BuildDate       *int64  `json:"buildDate,omitempty"`
	ExactCommitHash *bool   `json:"exactCommitHash,omitempty"`
}

// ArtifactListResp is the paged artifact list response body.
type ArtifactListResp struct {
	Total uint64         `json:"total"`
	Data  []ArtifactInfo `json:"data"`
}

// ArtifactLookupOptions holds reusable filters for name-based lookups.
type ArtifactLookupOptions struct {
	ModulePath      string
	ArtifactType    string
	Platform        *string
	SemanticVersion string
	IncludeVirtual  *bool
	ProjectName     string
}

// ArtifactChildHashInfo describes one dependency node extracted from artifact lineage.
type ArtifactChildHashInfo struct {
	ID         *uint64 `json:"id,omitempty"`
	ParentID   *uint64 `json:"parentId,omitempty"`
	Name       *string `json:"name,omitempty"`
	Type       *string `json:"type,omitempty"`
	CommitHash *string `json:"commitHash,omitempty"`
	ModulePath *string `json:"modulePath,omitempty"`
}

// ArtifactChildHashesInfo is a helper view built from artifact detail responses.
type ArtifactChildHashesInfo struct {
	RootArtifactID   *uint64                 `json:"rootArtifactId,omitempty"`
	RootArtifactName *string                 `json:"rootArtifactName,omitempty"`
	RootArtifactType *string                 `json:"rootArtifactType,omitempty"`
	RootCommitHash   *string                 `json:"rootCommitHash,omitempty"`
	ChildHashes      []ArtifactChildHashInfo `json:"childHashes,omitempty"`
}

// ArtifactTagSchemaInfo contains schema metadata.
type ArtifactTagSchemaInfo struct {
	Version string `json:"version"`
	Content string `json:"content"`
}

// Document parses Content into a schema document. It returns an empty map when
// Content is empty.
func (s ArtifactTagSchemaInfo) Document() (map[string]any, error) {
	if s.Content == "" {
		return map[string]any{}, nil
	}
	parsed, err := ParseJSON(s.Content)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return map[string]any{}, nil
	}
	return parsed, nil
}

// LabelMode returns the label grammar declared by the schema document.
// Documents without label_mode are treated as legacy.
func (s ArtifactTagSchemaInfo) LabelMode() (string, error) {
	document, err := s.Document()
	if err != nil {
		return "", err
	}
	mode, _ := document["label_mode"].(string)
	if mode == "" {
		return LabelModeLegacy, nil
	}
	return mode, nil
}

// UsesFlatLabels reports whether the schema uses the flat label grammar.
func (s ArtifactTagSchemaInfo) UsesFlatLabels() (bool, error) {
	mode, err := s.LabelMode()
	if err != nil {
		return false, err
	}
	return mode == LabelModeFlat, nil
}

// ArtifactTagPatchReq patches artifact tags with JSON Merge Patch semantics.
type ArtifactTagPatchReq struct {
	ArtifactID       uint64 `json:"artifactId"`
	TagsPatch        string `json:"tagsPatch"`
	TagSchemaVersion string `json:"tagSchemaVersion,omitempty"`
}

// JfrogTokenInfo contains a JFrog access token response.
type JfrogTokenInfo struct {
	TokenID     string `json:"token_id"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	URL         string `json:"url"`
}

// ArtifactDownloadURLInfo contains a signed download URL.
type ArtifactDownloadURLInfo struct {
	DownloadURL string `json:"downloadUrl"`
	ExpireTime  string `json:"expireTime"`
	FileName    string `json:"fileName"`
	FilePath    string `json:"filePath"`
}

// ArtifactExistenceInfo describes whether an artifact exists for a lookup target.
type ArtifactExistenceInfo struct {
	CommitHash string  `json:"commitHash"`
	Exists     bool    `json:"exists"`
	ArtifactID *uint64 `json:"artifactId,omitempty"`
	Name       *string `json:"name,omitempty"`
}

// BatchArtifactExistenceResp is the batch existence check response body.
type BatchArtifactExistenceResp struct {
	Data []ArtifactExistenceInfo `json:"data"`
}

// GetArtifactByCommitHashReq looks up an artifact by exact commit hash.
type GetArtifactByCommitHashReq struct {
	CommitHash      string  `json:"commitHash"`
	ModulePath      *string `json:"modulePath,omitempty"`
	ArtifactType    *string `json:"type,omitempty"`
	Platform        *string `json:"platform,omitempty"`
	SemanticVersion *string `json:"semanticVersion,omitempty"`
	IsVirtual       *bool   `json:"isVirtual,omitempty"`
	ProjectName     *string `json:"projectName,omitempty"`
}

// BatchCheckArtifactsExistReq performs a batch existence check by exact commit hashes.
type BatchCheckArtifactsExistReq struct {
	CommitHashes    []string `json:"commitHashes"`
	ModulePath      *string  `json:"modulePath,omitempty"`
	ArtifactType    *string  `json:"type,omitempty"`
	Platform        *string  `json:"platform,omitempty"`
	SemanticVersion *string  `json:"semanticVersion,omitempty"`
	IsVirtual       *bool    `json:"isVirtual,omitempty"`
	ProjectName     *string  `json:"projectName,omitempty"`
}

// ArtifactVersionMetadataInfo contains version metadata fetched by commit hash.
type ArtifactVersionMetadataInfo struct {
	ArtifactID       *uint64        `json:"artifactId,omitempty"`
	CommitHash       *string        `json:"commitHash,omitempty"`
	Name             *string        `json:"name,omitempty"`
	MetadataPath     *string        `json:"metadataPath,omitempty"`
	MetadataFileName *string        `json:"metadataFileName,omitempty"`
	RawContent       *string        `json:"rawContent,omitempty"`
	Parsed           map[string]any `json:"-"`
}

// ArtifactDownloadPlan describes a resolved download plan for one artifact.
type ArtifactDownloadPlan struct {
	Artifact        *ArtifactInfo            `json:"artifact,omitempty"`
	Token           *JfrogTokenInfo          `json:"token,omitempty"`
	DownloadURL     *ArtifactDownloadURLInfo `json:"downloadUrl,omitempty"`
	TargetPath      string                   `json:"targetPath"`
	Checksum        string                   `json:"checksum,omitempty"`
	SkippedExisting bool                     `json:"skippedExisting,omitempty"`
}

// RepoDiff groups artifact commit differences by repository.
type RepoDiff struct {
	RepositoryID   uint64       `json:"repositoryId"`
	RepositoryName *string      `json:"repositoryName,omitempty"`
	RepositoryPath *string      `json:"repositoryPath,omitempty"`
	OlderCommit    *CommitInfo  `json:"olderCommit,omitempty"`
	NewerCommit    *CommitInfo  `json:"newerCommit,omitempty"`
	Commits        []CommitInfo `json:"commits,omitempty"`
}

// ArtifactCommitDiffInfo describes the commit diff between two artifacts.
type ArtifactCommitDiffInfo struct {
	OlderArtifactID  uint64     `json:"olderArtifactId"`
	NewerArtifactID  uint64     `json:"newerArtifactId"`
	ChangedRepoCount uint64     `json:"changedRepoCount"`
	RepoDiffs        []RepoDiff `json:"repoDiffs"`
}

// ArtifactCommitDiffOptions configures commit diff requests.
type ArtifactCommitDiffOptions struct {
	// CompareMode selects how commit lineage is compared.
	// Supported values are "direct" (default) and "distant".
	// An empty value keeps the server default.
	CompareMode string
}

// ArtifactChangelogOptions configures changelog commit list requests.
type ArtifactChangelogOptions struct {
	Page         *uint64
	PageSize     *uint64
	RepositoryID *uint64
	// CompareMode selects how commit lineage is compared.
	// Supported values are "direct" (default) and "distant".
	// An empty value keeps the server default.
	CompareMode string
}

// ArtifactChangelogFilesOptions configures changelog file list pagination.
type ArtifactChangelogFilesOptions struct {
	Page     *uint64
	PageSize *uint64
}

// ArtifactCommitDiffUnavailableRepo describes a repository that cannot produce changelog data.
type ArtifactCommitDiffUnavailableRepo struct {
	RepositoryID    *uint64     `json:"repositoryId,omitempty"`
	GitlabProjectID *string     `json:"gitlabProjectId,omitempty"`
	RepositoryName  *string     `json:"repositoryName,omitempty"`
	RepositoryPath  *string     `json:"repositoryPath,omitempty"`
	OlderCommit     *CommitInfo `json:"olderCommit,omitempty"`
	NewerCommit     *CommitInfo `json:"newerCommit,omitempty"`
	Reason          *string     `json:"reason,omitempty"`
}

// ArtifactChangelogCommitInfo describes one commit in an artifact changelog.
type ArtifactChangelogCommitInfo struct {
	ID              *uint64 `json:"id,omitempty"`
	GitlabProjectID *string `json:"gitlabProjectId,omitempty"`
	RepositoryID    *uint64 `json:"repositoryId,omitempty"`
	RepositoryName  *string `json:"repositoryName,omitempty"`
	RepositoryPath  *string `json:"repositoryPath,omitempty"`
	CommitID        *string `json:"commitId,omitempty"`
	CommitShortID   *string `json:"commitShortId,omitempty"`
	Title           *string `json:"title,omitempty"`
	Message         *string `json:"message,omitempty"`
	AuthorName      *string `json:"authorName,omitempty"`
	AuthorEmail     *string `json:"authorEmail,omitempty"`
	CommittedAt     *int64  `json:"committedAt,omitempty"`
	WebURL          *string `json:"webUrl,omitempty"`
}

// ArtifactChangelogInfo contains paged changelog commits between two artifacts.
type ArtifactChangelogInfo struct {
	OlderArtifactID              *uint64                             `json:"olderArtifactId,omitempty"`
	OlderArtifactName            *string                             `json:"olderArtifactName,omitempty"`
	OlderArtifactSemanticVersion *string                             `json:"olderArtifactSemanticVersion,omitempty"`
	NewerArtifactID              *uint64                             `json:"newerArtifactId,omitempty"`
	NewerArtifactName            *string                             `json:"newerArtifactName,omitempty"`
	NewerArtifactSemanticVersion *string                             `json:"newerArtifactSemanticVersion,omitempty"`
	ChangedRepoCount             uint64                              `json:"changedRepoCount"`
	UnavailableRepoCount         uint64                              `json:"unavailableRepoCount"`
	UnavailableRepos             []ArtifactCommitDiffUnavailableRepo `json:"unavailableRepos,omitempty"`
	Total                        uint64                              `json:"total"`
	Data                         []ArtifactChangelogCommitInfo       `json:"data"`
}

// ArtifactChangelogFileInfo describes one file touched by a changelog commit.
type ArtifactChangelogFileInfo struct {
	ID          *uint64 `json:"id,omitempty"`
	GitCommitID *uint64 `json:"gitCommitId,omitempty"`
	OldPath     *string `json:"oldPath,omitempty"`
	NewPath     *string `json:"newPath,omitempty"`
	AMode       *string `json:"aMode,omitempty"`
	BMode       *string `json:"bMode,omitempty"`
	NewFile     *bool   `json:"newFile,omitempty"`
	RenamedFile *bool   `json:"renamedFile,omitempty"`
	DeletedFile *bool   `json:"deletedFile,omitempty"`
}

// ArtifactChangelogFilesInfo contains paged files for one changelog commit.
type ArtifactChangelogFilesInfo struct {
	Total uint64                      `json:"total"`
	Data  []ArtifactChangelogFileInfo `json:"data"`
}

// IDsReq is a numeric ID list request.
type IDsReq struct {
	IDs []uint64 `json:"ids"`
}

// IDReq is a numeric ID request.
type IDReq struct {
	ID uint64 `json:"id"`
}

// ArtifactCommitDiffReq compares two artifacts.
type ArtifactCommitDiffReq struct {
	ArtifactIDA uint64  `json:"artifactIdA"`
	ArtifactIDB uint64  `json:"artifactIdB"`
	CompareMode *string `json:"compareMode,omitempty"`
}

// ArtifactChangelogReq requests paged changelog commits between two artifacts.
type ArtifactChangelogReq struct {
	ArtifactIDA  uint64  `json:"artifactIdA"`
	ArtifactIDB  uint64  `json:"artifactIdB"`
	Page         *uint64 `json:"page,omitempty"`
	PageSize     *uint64 `json:"pageSize,omitempty"`
	RepositoryID *uint64 `json:"repositoryId,omitempty"`
	CompareMode  *string `json:"compareMode,omitempty"`
}

// ArtifactChangelogFilesReq requests files for a changelog commit.
type ArtifactChangelogFilesReq struct {
	GitCommitID uint64  `json:"gitCommitId"`
	Page        *uint64 `json:"page,omitempty"`
	PageSize    *uint64 `json:"pageSize,omitempty"`
}

// ArtifactTagSchemaReq queries a tag schema by version.
type ArtifactTagSchemaReq struct {
	Version string `json:"version,omitempty"`
}

// JfrogTokenReq requests a JFrog token.
type JfrogTokenReq struct {
	ProjectName string `json:"projectName"`
}

// ArtifactDownloadURLReq requests an artifact download URL.
type ArtifactDownloadURLReq struct {
	ArtifactID   uint64 `json:"artifactId"`
	DownloadType string `json:"downloadType"`
}

// ArtifactTagUpdateReq updates only artifact tags.
type ArtifactTagUpdateReq struct {
	ID               uint64 `json:"id"`
	Tags             string `json:"tags"`
	TagSchemaVersion string `json:"tagSchemaVersion,omitempty"`
}

// ParseJSON parses a raw JSON string to a generic object.
func ParseJSON(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
