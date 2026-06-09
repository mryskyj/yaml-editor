# Test Plan

## 目的

Go structを唯一のスキーマ定義として扱うYAML編集アプリについて、バックエンドの業務ロジックを中心にテストする。

---

## テスト方針

- Goバックエンドは `go test -mod=vendor ./...` で単体テストを実行する
- 機能追加時は対象パッケージに単体テストを追加する
- バグ修正時は回帰テストを追加する
- 不正な入力でpanicしないことを確認する
- フロントエンドの主要操作はMVPでは手動確認を基本とし、将来E2Eテストを追加する

---

## 対象範囲

### Schema model

- FieldがYAMLキー名、型、必須、説明、デフォルト値、enum、子要素を保持できる
- direct childをキー名で検索できる
- nil receiverでもpanicしない
- scalar typeを判定できる

### Schema parser

- yamlタグからキー名を取得できる
- `yaml:"-"` のフィールドを除外できる
- `yaml` タグがないフィールドを除外できる
- `json` / `xml` 専用フィールドが混在しても解析対象にしない
- YAML対象フィールドに `json` タグが併記されていても `yaml` タグのキー名で解析できる
- `required` タグを取得できる
- `desc` タグを取得できる
- `default` タグを取得できる
- `enum` タグをカンマ区切りで取得できる
- ネストしたstructを解析できる
- slice / array / mapを解析できる
- string / bool / int系 / float系を解析できる
- 外部Goソースフォルダから対象structを解析できる
- ルートstruct名を省略した場合に、単一のroot候補を自動検出できる
- root候補が複数ある場合は曖昧なスキーマとして明示エラーにできる
- root型名が `Config` ではない別サンプルスキーマを読み込み、組み込みサンプルとは異なる補完候補を返せる
- フォルダ内の複数Goソースファイルに分かれたstructを解析できる
- ファイルをまたぐ名前付きstruct参照を解決し、子フィールドを内部スキーマに展開できる
- `*_test.go` を外部Goソース読み込み対象から除外できる
- 存在しない外部Goソースフォルダや対象structを明示エラーにできる
- import先パッケージの型、`pkg.Type` 形式、type alias、generic型、循環参照を未対応として明示エラーにできる
- `app/sampleschema` に分割したサンプルスキーマをGo型名固定なしで解析できる
- `app/sampleschema` にJSON/XML専用構造体が混在してもYAMLタグ付きフィールドだけを解析できる
- `app/rootschema` に分割した組み込みRootスキーマから `File` をrootとして解析できる
- 引数指定なしのApp serviceが `File` rootの `schema_version`, `common`, `scenario` を登録できる
- `common.dates` が `dayN` の値として `date` とbool型の `holiday` を持つ構造として解析できる
- `common.schedules` が `runN` の値としてint型を持つmap構造として解析できる

### Schema registry

- root schemaを登録できる
- 外部Goソースフォルダからroot schemaを登録できる
- 登録済みroot schemaを取得できる
- 未登録状態を明示的に扱える

### YAML parser

- 正常なYAMLを `yaml.Node` に解析できる
- 構文エラーを診断情報へ変換できる
- 行番号と列番号を保持できる
- Anchor定義は許可し、Alias参照は未対応として検出できる

### Validator

- 未定義キーを検出できる
- 型不一致を検出できる
- 必須項目不足を検出できる
- enum不一致を検出できる
- ネスト不一致を検出できる
- `tool` の値が参照用サンプルスキーマに存在しない場合に検出できる
- `args` 配下の未定義キー、型不一致、必須項目不足、enum不一致を、選択されたtool structに基づいて検出できる
- `tool` が未指定または未解決の場合に `args` 配下でpanicせず明示的に扱える
- 不正なYAMLでもpanicしない
- `common.schedules` のAnchor付きrun値を正しいint値として検証できる

### Completion provider

