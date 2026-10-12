// Package services provides business logic for the MiniEye Intranet API.
package services

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hujia-team/intranet-sdk/client"
	"github.com/hujia-team/intranet-sdk/models"
	"github.com/hujia-team/intranet-sdk/utils"
	jfrogartifactory "github.com/jfrog/jfrog-client-go/artifactory"
	jfrogAuth "github.com/jfrog/jfrog-client-go/artifactory/auth"
	jfrogServices "github.com/jfrog/jfrog-client-go/artifactory/services"
	jfrogUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	jfrogConfig "github.com/jfrog/jfrog-client-go/config"
)

// ArtifactService defines artifact-management operations.
type ArtifactService interface {
	CreateArtifact(artifact *models.ArtifactInfo) (*models.BaseMsgResp, error)
	UpdateArtifact(artifact *models.ArtifactInfo) (*models.BaseMsgResp, error)
	DeleteArtifacts(ids []uint64) (*models.BaseMsgResp, error)
	ListArtifacts(req *models.ArtifactListReq) (*models.ArtifactListResp, error)
	GetArtifactByID(id uint64) (*models.ArtifactInfo, error)
	GetArtifactByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (*models.ArtifactInfo, error)
	GetArtifactByName(name string, lookup *models.ArtifactLookupOptions) (*models.ArtifactInfo, error)
	CheckExistsByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (bool, error)
	CheckExistsByName(name string, lookup *models.ArtifactLookupOptions) (bool, error)
	PrepareDownloadByArtifactID(artifactID uint64, destination string) (*models.ArtifactDownloadPlan, error)
	PrepareDownloadByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions, destination string) (*models.ArtifactDownloadPlan, error)
	DownloadByArtifactID(artifactID uint64, destination string) (*models.ArtifactDownloadPlan, error)
	DownloadByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions, destination string) (*models.ArtifactDownloadPlan, error)
	DownloadByName(name string, lookup *models.ArtifactLookupOptions, destination string) (*models.ArtifactDownloadPlan, error)
	GetVersionMetadataByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (*models.ArtifactVersionMetadataInfo, error)
	GetChildArtifactHashesByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (*models.ArtifactChildHashesInfo, error)
	GetArtifactCommitDiff(artifactIDA, artifactIDB uint64, options ...*models.ArtifactCommitDiffOptions) (*models.ArtifactCommitDiffInfo, error)
	GetArtifactChangelog(artifactIDA, artifactIDB uint64, options *models.ArtifactChangelogOptions) (*models.ArtifactChangelogInfo, error)
	GetArtifactChangelogFiles(gitCommitID uint64, options *models.ArtifactChangelogFilesOptions) (*models.ArtifactChangelogFilesInfo, error)
	GetArtifactTagSchema(version string) (*models.ArtifactTagSchemaInfo, error)
	GetArtifactTagSchemaJSON(version string) (map[string]any, error)
	GetArtifactLabelMode(version string) (string, error)
	UsesFlatLabels(version string) (bool, error)
	GetJfrogToken(projectName string) (*models.JfrogTokenInfo, error)
	GetJfrogTokenByArtifactName(name string, lookup *models.ArtifactLookupOptions) (*models.JfrogTokenInfo, error)
	GetArtifactDownloadURL(artifactID uint64, downloadType string) (*models.ArtifactDownloadURLInfo, error)
	GetArtifactDownloadURLByName(name string, lookup *models.ArtifactLookupOptions, downloadType string) (*models.ArtifactDownloadURLInfo, error)
	GetParsedArtifactTags(artifactID uint64) (map[string]any, error)
	UpdateArtifactTags(artifactID uint64, tags map[string]any, tagSchemaVersion string) (*models.BaseMsgResp, error)
	PatchArtifactTags(artifactID uint64, tagsPatch map[string]any, tagSchemaVersion string) (*models.BaseMsgResp, error)
	PatchArtifactLabels(artifactID uint64, labels map[string]any, tagSchemaVersion string) (*models.BaseMsgResp, error)
	ListArtifactsByLabels(labels map[string]any, tagSchemaVersion string, page, pageSize uint64) (*models.ArtifactListResp, error)
	ParseArtifactTags(tags string, schema any) (map[string]any, error)
}

