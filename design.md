# YAML Struct Editor Design

## 目的

Go言語で定義されたstructを唯一のスキーマ定義として扱い、YAMLファイルの編集、補完、検証、スキーマ表示を行うデスクトップGUIアプリを実装する。

JSON Schemaは生成・利用しない。

---

## 全体構成

アプリケーションはWails v3を使い、UIをフロントエンド、スキーマ解析やYAML検証などの業務ロジックをGoバックエンドに分離する。

```text
Frontend
  File tabs
  Monaco Editor
  File toolbar
  Error list
  Schema pane
        |
        | Wails binding
        v
Backend
  App service
  File service
  Schema registry
  YAML parser
  Validator
  Completion provider
```

---

## 技術構成

### フロントエンド

- Wails v3 frontend
- Monaco Editor
- YAMLシンタックスハイライト
- Monaco diagnosticsによる赤線表示
- Monaco completion providerによる補完表示

### バックエンド

- Go
- `reflect`によるstruct解析
- `gopkg.in/yaml.v3`によるYAML解析
- Go標準ライブラリによるファイル操作
- Go testingによる単体テスト

### 配布用ビルド

- macOSはGoバイナリを直接配布せず、`.app` バンドルに格納する
- macOSで生のGoバイナリをFinderから直接起動するとTerminalが表示されるため、ユーザー起動対象は `.app` とする
- Windowsは `-ldflags="-H windowsgui"` を指定してGUIアプリとしてビルドする
- 配布用ビルド前にfrontend buildを実行し、埋め込みアセットを最新化する
- frontend buildではMonaco Editorの `min/vs` アセットを `frontend/dist/vs` にコピーし、production UIは `/vs` から読み込む
- 配布用ビルドはWailsのproduction tagを付け、開発用ログ出力を抑制する
- Windows向けには、事前生成済みの `frontend/dist` を使ってnpmなしでexeを作成する `scripts/build-windows-offline.ps1` を用意する
- npmなしのWindowsビルドはUIを再生成しないため、接続端末で作成した `frontend/dist` をそのまま埋め込む
- GitHub Actionsではタグ `v*` のpushを契機にmacOSとWindowsの配布用ビルドを作成する
- featureブランチ検証用に、`workflow_dispatch` でWindows exe artifactだけを作成するGitHub Actionsを用意する
- GitHub ActionsのReleaseビルドは `go test ./...` を実行してから成果物を作成する
- macOS成果物は `YAML Struct Editor.app` を `YAML-Struct-Editor-macOS.zip` としてReleaseへ添付する
- Windows成果物は `yaml-struct-editor.exe` を `YAML-Struct-Editor-Windows.zip` としてReleaseへ添付する
- 署名とnotarizationは現時点では未対応とし、正式配布時の追加項目とする

---

## ディレクトリ構成

```text
.
├── app/
│   ├── app.go
│   └── sampleschema/
│       ├── root.go
│       ├── server.go
│       ├── app.go
│       ├── aws.go
│       ├── cloudformation.go
│       ├── ecs.go
│       ├── ssm.go
│       ├── observability.go
│       ├── deployment.go
│       ├── security.go
│       └── foreign.go
├── internal/
│   ├── schema/
│   │   ├── field.go
│   │   ├── parser.go
│   │   └── registry.go
│   ├── yamlx/
│   │   ├── parse.go
│   │   └── position.go
│   ├── validator/
│   │   ├── diagnostic.go
│   │   └── validator.go
│   ├── completion/
│   │   ├── completion.go
│   │   └── provider.go
│   └── file/
│       ├── recent.go
│       └── service.go
├── frontend/
│   └── src/
│       ├── components/
│       ├── editor/
│       └── app/
├── requirements.md
├── design.md
├── tasks.md
└── test_plan.md
```

責務ごとに分離し、1ファイル500行以内を目安にする。

---

## バックエンド設計

### App service

Wailsから呼び出される公開APIを提供する。

主な責務:

- YAMLファイルの新規作成
- YAMLファイルを開く
- YAMLファイルを保存する
- 最近開いたファイルの取得
- YAML本文の検証
- カーソル位置に応じた補完候補の取得
- 登録済みスキーマ情報の取得

UI向けAPIは表示に必要なデータだけを返し、検証や補完の判断はバックエンド側で行う。