- カーソル位置に応じて候補を返せる
- YAMLキー名を候補に含める
- Go型、説明、必須/任意、デフォルト値、enumを候補に含める
- 同一階層に存在するキーを重複候補として返さない
- 外部Goソースの複数ファイル間struct参照から展開された子フィールドを候補に含める
- `tool` の値位置では参照用サンプルスキーマの `<パッケージ名>.<構造体名>` を候補に含める
- `args` 配下では同一階層の `tool` 値に対応するstructのYAML対象フィールドをキー候補に含める

### File service

- UTF-8のYAMLファイルを開ける
- UTF-8のYAMLファイルを保存できる
- 新規ファイル状態を作成できる
- 最近開いたファイルを保存できる
- 最近開いたファイルを取得できる

### Release build

- タグ `v*` のpushでGitHub ActionsのRelease workflowが起動する
- macOS runnerで `.app` バンドルを作成できる
- Windows runnerでGUIサブシステムの `.exe` を作成できる
- Releaseビルド前に `go test -mod=vendor ./...` が実行される
- ReleaseビルドはGo依存を `vendor/` から解決できる
- macOS成果物 `YAML-Struct-Editor-macOS.zip` がReleaseへ添付される
- Windows成果物 `YAML-Struct-Editor-Windows.zip` がReleaseへ添付される

### Windows npmなしビルド

- `frontend/dist/index.html` が存在する場合、`scripts\build-windows-offline.ps1` でnpmを実行せずWindows exeを作成できる
- `scripts\build-windows-offline.ps1` はGo依存を `vendor/` から解決してWindows exeを作成できる
- `frontend/dist/vs/loader.js` が存在し、配布用exeがMonaco Editorのloaderを外部CDNへ依存せず読み込める
- `frontend/dist/index.html` が存在しない場合、`scripts\build-windows-offline.ps1` は明示エラーを出す
- npmなしビルドでもWindows GUIサブシステム指定によりコンソールが表示されない

### Feature branch Windows artifact

- GitHub Actionsの `Windows Artifact` workflowを手動実行できる
- 手動実行時に対象ブランチを選択できる
- Releaseを作成せず、Windows exe zipをartifactとして取得できる
- artifact生成前に `go test -mod=vendor ./...` が実行される

### Frontend tab state

- 新規作成時に未保存タブを追加できる
- 複数ファイルをタブとして保持できる
- タブ切り替え時にYAML本文、カーソル位置、検証診断を切り替えられる
- 同じファイルパスを開いた場合は既存タブをアクティブにできる
- 保存先未確定または未保存変更があるタブをdirty状態として判定できる
- 保存先未確定または未保存変更があるタブを閉じる場合に確認できる
- 最後のタブを閉じた場合に空の未保存タブを作成できる

### Frontend shortcuts

- `Cmd/Ctrl + N` で新規タブを作成できる
- `Cmd/Ctrl + O` でYAMLファイルを開ける
- `Cmd/Ctrl + S` でアクティブなタブを保存できる
- `Cmd/Ctrl + W` でアクティブなタブを閉じられる
- 保存先未確定または未保存変更があるタブを `Cmd/Ctrl + W` で閉じる場合に確認できる
- `Cmd/Ctrl + Tab` で次のタブへ切り替えられる
- `Cmd/Ctrl + Shift + Tab` で前のタブへ切り替えられる
- `Esc` でアプリ内確認ダイアログを閉じられる
- Monaco Editor標準の編集ショートカットが維持される

### 今後の改善候補

今後の改善候補を実装する場合は、対象機能に応じて以下のテストを追加する。

- 最近開いたファイルを起動時にバックエンドから読み込み、開く・保存の操作後に表示が更新される
- 開いたYAMLファイルのパスがタブ状態に保持され、保存済みファイルの再保存では保存先ダイアログを表示しない
- Wails API、スキーマ取得、検証、補完に失敗した場合、UIが空結果として黙殺せずエラー状態を表示する
- Completion providerが配列内、4スペースインデント、未完成YAMLでも可能な範囲で正しい階層の候補を返す
- Monaco diagnosticsが診断の開始位置と終了位置に合わせて適切な範囲をマークする
- フロントエンドのタブ状態、API正規化、ショートカット、保存フローを自動テストで検証する
- 外部スキーマ読み込み失敗時に、GUI上で原因を確認できる
- 複数スキーマを扱う場合、選択中スキーマに応じて検証、補完、スキーマペインが切り替わる
- `tool` の選択値が変わった場合、`args` 配下の補完候補と検証結果が対応するstructに切り替わる
- macOSビルド時のlinker警告について、原因と許容可否を記録し、必要に応じて `MACOSX_DEPLOYMENT_TARGET` とCI設定を調整する

