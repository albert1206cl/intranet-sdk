# 制品能力使用说明

本文档只聚焦 `sdk.Artifact`。

## 能力列表

当前 Go SDK 已提供这些制品相关能力：

- `sdk.Artifact.ListArtifacts`
- `sdk.Artifact.GetArtifactByID`
- `sdk.Artifact.GetArtifactByName`
- `sdk.Artifact.GetArtifactByCommitHash`
- `sdk.Artifact.CheckExistsByCommitHash`
- `sdk.Artifact.CheckExistsByName`
- `sdk.Artifact.PrepareDownloadByArtifactID`
- `sdk.Artifact.DownloadByArtifactID`
- `sdk.Artifact.PrepareDownloadByCommitHash`
- `sdk.Artifact.DownloadByCommitHash`
- `sdk.Artifact.DownloadByName`
- `sdk.Artifact.GetVersionMetadataByCommitHash`
- `sdk.Artifact.GetChildArtifactHashesByCommitHash`
- `sdk.Artifact.GetArtifactCommitDiff`
- `sdk.Artifact.GetArtifactChangelog`
- `sdk.Artifact.GetArtifactChangelogFiles`
- `sdk.Artifact.GetArtifactTagSchema`
- `sdk.Artifact.GetArtifactLabelMode`
- `sdk.Artifact.UsesFlatLabels`
- `sdk.Artifact.ParseArtifactTags`
- `sdk.Artifact.GetParsedArtifactTags`
- `sdk.Artifact.PatchArtifactTags`
- `sdk.Artifact.PatchArtifactLabels`
- `sdk.Artifact.ListArtifactsByLabels`
- `sdk.Artifact.GetJfrogToken`
- `sdk.Artifact.GetArtifactDownloadURL`
- `sdk.Artifact.GetArtifactDownloadURLByName`

这份文档不展开每个返回字段。日常开发建议直接看 `sdk.Artifact`
接口、`models` 类型定义和测试用例，再让 AI 基于签名生成业务代码。

## 按名称获取制品详情

```go
artifact, err := sdk.Artifact.GetArtifactByName(
	"D4Q2_V00.29.00_202603272037_release_snapshot_aarch64",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
)
if err != nil {
	return err
}

fmt.Printf("artifact id: %d\n", *artifact.ID)
fmt.Printf("commit hash: %s\n", *artifact.CommitHash)
```

## 按 commit hash 获取根制品

```go
artifact, err := sdk.Artifact.GetArtifactByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
)
if err != nil {
	return err
}

fmt.Printf("root artifact: %s (%s)\n", *artifact.Name, *artifact.Type)
```

注意：

- SDK 现在直接调用服务端精确 `commit_hash` 查询接口
- 建议传 `ArtifactType`
  常见值是 `pkg`、`app`、`mcu`、`bsp`、`data_proto`
- `ArtifactLookupOptions.Platform` 是 `*string`
  - `nil` 表示不按平台过滤
  - `&"linux"` 表示按 `linux` 过滤
  - `&""` 表示显式匹配空平台制品

## 文件名与展示名

- `artifact.name` 是展示名，不要求包含 `commit_hash`
- 制品仓库里的物理文件名仍要求在文件末尾、扩展名前包含 `commit_hash`
- 例如：
  - `data_proto_G1Q3_aarch64-015ce81dc113ad8bf3d5c0c5c4126b16.tar.gz`
  - `1.0.0-9a571910494efb10404e75dab6b7b671.tar.gz`
- 当你在 SDK 里按名称查询时，传的是展示名；当你解析物理文件或构建消息时，要以文件名末尾的 hash 作为 `commit_hash` 来源之一

## 存在性检查

```go
exists, err := sdk.Artifact.CheckExistsByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
)
if err != nil {
	return err
}

fmt.Printf("exists by commit hash: %v\n", exists)
```

说明：

- `CheckExistsByCommitHash` 判断的是“实体制品是否存在”
- 只有服务端能定位到制品，且返回的 `fullPath` 和 `fileHash` 都非空时，SDK 才会返回 `true`
- 仅存在虚拟制品、聚合制品，或者 `fullPath` / `fileHash` 缺失的残缺记录时，会返回 `false`

如果要显式查询 `platform=""` 的制品，可以这样传：

```go
emptyPlatform := ""
artifact, err := sdk.Artifact.GetArtifactByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "app",
		Platform:     &emptyPlatform,
	},
)
```