初期サンプルスキーマのGo structは `app/sampleschema` に配置し、root定義と第一階層の構造体ごとにファイルを分離する。
サンプルスキーマはAWS CloudFormation / ECS / SSMなど実在の設定ファイルを参考にした構成を含める。
JSON/XML専用の不要構造体も混在させ、Struct parserがYAMLタグ付きフィールドだけを取り込むことを確認できるようにする。

### Schema registry

アプリ起動時にGo structを登録し、解析済みスキーマを保持する。

外部Goソースフォルダが指定された場合は、指定フォルダ内のGoソースファイルから対象structを静的解析して登録する。

MVPでは複数スキーマ切り替えを行わないため、アプリ内で利用するルートスキーマは1つとする。

主な責務:

- structの登録
- 指定フォルダ内の外部Goソースファイルからstructを登録
- 解析済みスキーマの保持
- ルートスキーマの取得

### Struct parser

`reflect`または`go/parser`でGo structを解析し、YAML編集用の内部スキーマに変換する。

組み込みサンプルスキーマは`reflect`で解析する。

起動時の外部Goソース読み込みでは、Goコードをコンパイル・実行せず、指定フォルダ直下の`.go`ファイル内のstruct定義とタグだけを`go/parser`で静的解析する。

`*_test.go` は読み込み対象外とする。

起動時スキーマ設定:

- `--schema-dir`: 外部Goソースファイルを含むフォルダ
- `--schema-type`: ルートstruct名。省略時は自動検出する

`--schema-dir` 未指定時は組み込みの `app/sampleschema` ソースを読み込む。
組み込みサンプルスキーマは特定のGo型名を直接参照せず、`app/sampleschema` 配下のGoソースを静的解析して登録する。
そのためサンプルスキーマを差し替える場合も、root struct名を `Config` に固定する必要はない。
root schemaはYAML文書全体の検証・補完に必要な内部概念として保持する。

外部Goソースフォルダ解析では、読み込み対象ファイル内のトップレベル `type Xxx struct` をすべて収集してからルートstructを解析する。
`--schema-type` が省略された場合は、YAMLタグ付きフィールドを持ち、他structから参照されていないstructをrootとして自動検出する。
root候補が複数ある場合は曖昧なためエラーとし、`--schema-type` による明示指定を求める。
フィールド型が同一フォルダ内で定義された名前付きstructを参照している場合は、そのstruct定義へ解決して内部スキーマに展開する。

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

この場合、内部スキーマでは `server` の子要素として `host` と `port` を保持する。
Completion providerとSchema paneはこの解決済み内部スキーマを参照するため、ファイルをまたぐstruct依存も入力可能キーとして提示できる。

対応型:

- struct
- ネストしたstruct
- slice
- array
- map
- string
- bool
- int系
- float系

取得するタグ:

- `yaml`
- `required`
- `desc`
- `default`
- `enum`

`yaml:"-"`のフィールドは対象外とする。
`yaml`タグがないフィールドも対象外とし、`json`タグや`xml`タグだけを持つフィールドは解析しない。
YAML対象フィールドに`json`タグなどが併記されていても、内部スキーマ名には`yaml`タグのキー名を使う。

外部Goソース読み込みでは以下を未対応とし、検出時は明示的なエラーとして扱う。

- サブディレクトリ配下のGoファイル
- import先パッケージの型
- `pkg.Type` 形式の外部型参照
- type aliasのみで定義された型
- generic型
- 循環参照するstruct

### Schema model

```go
type Field struct {
    Name        string
    Type        FieldType
    Required    bool
    Description string
    Default     string
    Enum        []string
    Children    []*Field
    Item        *Field
    MapKeyType   FieldType
    MapValue     *Field
}
```

`Children`はstructの子フィールド、`Item`はsliceまたはarrayの要素、`MapValue`はmapの値を表す。

### YAML parser

`gopkg.in/yaml.v3`でYAML文字列を解析し、`yaml.Node`を取得する。

主な責務:

- YAML構文エラーの検出
- `yaml.Node`の行番号・列番号の保持
- ValidatorやCompletion providerが利用しやすい探索関数の提供

YAML AnchorとAliasはMVP対象外のため、検出した場合は未対応のスキーマエラーとして扱う。

### 今後の改善方針

MVP完了後の改善では、既存の責務分離を維持したまま、実利用時の信頼性と操作性を優先する。

