package tests

import (
	"testing"

	"github.com/hujia-team/intranet-sdk/models"
)

// TestImageListFlow 通过真实网关验证 ListImages：STS 认证、query 编码与响应解析。
func TestImageListFlow(t *testing.T) {
	client, err := NewTestClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	t.Log("=== 测试 ListImages（按代码仓过滤） ===")
	repo := "devops/sre-agent"
	pageNum, pageSize := uint64(1), uint64(5)
	resp, err := client.Image.ListImages(&models.ImageListReq{
		GitRepo:  &repo,
		PageNum:  &pageNum,
		PageSize: &pageSize,
	})
	if err != nil {
		t.Fatalf("ListImages 失败: %v", err)
	}
	t.Logf("total=%d, 本页 %d 条", resp.Total, len(resp.Items))
	for _, item := range resp.Items {
		tag := ""
		if item.ImgTag != nil {
			tag = *item.ImgTag
		}
		t.Logf("  id=%d %s/%s:%s digest=%s", item.ID, item.ImgProject, item.ImgRepo, tag, item.Digest)
	}
	if resp.Total == 0 || len(resp.Items) == 0 {
		t.Fatal("期望 devops/sre-agent 至少有一条镜像记录")
	}

	t.Log("=== 测试 ListImages（范围查询，省略 Tag/Digest） ===")
	fullname := "docker.minieye.tech/pdcl/trainer"
	ranged, err := client.Image.ListImages(&models.ImageListReq{ImgFullname: &fullname})
	if err != nil {
		t.Fatalf("ListImages 范围查询失败: %v", err)
	}
	t.Logf("trainer 仓库 total=%d", ranged.Total)
	if ranged.Total == 0 {
		t.Fatal("期望 trainer 仓库至少有一条镜像记录")
	}
}

// TestImageCompareFlow 通过真实网关验证 CompareImages。
// 依赖服务端 data 对象化改造上线（否则解析失败属预期，见测试内注释）。
func TestImageCompareFlow(t *testing.T) {
	client, err := NewTestClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	t.Log("=== 测试 CompareImages（两个不同镜像） ===")
	diff, err := client.Image.CompareImages(&models.ImageDiffReq{
		ImgFullnameA: "docker.minieye.tech/pdcl/trainer:v0.1-3a931714407604cac76b60b21ae6573e42ccad85",
		ImgFullnameB: "docker.minieye.tech/pdcl/trainer:v0.1-2a03e239b9b7bec604c6b155f2a0bd661e911c2c",
	})
	if err != nil {
		t.Fatalf("CompareImages 失败: %v（若服务端尚未部署 data 对象化改造，此失败属预期）", err)
	}
	t.Logf("git_repo=%s, 旧镜像 id=%d, 新镜像 id=%d, total_commits=%d, 本页 %d 条",
		diff.GitRepo, diff.ImgIDA, diff.ImgIDB, diff.TotalCommits, len(diff.Commits))
	if diff.GitRepo != "pdcl/trainer" {
		t.Fatalf("git_repo = %q, 期望 pdcl/trainer", diff.GitRepo)
	}
	if len(diff.Commits) == 0 || diff.TotalCommits == 0 {
		t.Fatal("期望两个不同镜像之间存在 Changelog 差异")
	}

	t.Log("=== 测试 CompareImages（同一镜像，期望空 Changelog） ===")
	same, err := client.Image.CompareImages(&models.ImageDiffReq{
		ImgFullnameA: "docker.minieye.tech/pdcl/trainer:v0.1-2a03e239b9b7bec604c6b155f2a0bd661e911c2c",
		ImgFullnameB: "docker.minieye.tech/pdcl/trainer:v0.1-2a03e239b9b7bec604c6b155f2a0bd661e911c2c",
	})
	if err != nil {
		t.Fatalf("CompareImages 同镜像比较失败: %v", err)
	}
	if same.TotalCommits != 0 || len(same.Commits) != 0 {
		t.Fatalf("同镜像比较期望空 Changelog, got total=%d len=%d（依赖同 Commit 空差异修复上线）",
			same.TotalCommits, len(same.Commits))
	}
}
