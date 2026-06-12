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
- `yaml` タグに `omitempty` があるフィールドを任意項目、ないフィールドを必須項目として解析できる
- `json` / `xml` 専用フィールドが混在しても解析対象にしない
- YAML対象フィールドに `json` タグが併記されていても `yaml` タグのキー名で解析できる
- `desc` タグを取得できる
- `default` タグを取得できる
- `enum` タグをカンマ区切りで取得できる
- ネストしたstructを解析できる
- slice / array / mapを解析できる
- string / bool / int系 / float系を解析できる
- `type Mode string` のような名前付きスカラー型を基底スカラー型として解析できる
- `const`、メソッド、YAML対象外structが混在してもYAML対象structを解析できる
- 外部Goソースフォルダから対象structを解析できる
- ルートstruct名を省略した場合に、単一のroot候補を自動検出できる
- root候補が複数ある場合は曖昧なスキーマとして明示エラーにできる
- root型名が `Config` ではない別サンプルスキーマを読み込み、組み込みサンプルとは異なる補完候補を返せる
- フォルダ内の複数Goソースファイルに分かれたstructを解析できる
- ファイルをまたぐ名前付きstruct参照を解決し、子フィールドを内部スキーマに展開できる
- `pkg.Type` 形式のimport先パッケージ参照をGOPATH配下から解決し、参照先のstructと名前付きスカラー型を内部スキーマに展開できる
- `*_test.go` を外部Goソース読み込み対象から除外できる
- 存在しない外部Goソースフォルダや対象structを明示エラーにできる
- GOPATH配下から解決できないimport先パッケージの型、標準ライブラリパッケージ型、type alias、generic型、循環参照を未対応として明示エラーにできる
- `app/sampleschema` に分割したサンプルスキーマをGo型名固定なしで解析できる
- `app/sampleschema` にJSON/XML専用構造体が混在してもYAMLタグ付きフィールドだけを解析できる
- `app/sampleschema` に名前付きスカラー型、const、メソッド、YAML対象外struct、slice、mapが混在してもtool schemaを解析できる
- `app/sampleschema` に `type AddAccounts []AddAccount` のような名前付きsliceがあってもtool schemaとして解析できる
- `app/rootschema` に分割した組み込みRootスキーマから `File` をrootとして解析できる
- `app/rootschema` の `omitempty` 指定に基づいてRootスキーマのRequired/Optionを解析できる
- `common.schema_version` は任意項目として解析できる
- 引数指定なしのApp serviceが `File` rootの `schema_version`, `common`, `scenario` を登録できる
- App serviceが起動後に選択された外部Goソースフォルダを読み込み、組み込みroot schemaを維持したままtool schemaだけを差し替えられる
- 外部スキーマ読み込みに失敗した場合はエラーを返し、既存スキーマを維持できる
- `common.dates` が `dayN` の値として `date` とbool型の `holiday` を持つ構造として解析できる
- `common.number_of_days` がint64由来のint型必須項目として解析できる
- `common.schedules` が `runN` の値としてint型を持つmap構造として解析できる

### Schema registry

- root schemaを登録できる
- 外部Goソースフォルダからroot schemaを登録できる
- 登録済みroot schemaを取得できる
- UI表示用Schema APIではroot schemaを除外し、参照用tool schemaを返せる
- 外部Goソースフォルダ指定時は、UI表示用Schema APIが外部tool schemaを返し、組み込みsampleschemaを含めない
- Schema paneのCurrent表示は、通常はRootスキーマ上のカーソル位置に対応する構造体名とフィールド一覧を表示できる
- Schema paneのCurrent表示は、`scenario:` で `Scenario` のフィールド一覧、`steps:` およびstep要素内で `Step` のフィールド一覧を表示できる
- Schema paneのCurrent表示は、`action` 欄では `tool` 行のみ参照用tool schema一覧を表示し、それ以外では `Action` のフィールド一覧を表示できる
- Schema paneのCurrent表示は、`args` 内では選択toolのフィールド一覧を表示できる
- Schema paneのAvailable keysはキー名、型、`Required` / `Option` を表示できる
- tool schema自体がsliceの場合、Schema paneのCurrent表示は1要素分のフィールド一覧を表示できる
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
- tool schema自体がsliceの場合、`args` をYAMLリストとして検証できる
- `tool` が未指定または未解決の場合に `args` 配下でpanicせず明示的に扱える
- 不正なYAMLでもpanicしない
- `common.schedules` のAnchor付きrun値を正しいint値として検証できる
- `common: !include "relative/path.yaml"` をCommon schemaとして検証できる
- include先が `common` の中身だけの場合とトップレベル `common:` を持つ場合の両方をCommon schemaとして検証できる
- 実ファイルパスを保持しているタブでは、親YAMLファイルの場所を基準にinclude相対パスを解決できる
- include先が未解決、絶対パス、構文エラー、またはネストincludeを含む場合に診断できる

