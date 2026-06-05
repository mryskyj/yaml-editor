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
- [ ] 外部Goソーススキーマ読み込みを実装する
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
- [x] 複数ファイルタブの状態管理を実装する
- [x] ファイルタブUIを実装する
- [ ] タブ単位の保存・検証・補完連携を実装する
- [x] ショートカットキー要件を整理する
- [x] ショートカットキーを実装する
- [x] GitHub Actionsリリースビルドを実装する
- [x] Windows npmなしビルドを実装する
- [x] FeatureブランチWindows artifactビルドを実装する
- [x] Go toolchainを1.26.2へ更新する

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
- `yaml` タグがないフィールドを除外し、JSON/XML専用フィールドを解析対象外にした
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
- 初期サンプルスキーマのstruct定義をApp service本体から分離した
- 初期サンプルスキーマを `app/sampleschema` に移し、第一階層の構造体ごとにファイル分割した
- 初期サンプルスキーマをAWS CloudFormation / Auto Scaling / ECS / SSM風の複雑な構成へ拡張した
- 初期サンプルスキーマにJSON/XML専用構造体を混在させ、YAMLタグ付きフィールドだけが解析される単体テストを追加した
- 起動時にサンプルスキーマを `schema.Registry` へ登録した
- ファイル新規作成・開く・保存・最近開いたファイル取得APIを追加した
- YAML検証APIを追加した
- YAML補完APIを追加した
- 登録済みスキーマ取得APIを追加した
- App serviceの単体テストを追加した
- サンプルスキーマの第一階層が解析できる単体テストを追加した

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
- ウィンドウサイズが変わってもエラーメッセージが見切れないように折り返しとスクロールを調整した

### スキーマペイン表示を実装する

- スキーマ表示用の `SchemaPane` コンポーネントを追加した
- キー名、型、必須/任意、説明、デフォルト値、enum表示を追加した
- ネストしたスキーマを再帰表示できるようにした
- Monaco画面の右ペインを `SchemaPane` に置き換えた
- 右ペインをカーソル位置のコンテキスト表示、スキーマ検索、全体ツリーのタブ構成にした
- カーソル位置のパンくずと同階層で入力可能なキー一覧を表示するようにした
- `server:` のようなコンテナキーにカーソルがある場合は、その配下の入力可能キーを表示するようにした
- カーソルが第一階層に戻った場合は、第一階層で入力可能なキーを表示するようにした

### ファイル操作UIを実装する

- `FileToolbar` コンポーネントを追加した
- 新規作成、開く、保存、最近開いたファイル、Undo、Redoの操作UIを追加した
- 保存ボタンからWails API経由で `SaveFile` を呼び出すようにした
- 保存先未確定の新規ファイルではOSの保存ダイアログから保存先を選んで保存できるようにした
- 現在のファイル名表示を追加した
- Monaco画面のツールバーを `FileToolbar` に置き換えた

### 統合動作を確認する

- Wailsアプリにfrontend build出力のアセット配信設定を追加した
- Wailsアプリの起動URLをfrontend build出力のルートに合わせた
- frontend build出力をバイナリへ埋め込み、実行時の作業ディレクトリに依存しないようにした
- macOS配布用にTerminalを表示せず起動できる `.app` バンドル作成スクリプトを追加した
- macOSでは生の実行ファイルではなく `.app` を起動対象にすることを配布手順へ明記した
- Windows配布用にコンソールを表示しないGUIサブシステム指定のビルドスクリプトを追加した
- Monaco EditorからWails API経由で補完候補と検証診断を取得するようにした
- YAMLキー名やenum値の通常文字入力時にもMonaco補完候補を表示するようにした
- 入力変更時にMonaco diagnosticsとエラー一覧を更新するようにした
- frontend buildを確認した
- Go全体テストを確認した
- Wailsエントリポイントのbuildを確認した

### 複数ファイルタブの状態管理を実装する

