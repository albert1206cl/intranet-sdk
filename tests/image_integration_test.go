package tests

import (
	"errors"
	"testing"

	"github.com/hujia-team/intranet-sdk/models"
	"github.com/hujia-team/intranet-sdk/services"
	"github.com/hujia-team/intranet-sdk/utils"
)

// 镜像数据不硬编码: 通过 devops/sre-agent 仓库动态发现真实镜像,
// 使本测试可运行在任意环境(生产/测试), 不依赖特定库里的镜像记录。
const imageDiscoveryRepo = "devops/sre-agent"

// discoverImages 列出发现仓库下最新的镜像记录。
func discoverImages(t *testing.T) []models.ImageInfo {
	t.Helper()
	client, err := NewTestClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	repo := imageDiscoveryRepo
	pageSize := uint64(20)
	resp, err := client.Image.ListImages(&models.ImageListReq{
		GitRepo:  &repo,
		PageSize: &pageSize,
	})
	if err != nil {
		t.Fatalf("发现镜像失败: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatalf("仓库 %s 无镜像记录, 无法继续验证", imageDiscoveryRepo)
	}
	return resp.Items
}

// TestImageListFlow 通过真实网关验证 ListImages: STS 认证、query 编码与响应解析。
func TestImageListFlow(t *testing.T) {
	client, err := NewTestClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	t.Log("=== 测试 ListImages（按代码仓过滤） ===")
	repo := imageDiscoveryRepo
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
	for i, item := range resp.Items {
		if i >= 3 {
			break
		}
		tag := "<无>"
		if item.ImgTag != nil {
			tag = *item.ImgTag
		}
		t.Logf("  id=%d %s:%s", item.ID, item.ImgRepo, tag)
	}
	if resp.Total == 0 || len(resp.Items) == 0 {
		t.Fatal("期望仓库至少有一条镜像记录")
	}

	t.Log("=== 测试 ListImages（范围查询，省略 Tag/Digest） ===")
	first := resp.Items[0]
	fullname := first.ImgRegistry + "/" + first.ImgProject + "/" + first.ImgRepo
	ranged, err := client.Image.ListImages(&models.ImageListReq{ImgFullname: &fullname})
	if err != nil {
		t.Fatalf("ListImages 范围查询失败: %v", err)
	}
	t.Logf("仓库 %s total=%d", fullname, ranged.Total)
	if ranged.Total == 0 {
		t.Fatal("期望范围查询至少命中一条记录")
	}
}

// TestImageCompareFlow 通过真实网关验证 CompareImages 的完整契约:
// 对象形态、含新不含旧、A/B 角色重排、同镜像空 Changelog 与 Digest 前缀定位。
// 依赖服务端 data 对象化改造与同 Commit 空差异修复上线。
func TestImageCompareFlow(t *testing.T) {
	client, err := NewTestClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	items := discoverImages(t)

	imgA, imgB, diff, ok := imagePickComparablePair(t, client.Image, items)
	if !ok {
		t.Fatal("所有相邻 commit 对均不可比较(历史分叉), 请补充同分支的镜像数据")
	}
	nameA, nameB := imageFullnameOf(imgA), imageFullnameOf(imgB)
	t.Logf("选用镜像: A=%s, B=%s", nameA, nameB)

	if diff.GitRepo != imageDiscoveryRepo {
		t.Fatalf("git_repo = %q, 期望 %s", diff.GitRepo, imageDiscoveryRepo)
	}
	t.Logf("旧镜像 id=%d, 新镜像 id=%d, total_commits=%d", diff.ImgIDA, diff.ImgIDB, diff.TotalCommits)
	if diff.TotalCommits == 0 || len(diff.Commits) == 0 {
		t.Fatal("期望两个不同镜像之间存在 Changelog 差异")
	}
	for _, commit := range diff.Commits {
		t.Logf("  %s %s (%s)", commit.ShortHash, commit.Title, commit.Author)
	}
	if last := diff.Commits[len(diff.Commits)-1]; last.Hash != diff.CommitB.Hash {
		t.Fatalf("commits 末条 %s 应为新镜像 Commit %s(含新不含旧)", last.ShortHash, diff.CommitB.ShortHash)
	}

	t.Log("=== 测试 A/B 交换后角色重排 ===")
	swapped, err := client.Image.CompareImages(&models.ImageDiffReq{ImgFullnameA: nameB, ImgFullnameB: nameA})
	if err != nil {
		t.Fatalf("CompareImages 交换比较失败: %v", err)
	}
	if swapped.ImgIDA != diff.ImgIDA || swapped.ImgIDB != diff.ImgIDB {
		t.Fatalf("交换入参后 img_id_a=%d/img_id_b=%d, 期望仍为 %d/%d(旧/新)",
			swapped.ImgIDA, swapped.ImgIDB, diff.ImgIDA, diff.ImgIDB)
	}
	t.Logf("img_id_a=%d 仍为旧侧, 角色重排生效", swapped.ImgIDA)

	t.Log("=== 测试同镜像比较(期望空 Changelog) ===")
	same, err := client.Image.CompareImages(&models.ImageDiffReq{ImgFullnameA: nameA, ImgFullnameB: nameA})
	if err != nil {
		t.Fatalf("CompareImages 同镜像比较失败: %v", err)
	}
	if same.TotalCommits != 0 || len(same.Commits) != 0 {
		t.Fatalf("同镜像比较期望空 Changelog, got total=%d len=%d", same.TotalCommits, len(same.Commits))
	}
	t.Log("total_commits=0, commits 为空")

	t.Log("=== 测试 Digest 前 8 位定位 ===")
	prefix := imgB.Digest[len("sha256:"):]
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	digestRef := imgB.ImgRegistry + "/" + imgB.ImgProject + "/" + imgB.ImgRepo + "@sha256:" + prefix
	byDigest, err := client.Image.CompareImages(&models.ImageDiffReq{ImgFullnameA: digestRef, ImgFullnameB: nameA})
	if err != nil {
		t.Fatalf("CompareImages Digest 前缀定位失败: %v", err)
	}
	if byDigest.GitRepo != imageDiscoveryRepo {
		t.Fatalf("Digest 前缀命中 git_repo=%q, 期望 %s", byDigest.GitRepo, imageDiscoveryRepo)
	}
	t.Logf("引用 %s 命中成功", digestRef)
}

// TestImageErrorPaths 验证业务错误与本地参数校验两条错误路径。
func TestImageErrorPaths(t *testing.T) {
	client, err := NewTestClient()
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	t.Log("=== 测试不存在的镜像返回业务错误 ===")
	_, err = client.Image.CompareImages(&models.ImageDiffReq{
		ImgFullnameA: "docker.minieye.tech/" + imageDiscoveryRepo + ":v9.9-not-exist",
		ImgFullnameB: "docker.minieye.tech/" + imageDiscoveryRepo + ":v9.9-not-exist-b",
	})
	var sdkErr *utils.SDKError
	if !errors.As(err, &sdkErr) || sdkErr.Code != utils.ErrCodeAPIError {
		t.Fatalf("期望 ErrCodeAPIError 业务错误, got %v", err)
	}
	t.Logf("SDKError(APIError): %s", sdkErr.Message)

	t.Log("=== 测试缺失 fullname 本地校验拦截 ===")
	_, err = client.Image.CompareImages(&models.ImageDiffReq{ImgFullnameB: "docker.minieye.tech/x/y:v1"})
	if !errors.As(err, &sdkErr) || sdkErr.Code != utils.ErrCodeInvalidInput {
		t.Fatalf("期望 ErrCodeInvalidInput 本地校验错误, got %v", err)
	}
	t.Logf("SDKError(InvalidInput): %s", sdkErr.Message)
}

// imagePickComparablePair 在带 Tag、commit 不同的相邻镜像对中找到第一对
// 可比较(祖先关系)的组合; 仓库存在分支开发时, 历史分叉的对会被跳过。
func imagePickComparablePair(t *testing.T, client services.ImageService, items []models.ImageInfo) (models.ImageInfo, models.ImageInfo, *models.ImageDiffData, bool) {
	t.Helper()
	tagged := make([]models.ImageInfo, 0, len(items))
	for _, item := range items {
		if item.ImgTag != nil && *item.ImgTag != "" && item.GitCommitHash != nil && *item.GitCommitHash != "" {
			tagged = append(tagged, item)
		}
	}
	for i := 0; i+1 < len(tagged); i++ {
		a, b := tagged[i], tagged[i+1]
		if *a.GitCommitHash == *b.GitCommitHash {
			continue
		}
		diff, err := client.CompareImages(&models.ImageDiffReq{
			ImgFullnameA: imageFullnameOf(a), ImgFullnameB: imageFullnameOf(b),
		})
		if err != nil {
			t.Logf("跳过 %s vs %s: %v", *a.GitCommitHash, *b.GitCommitHash, err)
			continue
		}
		return a, b, diff, true
	}
	var zero models.ImageInfo
	return zero, zero, nil, false
}

func imageFullnameOf(item models.ImageInfo) string {
	return item.ImgRegistry + "/" + item.ImgProject + "/" + item.ImgRepo + ":" + *item.ImgTag
}