type artifactService struct {
	httpClient       *client.HTTPClient
	downloadArtifact func(token *models.JfrogTokenInfo, filePath, targetDir string) error
}

// NewArtifactService creates a new artifact service.
func NewArtifactService(httpClient *client.HTTPClient) ArtifactService {
	return &artifactService{
		httpClient:       httpClient,
		downloadArtifact: downloadWithJFrog,
	}
}

func (s *artifactService) CreateArtifact(artifact *models.ArtifactInfo) (*models.BaseMsgResp, error) {
	var response models.BaseMsgResp
	utils.Debug("Creating artifact")
	if err := s.httpClient.Post("/aiplorer/artifact/create", artifact, &response); err != nil {
		return nil, utils.NewAPIError("failed to create artifact", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response, nil
}

func (s *artifactService) UpdateArtifact(artifact *models.ArtifactInfo) (*models.BaseMsgResp, error) {
	var response models.BaseMsgResp
	utils.Debug("Updating artifact")
	if err := s.httpClient.Post("/aiplorer/artifact/update", artifact, &response); err != nil {
		return nil, utils.NewAPIError("failed to update artifact", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response, nil
}

func (s *artifactService) DeleteArtifacts(ids []uint64) (*models.BaseMsgResp, error) {
	var response models.BaseMsgResp
	utils.Debug("Deleting artifacts: %v", ids)
	if err := s.httpClient.Post("/aiplorer/artifact/delete", &models.IDsReq{IDs: ids}, &response); err != nil {
		return nil, utils.NewAPIError("failed to delete artifacts", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response, nil
}

func (s *artifactService) ListArtifacts(req *models.ArtifactListReq) (*models.ArtifactListResp, error) {
	var response struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data models.ArtifactListResp `json:"data"`
	}
	utils.Debug("Listing artifacts")
	if err := s.httpClient.Post("/aiplorer/artifact/list", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to list artifacts", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactByID(id uint64) (*models.ArtifactInfo, error) {
	var response struct {
		Code int                 `json:"code"`
		Msg  string              `json:"msg"`
		Data models.ArtifactInfo `json:"data"`
	}
	utils.Debug("Getting artifact by ID: %d", id)
	if err := s.httpClient.Post("/aiplorer/artifact", &models.IDReq{ID: id}, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact by id", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactByName(name string, lookup *models.ArtifactLookupOptions) (*models.ArtifactInfo, error) {
	req := &models.ArtifactListReq{Page: 1, PageSize: 100, Name: &name}
	if lookup != nil {
		if lookup.ModulePath != "" {
			req.ModulePath = &lookup.ModulePath
		}
		if lookup.ArtifactType != "" {
			req.Type = &lookup.ArtifactType
		}
		if lookup.Platform != nil {
			req.Platform = lookup.Platform
		}
		if lookup.SemanticVersion != "" {
			req.SemanticVersion = &lookup.SemanticVersion
		}
		if lookup.ProjectName != "" {
			req.ProjectName = &lookup.ProjectName
		}
		req.IsVirtual = lookup.IncludeVirtual
	}
	result, err := s.ListArtifacts(req)
	if err != nil {
		return nil, err
	}
	var matched []models.ArtifactInfo
	for _, item := range result.Data {
		if item.Name != nil && *item.Name == name {
			matched = append(matched, item)
		}
	}
	if len(matched) == 0 {
		return nil, utils.NewAPIError(fmt.Sprintf("artifact not found by name: %s", name), nil)
	}
	if len(matched) > 1 {
		return nil, utils.NewAPIError(fmt.Sprintf("multiple artifacts found by name: %s", name), nil)
	}
	if matched[0].ID == nil {
		return nil, utils.NewAPIError(fmt.Sprintf("artifact id missing for artifact: %s", name), nil)
	}
	return s.GetArtifactByID(*matched[0].ID)
}

func (s *artifactService) GetArtifactByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (*models.ArtifactInfo, error) {
	var response struct {
		Code int                 `json:"code"`
		Msg  string              `json:"msg"`
		Data models.ArtifactInfo `json:"data"`
	}
	req := buildCommitHashLookupRequest(commitHash, lookup)
	if err := s.httpClient.Post("/aiplorer/artifact/by-commit-hash", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact by commit hash", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) CheckExistsByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (bool, error) {
	artifact, err := s.GetArtifactByCommitHash(commitHash, lookup)
	if err != nil {
		if strings.Contains(err.Error(), "artifact not found by commit hash") {
			return false, nil
		}
		return false, err
	}
	return artifact.FullPath != nil && *artifact.FullPath != "" &&
		artifact.FileHash != nil && *artifact.FileHash != "", nil
}

func (s *artifactService) CheckExistsByName(name string, lookup *models.ArtifactLookupOptions) (bool, error) {
	_, err := s.GetArtifactByName(name, lookup)
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "artifact not found by name") {
		return false, nil
	}
	return false, err
}

func (s *artifactService) PrepareDownloadByArtifactID(artifactID uint64, destination string) (*models.ArtifactDownloadPlan, error) {
	artifact, err := s.GetArtifactByID(artifactID)
	if err != nil {
		return nil, err
	}
	return s.prepareDownload(artifact, destination)
}

func (s *artifactService) PrepareDownloadByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions, destination string) (*models.ArtifactDownloadPlan, error) {
	artifact, err := s.GetArtifactByCommitHash(commitHash, lookup)
	if err != nil {
		return nil, err
	}
	return s.prepareDownload(artifact, destination)
}

func (s *artifactService) DownloadByArtifactID(artifactID uint64, destination string) (*models.ArtifactDownloadPlan, error) {
	plan, err := s.PrepareDownloadByArtifactID(artifactID, destination)
	if err != nil {
		return nil, err
	}
	return s.executeDownloadPlan(plan)
}

func (s *artifactService) DownloadByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions, destination string) (*models.ArtifactDownloadPlan, error) {
	plan, err := s.PrepareDownloadByCommitHash(commitHash, lookup, destination)
	if err != nil {
		return nil, err
	}
	return s.executeDownloadPlan(plan)
}

func (s *artifactService) executeDownloadPlan(plan *models.ArtifactDownloadPlan) (*models.ArtifactDownloadPlan, error) {
	if plan.Token == nil || plan.DownloadURL == nil {
		return nil, utils.NewAPIError("download plan is incomplete", nil)
	}
	targetDir := filepath.Dir(plan.TargetPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, utils.NewInternalError("failed to create download target directory", err)
	}
	skipped, err := skipDownloadIfExisting(plan.TargetPath, plan.Checksum)
	if err != nil {
		return nil, err
	}
	if skipped {
		plan.SkippedExisting = true
		return plan, nil
	}
	if err := s.downloadArtifact(plan.Token, plan.DownloadURL.FilePath, targetDir); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *artifactService) DownloadByName(name string, lookup *models.ArtifactLookupOptions, destination string) (*models.ArtifactDownloadPlan, error) {
	artifact, err := s.GetArtifactByName(name, lookup)
	if err != nil {
		return nil, err
	}
	if artifact.ID == nil {
		return nil, utils.NewAPIError(fmt.Sprintf("artifact id missing for artifact: %s", name), nil)
	}
	return s.DownloadByArtifactID(*artifact.ID, destination)
}

func (s *artifactService) prepareDownload(artifact *models.ArtifactInfo, destination string) (*models.ArtifactDownloadPlan, error) {
	if artifact == nil {
		return nil, utils.NewAPIError("artifact is nil", nil)
	}
	if artifact.ID == nil {
		return nil, utils.NewAPIError("artifact id is empty", nil)
	}
	if artifact.ProjectName == nil || *artifact.ProjectName == "" {
		return nil, utils.NewAPIError(fmt.Sprintf("artifact project_name is empty for artifact id: %d", *artifact.ID), nil)
	}

	token, err := s.GetJfrogToken(*artifact.ProjectName)
	if err != nil {
		return nil, err
	}
	downloadURL, err := s.GetArtifactDownloadURL(*artifact.ID, "artifact")
	if err != nil {
		return nil, err
	}

	targetPath := resolveDownloadTarget(destination, downloadURL.FileName)
	checksum := ""
	if artifact.FileHash != nil {
		checksum = *artifact.FileHash
	}
	return &models.ArtifactDownloadPlan{
		Artifact:    artifact,
		Token:       token,
		DownloadURL: downloadURL,
		TargetPath:  targetPath,
		Checksum:    checksum,
	}, nil
}

func (s *artifactService) GetVersionMetadataByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (*models.ArtifactVersionMetadataInfo, error) {
	var response struct {
		Code int                                `json:"code"`
		Msg  string                             `json:"msg"`
		Data models.ArtifactVersionMetadataInfo `json:"data"`
	}
	req := buildCommitHashLookupRequest(commitHash, lookup)
	if err := s.httpClient.Post("/aiplorer/artifact/version-metadata", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact version metadata", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	if response.Data.RawContent != nil && *response.Data.RawContent != "" {
		parsed, err := parseVersionMetadata(*response.Data.RawContent, valueOrEmpty(response.Data.MetadataFileName))
		if err != nil {
			return nil, err
		}
		response.Data.Parsed = parsed
	}
	return &response.Data, nil
}

func (s *artifactService) GetChildArtifactHashesByCommitHash(commitHash string, lookup *models.ArtifactLookupOptions) (*models.ArtifactChildHashesInfo, error) {
	artifact, err := s.GetArtifactByCommitHash(commitHash, lookup)
	if err != nil {
		return nil, err
	}

	result := &models.ArtifactChildHashesInfo{
		RootArtifactID:   artifact.ID,
		RootArtifactName: artifact.Name,
		RootArtifactType: artifact.Type,
		RootCommitHash:   artifact.CommitHash,
		ChildHashes:      make([]models.ArtifactChildHashInfo, 0, len(artifact.Dependencies)),
	}
	for _, dep := range artifact.Dependencies {
		result.ChildHashes = append(result.ChildHashes, models.ArtifactChildHashInfo{
			ID:         dep.ID,
			ParentID:   dep.ParentID,
			Name:       dep.Name,
			Type:       dep.Type,
			CommitHash: dep.CommitHash,
			ModulePath: dep.ModulePath,
		})
	}
	return result, nil
}

func (s *artifactService) GetArtifactCommitDiff(artifactIDA, artifactIDB uint64, options ...*models.ArtifactCommitDiffOptions) (*models.ArtifactCommitDiffInfo, error) {
	var response struct {
		Code int                           `json:"code"`
		Msg  string                        `json:"msg"`
		Data models.ArtifactCommitDiffInfo `json:"data"`
	}
	req := &models.ArtifactCommitDiffReq{
		ArtifactIDA: artifactIDA,
		ArtifactIDB: artifactIDB,
	}
	if len(options) > 0 && options[0] != nil {
		req.CompareMode = compareModePtr(options[0].CompareMode)
	}
	if err := s.httpClient.Post("/aiplorer/artifact/commit-diff", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact commit diff", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactChangelog(artifactIDA, artifactIDB uint64, options *models.ArtifactChangelogOptions) (*models.ArtifactChangelogInfo, error) {
	var response struct {
		Code int                          `json:"code"`
		Msg  string                       `json:"msg"`
		Data models.ArtifactChangelogInfo `json:"data"`
	}
	req := &models.ArtifactChangelogReq{
		ArtifactIDA: artifactIDA,
		ArtifactIDB: artifactIDB,
	}
	if options != nil {
		req.Page = options.Page
		req.PageSize = options.PageSize
		req.RepositoryID = options.RepositoryID
		req.CompareMode = compareModePtr(options.CompareMode)
	}
	if err := s.httpClient.Post("/aiplorer/artifact/commit-diff-v2", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact changelog", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactChangelogFiles(gitCommitID uint64, options *models.ArtifactChangelogFilesOptions) (*models.ArtifactChangelogFilesInfo, error) {
	var response struct {
		Code int                               `json:"code"`
		Msg  string                            `json:"msg"`
		Data models.ArtifactChangelogFilesInfo `json:"data"`
	}
	req := &models.ArtifactChangelogFilesReq{GitCommitID: gitCommitID}
	if options != nil {
		req.Page = options.Page
		req.PageSize = options.PageSize
	}
	if err := s.httpClient.Post("/aiplorer/artifact/commit-diff-files", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact changelog files", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactTagSchema(version string) (*models.ArtifactTagSchemaInfo, error) {
	var response struct {
		Code int                          `json:"code"`
		Msg  string                       `json:"msg"`
		Data models.ArtifactTagSchemaInfo `json:"data"`
	}
	req := &models.ArtifactTagSchemaReq{}
	if version != "" {
		req.Version = version
	}
	if err := s.httpClient.Post("/aiplorer/artifact/tag-schema", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact tag schema", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactTagSchemaJSON(version string) (map[string]any, error) {
	schema, err := s.GetArtifactTagSchema(version)
	if err != nil {
		return nil, err
	}
	return models.ParseJSON(schema.Content)
}

// GetArtifactLabelMode returns the label grammar declared by the artifact tag
// schema document: models.LabelModeFlat for artifact-tag 0.3.0 and later flat
// label schemas, models.LabelModeLegacy for the original nested grammar.
func (s *artifactService) GetArtifactLabelMode(version string) (string, error) {
	schema, err := s.GetArtifactTagSchema(version)
	if err != nil {
		return "", err
	}
	mode, err := schema.LabelMode()
	if err != nil {
		return "", utils.NewAPIError("failed to decode artifact tag schema", err)
	}
	return mode, nil
}

// UsesFlatLabels reports whether the artifact tag schema uses the flat label
// grammar.
func (s *artifactService) UsesFlatLabels(version string) (bool, error) {
	mode, err := s.GetArtifactLabelMode(version)
	if err != nil {
		return false, err
	}
	return mode == models.LabelModeFlat, nil
}

func (s *artifactService) GetJfrogToken(projectName string) (*models.JfrogTokenInfo, error) {
	var response struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data models.JfrogTokenInfo `json:"data"`
	}
	if err := s.httpClient.Post("/aiplorer/jfrog/token", &models.JfrogTokenReq{ProjectName: projectName}, &response); err != nil {
		return nil, utils.NewAPIError("failed to get jfrog token", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetJfrogTokenByArtifactName(name string, lookup *models.ArtifactLookupOptions) (*models.JfrogTokenInfo, error) {
	artifact, err := s.GetArtifactByName(name, lookup)
	if err != nil {
		return nil, err
	}
	if artifact.ProjectName == nil || *artifact.ProjectName == "" {
		return nil, utils.NewAPIError(fmt.Sprintf("artifact project_name is empty: %s", name), nil)
	}
	return s.GetJfrogToken(*artifact.ProjectName)
}

func (s *artifactService) GetArtifactDownloadURL(artifactID uint64, downloadType string) (*models.ArtifactDownloadURLInfo, error) {
	if downloadType == "" {
		downloadType = "artifact"
	}
	var response struct {
		Code int                            `json:"code"`
		Msg  string                         `json:"msg"`
		Data models.ArtifactDownloadURLInfo `json:"data"`
	}
	if err := s.httpClient.Post("/aiplorer/artifact/download-url", &models.ArtifactDownloadURLReq{
		ArtifactID:   artifactID,
		DownloadType: downloadType,
	}, &response); err != nil {
		return nil, utils.NewAPIError("failed to get artifact download url", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func (s *artifactService) GetArtifactDownloadURLByName(name string, lookup *models.ArtifactLookupOptions, downloadType string) (*models.ArtifactDownloadURLInfo, error) {
	artifact, err := s.GetArtifactByName(name, lookup)
	if err != nil {
		return nil, err
	}
	if artifact.ID == nil {
		return nil, utils.NewAPIError(fmt.Sprintf("artifact id missing for artifact: %s", name), nil)
	}
	return s.GetArtifactDownloadURL(*artifact.ID, downloadType)
}

func (s *artifactService) GetParsedArtifactTags(artifactID uint64) (map[string]any, error) {
	artifact, err := s.GetArtifactByID(artifactID)
	if err != nil {
		return nil, err
	}
	if artifact.Tags == nil || *artifact.Tags == "" {
		return map[string]any{}, nil
	}
	version := ""
	if artifact.TagSchemaVersion != nil {
		version = *artifact.TagSchemaVersion
	}
	schema, err := s.GetArtifactTagSchema(version)
	if err != nil {
		return nil, err
	}
	return s.ParseArtifactTags(*artifact.Tags, schema)
}

func (s *artifactService) UpdateArtifactTags(artifactID uint64, tags map[string]any, tagSchemaVersion string) (*models.BaseMsgResp, error) {
	if tagSchemaVersion == "" {
		if rawVersion, ok := tags["schema_version"].(string); ok {
			tagSchemaVersion = rawVersion
		}
	}
	if tagSchemaVersion == "" {
		artifact, err := s.GetArtifactByID(artifactID)
		if err != nil {
			return nil, err
		}
		if artifact.TagSchemaVersion != nil {
			tagSchemaVersion = *artifact.TagSchemaVersion
		}
	}
	schema, err := s.GetArtifactTagSchema(tagSchemaVersion)
	if err != nil {
		return nil, err
	}
	if _, err := s.ParseArtifactTags(mustJSON(tags), schema); err != nil {
		return nil, err
	}
	return s.UpdateArtifact(&models.ArtifactInfo{
		ID:               &artifactID,
		Tags:             stringPtr(mustJSON(tags)),
		TagSchemaVersion: stringPtr(tagSchemaVersion),
	})
}

// PatchArtifactTags patches artifact tags with JSON Merge Patch semantics
// through the existing patch-tags endpoint.
func (s *artifactService) PatchArtifactTags(artifactID uint64, tagsPatch map[string]any, tagSchemaVersion string) (*models.BaseMsgResp, error) {
	if tagsPatch == nil {
		return nil, utils.NewAPIError("tagsPatch must not be nil", nil)
	}
	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	req := &models.ArtifactTagPatchReq{
		ArtifactID:       artifactID,
		TagsPatch:        mustJSON(tagsPatch),
		TagSchemaVersion: tagSchemaVersion,
	}
	if err := s.httpClient.Post("/aiplorer/artifact/patch-tags", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to patch artifact tags", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &models.BaseMsgResp{Msg: response.Msg}, nil
}

// PatchArtifactLabels patches flat artifact labels for schemas that declare
// label_mode "flat" (artifact-tag 0.3.0 and later). The schema version is
// resolved from the artifact when not given explicitly.
//
// Every submitted key follows JSON Merge Patch semantics: new values overwrite
// the current ones and nil deletes custom.* keys. The backend rejects nil for
// eval.* and gate.* keys, and requires a matching urls key in the same request
// whenever an eval.* value key is submitted.
func (s *artifactService) PatchArtifactLabels(artifactID uint64, labels map[string]any, tagSchemaVersion string) (*models.BaseMsgResp, error) {
	if len(labels) == 0 {
		return nil, utils.NewAPIError("labels must not be empty", nil)
	}
	version := tagSchemaVersion
	if version == "" {
		artifact, err := s.GetArtifactByID(artifactID)
		if err != nil {
			return nil, err
		}
		if artifact.TagSchemaVersion != nil {
			version = *artifact.TagSchemaVersion
		}
	}
	if version == "" {
		return nil, utils.NewAPIError("tagSchemaVersion is required to patch artifact labels", nil)
	}

	flat, err := s.UsesFlatLabels(version)
	if err != nil {
		return nil, err
	}
	if !flat {
		return nil, utils.NewAPIError(fmt.Sprintf(
			"artifact tag schema %s does not declare label_mode %q; use UpdateArtifactTags for legacy schemas",
			version, models.LabelModeFlat,
		), nil)
	}

	return s.PatchArtifactTags(artifactID, map[string]any{
		"schema_version": version,
		"labels":         labels,
	}, version)
}

// ListArtifactsByLabels lists artifacts matching flat artifact labels. The
// backend requires tagSchemaVersion whenever a tag filter is present and
// evaluates every provided label with AND semantics.
func (s *artifactService) ListArtifactsByLabels(labels map[string]any, tagSchemaVersion string, page, pageSize uint64) (*models.ArtifactListResp, error) {
	if tagSchemaVersion == "" {
		return nil, utils.NewAPIError("tagSchemaVersion is required to search artifacts by labels", nil)
	}
	if len(labels) == 0 {
		return nil, utils.NewAPIError("labels must not be empty", nil)
	}

	flat, err := s.UsesFlatLabels(tagSchemaVersion)
	if err != nil {
		return nil, err
	}
	if !flat {
		return nil, utils.NewAPIError(fmt.Sprintf(
			"artifact tag schema %s does not declare label_mode %q; use a nested tag filter for legacy schemas",
			tagSchemaVersion, models.LabelModeFlat,
		), nil)
	}

	tagFilter := mustJSON(map[string]any{"labels": labels})
	return s.ListArtifacts(&models.ArtifactListReq{
		Page:             page,
		PageSize:         pageSize,
		TagSchemaVersion: stringPtr(tagSchemaVersion),
		TagFilter:        &tagFilter,
	})
}

func (s *artifactService) ParseArtifactTags(tags string, schema any) (map[string]any, error) {
	parsedTags, err := models.ParseJSON(tags)
	if err != nil {
		return nil, utils.NewAPIError("failed to decode artifact tags", err)
	}
	var parsedSchema map[string]any
	switch v := schema.(type) {
	case *models.ArtifactTagSchemaInfo:
		parsedSchema, err = models.ParseJSON(v.Content)
	case models.ArtifactTagSchemaInfo:
		parsedSchema, err = models.ParseJSON(v.Content)
	case string:
		parsedSchema, err = models.ParseJSON(v)
	case []byte:
		parsedSchema, err = models.ParseJSON(string(v))
	case map[string]any:
		parsedSchema = v
	default:
		return nil, utils.NewAPIError("unsupported artifact tag schema type", nil)
	}
	if err != nil {
		return nil, utils.NewAPIError("failed to decode artifact tag schema", err)
	}
	// Schema documents carry "schema_version"; "version" is kept as a fallback
	// for older or third-party schema payloads.
	schemaVersion, hasSchemaVersion := parsedSchema["schema_version"].(string)
	if !hasSchemaVersion {
		schemaVersion, hasSchemaVersion = parsedSchema["version"].(string)
	}
	tagSchemaVersion, hasTagSchemaVersion := parsedTags["schema_version"].(string)
	if hasSchemaVersion && hasTagSchemaVersion && schemaVersion != tagSchemaVersion {
		return nil, utils.NewAPIError("artifact tag schema version does not match tag schema_version", nil)
	}
	return parsedTags, nil
}

func (s *artifactService) batchCheckArtifactsExist(req *models.BatchCheckArtifactsExistReq) (*models.BatchArtifactExistenceResp, error) {
	var response struct {
		Code int                               `json:"code"`
		Msg  string                            `json:"msg"`
		Data models.BatchArtifactExistenceResp `json:"data"`
	}
	if err := s.httpClient.Post("/aiplorer/artifact/batch-exists", req, &response); err != nil {
		return nil, utils.NewAPIError("failed to batch check artifact existence", err)
	}
	if response.Code != 0 {
		return nil, utils.NewAPIError(response.Msg, nil)
	}
	return &response.Data, nil
}

func buildCommitHashLookupRequest(commitHash string, lookup *models.ArtifactLookupOptions) *models.GetArtifactByCommitHashReq {
	req := &models.GetArtifactByCommitHashReq{CommitHash: commitHash}
	if lookup == nil {
		return req
	}
	if lookup.ModulePath != "" {
		req.ModulePath = &lookup.ModulePath
	}
	if lookup.ArtifactType != "" {
		req.ArtifactType = &lookup.ArtifactType
	}
	if lookup.Platform != nil {
		req.Platform = lookup.Platform
	}
	if lookup.SemanticVersion != "" {
		req.SemanticVersion = &lookup.SemanticVersion
	}
	if lookup.IncludeVirtual != nil {
		req.IsVirtual = lookup.IncludeVirtual
	}
	if lookup.ProjectName != "" {
		req.ProjectName = &lookup.ProjectName
	}
	return req
}

func resolveDownloadTarget(destination, fileName string) string {
	if destination == "" {
		return fileName
	}
	if strings.HasSuffix(destination, string(os.PathSeparator)) {
		return filepath.Join(destination, fileName)
	}
	info, err := os.Stat(destination)
	if err == nil && info.IsDir() {
		return filepath.Join(destination, fileName)
	}
	if filepath.Ext(destination) != "" {
		return destination
	}
	return filepath.Join(destination, fileName)
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func downloadWithJFrog(token *models.JfrogTokenInfo, filePath, targetDir string) error {
	rtDetails := jfrogAuth.NewArtifactoryDetails()
	baseURL := strings.TrimRight(token.URL, "/")
	if !strings.HasSuffix(baseURL, "/artifactory") {
		baseURL += "/artifactory"
	}
	rtDetails.SetUrl(baseURL)
	rtDetails.SetAccessToken(token.AccessToken)

	serviceConfig, err := jfrogConfig.NewConfigBuilder().SetServiceDetails(rtDetails).Build()
	if err != nil {
		return utils.NewInternalError("failed to build jfrog service config", err)
	}

	manager, err := jfrogartifactory.New(serviceConfig)
	if err != nil {
		return utils.NewInternalError("failed to create jfrog client", err)
	}

	params := jfrogServices.NewDownloadParams()
	params.CommonParams = &jfrogUtils.CommonParams{
		Pattern:   filePath,
		Recursive: false,
		Target:    targetDir + string(os.PathSeparator),
	}
	params.Flat = true

	downloaded, failed, err := manager.DownloadFiles(params)
	if err != nil {
		return utils.NewInternalError("failed to download artifact with jfrog client", err)
	}
	if downloaded == 0 || failed > 0 {
		return utils.NewInternalError("jfrog download did not complete successfully", nil)
	}
	return nil
}

func parseVersionMetadata(rawContent, fileName string) (map[string]any, error) {
	lowerName := strings.ToLower(fileName)
	if strings.HasSuffix(lowerName, ".json") {
		return models.ParseJSON(rawContent)
	}
	if strings.HasSuffix(lowerName, ".xml") {
		return parseXMLMetadata(rawContent)
	}

	parsed, err := models.ParseJSON(rawContent)
	if err == nil {
		return parsed, nil
	}
	return parseXMLMetadata(rawContent)
}

type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Text    string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

func parseXMLMetadata(rawContent string) (map[string]any, error) {
	var root xmlNode
	if err := xml.Unmarshal([]byte(rawContent), &root); err != nil {
		return nil, utils.NewAPIError("failed to parse version metadata", err)
	}
	return map[string]any{
		root.XMLName.Local: xmlNodeToAny(root),
	}, nil
}

func xmlNodeToAny(node xmlNode) any {
	result := map[string]any{}
	for _, attr := range node.Attrs {
		result["@"+attr.Name.Local] = attr.Value
	}
	text := strings.TrimSpace(node.Text)
	if len(node.Nodes) == 0 {
		if len(result) == 0 {
			return text
		}
		if text != "" {
			result["#text"] = text
		}
		return result
	}
	for _, child := range node.Nodes {
		value := xmlNodeToAny(child)
		existing, exists := result[child.XMLName.Local]
		if !exists {
			result[child.XMLName.Local] = value
			continue
		}
		switch typed := existing.(type) {
		case []any:
			result[child.XMLName.Local] = append(typed, value)
		default:
			result[child.XMLName.Local] = []any{typed, value}
		}
	}
	if text != "" {
		result["#text"] = text
	}
	return result
}

func skipDownloadIfExisting(targetPath, checksum string) (bool, error) {
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, utils.NewInternalError("failed to stat existing artifact file", err)
	}
	if info.IsDir() {
		return false, nil
	}
	if checksum == "" {
		return true, nil
	}
	matched, err := verifyFileHash(targetPath, checksum)
	if err != nil {
		return false, err
	}
	return matched, nil
}

func verifyFileHash(targetPath, expected string) (bool, error) {
	file, err := os.Open(targetPath)
	if err != nil {
		return false, utils.NewInternalError("failed to open existing artifact file", err)
	}
	defer file.Close()

	hasher, err := newHasher(expected)
	if err != nil {
		return false, err
	}
	if _, err := io.Copy(hasher, file); err != nil {
		return false, utils.NewInternalError("failed to hash existing artifact file", err)
	}
	actual := fmt.Sprintf("%x", hasher.Sum(nil))
	return strings.EqualFold(actual, strings.TrimSpace(expected)), nil
}

func newHasher(expected string) (hashWriter, error) {
	switch len(strings.TrimSpace(expected)) {
	case 32:
		return md5.New(), nil
	case 40:
		return sha1.New(), nil
	case 64:
		return sha256.New(), nil
	case 128:
		return sha512.New(), nil
	default:
		return nil, utils.NewInvalidInputError("unsupported artifact checksum length", nil)
	}
}

type hashWriter interface {
	io.Writer
	Sum(b []byte) []byte
}

func mustJSON(v any) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

func stringPtr(v string) *string {
	return &v
}

// compareModePtr keeps the field absent when the caller did not pick a mode,
// so the server keeps applying its default strategy.
func compareModePtr(mode string) *string {
	if mode == "" {
		return nil
	}
	return stringPtr(mode)
}
