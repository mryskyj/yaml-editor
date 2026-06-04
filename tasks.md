# Tasks

## 実施ルール

- 一度に実施するタスクは1つのみ
- 完了したタスクだけ `[x]` に更新する
- テスト成功とドキュメント更新を確認してから完了扱いにする

---

## タスク一覧

- [x] 開発計画とテスト計画を整備する
- [x] Goプロジェクトの最小構成を作成する
- [x] Schema modelを実装する
- [x] Struct parserを実装する
- [x] Schema registryを実装する
- [x] YAML parserを実装する
- [x] Validatorの基本診断を実装する
- [x] Validatorの必須項目とenum診断を実装する
- [x] Completion providerを実装する
- [x] File serviceを実装する
- [x] App serviceを実装する
- [x] Wails v3アプリの最小構成を作成する
- [x] Monaco Editor画面を実装する
- [x] エラー一覧表示を実装する
- [x] スキーマペイン表示を実装する
- [x] ファイル操作UIを実装する
- [x] 統合動作を確認する

---

## 現在の完了内容

### 開発計画とテスト計画を整備する

- `tasks.md` を作成した
- `test_plan.md` を作成した
- requirements.md と design.md に沿って実装順序を整理した

### Goプロジェクトの最小構成を作成する

- `go.mod` を作成した
- `app` パッケージを作成した
- `internal` 配下に設計どおりのパッケージを作成した
- `app.New` の単体テストを追加した

### Schema modelを実装する

- `schema.FieldType` を定義した
- `schema.Field` を定義した
- 子フィールド検索用の `FindChild` を追加した
- スカラー型判定用の `FieldType.IsScalar` を追加した
- Schema modelの単体テストを追加した

### Struct parserを実装する

- `reflect` でGo structを `schema.Field` に変換する `schema.Parse` を追加した
- `yaml`, `required`, `desc`, `default`, `enum` タグ解析を追加した
- struct / slice / array / map / string / bool / int系 / float系に対応した
- `yaml:"-"` と未exported fieldを除外した
- 未対応型を明示的なエラーとして返すようにした
- Struct parserの単体テストを追加した

### Schema registryを実装する

- root schemaを保持する `schema.Registry` を追加した
- Go structを登録して内部スキーマへ変換する `Register` を追加した
- 登録済みroot schemaを取得する `Root` を追加した
- 未登録状態とnil receiverを明示エラーにした
- Schema registryの単体テストを追加した

### YAML parserを実装する

- YAML ASTと行番号・列番号を取得するため `gopkg.in/yaml.v3` を追加した
- YAML文字列を `yaml.Node` に解析する `yamlx.Parse` を追加した
- YAML構文エラーを行番号・列番号付き診断へ変換した
- `yaml.Node` の位置情報を取得する `NodePosition` を追加した
- YAML Anchor / AliasをMVP未対応診断として検出した
- YAML parserの単体テストを追加した

### Validatorの基本診断を実装する

- Validator用の `Diagnostic` と `Severity` を追加した
- YAML parser診断をValidator診断へ変換した
- 未定義キー診断を追加した
- 型不一致診断を追加した
- ネスト不一致診断を追加した
- struct / slice / array / map の子要素検証を追加した
- Validator基本診断の単体テストを追加した

### Validatorの必須項目とenum診断を実装する

- 同一階層のキー存在確認を追加した
- 必須項目不足診断を追加した
- enum不一致診断を追加した
- 必須項目不足とenum不一致の単体テストを追加した

### Completion providerを実装する

- 補完候補 `completion.Candidate` を追加した
- カーソル行のインデントから現在階層を推定する `completion.Provide` を追加した
- YAMLキー名、Go型、説明、必須、デフォルト値、enumを候補に含めた
- enum値入力位置では許可値を補完候補として返すようにした
- 同一階層に存在するキーを補完候補から除外した
- Completion providerの単体テストを追加した

### File serviceを実装する

- UTF-8 YAMLファイルを開く `file.Service.Open` を追加した
- UTF-8 YAMLファイルを保存する `file.Service.Save` を追加した
- 新規ファイル状態を作る `NewDocument` を追加した
- 最近開いたファイルをJSONで保存・取得する `RecentStore` を追加した
- File serviceとRecentStoreの単体テストを追加した

### App serviceを実装する

- 初期サンプルスキーマをGo structとして定義した
- 起動時にサンプルスキーマを `schema.Registry` へ登録した
- ファイル新規作成・開く・保存・最近開いたファイル取得APIを追加した
- YAML検証APIを追加した
- YAML補完APIを追加した
- 登録済みスキーマ取得APIを追加した
- App serviceの単体テストを追加した

### Wails v3アプリの最小構成を作成する

- Wails v3 Application API依存を追加した
- `cmd/yaml-struct-editor/main.go` を追加した
- App serviceをWails serviceとして登録した
- メインウィンドウの最小設定を追加した

### Monaco Editor画面を実装する

- Vite + React + TypeScriptのfrontend構成を追加した
- Monaco Editorを使ったYAML編集画面を追加した
- 行番号、折りたたみ、Undo / Redo、YAMLシンタックスハイライトを有効にした
- ツールバー、スキーマペイン、エラー一覧の初期レイアウトを追加した
- frontend buildを確認した

### エラー一覧表示を実装する

- 診断一覧用の `ErrorList` コンポーネントを追加した
- 行番号・列番号・メッセージ表示を追加した
- エラークリック時にMonaco Editorの該当位置へ移動する処理を追加した
- 空の診断状態でもレイアウトが崩れないようにした

### スキーマペイン表示を実装する

- スキーマ表示用の `SchemaPane` コンポーネントを追加した
- キー名、型、必須/任意、説明、デフォルト値、enum表示を追加した
- ネストしたスキーマを再帰表示できるようにした
- Monaco画面の右ペインを `SchemaPane` に置き換えた

### ファイル操作UIを実装する

- `FileToolbar` コンポーネントを追加した
- 新規作成、開く、保存、最近開いたファイル、Undo、Redoの操作UIを追加した
- 現在のファイル名表示を追加した
- Monaco画面のツールバーを `FileToolbar` に置き換えた

### 統合動作を確認する

- Wailsアプリにfrontend build出力のアセット配信設定を追加した
- Wailsアプリの起動URLをfrontend build出力のルートに合わせた
- Monaco EditorからWails API経由で補完候補と検証診断を取得するようにした
- 入力変更時にMonaco diagnosticsとエラー一覧を更新するようにした
- frontend buildを確認した
- Go全体テストを確認した
- Wailsエントリポイントのbuildを確認した