---

## 手動確認

### Frontend

- Monaco EditorでYAMLを編集できる
- `dates:` で改行すると `day1.date` と `day1.holiday` が自動入力される
- 自動入力された `day1` ブロックは1回のUndoで取り消せる
- `dayN.holiday` に `true` または `false` を入力して改行すると次の `dayN+1.date` と `dayN+1.holiday` が補完候補として表示される
- `dayN+1` 補完候補を確定した場合だけ次の日付ブロックが挿入される
- 直前の `dayN.date` が `YYYY-MM-DD` 形式の場合、補完候補の `dayN+1.date` に翌日が入力される
- `schedules:` で改行すると登録済みの `run1`, `run2`, `run3` 情報が自動入力される
- Schedulesメニューからschedule登録情報を変更でき、変更後の内容が次回以降の `schedules:` 自動入力に使われる
- 複数ファイルをタブで開き、タブ切り替えで表示内容が変わる
- タブごとに未保存状態が表示される
- 保存先未確定または未保存変更があるタブを閉じるとアプリ内確認ダイアログが表示される
- SaveボタンからOSの保存ダイアログで保存先を指定してYAMLファイルを保存できる
- Saveボタンはアクティブなタブだけを保存する
- ショートカットキーで新規、開く、保存、閉じる、タブ切り替え、確認ダイアログキャンセルが実行できる
- 行番号が表示される
- 折りたたみが利用できる
- Undo / Redoが利用できる
- YAMLシンタックスハイライトが表示される
- 補完候補が表示される
- 検証エラーが赤線で表示される
- タブ切り替え時にMonaco diagnosticsとエラー一覧がアクティブなタブの内容に切り替わる
- エラー一覧をクリックすると該当位置へ移動する
- ウィンドウサイズを変更してもエラー一覧のメッセージが見切れず、必要に応じて一覧内でスクロールできる
- スキーマペインにカーソル位置のキー名、型、必須/任意、説明、デフォルト値、enumが表示される
- スキーマペインに親階層のパンくずと同階層の入力可能キーが表示される
- タブ切り替え時にスキーマペインがアクティブなタブのカーソル位置に追従する
- スキーマペインでスキーマ全体を検索できる
- スキーマペインで全体ツリーを補助的に確認できる

---

## 実行コマンド

```sh
go test -mod=vendor ./...
```

フロントエンド実装後は、Wails v3の開発サーバーまたはビルドコマンドで手動確認を行う。

配布用ビルドの確認:

```sh
scripts/build-macos-app.sh
open "dist/YAML Struct Editor.app"
```

macOSでは `./yaml-struct-editor` のような生の実行ファイルではなく、`.app` をFinderまたは `open` で起動し、Terminalが表示されないことを確認する。

Windows配布用exeはWindows環境で以下を実行し、起動時にコンソールが表示されないことを確認する。

```powershell
scripts\build-windows.ps1
```

npmなしでWindows配布用exeを作る場合は、事前生成済みの `frontend/dist` を含めた状態で以下を実行する。

```powershell
scripts\build-windows-offline.ps1
```

Go toolchainを更新した場合は、以下でGoバージョンとテストを確認する。

```sh
go version
go test -mod=vendor ./...
```

GitHub ActionsのReleaseビルド確認:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Releaseに以下の成果物が添付されることを確認する。

- `YAML-Struct-Editor-macOS.zip`
- `YAML-Struct-Editor-Windows.zip`

FeatureブランチのWindows exeを確認する場合は、GitHub Actionsの `Windows Artifact` workflowを手動実行し、対象ブランチを選んでartifactを取得する。
