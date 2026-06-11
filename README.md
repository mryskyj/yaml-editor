# YAML Struct Editor

Go structをスキーマ定義として使うYAML編集用デスクトップGUIアプリです。

JSON Schemaは使わず、Go structの `yaml` タグと補助タグをもとに、YAMLの補完、検証、スキーマ表示を行います。

## Features

- Monaco EditorによるYAML編集
- 配布用ビルドではMonaco Editor実行時アセットを同梱
- YAML構文エラーの診断
- Go struct由来の未定義キー、型不一致、必須項目不足、enum不一致の検証
- YAMLキー、型、説明、必須、デフォルト値、enumを含む補完候補
- 複数ファイルタブ、未保存状態表示、閉じる前の確認
- 起動時と新規作成時のRootスキーマ必須キー自動補填
- Rootスキーマ初期値の `app/rootschema/defaults.yaml` 定義
- 最近開いたファイルのバックエンド履歴表示とRecentからの再オープン
- Copy / Cut / Paste、Undo / Redo、主要ファイル操作のショートカット
- Toolbarボタンのホバー/フォーカス時ショートカット表示
- スキーマペインでのカーソル位置コンテキスト表示、検索、ツリー表示
- 組み込みRootスキーマと参照用tool schema
- `common.dates`、`common.schedules`、`scenario.steps` の入力補助
- `day_ref` の動的day候補補完
- `tool` 値に応じた `args` の補完・検証
- 起動時オプションによる外部Goソーススキーマ読み込み
- macOS `.app` とWindows GUI exeの配布用ビルド

## Requirements

- Go 1.26.2
- Node.js 24
- npm

## Development

依存関係をインストールします。

```sh
cd frontend
npm install
```

フロントエンドをビルドします。

```sh
cd frontend
npm run build
```

このbuildでは `frontend/dist/vs` にMonaco Editorの実行時アセットも生成されます。
配布用exeはこの同梱アセットを使うため、Monacoの読み込みにCDN接続は不要です。

Go依存ライブラリは `vendor/` に同梱しています。
配布用ビルドとCIテストは `-mod=vendor` を指定し、Go module downloadに依存しません。

アプリを起動します。

```sh
go run ./cmd/yaml-struct-editor
```

## External Schema

外部Goソースをスキーマとして使う場合は、Goファイルを含むディレクトリを指定します。
YAML文書全体を表すroot schemaは、指定ディレクトリ直下の `.go` ファイルから読み込みます。
ルートstruct名は省略でき、YAMLタグ付きフィールドを持ち、他structから参照されていないstructが1つだけの場合は自動検出されます。

```sh
go run ./cmd/yaml-struct-editor --schema-dir schemas/external-sample
```

root候補が複数ある場合は、`--schema-type` で明示します。

```sh
go run ./cmd/yaml-struct-editor --schema-dir schemas/external-sample --schema-type Config
```

root型名が `Config` ではないスキーマも利用できます。

```sh
go run ./cmd/yaml-struct-editor --schema-dir schemas/alternate-sample
```

root schemaの対象は、指定ディレクトリ直下の `.go` ファイルです。`*_test.go` とサブディレクトリは読み込み対象外です。
`tool` / `args` 用の参照スキーマは同じ `--schema-dir` から再帰的に読み込み、サブディレクトリを `.` 区切りの名前空間として扱います。
例えば `cloud/ecs` 配下の `RunTask` は `"cloud.ecs.RunTask"` として補完されます。

対応する型:

- struct
- nested struct
- slice
- array
- map
- string
- bool
- int系
- float系
- `type Mode string` のような名前付きスカラー型
- `type AddAccounts []AddAccount` のような名前付きslice

対応するタグ:

- `yaml`
- `desc`
- `default`
- `enum`

`yaml` タグがないフィールド、`yaml:"-"` のフィールド、`json` / `xml` タグだけのフィールドはYAML編集用スキーマに含めません。
`yaml` タグに `omitempty` があるフィールドは任意、ないフィールドは必須として扱います。

例:

```go
type Config struct {
	Server Server `yaml:"server" desc:"server settings"`
}

type Server struct {
	Host string `yaml:"host" default:"127.0.0.1"`
	Port int    `yaml:"port,omitempty" default:"8080"`
}
```

## Editing Helpers

`common.dates` では、`dates:` の直後で改行すると `day1.date` と `day1.holiday` を自動入力します。
`dayN.holiday` に `true` または `false` を入力して改行した場合は、次の `dayN+1` ブロックと `schedules` ブロックを補完候補として表示し、確定した場合だけ挿入します。

`common.schedules` では、`schedules:` の直後で改行すると登録済みの `runN` テンプレートを自動入力します。
テンプレートはToolbarの `Schedules` メニューから変更できます。

`scenario.steps` では、`steps:` の直後で改行すると最初のstepリスト要素を自動入力します。
テンプレートは `id`, `name`, `day_ref`, `schedule_ref`, `action.tool` までを含みます。
`day_ref` の値入力時は、編集中の `common.dates` に存在する `day1`, `day2` などを候補として表示し、候補詳細に `date` と `holiday` を表示します。

## Tool Args Completion

`action.tool` は参照用tool schemaから `"<namespace>.<Struct>"` 形式で補完します。
名前空間は `.` の前後で段階的に補完され、値はダブルクォート付きで挿入されます。

`tool` の構造体名を補完すると、同じ `action` 階層の `args` が選択toolに応じて自動入力されます。
既存の `args` がある状態で `tool` を補完し直した場合は、既存 `args` を新しいtool用の内容へ置き換えます。
tool自体がsliceの場合、`args` はYAMLリストとして補完されます。

## Test

```sh
go test -mod=vendor ./...
```

## Build

macOS配布用 `.app` を作成します。

```sh
scripts/build-macos-app.sh
```

作成後は以下で起動します。

```sh
open "dist/YAML Struct Editor.app"
```

Windows配布用exeはWindows環境で作成します。

```powershell
scripts\build-windows.ps1
```

事前生成済みの `frontend/dist` を使い、npmなしでWindows exeを作成する場合:

```powershell
scripts\build-windows-offline.ps1
```

このofflineビルドでは `frontend/dist` と `vendor/` を使うため、Node/npmとGo module downloadは不要です。

## Release

`v*` タグをpushすると、GitHub ActionsでmacOSとWindowsの配布用成果物を作成し、GitHub Releaseへ添付します。

```sh
git tag v0.1.0
git push origin v0.1.0
```

成果物:

- `YAML-Struct-Editor-macOS.zip`
- `YAML-Struct-Editor-Windows.zip`

FeatureブランチでWindows exeだけを確認する場合は、GitHub Actionsの `Windows Artifact` workflowを手動実行します。
