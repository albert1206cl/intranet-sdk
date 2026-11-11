package services

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/hujia-team/intranet-sdk/models"
	"github.com/hujia-team/intranet-sdk/utils"
)

// 六类业务场景的映射矩阵：每个服务端 code 都应有可 errors.Is 的 sentinel。
func TestMapImageDiffBusinessError(t *testing.T) {
	cases := []struct {
		name       string
		code       int
		msg        string
		sentinel   error
		sdkCode    utils.ErrorCode
		serverCode int
	}{
		{"invalid argument", 3, "image fullname must include tag or digest", ErrImageDiffInvalidArgument, utils.ErrCodeInvalidInput, 3},
		{"image not found", 5, "image not found", ErrImageDiffImageNotFound, utils.ErrCodeNotFound, 5},
		{"commit_not_found on NotFound", 5, "commit_not_found: git commit object not found on gitlab, source branch may have been deleted", ErrImageDiffCommitNotFound, utils.ErrCodeNotFound, 5},
		{"ambiguous reference", 6, "image reference is ambiguous", ErrImageDiffAmbiguous, utils.ErrCodeAPIError, 6},
		{"precondition failed", 9, "commits belong to different repositories", ErrImageDiffPrecondition, utils.ErrCodeAPIError, 9},
		{"gitlab permission denied", 7, "gitlab permission denied", ErrImageDiffUpstream, utils.ErrCodeAPIError, 7},
		{"gitlab unavailable", 14, "gitlab unavailable", ErrImageDiffUpstream, utils.ErrCodeAPIError, 14},
		{"deadline exceeded", 4, "context deadline exceeded", ErrImageDiffUpstream, utils.ErrCodeAPIError, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := mapImageDiffBusinessError(tc.code, tc.msg)
			var sdkErr *utils.SDKError
			if !errors.As(err, &sdkErr) {
				t.Fatalf("expected *utils.SDKError, got %T", err)
			}
			if sdkErr.Code != tc.sdkCode {
				t.Fatalf("sdk code = %v, want %v", sdkErr.Code, tc.sdkCode)
			}
			if sdkErr.ServerCode != tc.serverCode {
				t.Fatalf("server code = %d, want %d", sdkErr.ServerCode, tc.serverCode)
			}
			if !strings.Contains(sdkErr.Message, tc.msg) {
				t.Fatalf("message %q must contain original msg %q", sdkErr.Message, tc.msg)
			}
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("errors.Is miss for %v", tc.sentinel)
			}
		})
	}
}

// 未知 code 保留原始信息落入通用 APIError，不带 sentinel。
func TestMapImageDiffBusinessErrorUnknownCode(t *testing.T) {
	err := mapImageDiffBusinessError(13, "internal error")
	var sdkErr *utils.SDKError
	if !errors.As(err, &sdkErr) || sdkErr.Code != utils.ErrCodeAPIError || sdkErr.ServerCode != 13 {
		t.Fatalf("unexpected error: %v", err)
	}
	if sdkErr.Err != nil {
		t.Fatalf("unknown code must not carry a sentinel, got %v", sdkErr.Err)
	}
}

// 网关 TransError 的 i18n 回落格式为 "rpc error: code = NotFound desc = commit_not_found: ..."，
// 标识子串仍存在，映射必须命中。
func TestCompareImagesErrorMatchesI18nFallbackFormat(t *testing.T) {
	service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":5,"msg":"rpc error: code = NotFound desc = commit_not_found: git commit object not found on gitlab, source branch may have been deleted"}`))
	})

	_, err := service.CompareImages(&models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b"})
	if !errors.Is(err, ErrImageDiffCommitNotFound) {
		t.Fatalf("expected ErrImageDiffCommitNotFound via errors.Is, got %v", err)
	}
}

func TestCompareImagesSentinelsEndToEnd(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		sentinel error
	}{
		{"invalid argument", `{"code":3,"msg":"image fullname must include tag or digest"}`, ErrImageDiffInvalidArgument},
		{"ambiguous", `{"code":6,"msg":"image reference is ambiguous"}`, ErrImageDiffAmbiguous},
		{"upstream unavailable", `{"code":14,"msg":"gitlab unavailable"}`, ErrImageDiffUpstream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newImageTestService(t, func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := service.CompareImages(&models.ImageDiffReq{ImgFullnameA: "a", ImgFullnameB: "b"})
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("expected sentinel hit, got %v", err)
			}
		})
	}
}