- ファイル操作は、最近開いたファイルとタブのファイルパスをバックエンドの実データと同期させる
- Wails API、スキーマ取得、検証、補完の失敗は、空結果として扱わずUIで確認できる状態にする
- 補完の階層推定は、単純なインデント推定だけに依存せず、YAML parserの位置情報を活用する
- Monaco diagnosticsは、診断の開始位置と終了位置に合わせて表示範囲を制御する
- フロントエンドの主要な状態遷移と操作は、自動テストで回帰を検出できるようにする
- 複数スキーマ対応を追加する場合も、検証・補完の判断はバックエンド側に置く
- macOS配布前にlinker警告を確認し、deployment targetとビルド環境の設定を明示する

### Validator

YAML ASTと内部スキーマを比較し、診断情報を返す。

検出対象:

- YAML構文エラー
- 未定義キー
- 型不一致
- 必須項目不足
- enum不一致
- ネスト不一致

診断情報はMonacoで扱いやすい形式に変換できるよう、位置情報とメッセージを含める。

```go
type Diagnostic struct {
    Severity string
    Message  string
    Line     int
    Column   int
    EndLine  int
    EndColumn int
}
```

### Completion provider

YAML本文とカーソル位置から、現在のパスを推定し、該当スキーマの候補を返す。

候補内容:

- YAMLキー名
- Go型
- 説明
- 必須または任意
- デフォルト値
- enum候補

既に同じ階層に存在するキーは、重複候補として出さない。

### File service

OS差異を吸収し、UTF-8テキストとしてYAMLファイルを読み書きする。

主な責務:

- ファイルを開く
- ファイルを保存する
- 新規ファイル状態を作る
- 最近開いたファイルを保存・取得する

最近開いたファイルはユーザー設定領域にJSON形式で保存する。

---

## フロントエンド設計

### 画面構成

```text
+--------------------------------------------------+
| Toolbar                                          |
+-------------------------------+------------------+
| File Tabs                                        |
+-------------------------------+------------------+
| Monaco Editor                 | Schema Pane      |
|                               |                  |
|                               |                  |
+-------------------------------+------------------+
| Error List                                       |
+--------------------------------------------------+
```

### Toolbar

主な操作:

- 新規作成
- 開く
- 保存
- 最近開いたファイル
- Undo
- Redo

保存はWails API経由でApp serviceの `SaveFile` を呼び出す。
保存先パスが未確定の場合は、Wails runtimeの保存ダイアログで保存先パスを取得してから保存する。

### Keyboard Shortcuts

フロントエンドでアプリ全体のキーボードイベントを扱い、ToolbarとFile Tabsの操作へ接続する。

OSごとの主修飾キー:

- macOS: `Meta`
- Windows / Linux: `Ctrl`

対応表:

| Shortcut | Action |
| --- | --- |
| `Cmd/Ctrl + N` | 新規タブを作成する |
| `Cmd/Ctrl + O` | YAMLファイルを開く |
| `Cmd/Ctrl + S` | アクティブなタブを保存する |
| `Cmd/Ctrl + W` | アクティブなタブを閉じる |
| `Cmd/Ctrl + Tab` | 次のタブへ切り替える |
| `Cmd/Ctrl + Shift + Tab` | 前のタブへ切り替える |
| `Esc` | アプリ内確認ダイアログを閉じる |

ショートカット処理の方針:

- アプリ操作に対応するキー入力は `preventDefault` してブラウザ/WebView標準動作を抑制する
- `Cmd/Ctrl + W` はアクティブなタブの閉じる操作として扱い、保存先未確定または未保存変更があればアプリ内確認ダイアログを表示する
- `Cmd/Ctrl + O` はToolbarのOpenと同じファイル選択処理を呼び出す
- `Cmd/Ctrl + S` はToolbarのSaveと同じ保存処理を呼び出す
- `Esc` はアプリ内確認ダイアログが表示されている場合だけキャンセルとして扱う
- Monaco Editor標準の編集ショートカットは上書きしない

### File Tabs

複数のYAMLファイルを同時に開き、タブでアクティブなドキュメントを切り替える。

タブごとに保持する状態:

- タブID
- ファイルパス
- 表示名
- YAML本文
- 保存状態
- カーソル位置
- 検証診断

タブ操作:

