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
- `frontend/dist` が事前生成済みの場合は、npmを使わずGoだけでWindows exeをビルドできる

### macOS

- YAML Struct Editor.app
- 配布用アプリは `.app` バンドルとして作成し、Finder起動時にTerminalを表示しない
- `go build` で作成した生の実行ファイルは配布形式ではなく、Finderから直接起動するとTerminalが表示される

### GitHub Releases

- タグ `v*` をpushした場合にGitHub Actionsで配布用ビルドを作成する
- macOS向け成果物は `.app` バンドルをzip化してReleaseへ添付する
- Windows向け成果物はGUIサブシステムのexeをzip化してReleaseへ添付する
- ReleaseビルドではGoテストを実行し、テスト失敗時は成果物を公開しない
- featureブランチ検証用に、手動実行でWindows exeのartifactだけを生成できる

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
- 複数のYAMLファイルを同時に開き、タブで切り替えて表示する
- YAMLファイルを保存する
- 保存先が未確定の新規ファイルは、保存時にOSの保存ダイアログで保存先を指定する
- YAMLファイルを新規作成する
- 新規作成した未保存ファイルもタブとして扱う
- タブごとに編集中の内容、ファイルパス、保存状態、検証結果を保持する
- 保存先が未確定、または未保存変更があるタブを閉じる場合は確認する
- 最近開いたファイルを保持する

---

### YAML編集

- Monaco Editorで編集する
- アクティブなタブのYAML本文をMonaco Editorに表示する
- 行番号を表示する
- 折りたたみをサポートする
- Undo / Redoをサポートする
- YAMLシンタックスハイライトを表示する

---

### ショートカットキー

主要なファイル操作とタブ操作をキーボードから実行できる。

修飾キーはOSごとに以下を使う。

- macOS: `Cmd`
- Windows / Linux: `Ctrl`

対応する操作:

- `Cmd/Ctrl + N`: 新規タブを作成する
- `Cmd/Ctrl + O`: YAMLファイルを開く
- `Cmd/Ctrl + S`: アクティブなタブを保存する
- `Cmd/Ctrl + W`: アクティブなタブを閉じる
- `Cmd/Ctrl + Tab`: 次のタブへ切り替える
- `Cmd/Ctrl + Shift + Tab`: 前のタブへ切り替える
- `Esc`: アプリ内確認ダイアログを閉じる

未保存タブをショートカットで閉じる場合も、タブの閉じるボタンと同じ確認を表示する。

Monaco Editor標準の編集ショートカットは妨げない。

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

`yaml` タグがないフィールドはYAML編集用スキーマの対象外とする。
`json` や `xml` などYAML以外のタグが付いたフィールドが混在していても、`yaml` タグがあるフィールドだけを解析対象にする。

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
- アクティブなタブのエラー一覧に表示
- エラークリックで該当位置へジャンプ

---

### スキーマ表示

画面右ペインには、アクティブなタブのカーソル位置に応じたコンテキスト情報を優先表示する。
スキーマ全体の一覧は常時全量表示しない。

表示内容

- 現在カーソル位置に対応するキー名
- 型
- 必須／任意
- 説明
- デフォルト値
- enum
- 親階層のパンくず
- 同一階層で入力可能なキー一覧
- スキーマ全体検索
- 補助的に参照できる全体ツリー

---

## Structタグ仕様

例

Port int yaml:"port" required:"true" default:"8080" desc:"待受ポート"

### yaml

YAMLキー名

YAML編集用スキーマへの取り込み対象を示すタグ。
`yaml:"-"` または `yaml` タグなしのフィールドは取り込まない。

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

初期サンプルスキーマは `app/sampleschema` にGo structとして定義する。
root定義と第一階層の構造体ごとにファイルを分離し、AWSなど実在の設定ファイルを参考にした複雑な構成を扱えることを確認する。

Configの第一階層:

- server
- app
- aws
- cloudformation
- ecs
- ssm
- observability
- deployment
- security

`server` の例:

- host
- port

`app` の例:

- mode

`mode` の許可値:

- dev
- stg
- prod

サンプルスキーマにはJSON/XML専用の構造体も混在させる。
ただしYAML編集用スキーマには、`yaml` タグがあるフィールドだけを取り込む。

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
