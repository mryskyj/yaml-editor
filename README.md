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
- スキーマペインでのカーソル位置コンテキスト表示、検索、ツリー表示
- 組み込みサンプルスキーマ
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

外部スキーマの対象は、指定ディレクトリ直下の `.go` ファイルです。`*_test.go` とサブディレクトリは読み込み対象外です。

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

対応するタグ:

- `yaml`
- `required`
- `desc`
- `default`
- `enum`

例:

```go
type Config struct {
	Server Server `yaml:"server" required:"true" desc:"server settings"`
}

type Server struct {
	Host string `yaml:"host" required:"true" default:"127.0.0.1"`
	Port int    `yaml:"port" required:"true" default:"8080"`
}
```

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
