// Package models defines the data structures used in the MiniEye Intranet API.
package models

// ImageListReq contains optional filters for the OCI image list endpoint.
// A nil field is omitted from the query string; a non-nil pointer is sent even
// when it contains a zero value.
type ImageListReq struct {
	PageNum            *uint64 `json:"page_num,omitempty"`
	PageSize           *uint64 `json:"page_size,omitempty"`
	GitRepo            *string `json:"git_repo,omitempty"`
	GitBranch          *string `json:"git_branch,omitempty"`
	ImgFullname        *string `json:"img_fullname,omitempty"`
	Digest             *string `json:"digest,omitempty"`
	GitCommitShortHash *string `json:"git_commit_short_hash,omitempty"`
	PipelineID         *uint64 `json:"pipeline_id,omitempty"`
	JobID              *uint64 `json:"job_id,omitempty"`
}

// ImageInfo describes OCI image metadata.
type ImageInfo struct {
	ID            uint64  `json:"id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	ImgRegistry   string  `json:"img_registry"`
	ImgProject    string  `json:"img_project"`
	ImgRepo       string  `json:"img_repo"`
	Digest        string  `json:"digest"`
	OS            *string `json:"os,omitempty"`
	Arch          *string `json:"arch,omitempty"`
	ImgTag        *string `json:"img_tag,omitempty"`
	JobURL        *string `json:"job_url,omitempty"`
	PipelineURL   *string `json:"pipeline_url,omitempty"`
	GitRepo       *string `json:"git_repo,omitempty"`
	GitBranch     *string `json:"git_branch,omitempty"`
	GitCommitHash *string `json:"git_commit_hash,omitempty"`
	GitCommitAt   *string `json:"git_commit_at,omitempty"`
}

// ImageListResp is the paginated OCI image list returned by the API.
type ImageListResp struct {
	Total uint64      `json:"total"`
	Items []ImageInfo `json:"items"`
}

// ImageDiffReq contains the two image full names to compare. The full names
// are required; a nil pagination field is omitted from the query string, while
// a non-nil pointer must be greater than zero as enforced by the server.
type ImageDiffReq struct {
	ImgFullnameA string  `json:"img_fullname_a,omitempty"`
	ImgFullnameB string  `json:"img_fullname_b,omitempty"`
	PageNum      *uint64 `json:"page_num,omitempty"`
	PageSize     *uint64 `json:"page_size,omitempty"`
}

// ImageDiffCommitInfo describes one commit between the two compared images.
type ImageDiffCommitInfo struct {
	ID        *uint64 `json:"id,omitempty"`
	CommitAt  string  `json:"commit_at"`
	Hash      string  `json:"hash"`
	ShortHash string  `json:"short_hash"`
	Branch    *string `json:"branch,omitempty"`
	Title     string  `json:"title"`
	Message   string  `json:"message"`
	Author    string  `json:"author"`
	WebURL    *string `json:"web_url,omitempty"`
}

// ImageDiffData describes the commit difference between one image pair.
// Commits is the paginated commit list ordered from CommitA to CommitB, and
// TotalCommits is the total count before pagination.
type ImageDiffData struct {
	ImgIDA       uint64                `json:"img_id_a"`
	ImgIDB       uint64                `json:"img_id_b"`
	ImgFullnameA string                `json:"img_fullname_a"`
	ImgFullnameB string                `json:"img_fullname_b"`
	GitRepo      string                `json:"git_repo"`
	CommitA      ImageDiffCommitInfo   `json:"commit_a"`
	CommitB      ImageDiffCommitInfo   `json:"commit_b"`
	Commits      []ImageDiffCommitInfo `json:"commits"`
	TotalCommits uint64                `json:"total_commits"`
}