- 新規作成時は未保存タブを追加してアクティブにする
- ファイルを開いた場合は新しいタブとして追加する
- 既に同じパスのタブが開いている場合は、そのタブをアクティブにする
- タブを閉じる場合、保存先未確定または未保存変更があればアプリ内ダイアログで確認する
- 最後のタブを閉じた場合は空の未保存タブを作成する
- 保存後はタブのファイルパス、表示名、保存状態、最近開いたファイルを更新する

Monaco Editor、Error List、Schema Paneはアクティブなタブの状態だけを表示する。

### Monaco Editor

主な責務:

- YAML編集
- 行番号表示
- 折りたたみ
- Undo / Redo
- YAMLシンタックスハイライト
- 補完候補表示
- 検証結果の赤線表示

YAML本文が変更されたら、アクティブなタブの本文を更新し、短い遅延を入れてバックエンド検証を呼び出す。
検証結果はアクティブなタブの診断として保持する。

### Error List

アクティブなタブの検証結果を一覧表示する。

表示内容:

- エラーメッセージ
- 行番号
- 列番号

エラーをクリックするとMonaco Editorの該当位置へ移動する。

### Schema Pane

アクティブなタブのカーソル位置に応じたスキーマ情報を優先表示する。

表示内容:

- 現在カーソル位置に対応するキー名
- 型
- 必須または任意
- 説明
- デフォルト値
- enum候補
- 親階層のパンくず
- 同一階層で入力可能なキー一覧
- 入力済みキーの状態

スキーマ全体は常時全量表示せず、検索ビューとツリービューで補助的に参照できるようにする。

---

## データフロー

### 起動時

1. Goバックエンドで対象structを登録する
2. Struct parserが内部スキーマを生成する
3. Schema registryに内部スキーマを保存する
4. フロントエンドが空の未保存タブを作成する
5. フロントエンドがスキーマ情報を取得して右ペインに表示する

### YAML編集時

1. ユーザーがアクティブなタブのMonaco EditorでYAMLを編集する
2. フロントエンドがアクティブなタブのYAML本文を更新する
3. フロントエンドがバックエンドへYAML本文を送信する
4. YAML parserが構文解析する
5. Validatorがスキーマと照合する
6. フロントエンドがアクティブなタブの診断、Monaco diagnostics、エラー一覧を更新する
7. 右ペインがアクティブなタブのカーソル位置に応じたスキーマ情報と同階層候補を更新する

### タブ切り替え時

1. ユーザーがタブを選択する
2. フロントエンドがアクティブなタブIDを更新する
3. Monaco Editorに選択されたタブのYAML本文を表示する
4. Monaco diagnostics、エラー一覧、Schema Paneを選択されたタブの状態で更新する

### ショートカット操作時

1. フロントエンドがOSに応じた主修飾キー付き入力を検出する
2. 対応するToolbarまたはFile Tabsの既存ハンドラを呼び出す
3. アクティブなタブの状態を更新する
4. 必要に応じて保存ダイアログ、ファイル選択、タブ閉じる確認を表示する

### ファイルを開く時

1. ユーザーがファイルを開く
2. フロントエンドまたはバックエンドがYAML本文とファイルパスを取得する
3. 同じファイルパスのタブが既にあれば、そのタブをアクティブにする
4. 未オープンのファイルであれば新しいタブを追加してアクティブにする

### 保存時

1. ユーザーがSaveを実行する
2. アクティブなタブにファイルパスがなければWails runtimeの保存ダイアログで保存先を取得する
3. フロントエンドがApp serviceの `SaveFile` にファイルパスとYAML本文を渡す
4. 保存成功後、アクティブなタブのファイルパス、表示名、保存状態、最近開いたファイルを更新する

### タブを閉じる時

1. ユーザーがタブを閉じる
2. 保存先未確定または未保存変更がある場合はアプリ内ダイアログで確認する
3. 確認後にタブを閉じる
4. 閉じたタブがアクティブだった場合は隣接タブをアクティブにする
5. タブがなくなった場合は空の未保存タブを作成する

### 補完時

1. Monaco Editorが補完要求を発行する
2. フロントエンドがアクティブなタブのYAML本文とカーソル位置をバックエンドへ送信する
3. Completion providerが登録済みの解決済み内部スキーマから現在位置に合う候補を生成する
4. Monaco Editorに候補を表示する

---

## エラーハンドリング

