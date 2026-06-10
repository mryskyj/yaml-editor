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
- Go依存ライブラリは `vendor/` に同梱し、Go module downloadなしでビルドできる

### macOS

- YAML Struct Editor.app
- 配布用アプリは `.app` バンドルとして作成し、Finder起動時にTerminalを表示しない
- `go build` で作成した生の実行ファイルは配布形式ではなく、Finderから直接起動するとTerminalが表示される

### GitHub Releases

- タグ `v*` をpushした場合にGitHub Actionsで配布用ビルドを作成する
- macOS向け成果物は `.app` バンドルをzip化してReleaseへ添付する
- Windows向け成果物はGUIサブシステムのexeをzip化してReleaseへ添付する
- ReleaseビルドではGoテストを実行し、テスト失敗時は成果物を公開しない
- Releaseビルドでは `vendor/` のGo依存を使ってテストと配布用ビルドを実行する
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
- 配布用buildではMonaco Editorの実行時アセットを `frontend/dist` に同梱し、CDNや外部ネットワークに依存しない
- アクティブなタブのYAML本文をMonaco Editorに表示する
- 行番号を表示する
- 折りたたみをサポートする
- Undo / Redoをサポートする
- `Shift + 矢印キー` で文字単位・行単位の範囲選択ができる
- 範囲選択したテキストをコピー、切り取り、貼り付けできる
- 範囲未選択でコピーまたは切り取りを実行した場合は現在行を対象にする
- 貼り付けは現在の選択範囲を置き換え、範囲未選択の場合はカーソル位置へ挿入する
- 右クリックメニューからコピー、切り取り、貼り付けを実行できる
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
- `Cmd/Ctrl + C`: 選択範囲または現在行をコピーする
- `Cmd/Ctrl + X`: 選択範囲または現在行を切り取る
- `Cmd/Ctrl + V`: クリップボードのテキストを貼り付ける
- `Shift + ←/→/↑/↓`: カーソルを移動しながら範囲選択する
- `Cmd/Ctrl + Tab`: 次のタブへ切り替える
- `Cmd/Ctrl + Shift + Tab`: 前のタブへ切り替える
- `Esc`: アプリ内確認ダイアログを閉じる

未保存タブをショートカットで閉じる場合も、タブの閉じるボタンと同じ確認を表示する。

Monaco Editor標準の編集ショートカットは妨げない。

---

### Struct連携

アプリ起動時にGo structを登録する。

組み込みサンプルスキーマに加えて、起動時に指定フォルダ内の外部Goソースファイルから対象structを読み込み、登録できるようにする。

外部Goソースファイルは起動時に1回だけ静的解析する。
Goコードのコンパイル、実行、動的更新は行わない。

起動時の指定:

- `--schema-dir`: Goソースファイルを含むフォルダ
- `--schema-type`: ルートstruct名。省略時は自動検出する

`--schema-dir` を指定した場合、指定フォルダ直下の複数 `.go` ファイルを読み込む。
フォルダ内のファイルをまたいでstructや型を参照している場合も、同一フォルダ内で定義された名前付きstructであれば依存関係を解決し、YAML補完・検証・スキーマ表示に反映する。
YAML文書全体を表すroot schemaは内部的に必要だが、struct名を `Config` などの固定名にする必要はない。
`--schema-type` を省略した場合は、YAMLタグ付きフィールドを持ち、他structから参照されていないstructをrootとして自動検出する。
root候補が複数ある場合のみ、`--schema-type` で明示する。

例:

```go
// config.go
type Config struct {
    Server Server `yaml:"server"`
}

// server.go
type Server struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
```

この場合、`server:` 配下では `host` と `port` を入力可能なキーとして提示する。

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

#### tool / args 連動スキーマ

YAML文書全体の基本構造を表すroot structはあらかじめ用意する。
通常はroot structに沿って入力可能なキーと値を補完・検証する。

