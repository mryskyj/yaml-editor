# YAML Struct Editor Design

## 目的

Go言語で定義されたstructを唯一のスキーマ定義として扱い、YAMLファイルの編集、補完、検証、スキーマ表示を行うデスクトップGUIアプリを実装する。

JSON Schemaは生成・利用しない。

---

## 全体構成

アプリケーションはWails v3を使い、UIをフロントエンド、スキーマ解析やYAML検証などの業務ロジックをGoバックエンドに分離する。

```text
Frontend
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
- 配布用ビルドはWailsのproduction tagを付け、開発用ログ出力を抑制する

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

MVPでは複数スキーマ切り替えを行わないため、アプリ内で利用するルートスキーマは1つとする。

主な責務:

- structの登録
- 解析済みスキーマの保持
- ルートスキーマの取得

### Struct parser

`reflect`でGo structを解析し、YAML編集用の内部スキーマに変換する。

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
保存先パスが未確定の場合は、保存時にユーザーへ保存先パスの入力を求める。

### Monaco Editor

主な責務:

- YAML編集
- 行番号表示
- 折りたたみ
- Undo / Redo
- YAMLシンタックスハイライト
- 補完候補表示
- 検証結果の赤線表示

YAML本文が変更されたら、短い遅延を入れてバックエンド検証を呼び出す。

### Error List

検証結果を一覧表示する。

表示内容:

- エラーメッセージ
- 行番号
- 列番号

エラーをクリックするとMonaco Editorの該当位置へ移動する。

### Schema Pane

カーソル位置に応じたスキーマ情報を優先表示する。

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
4. フロントエンドがスキーマ情報を取得して右ペインに表示する

### YAML編集時

1. ユーザーがMonaco EditorでYAMLを編集する
2. フロントエンドがバックエンドへYAML本文を送信する
3. YAML parserが構文解析する
4. Validatorがスキーマと照合する
5. フロントエンドがMonaco diagnosticsとエラー一覧を更新する
6. 右ペインがカーソル位置のスキーマ情報と同階層候補を更新する

### 補完時

1. Monaco Editorが補完要求を発行する
2. フロントエンドがYAML本文とカーソル位置をバックエンドへ送信する
3. Completion providerが現在位置に合う候補を生成する
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
- プラグイン機構
- 複数スキーマ切り替え

---

## 未決定事項

### 登録するGo struct

MVPではアプリ起動時に単一のGo structを登録する設計とし、現在は `app/sampleschema.Config` を登録する。

### 最近開いたファイルの保存場所

OSごとのユーザー設定ディレクトリ配下に保存する。

具体的なパス取得方法はWails v3またはGo標準ライブラリの利用可能APIを確認して実装時に決定する。