- YAML構文エラーが発生してもアプリを終了しない
- バックエンドAPIはエラーを無視せず、呼び出し元へ返す
- フロントエンドはバックエンドエラーをエラー一覧または通知として表示する
- `panic`を前提にした処理を書かない
- 不明な型や未対応ノードは明示的なエラーとして扱う

---

## パフォーマンス設計

### 起動

起動時に行う処理はstruct登録と軽量な設定読み込みに限定し、3秒以内を目標にする。

### バリデーション

1000行程度のYAMLを1秒以内で検証する。

方針:

- YAML ASTを1回解析してValidatorで再利用する
- スキーマフィールドはキー名で検索できるようmapを併用する
- 編集ごとの検証呼び出しはフロントエンド側で短くデバウンスする

### 補完

補完表示は200ms以内を目標にする。

方針:

- 登録済みスキーマを再解析しない
- カーソル位置周辺から必要なパスだけを推定する
- 候補生成は現在階層のフィールドに限定する

---

## テスト方針

Goバックエンドの業務ロジックを中心に単体テストを作成する。

対象:

- structタグ解析
- 外部Goソースフォルダ解析
- 外部Goソースの複数ファイル間struct参照解決
- ネストしたstruct解析
- slice / array / map解析
- YAML構文エラー検出
- 未定義キー検出
- 型不一致検出
- 必須項目不足検出
- enum不一致検出
- 補完候補生成
- 最近開いたファイルの保存・取得

フロントエンドは主要操作を手動確認または将来のE2Eテスト対象とする。

---

## Windows npmなしビルド

オフライン端末などnpm環境を作りたくないWindows環境では、接続端末で事前生成した `frontend/dist` を利用してGo buildだけを実行する。
`frontend/dist` にはReactアプリ本体に加えてMonaco Editorの `/vs` アセットも含める。
これにより配布用exeはMonacoのloaderやworkerを外部CDNから取得しない。

```powershell
scripts\build-windows-offline.ps1
```

前提:

- `frontend/dist/index.html` が存在する
- `frontend/dist/vs/loader.js` が存在する
- Go buildに必要なGo環境がある
- UIを変更した場合は接続端末で `npm run build` を実行し、生成済み `frontend/dist` を反映してから使う

このスクリプトは `npm run build` を実行しない。

---

## リリース手順

GitHub Releases向けの配布用ビルドはGitHub Actionsで実行する。

リリース作成手順:

```sh
git tag v0.1.0
git push origin v0.1.0
```

タグpush後、`.github/workflows/release.yml` が以下を実行する。

1. macOS runnerで `scripts/build-macos-app.sh` を実行する
2. Windows runnerで `scripts/build-windows.ps1` を実行する
3. 各OSの成果物をzip化する
4. GitHub Releaseを作成し、zipを添付する

---

## FeatureブランチのWindows artifact生成

Releaseを作成せずにfeatureブランチのWindows exeだけを確認したい場合は、GitHub Actionsの `Windows Artifact` workflowを手動実行する。

手順:

1. GitHubのActions画面で `Windows Artifact` を選ぶ
2. `Run workflow` から対象ブランチを選ぶ
3. workflow完了後、Artifactsから `YAML-Struct-Editor-Windows` をダウンロードする

このworkflowはGitHub Releaseを作成しない。

---

## 依存ライブラリ

MVPで利用する外部ライブラリは以下に限定する。

- Wails v3
- Monaco Editor
- `gopkg.in/yaml.v3`

追加ライブラリが必要になった場合は、必要性、標準ライブラリで代替できない理由、既存ライブラリとの重複有無を確認してから追加する。

---

## MVP対象外

requirements.mdに従い、以下は設計対象外とする。

- JSON Schema生成
- YAML Anchor
- YAML Alias
- Language Server
- VS Code Extension
- Goコードの動的ロード
- 外部Goソースの実行またはコンパイル
- 外部Goソースの動的更新監視
- プラグイン機構
- 複数スキーマ切り替え

---

## 未決定事項

### 登録するGo struct

MVPではアプリ起動時に単一のroot schemaを登録する。
組み込みサンプルは `app/sampleschema` 配下のGoソースからrootを自動検出して登録する。
外部スキーマを指定した場合も、root候補が1つであれば `--schema-type` を省略できる。

### 最近開いたファイルの保存場所

OSごとのユーザー設定ディレクトリ配下に保存する。

具体的なパス取得方法はWails v3またはGo標準ライブラリの利用可能APIを確認して実装時に決定する。