root struct配下に `tool` と `args` の組み合わせがある場合、`tool` の値は参照用サンプルスキーマに含まれるGo structだけを選択肢とする。
`tool` に入力できる値は `"<パッケージ名>.<構造体名>"` 形式とし、ダブルクォートで囲む。
`tool` の補完は `.` の前後で分割し、パッケージ名候補を確定した後に同パッケージ内の構造体名候補を提示する。
引数指定なしの場合は組み込みの `app/sampleschema`、`--schema-dir` 指定時は指定フォルダ内のGo structを参照用スキーマとして扱う。

`args` の値は、同じ階層の `tool` で選択された構造体に応じて動的に解釈する。
`args` 配下では、選択された構造体のYAML対象フィールドをキーとして補完・検証する。
キー名は既存のstruct解析方針に合わせて `yaml` タグ名を使い、`yaml` タグがないフィールドは対象外とする。

例:

```yaml
steps:
  - tool: "sampletools.CopyFile"
    args:
      source: ./input.yaml
      destination: ./output.yaml
```

この場合、`args` 配下では `sampletools.CopyFile` 構造体のYAML対象フィールドだけを入力可能なキーとして提示する。

外部Goソース読み込みの対象外:

- サブディレクトリ配下のGoファイル
- import先パッケージの型
- `pkg.Type` 形式の外部型参照
- type aliasのみで定義された型
- generic型
- 循環参照するstruct

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

slice / array のリスト要素配下では、要素structのキーを候補として表示する。
例として `scenario.steps` の `- id:` 配下では、同じstep要素の `name`, `day_ref`, `schedule_ref`, `action` などを候補に出す。

組み込みRootスキーマの `common.dates` では、日付エントリを `day1`, `day2` のように連番キーで表す。
`dates:` で改行した場合は、`day1` とその配下の `date`, `holiday` を自動入力する。
`dayN.holiday` に `true` または `false` を入力して改行した場合は、次の `dayN+1` とその配下の `date`, `holiday` を補完候補として表示する。
ユーザーが候補を確定した場合のみ `dayN+1` ブロックを挿入する。
直前の `dayN.date` が `YYYY-MM-DD` 形式の場合、補完候補の `date` には翌日を設定する。
自動入力された `day1` ブロックは1回のUndoで取り消せる。

組み込みRootスキーマの `common.schedules` では、実行情報を `run1`, `run2` のように連番キーで表す。
`schedules:` で改行した場合は、登録済みのschedule情報をすべて自動入力する。
初期値は以下とする。

```yaml
schedules:
    run1: &run1 1 #BOD
    run2: &run2 2 #あいうえお
    run3: &run3 3 #かきくけこ
```

schedule情報は通常変更しないが、必要に応じてアプリ内メニューから登録内容を変更できる。
変更後の登録内容は、次回以降の `schedules:` 改行時の自動入力に使う。

組み込みRootスキーマの `scenario.steps` では、step情報をリストで表す。
`steps:` で改行した場合は、最初のstep要素として `id`, `name`, `day_ref`, `schedule_ref`, `action.tool`, `action.args` を自動入力する。

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
- YAML Aliasによる参照
- Language Server
- VS Code Extension
- Goコードの動的ロード
- 外部Goソースの実行またはコンパイル
- 外部Goソースの動的更新監視
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

## 組み込みRootスキーマ

組み込みRootスキーマは `app/rootschema` にGo structとして定義する。
引数指定なしで起動した場合は `app/rootschema/scenario.go` の `File` をYAML文書全体のroot schemaとして登録する。

Fileの第一階層:

- schema_version
- common
- scenario

`common` の `dates` は、以下のように `dayN` 配下へ日付と祝日フラグを持つ。

```yaml
dates:
    day1:
        date: "2026-03-01"
        holiday: false
    day2:
        date: "2026-03-02"
        holiday: false
```

`common` の `schedules` は、以下のように `runN` 配下へAnchor付きの実行番号を持つ。

```yaml
schedules:
    run1: &run1 1 #BOD
    run2: &run2 2 #あいうえお
    run3: &run3 3 #かきくけこ
```

`scenario` の例:

- id
- name
- description
- docs
- steps

`steps` の例:

- id
- name
- day_ref
- schedule_ref
- action
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