### Completion provider

- カーソル位置に応じて候補を返せる
- YAMLキー名を候補に含める
- Go型、説明、必須/任意、デフォルト値、enumを候補に含める
- 同一階層に存在するキーを重複候補として返さない
- `scenario.steps` のようなslice / array要素配下で要素structのキー候補を返せる
- 外部Goソースの複数ファイル間struct参照から展開された子フィールドを候補に含める
- `tool` の値位置では参照用サンプルスキーマの識別子を `.` の前後で分割し、パッケージ名候補と構造体名候補を段階的に含め、UIではダブルクォート付きで挿入できる
- 参照用サンプルスキーマのサブディレクトリを再帰的に読み込み、`cloud.ecs.RunTask` のような階層付きtool識別子を候補に含める
- `tool` の値位置では階層付きtool識別子を `cloud.`、`ecs.`、`RunTask` のように段階的に含める
- `tool` の構造体名補完を確定した場合、同階層に `args` がなければ `args:` と選択された構造体のYAML対象フィールド名を自動入力できる
- `tool` の構造体名を補完し直した場合、既存の `args` ブロックを修正後のtoolに応じた内容へ置き換えられる
- `args` 配下では同一階層の `tool` 値に対応するstructのYAML対象フィールドをキー候補に含める
- `args` 補完候補にslice / arrayの要素スキーマを含め、struct要素をYAMLリスト形式で展開できる
- tool schema自体がsliceの場合、`args` 自動補完をYAMLリスト形式で展開できる
- リスト要素ブロックの末尾で改行した場合、内側リストと親リストの次要素候補を対象名付きで提示できる
- `scenario.steps[].day_ref` の値位置では、現在のYAML本文の `common.dates` に存在するdayキーを候補に含め、型情報ではなく `date` と `holiday` を参照情報として表示できる
- `scenario.steps[].schedule_ref` の値位置では、現在のYAML本文の `common.schedules` に存在するrunキーを候補に含め、型情報ではなくrun値とコメントを参照情報として表示できる
- `common` がinclude定義の場合でも、include先の `dates` / `schedules` から `day_ref` / `schedule_ref` 候補を生成できる
- `common` の値位置で、include common fileの補完候補を表示できる

### File service

- UTF-8のYAMLファイルを開ける
- ToolbarのOpen操作でOSファイル選択ダイアログから選択した実ファイルパスをタブに保持できる
- UTF-8のYAMLファイルを保存できる
- 新規ファイル状態を作成できる
- 新規ファイル状態にはRootスキーマの必須キーだけを初期YAMLとして補填し、任意キーを含めない
- 新規ファイル状態の初期YAMLにはRoot `schema_version: "1.0.0"`、`common.dates.day1`、`common.number_of_days: 1`、デフォルトの `common.schedules.run1` から `run3` を含める
- 新規ファイル状態の初期YAMLでは `common` をインライン定義として補填できる
- 新規ファイル状態の初期YAMLでは任意キーの `day_ref` と `schedule_ref` を省略できる
- 初期YAMLとScheduleテンプレートの既定値は `app/rootschema/defaults.yaml` から取得できる
- 保存済みSchedulesテンプレートがある場合、起動時と新規作成時の初期YAMLの `common.schedules` に同じ内容が反映される
- 最近開いたファイルを保存できる
- 最近開いたファイルを取得できる
- 起動時にバックエンドの最近開いたファイル一覧を読み込み、固定のダミー値を表示しない
- Recentから選択した保存済みパスをバックエンド経由で開ける

### Release build

- タグ `v*` のpushでGitHub ActionsのRelease workflowが起動する
- macOS runnerで `.app` バンドルを作成できる
- Windows runnerでGUIサブシステムの `.exe` を作成できる
- Releaseビルド前に `go test -mod=vendor ./...` が実行される
- ReleaseビルドはGo依存を `vendor/` から解決できる
- macOS成果物 `YAML-Struct-Editor-macOS.zip` がReleaseへ添付される
- Windows成果物 `YAML-Struct-Editor-Windows.zip` がReleaseへ添付される

### External schema startup selection

- 起動後に外部スキーマ読み込み確認を表示できる
- 起動後の確認はWailsのネイティブQuestionダイアログで表示できる
- Yesを選んだ場合にOSフォルダ選択ダイアログを表示できる
- 選択したフォルダをApp serviceへ渡して外部tool schemaを読み込める
- Noまたはキャンセルを選んだ場合は組み込みスキーマのまま起動できる
- 外部スキーマ読み込みに失敗した場合はエラーを表示し、組み込みスキーマのまま継続できる

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
- 起動時と新規作成時の未保存タブにはRootスキーマの必須キーだけが初期YAMLとして表示される
- 複数ファイルをタブとして保持できる
- タブ切り替え時にYAML本文、カーソル位置、検証診断を切り替えられる
- 同じファイルパスを開いた場合は既存タブをアクティブにできる
- 保存先未確定または未保存変更があるタブをdirty状態として判定できる
- 保存先未確定または未保存変更があるタブを閉じる場合に確認できる
- 最後のタブを閉じた場合にタブ0件の空状態へ遷移できる
- タブ0件状態でもNewとOpenで新しいアクティブタブを作成できる
- タブ0件状態でSaveやCloseショートカットを実行してもpanicせず何もしない

