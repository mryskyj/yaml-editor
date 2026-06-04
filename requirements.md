# YAML Struct Editor Requirements

## 目的

Go言語で定義された構造体をもとに、YAMLファイルを補完・検証できるデスクトップGUIアプリを作る。

JSON Schemaは使用しない。

Go structを唯一の正しい定義として扱い、YAML編集時の補完・検証・ドキュメント表示を行う。

---

## 技術方針

- GUI: Wails v3
- Editor: Monaco Editor
- Backend: Go
- YAML解析: gopkg.in/yaml.v3
- Struct解析: reflect
- テスト: Go testing
- ソース管理: Git

---

## 対象OS

### 対応対象

- Windows 11
- macOS (Apple Silicon)
- macOS (Intel)

### 将来対応

- Linux (Ubuntu)

---

## 配布形式

### Windows

- yaml-struct-editor.exe
- 配布用exeはGUIサブシステムでビルドし、起動時にコンソールを表示しない

### macOS

- YAML Struct Editor.app
- 配布用アプリは `.app` バンドルとして作成し、Finder起動時にTerminalを表示しない
- `go build` で作成した生の実行ファイルは配布形式ではなく、Finderから直接起動するとTerminalが表示される

---

## 対象ユーザー

- Goアプリケーション開発者
- YAMLベースの設定ファイルを扱う開発者
- DevOpsエンジニア
- インフラエンジニア

---

## MVP機能

### ファイル操作

- YAMLファイルを開く
- YAMLファイルを保存する
- YAMLファイルを新規作成する
- 最近開いたファイルを保持する

---

### YAML編集

- Monaco Editorで編集する
- 行番号を表示する
- 折りたたみをサポートする
- Undo / Redoをサポートする
- YAMLシンタックスハイライトを表示する

---

### Struct連携

アプリ起動時にGo structを登録する。

対象

- struct
- structネスト
- slice
- array
- map
- string
- bool
- int
- float

取得対象タグ

- yaml
- required
- desc
- default
- enum

---

### 補完

カーソル位置に応じて候補を表示する。

候補表示内容

- YAMLキー名
- Go型
- 説明
- 必須／任意
- デフォルト値
- enum候補

---

### バリデーション

以下を検出する。

#### YAML構文エラー

例

- インデント不正
- コロン不足

#### スキーマエラー

- 未定義キー
- 型不一致
- 必須項目不足
- enum不一致
- ネスト不一致

---

### エラー表示

エラー発生時

- Monaco上に赤線表示
- エラー一覧に表示
- エラークリックで該当位置へジャンプ

---

### スキーマ表示

画面右ペインに現在のスキーマを表示する。

表示内容

- キー名
- 型
- 必須／任意
- 説明
- デフォルト値
- enum

---

## Structタグ仕様

例

Port int yaml:"port" required:"true" default:"8080" desc:"待受ポート"

### yaml

YAMLキー名

### required

必須フラグ

### desc

説明文

### default

デフォルト値

### enum

許可値一覧

例

enum:"dev,stg,prod"

---

## MVP対象外

以下は実装しない。

- JSON Schema生成
- YAML Anchor
- YAML Alias
- Language Server
- VS Code Extension
- Goコードの動的ロード
- プラグイン機構
- 複数スキーマ切り替え

---

## パフォーマンス要件

### 起動時間

3秒以内

### バリデーション

1000行程度のYAMLで1秒以内

### 補完表示

200ms以内

---

## 品質要件

- panicしない
- 不正なYAMLでクラッシュしない
- UTF-8をサポートする
- 日本語をサポートする
- WindowsとmacOSで同等動作する

---

## 初期サンプルスキーマ

Config

- server
- app

server

- host
- port

app

- mode

mode

- dev
- stg
- prod

---

## 成功条件

以下が実現できればMVP完了とする。

1. YAMLファイルを開ける
2. YAMLを編集できる
3. Go structから補完できる
4. 未定義キーを検出できる
5. 型不一致を検出できる
6. 必須項目不足を検出できる
7. エラー位置へジャンプできる
8. Windows/macOSで動作する