- 複数のYAMLドキュメントをタブID単位で保持する
- タブごとにファイルパス、表示名、YAML本文、保存状態、カーソル位置、検証診断を保持する
- 新規作成時は未保存タブを追加する
- 同じファイルパスを開いた場合は既存タブをアクティブにする
- 最後のタブを閉じた場合は空の未保存タブを作成する
- タブ状態管理を `frontend/src/editor/tabs.ts` に分離した
- EditorShellの本文、保存、カーソル、診断状態をアクティブタブ経由で保持するようにした

### ファイルタブUIを実装する

- Monaco Editor上部に `FileTabs` タブバーを追加した
- タブクリックでアクティブなドキュメントを切り替えるようにした
- 保存先未確定または未保存変更があるタブは太字とドットで視覚的に区別するようにした
- タブの閉じる操作を追加した
- 保存先未確定または未保存変更があるタブを閉じる場合はアプリ内ダイアログで確認するようにした
- タブ切り替え時に保存済みカーソル位置へ戻すようにした

### タブ単位の保存・検証・補完連携を実装する

- Saveはアクティブなタブだけを保存する
- 保存後はアクティブなタブのファイルパス、表示名、保存状態、最近開いたファイルを更新する
- 入力変更時の検証診断をアクティブなタブへ保存する
- タブ切り替え時にMonaco diagnostics、エラー一覧、Schema Paneを切り替える
- 補完はアクティブなタブのYAML本文とカーソル位置を使う

### ショートカットキー要件を整理する

- macOSは `Cmd`、Windows / Linuxは `Ctrl` を主修飾キーとして扱うことを整理した
- 新規、開く、保存、閉じる、次タブ、前タブ、確認ダイアログキャンセルを対象にした
- 未保存タブをショートカットで閉じる場合もアプリ内確認ダイアログを表示することを整理した
- Monaco Editor標準の編集ショートカットは妨げない方針にした

### ショートカットキーを実装する

- `Cmd/Ctrl + N` で新規タブを作成するようにした
- `Cmd/Ctrl + O` でYAMLファイルを開くようにした
- `Cmd/Ctrl + S` でアクティブなタブを保存するようにした
- `Cmd/Ctrl + W` でアクティブなタブを閉じるようにした
- `Cmd/Ctrl + W` で未保存タブを閉じる場合はアプリ内確認ダイアログを表示するようにした
- `Cmd/Ctrl + Tab` で次のタブへ切り替えるようにした
- `Cmd/Ctrl + Shift + Tab` で前のタブへ切り替えるようにした
- `Esc` でアプリ内確認ダイアログを閉じるようにした
- ショートカット処理をToolbarとFile Tabsの既存操作へ接続した

### GitHub Actionsリリースビルドを実装する

- タグ `v*` のpushで起動する `.github/workflows/release.yml` を追加した
- macOS runnerで `scripts/build-macos-app.sh` を実行するようにした
- Windows runnerで `scripts/build-windows.ps1` を実行するようにした
- Releaseビルド前に `go test ./...` を実行するようにした
- macOS `.app` とWindows `.exe` をzip化してGitHub Releaseへ添付するようにした
- GitHub Release作成手順を `design.md` に追記した

### Windows npmなしビルドを実装する

- `scripts/build-windows-offline.ps1` を追加した
- `frontend/dist/index.html` が存在することをビルド前に確認するようにした
- npmを実行せず、事前生成済み `frontend/dist` をGoのembed対象として使うようにした
- GUIサブシステム指定 `-H windowsgui` は通常のWindowsビルドと同じにした
- npmなしビルド手順を `design.md` と `test_plan.md` に追記した

### FeatureブランチWindows artifactビルドを実装する

- 手動実行用の `.github/workflows/windows-artifact.yml` を追加した
- 任意ブランチを選んでWindows exeをartifact生成できるようにした
- Releaseは作成せず、GitHub Actions artifactだけをアップロードするようにした
- workflow内で `go test ./...` を実行してからWindows exeを作成するようにした
- 手順を `design.md` と `test_plan.md` に追記した

### Go toolchainを1.26.2へ更新する

- `go.mod` のGo directiveを `1.26.2` に更新した
- GitHub Actionsの `actions/setup-go` は `go-version-file: go.mod` を参照するため、CIのGo toolchainも `1.26.2` へ更新される
- `go test ./...` でGoテストを確認した