### Frontend shortcuts

- `Cmd/Ctrl + N` で新規タブを作成できる
- `Cmd/Ctrl + O` でYAMLファイルを開ける
- `Cmd/Ctrl + S` でアクティブなタブを保存できる
- `Cmd/Ctrl + W` でアクティブなタブを閉じられる
- `Cmd/Ctrl + C` で選択範囲または現在行をコピーできる
- `Cmd/Ctrl + X` で選択範囲または現在行を切り取れる
- `Cmd/Ctrl + V` でクリップボードのテキストを貼り付けられる
- `Shift + ←/→/↑/↓` で範囲選択しながらカーソル移動できる
- 保存先未確定または未保存変更があるタブを `Cmd/Ctrl + W` で閉じる場合に確認できる
- `Cmd/Ctrl + Tab` で次のタブへ切り替えられる
- `Cmd/Ctrl + Shift + Tab` で前のタブへ切り替えられる
- `Esc` でアプリ内確認ダイアログを閉じられる
- Monaco Editor標準の編集ショートカットが維持される
- ショートカットキーが対応するToolbarボタンにマウスホバーまたはフォーカスするとショートカット表示が出る

### 今後の改善候補

今後の改善候補を実装する場合は、対象機能に応じて以下のテストを追加する。

- OSファイル選択で開いたYAMLファイルの実パスがタブ状態に保持され、保存済みファイルの再保存では保存先ダイアログを表示しない
- Wails API、スキーマ取得、検証、補完に失敗した場合、UIが空結果として黙殺せずエラー状態を表示する
- Completion providerが配列内、4スペースインデント、未完成YAMLでも可能な範囲で正しい階層の候補を返す
- Monaco diagnosticsが診断の開始位置と終了位置に合わせて適切な範囲をマークする
- フロントエンドのタブ状態、API正規化、ショートカット、保存フローを自動テストで検証する
- 外部スキーマ読み込み失敗時に、GUI上で原因を確認できる
- 複数スキーマを扱う場合、選択中スキーマに応じて検証、補完、スキーマペインが切り替わる
- macOSビルド時のlinker警告について、原因と許容可否を記録し、必要に応じて `MACOSX_DEPLOYMENT_TARGET` とCI設定を調整する

---

## 手動確認

### Frontend

- Monaco EditorでYAMLを編集できる
- ToolbarのCopyで選択範囲または現在行をコピーできる
- ToolbarのCutで選択範囲または現在行を切り取れる
- ToolbarのPasteでクリップボードのテキストを選択範囲またはカーソル位置へ貼り付けられる
- 右クリックメニューのCopy / Cut / PasteでToolbarと同じクリップボード操作を実行できる
- `Shift + ←/→/↑/↓` で範囲選択でき、選択範囲をCopy/Cut/Paste操作の対象にできる
- `dates:` で改行すると `day1.date` と `day1.holiday` が自動入力される
- 自動入力された `day1` ブロックは1回のUndoで取り消せる
- `dayN.holiday` に `true` または `false` を入力して改行すると次の `dayN+1.date` と `dayN+1.holiday` が補完候補として表示される
- `dayN.holiday` に `true` または `false` を入力して改行すると `dates` と同階層の `schedules` ブロックも補完候補として表示される
- `dayN+1` 補完候補を確定した場合だけ次の日付ブロックが挿入される
- 直前の `dayN.date` が `YYYY-MM-DD` 形式の場合、補完候補の `dayN+1.date` に翌日が入力される
- `schedules:` で改行すると登録済みの `run1`, `run2`, `run3` 情報が自動入力される
- Schedulesメニューからschedule登録情報を変更でき、変更後の内容が次回以降の `schedules:` 自動入力に使われる
- `steps:` で改行すると最初のstepリスト要素が自動入力される
- step内のキー補完候補が存在する場合、次のstep追加候補が混在しない
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
- `day_ref` などのキー入力中に本文内の `day1` のような単語候補が混在しない
- タブ切り替え時にMonaco diagnosticsとエラー一覧がアクティブなタブの内容に切り替わる
- エラー一覧をクリックすると該当位置へ移動する
- エラー一覧の行番号とメッセージを範囲選択し、OS標準のコピー操作でコピーできる
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
