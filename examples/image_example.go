package main

import (
	"fmt"
	"os"

	intranet "github.com/hujia-team/intranet-sdk"
	"github.com/hujia-team/intranet-sdk/models"
)

// 镜像元数据查询与差异比较示例:
//
//	INTRANET_ACCESS_KEY_ID=xxx INTRANET_ACCESS_KEY_SECRET=yyy \
//	  go run examples/image_example.go
//
// 可选 INTRANET_BASE_URL 指定网关地址(默认生产环境, 如
// https://i-test.minieye.tech/sys-api 指向测试环境)。
func main() {
	accessKeyID := os.Getenv("INTRANET_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("INTRANET_ACCESS_KEY_SECRET")
	if accessKeyID == "" || accessKeySecret == "" {
		fmt.Println("错误: 缺少STS凭证环境变量")
		fmt.Println("请设置 INTRANET_ACCESS_KEY_ID 和 INTRANET_ACCESS_KEY_SECRET 环境变量")
		os.Exit(1)
	}

	options := []intranet.Option{
		intranet.WithAccessKeyID(accessKeyID),
		intranet.WithAccessKeySecret(accessKeySecret),
		intranet.WithUserAgent("intranet-sdk-image-example/1.0"),
	}
	if baseURL := os.Getenv("INTRANET_BASE_URL"); baseURL != "" {
		options = append(options, intranet.WithBaseURL(baseURL))
	}

	sdk, err := intranet.NewClient(options...)
	if err != nil {
		fmt.Printf("初始化SDK失败: %v\n", err)
		os.Exit(1)
	}

	// 1. 按代码仓查询镜像列表
	repo := "devops/sre-agent"
	pageNum, pageSize := uint64(1), uint64(5)
	resp, err := sdk.Image.ListImages(&models.ImageListReq{
		GitRepo:  &repo,
		PageNum:  &pageNum,
		PageSize: &pageSize,
	})
	if err != nil {
		fmt.Printf("查询镜像列表失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("仓库 %s 共 %d 条镜像记录, 本页 %d 条:\n", repo, resp.Total, len(resp.Items))
	for _, item := range resp.Items {
		tag := "<无>"
		if item.ImgTag != nil {
			tag = *item.ImgTag
		}
		fmt.Printf("  id=%d %s/%s:%s digest=%s\n", item.ID, item.ImgProject, item.ImgRepo, tag, item.Digest)
	}

	// 2. 比较两条 commit 不同的镜像, 输出 Changelog
	imgA, imgB, ok := pickComparablePair(sdk, resp.Items)
	if !ok {
		fmt.Println("本页未找到 commit 不同的两条带 Tag 镜像, 跳过差异比较")
		return
	}
	diff, err := sdk.Image.CompareImages(&models.ImageDiffReq{
		ImgFullnameA: fullnameOf(imgA),
		ImgFullnameB: fullnameOf(imgB),
	})
	if err != nil {
		fmt.Printf("镜像差异比较失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n镜像差异(%s -> %s):\n", diff.ImgFullnameA, diff.ImgFullnameB)
	fmt.Printf("  git_repo=%s, total_commits=%d\n", diff.GitRepo, diff.TotalCommits)
	fmt.Printf("  旧镜像关联 Commit: %s %s\n", diff.CommitA.ShortHash, diff.CommitA.Title)
	fmt.Printf("  新镜像关联 Commit: %s %s\n", diff.CommitB.ShortHash, diff.CommitB.Title)
	for i, commit := range diff.Commits {
		fmt.Printf("  [%d] %s %s (%s)\n", i+1, commit.ShortHash, commit.Title, commit.Author)
	}
	fmt.Println("\n提示: 响应中的 *_a 恒为旧侧, 由服务端按 Commit 时间重排, 不保证与入参顺序一致")
}

// pickComparablePair 在带 Tag、commit 不同的相邻镜像对中找到第一对
// 可比较(祖先关系)的组合; 仓库存在分支开发时, 历史分叉的对会被跳过。
func pickComparablePair(sdk *intranet.Client, items []models.ImageInfo) (models.ImageInfo, models.ImageInfo, bool) {
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
		if _, err := sdk.Image.CompareImages(&models.ImageDiffReq{
			ImgFullnameA: fullnameOf(a), ImgFullnameB: fullnameOf(b),
		}); err != nil {
			continue
		}
		return a, b, true
	}
	var zero models.ImageInfo
	return zero, zero, false
}

func fullnameOf(item models.ImageInfo) string {
	return item.ImgRegistry + "/" + item.ImgProject + "/" + item.ImgRepo + ":" + *item.ImgTag
}
