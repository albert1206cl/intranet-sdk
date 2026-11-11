package tests

import (
	"errors"
	"testing"

	"github.com/hujia-team/intranet-sdk"
	"github.com/hujia-team/intranet-sdk/models"
)

// 验证 CompareImages 的业务错误 sentinel 映射（api_design v5.5.0 §5.4.2）。
// 环境变量驱动（INTRANET_BASE_URL 等），未配置时由 NewTestClient 报错跳过。
func TestCompareImagesSentinelMapping(t *testing.T) {
	client, err := NewTestClient()
	if err != nil {
		t.Skipf("测试环境未配置: %v", err)
	}

	// 场景 1：跨仓库比较 → FailedPrecondition → ErrImageDiffPrecondition
	cross := &models.ImageDiffReq{
		ImgFullnameA: "docker.minieye.tech/pdcl/trainer:v0.1-2a03e239b9b7bec604c6b155f2a0bd661e911c2c",
		ImgFullnameB: "docker.minieye.tech/devops/muster-collector:mr-81-64597c1",
	}
	if _, err := client.Image.CompareImages(cross); !errors.Is(err, intranet.ErrImageDiffPrecondition) {
		t.Fatalf("跨仓库场景应命中 ErrImageDiffPrecondition, got %v", err)
	}

	// 场景 2：镜像不存在 → NotFound → ErrImageDiffImageNotFound
	missing := &models.ImageDiffReq{
		ImgFullnameA: "docker.minieye.tech/pdcl/trainer:v0.1-does-not-exist",
		ImgFullnameB: "docker.minieye.tech/pdcl/trainer:v0.1-2a03e239b9b7bec604c6b155f2a0bd661e911c2c",
	}
	if _, err := client.Image.CompareImages(missing); !errors.Is(err, intranet.ErrImageDiffImageNotFound) {
		t.Fatalf("镜像不存在场景应命中 ErrImageDiffImageNotFound, got %v", err)
	}

	// 场景 3：跨分支（历史分叉）正常返回 distant Changelog
	forked := &models.ImageDiffReq{
		ImgFullnameA: "docker.minieye.tech/devops/muster-collector:mr-81-64597c1",
		ImgFullnameB: "docker.minieye.tech/devops/muster-collector:mr-82-c22d246",
	}
	result, err := client.Image.CompareImages(forked)
	if err != nil {
		t.Fatalf("跨分支比较应成功: %v", err)
	}
	if result.GitRepo != "lvlin/muster-backend" || result.TotalCommits == 0 {
		t.Fatalf("跨分支结果异常: repo=%s total=%d", result.GitRepo, result.TotalCommits)
	}
	for _, c := range result.Commits {
		if c.Hash == result.CommitA.Hash {
			t.Fatal("distant 模式下 older 不应出现在 commits 列表")
		}
	}
}
