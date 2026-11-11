// Package services - image diff business error mapping.
//
// 服务端（intranet-aiplorer）在 gRPC code 之上以稳定 msg 标识区分镜像差异
// 比较的业务错误分支。本文件将该约定映射为 sentinel error，调用方使用
// errors.Is 判定，无需匹配 msg 文本。
package services

import (
	"errors"
	"strings"

	"github.com/hujia-team/intranet-sdk/utils"
)

// 服务端 gRPC code（镜像差异接口透传使用）。
const (
	serverCodeInvalidArgument    = 3
	serverCodeDeadlineExceeded   = 4
	serverCodeNotFound           = 5
	serverCodeAlreadyExists      = 6
	serverCodePermissionDenied   = 7
	serverCodeFailedPrecondition = 9
	serverCodeUnavailable        = 14
)

// 稳定 msg 标识（服务端 status message 以标识串开头；网关 i18n 回落格式
// "rpc error: code = ... desc = <标识>: ..." 中同样包含标识，故用子串匹配）。
const tokenCommitNotFound = "commit_not_found"

var (
	// ErrImageDiffInvalidArgument：请求参数非法（镜像引用缺少 Tag/Digest、格式错误、分页参数越界）。
	ErrImageDiffInvalidArgument = errors.New("image diff: invalid argument")
	// ErrImageDiffAmbiguous：镜像引用命中多条记录（Digest 前缀过短），需加长前缀后重查。
	ErrImageDiffAmbiguous = errors.New("image diff: image reference is ambiguous")
	// ErrImageDiffCommitNotFound：镜像关联的 Commit 在 GitLab 上不可达，
	// 典型诱因是构建分支被删除后 Commit 对象被回收。
	ErrImageDiffCommitNotFound = errors.New("image diff: git commit not found on gitlab")
	// ErrImageDiffImageNotFound：镜像引用未命中任何镜像。
	ErrImageDiffImageNotFound = errors.New("image diff: image not found")
	// ErrImageDiffPrecondition：比较前置条件不满足（如跨仓库、Commit 信息不完整、无关联 Commit）。
	ErrImageDiffPrecondition = errors.New("image diff: precondition failed")
	// ErrImageDiffUpstream：上游 GitLab 错误（权限不足 / GitLab 不可用 / 超时）。
	// 是否可重试按 ServerCode 区分：4（超时）、14（不可用）可重试，7（权限）不可。
	ErrImageDiffUpstream = errors.New("image diff: upstream gitlab error")
)

// mapImageDiffBusinessError 将镜像差异接口的业务错误响应（code + msg）
// 映射为携带 sentinel 的 SDKError。稳定标识命中优先，gRPC code 兜底；
// 未命中任何已知 code 时保留原始 code/msg 落入通用 ErrCodeAPIError。
func mapImageDiffBusinessError(code int, msg string) error {
	switch {
	case code == serverCodeNotFound && strings.Contains(msg, tokenCommitNotFound):
		return newImageDiffError(utils.ErrCodeNotFound, ErrImageDiffCommitNotFound, code, msg)
	case code == serverCodeNotFound:
		return newImageDiffError(utils.ErrCodeNotFound, ErrImageDiffImageNotFound, code, msg)
	case code == serverCodeFailedPrecondition:
		return newImageDiffError(utils.ErrCodeAPIError, ErrImageDiffPrecondition, code, msg)
	case code == serverCodeInvalidArgument:
		return newImageDiffError(utils.ErrCodeInvalidInput, ErrImageDiffInvalidArgument, code, msg)
	case code == serverCodeAlreadyExists:
		return newImageDiffError(utils.ErrCodeAPIError, ErrImageDiffAmbiguous, code, msg)
	case code == serverCodePermissionDenied, code == serverCodeUnavailable, code == serverCodeDeadlineExceeded:
		return newImageDiffError(utils.ErrCodeAPIError, ErrImageDiffUpstream, code, msg)
	default:
		return newImageDiffError(utils.ErrCodeAPIError, nil, code, msg)
	}
}

func newImageDiffError(sdkCode utils.ErrorCode, sentinel error, serverCode int, msg string) *utils.SDKError {
	return &utils.SDKError{
		Code:       sdkCode,
		Message:    msg,
		Err:        sentinel,
		ServerCode: serverCode,
	}
}