也可以按名称做唯一性检查：

```go
exists, err := sdk.Artifact.CheckExistsByName(
	"D4Q2_V00.29.00_202603272037_release_snapshot_aarch64",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
)
```

## 下载计划与下载

推荐顺序：

- 已经拿到制品 `id` 时，优先使用 `PrepareDownloadByArtifactID` / `DownloadByArtifactID`
- 只有 `commit hash` 时，再使用 `PrepareDownloadByCommitHash` / `DownloadByCommitHash`
- `DownloadByName` 仅适合名称在当前筛选条件下唯一的场景

## 按 artifact ID 下载

这是 SDK 当前最稳妥的下载方式，因为不依赖名称或 `commit_hash` 的唯一性。

```go
plan, err := sdk.Artifact.DownloadByArtifactID(
	artifactID,
	"./downloads",
)
if err != nil {
	return err
}

fmt.Printf("downloaded to: %s\n", plan.TargetPath)
```

如果想先看下载计划，再决定是否真正下载：

```go
plan, err := sdk.Artifact.PrepareDownloadByArtifactID(
	artifactID,
	"./downloads",
)
if err != nil {
	return err
}

fmt.Printf("target path: %s\n", plan.TargetPath)
fmt.Printf("jfrog file path: %s\n", plan.DownloadURL.FilePath)
fmt.Printf("checksum: %s\n", plan.Checksum)
```

推荐先准备下载计划，再决定是否执行真正下载。

```go
plan, err := sdk.Artifact.PrepareDownloadByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
	"./downloads",
)
if err != nil {
	return err
}

fmt.Printf("target path: %s\n", plan.TargetPath)
fmt.Printf("jfrog file path: %s\n", plan.DownloadURL.FilePath)
fmt.Printf("checksum: %s\n", plan.Checksum)
```

执行下载时，SDK 内部会复用 JFrog 原生 Go client：

```go
plan, err := sdk.Artifact.DownloadByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
	"./downloads",
)
if err != nil {
	return err
}

fmt.Printf("downloaded to: %s\n", plan.TargetPath)
```

也可以按名称直接下载：

```go
_, err := sdk.Artifact.DownloadByName(
	"D4Q2_V00.29.00_202603272037_release_snapshot_aarch64",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
	"./downloads",
)
```

## 版本元数据

```go
metadata, err := sdk.Artifact.GetVersionMetadataByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
)
if err != nil {
	return err
}

fmt.Printf("metadata file: %s\n", *metadata.MetadataFileName)
fmt.Printf("raw size: %d\n", len(*metadata.RawContent))
fmt.Printf("parsed fields: %d\n", len(metadata.Parsed))
```

说明：

- `version.json` / `mcu_version.json` 会解析为 JSON map
- `bsp_version.xml` 会被转成通用 map 结构
- 原始内容始终保留在 `RawContent`

## 获取递归子制品 hashes

```go
childHashes, err := sdk.Artifact.GetChildArtifactHashesByCommitHash(
	"89a84fcee9c8db4c7d8ccb3547cfcc0a",
	&models.ArtifactLookupOptions{
		ArtifactType: "pkg",
	},
)
if err != nil {
	return err
}

fmt.Printf("child count: %d\n", len(childHashes.ChildHashes))
for _, item := range childHashes.ChildHashes {
	if item.Name != nil && item.CommitHash != nil {
		fmt.Printf("%s -> %s\n", *item.Name, *item.CommitHash)
	}
}
```

底层逻辑：

1. 调 `/aiplorer/artifact/by-commit-hash`
2. 服务端精确定位根制品
3. 再从详情里的递归 `dependencies` 提取所有子制品的 `commit_hash`

## 提交对比与 compare mode

`GetArtifactCommitDiff` 对比两个制品的提交历史，默认由服务端按 `direct` 处理。
当两个制品的提交历史出现分叉时，可以显式选择 `distant`，让服务端改用逐仓库的远距离对比：

```go
diff, err := sdk.Artifact.GetArtifactCommitDiff(
	olderArtifactID,
	newerArtifactID,
	&models.ArtifactCommitDiffOptions{CompareMode: "distant"},
)
if err != nil {
	return err
}

fmt.Printf("changed repos: %d\n", diff.ChangedRepoCount)
for _, repo := range diff.RepoDiffs {
	if repo.RepositoryName != nil {
		fmt.Printf("repo %s: %d commits\n", *repo.RepositoryName, len(repo.Commits))
	}
}
```

