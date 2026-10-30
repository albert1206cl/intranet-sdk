package services

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/hujia-team/intranet-sdk/models"
	"github.com/hujia-team/intranet-sdk/utils"
)

func TestCompareImages(t *testing.T) {
	service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/v1/image/diff" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		query := r.URL.Query()
		expected := url.Values{
			"img_fullname_a": {"registry.example.com/project/image:v1"},
			"img_fullname_b": {"registry.example.com/project/image:v2"},
			"page_num":       {"2"},
			"page_size":      {"10"},
		}
		if query.Encode() != expected.Encode() {
			t.Fatalf("unexpected query: got %s, want %s", query.Encode(), expected.Encode())
		}

		_, _ = w.Write([]byte(`{"code":0,"msg":"success","data":{"img_id_a":1,"img_id_b":2,"img_fullname_a":"registry.example.com/project/image:v1","img_fullname_b":"registry.example.com/project/image:v2","git_repo":"group/project","commit_a":{"id":11,"commit_at":"2026-09-01T00:00:00Z","hash":"aaa","short_hash":"a1","branch":"main","title":"init","message":"init","author":"alice","web_url":"https://git/aaa"},"commit_b":{"id":22,"commit_at":"2026-09-02T00:00:00Z","hash":"bbb","short_hash":"b2","title":"fix","message":"fix bug","author":"bob"},"commits":[{"id":33,"commit_at":"2026-09-03T00:00:00Z","hash":"ccc","short_hash":"c3","title":"mid","message":"middle change","author":"carol"}],"total_commits":3}}`))
	})

	result, err := service.CompareImages(&models.ImageDiffReq{
		ImgFullnameA: "registry.example.com/project/image:v1",
		ImgFullnameB: "registry.example.com/project/image:v2",
		PageNum:      uint64Ptr(2),
		PageSize:     uint64Ptr(10),
	})
	if err != nil {
		t.Fatalf("CompareImages error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	item := result
	if item.ImgIDA != 1 || item.ImgIDB != 2 || item.GitRepo != "group/project" || item.TotalCommits != 3 {
		t.Fatalf("unexpected diff item: %#v", item)
	}
	if item.CommitA.ID == nil || *item.CommitA.ID != 11 || item.CommitA.Branch == nil || *item.CommitA.Branch != "main" {
		t.Fatalf("unexpected commit_a: %#v", item.CommitA)
	}
	if item.CommitB.WebURL != nil {
		t.Fatalf("expected nil web_url on commit_b, got %#v", item.CommitB.WebURL)
	}
	if len(item.Commits) != 1 || item.Commits[0].Hash != "ccc" {
		t.Fatalf("unexpected commits: %#v", item.Commits)
	}
}

func TestCompareImagesRequiredAndPaginationValidation(t *testing.T) {
	service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("validation failures must not reach the server")
	})

	cases := []struct {
		name string
		req  *models.ImageDiffReq
		want string
	}{
		{"nil request", nil, "img_fullname_a and img_fullname_b are required"},
		{"missing fullname a", &models.ImageDiffReq{ImgFullnameB: "b"}, "img_fullname_a is required"},
		{"missing fullname b", &models.ImageDiffReq{ImgFullnameA: "a"}, "img_fullname_b is required"},
		{"zero page num", &models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b", PageNum: uint64Ptr(0)}, "page_num must be greater than 0"},
		{"zero page size", &models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b", PageSize: uint64Ptr(0)}, "page_size must be greater than 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CompareImages(tc.req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error %q, got %v", tc.want, err)
			}
			var sdkErr *utils.SDKError
			if !errors.As(err, &sdkErr) || sdkErr.Code != utils.ErrCodeInvalidInput {
				t.Fatalf("unexpected error type: %T %v", err, err)
			}
		})
	}
}

func TestCompareImagesNilData(t *testing.T) {
	service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Encode(); got != "img_fullname_a=a&img_fullname_b=b" {
			t.Fatalf("unexpected query: %s", got)
		}
		_, _ = w.Write([]byte(`{"code":0,"data":null}`))
	})

	result, err := service.CompareImages(&models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b"})
	if err != nil {
		t.Fatalf("CompareImages error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil empty result")
	}
	if result.Commits == nil || len(result.Commits) != 0 {
		t.Fatalf("expected empty commits, got %#v", result.Commits)
	}
}

func TestCompareImagesNormalizesNilCommits(t *testing.T) {
	service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"img_id_a":1,"img_id_b":2,"img_fullname_a":"a","img_fullname_b":"b","git_repo":"group/project","total_commits":0}}`))
	})

	result, err := service.CompareImages(&models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b"})
	if err != nil {
		t.Fatalf("CompareImages error: %v", err)
	}
	if result.Commits == nil || len(result.Commits) != 0 {
		t.Fatalf("expected normalized empty commits, got %#v", result.Commits)
	}
}

func TestCompareImagesBusinessError(t *testing.T) {
	service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":5,"msg":"image not found"}`))
	})

	_, err := service.CompareImages(&models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b"})
	if err == nil || !strings.Contains(err.Error(), "image not found") {
		t.Fatalf("expected business error, got %v", err)
	}
	var sdkErr *utils.SDKError
	if !errors.As(err, &sdkErr) || sdkErr.Code != utils.ErrCodeAPIError {
		t.Fatalf("unexpected error type: %T %v", err, err)
	}
}
