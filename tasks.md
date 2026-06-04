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
- [ ] Validatorの必須項目とenum診断を実装する
- [ ] Completion providerを実装する
- [ ] File serviceを実装する
- [ ] App serviceを実装する
- [ ] Wails v3アプリの最小構成を作成する
- [ ] Monaco Editor画面を実装する
- [ ] エラー一覧表示を実装する
- [ ] スキーマペイン表示を実装する
- [ ] ファイル操作UIを実装する
- [ ] 統合動作を確認する

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