`GetArtifactChangelog`（`/commit-diff-v2`）通过 `ArtifactChangelogOptions.CompareMode` 支持同一个取值。

注意：

- 不传 options，或 `CompareMode` 留空时，服务端按 `direct` 处理，与旧行为一致
- 只支持 `direct` 和 `distant`，其它取值服务端会返回参数错误

## 标签与 schema

```go
schema, err := sdk.Artifact.GetArtifactTagSchema("0.2.0")
if err != nil {
	return err
}

parsed, err := sdk.Artifact.ParseArtifactTags(rawTagsJSON, schema)
if err != nil {
	return err
}

fmt.Printf("parsed tag keys: %d\n", len(parsed))
```

也可以直接通过制品 ID 获取解析结果：

```go
parsed, err := sdk.Artifact.GetParsedArtifactTags(artifactID)
if err != nil {
	return err
}
```

### 扁平标签（0.3.0 起）

制品标签 0.3.0 起引入了扁平标签文法，schema 文档中通过
`"label_mode": "flat"` 声明，载荷固定为 `{schema_version, labels}`：

```json
{
  "schema_version": "0.3.0",
  "labels": {
    "eval.L1.sim.pass_rate.ratio": "0.9-1.0",
    "eval.L1.sim.urls": ["https://example.com/report"],
    "gate.L2": "PASS"
  }
}
```

先判断当前 schema 属于哪套文法。未声明 `label_mode` 的历史 schema
（0.1.0 / 0.2.0）按 `models.LabelModeLegacy` 处理：

```go
mode, err := sdk.Artifact.GetArtifactLabelMode("0.3.0") // models.LabelModeFlat
if err != nil {
	return err
}

flat, err := sdk.Artifact.UsesFlatLabels("0.3.0")
if err != nil {
	return err
}
```

写入扁平标签用 `PatchArtifactLabels`。它走 `patch-tags`，自动补上
`schema_version`，并在 schema 不是扁平文法时直接报错，避免把扁平键
误写进 legacy 制品：

```go
_, err = sdk.Artifact.PatchArtifactLabels(
	artifactID,
	map[string]any{
		"eval.L1.sim.pass_rate.ratio": "0.9-1.0",
		"eval.L1.sim.urls":            []any{"https://example.com/report"},
	},
	"0.3.0",
)
```

约束由服务端整体校验：`eval.*` / `gate.*` 不能用 `nil` 删除，
`custom.*` 可以；提交 `eval.*` 值键时同一请求必须带上对应的
`eval.{loop}.{module}.urls`。

按标签检索制品用 `ListArtifactsByLabels`，底层走 `artifact/list`
的 `tagFilter`，多个条件之间是 AND：

```go
result, err := sdk.Artifact.ListArtifactsByLabels(
	map[string]any{
		"gate.L2":                     "PASS",
		"eval.L1.sim.pass_rate.ratio": "0.9-1.0",
	},
	"0.3.0",
	1,
	20,
)
```

它同样会先确认目标 schema 是扁平文法，避免用扁平键查 legacy 制品
时静默返回空结果。需要做更自由的组合过滤时，可以直接给
`models.ArtifactListReq` 设置 `TagFilter` / `TagSchemaVersion`。

`PatchArtifactTags` 是底层入口，直接用 JSON Merge Patch 语义提交
`tagsPatch`；`PatchArtifactLabels` 和 `ListArtifactsByLabels` 都是
在此基础上针对扁平标签的语义封装。

## JFrog token 与下载地址

```go
token, err := sdk.Artifact.GetJfrogToken("D4Q2")
if err != nil {
	return err
}

downloadURL, err := sdk.Artifact.GetArtifactDownloadURL(artifactID, "artifact")
if err != nil {
	return err
}

fmt.Printf("jfrog token type: %s\n", token.TokenType)
fmt.Printf("download file: %s\n", downloadURL.FileName)
```

## 已验证样例

这些样例已经在正式服验证通过：

- `pkg` 根制品 `89a84fcee9c8db4c7d8ccb3547cfcc0a`
  根制品：`D4Q2_V00.29.00_202603272037_release_snapshot_aarch64`
  递归子制品 hash 数量：`20`
- `app` 根制品 `60e604273314bafa694b30438bcc0c9a`
  根制品：`app-apa/perception-1.0.0`
  递归子制品 hash 数量：`0`
