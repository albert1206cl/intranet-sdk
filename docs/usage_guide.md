# 使用指南

本文档只介绍 `intranet-sdk` 的通用使用方式，不展开某个具体服务的细节。

## 安装

```bash
go get github.com/hujia-team/intranet-sdk
```

## 初始化

推荐使用 STS 凭证初始化：

```go
import intranet "github.com/hujia-team/intranet-sdk"

sdk, err := intranet.NewClient(
	intranet.WithBaseURL("https://intranet.minieye.tech/sys-api"),
	intranet.WithAccessKeyID("your_access_key_id"),
	intranet.WithAccessKeySecret("your_access_key_secret"),
)
if err != nil {
	panic(err)
}
```

可用配置项：

- `WithBaseURL`
- `WithAPIKey`
- `WithUserAgent`
- `WithAccessKeyID`
- `WithAccessKeySecret`

## 服务入口

客户端当前暴露这些服务：

- `sdk.User`
- `sdk.Connector`
- `sdk.ApiKey`
- `sdk.Artifact`
- `sdk.ClawSkill`
- `sdk.Image`

使用制品查询、下载、标签或 changelog 能力时，建议先看
`sdk.Artifact` 接口和 `models` 类型定义，再让 AI 按签名生成脚本。

## 最小示例

```go
userInfo, err := sdk.User.GetUserInfo()
if err != nil {
	return err
}

if userInfo.Username != nil {
	fmt.Printf("username: %s\n", *userInfo.Username)
}
```

## OCI 镜像查询

通过 `sdk.Image.ListImages` 查询 OCI 镜像元数据。请求参数会编码为
`GET /api/v1/images` 的 URL query；传入 `nil` 时使用服务端默认分页。

支持的查询字段包括 `page_num`、`page_size`、`git_repo`、`git_branch`、
`img_fullname`、`digest`、`git_commit_short_hash`、`pipeline_id` 和 `job_id`。

```go
images, err := sdk.Image.ListImages(&models.ImageListReq{
	PageNum:  &pageNum,
	PageSize: &pageSize,
	GitRepo:  &gitRepo,
})
if err != nil {
	return err
}
fmt.Printf("image total: %d\n", images.Total)
```

## OCI 镜像差异对比

通过 `sdk.Image.CompareImages` 对比两个 OCI 镜像，获取两镜像对应提交
之间的 commit 差异。请求参数会编码为 `GET /api/v1/image/diff` 的
URL query。

`img_fullname_a` 和 `img_fullname_b` 为必填；`page_num` 和
`page_size` 用于对两镜像之间的 commit 列表分页，可选，设置时必须
大于 0（SDK 会在本地校验并快速失败）。返回单个 `models.ImageDiffData`
对象：`commit_a`/`commit_b` 按服务端判定的旧/新角色返回（不保证与
入参 A/B 顺序对应），`total_commits` 是分页前的 commit 总数，
`commits` 是当前页的 commit 列表（含新侧 Commit、不含旧侧）。

```go
diff, err := sdk.Image.CompareImages(&models.ImageDiffReq{
	ImgFullnameA: "registry.example.com/project/image:v1",
	ImgFullnameB: "registry.example.com/project/image:v2",
})
if err != nil {
	return err
}
fmt.Printf("repo %s: %d commits between images\n", diff.GitRepo, diff.TotalCommits)
```

## 错误处理

SDK 返回的主要是 `*utils.SDKError`：

```go
resp, err := sdk.User.GetUserInfo()
if err != nil {
	if sdkErr, ok := err.(*utils.SDKError); ok {
		fmt.Printf("sdk error: %s\n", sdkErr.Error())
	}
	return err
}
_ = resp
```

## 按场景查看专题文档

- Claw Skill 能力: [claw-skill-usage.md](./claw-skill-usage.md)
- 制品能力: [artifact_usage.md](./artifact_usage.md)
- ApiKey 能力: [apikey-usage.md](./apikey-usage.md)
- 架构说明: [architecture.md](./architecture.md)
