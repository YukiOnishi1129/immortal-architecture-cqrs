# 🏗️ CQRSアーキテクチャガイド - 読み書き分離の実装

> 💡 **このドキュメントのゴール**
> Clean Architecture（CRUD）を **CQRSにどう組み替えたか** が
> 具体的にわかること。実際のコードを見ながら読んでね。

---

## 🔄 Before / After を一枚絵で見よう

### Before: Clean Architecture（CRUD）

```
📱 画面
 │
 ▼
┌─────────────────────────────────────────────────┐
│ Controller（HTTP）                                │
│   NoteController                                 │
│     .List()   ← 読む                             │
│     .GetByID()← 読む                             │
│     .Create() ← 書く    ← 全部同じControllerに    │
│     .Update() ← 書く      まとまってる           │
│     .Delete() ← 書く                             │
└──────────┬──────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────┐
│ UseCase（NoteInteractor）                        │
│   .List()    ← 読む                              │
│   .Get()     ← 読む                              │
│   .Create()  ← 書く    ← 全部同じInteractorに     │
│   .Update()  ← 書く      まとまってる            │
│   .Delete()  ← 書く                              │
└──────────┬──────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────┐
│ Repository（NoteRepository）                     │
│   .List()    ← 読む（3テーブルJOIN）              │
│   .Get()     ← 読む（3テーブルJOIN）              │
│   .Create()  ← 書く                              │
│   .Update()  ← 書く    ← 全部同じRepositoryに     │
│   .Delete()  ← 書く      まとまってる            │
└──────────┬──────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────┐
│ PostgreSQL                                       │
│   notes / templates / accounts / sections / fields│
│   ← 読むのも書くのも同じテーブル                   │
└─────────────────────────────────────────────────┘
```

**問題：全部が1本の道を通ってる。読むも書くもごちゃ混ぜ。**

---

### After: CQRS（読み書き分離）

```
📱 画面
 │
 ├──── 読む（Query）──────────────── 書く（Command）────┐
 ▼                                    ▼                │
┌──────────────────────┐  ┌────────────────────────┐   │
│ Controller           │  │ Controller             │   │
│  .List()             │  │  .Create()             │   │
│  .GetByID()          │  │  .Update()             │   │
│                      │  │  .Delete()             │   │
│  読む専用！           │  │  .Publish()            │   │
└────────┬─────────────┘  └────────┬───────────────┘   │
         │                         │                    │
         ▼                         ▼                    │
┌──────────────────────┐  ┌────────────────────────┐   │
│ QueryUseCase         │  │ CommandUseCase         │   │
│  .List()             │  │  .Create()             │   │
│  .Get()              │  │  .Update()             │   │
│                      │  │  .Delete()             │   │
│  バリデーションなし    │  │  .ChangeStatus()       │   │
│  ただ取るだけ         │  │                        │   │
│                      │  │  バリデーションあり      │   │
│                      │  │  ドメインロジック実行    │   │
└────────┬─────────────┘  └────────┬───────────────┘   │
         │                         │                    │
         ▼                         ▼                    │
┌──────────────────────┐  ┌────────────────────────┐   │
│ QueryRepository      │  │ CommandRepository      │   │
│（Read Model専用）     │  │（Write用）              │   │
│  .List()             │  │  .Create()             │   │
│  .Get()              │  │  .Update()             │   │
│                      │  │  .Delete()             │   │
│  JOINなし！一発！     │  │  .ReplaceSections()    │   │
└────────┬─────────────┘  └──────┬──┬──────────────┘   │
         │                       │  │                   │
         ▼                       │  ▼                   │
┌──────────────────────┐  ┌──│──────────────────────┐  │
│ note_read_models     │  │  │ notes               │  │
│（読む専用テーブル）    │  │  │ templates           │  │
│                      │  │  │ accounts            │  │
│ JOINの結果が         │◀─┘  │ sections            │  │
│ 事前に入ってる！      │     │ fields              │  │
│                      │同期  │                     │  │
└──────────────────────┘     └─────────────────────┘  │
         PostgreSQL（同じDB内）                          │
```

---

## 🧩 ファイル構成 - 何がどこにある？

### CRUD時代のファイル構成

```
internal/
├── port/
│   └── note_port.go          ← NoteInputPort（読み書き混在）
│                                NoteRepository（読み書き混在）
├── usecase/
│   └── note_interactor.go    ← List/Get/Create/Update/Delete（全部入り）
│
├── adapter/
│   ├── http/controller/
│   │   └── note_controller.go ← 全エンドポイント（全部入り）
│   └── gateway/db/sqlc/
│       └── note_repository.go ← 読み書き両方のSQL
│
└── driver/
    ├── factory/
    │   ├── repository_factory.go  ← NoteRepoFactory（1つ）
    │   └── usecase_factory.go     ← NoteInputFactory（1つ）
    └── initializer/
        └── api/initializer.go     ← 全部ここで配線
```

### CQRS後のファイル構成（実際のコード）

```
internal/
├── port/
│   ├── note_command_port.go         ← Command系のポート
│   │                                   → NoteCommandInputPort
│   │                                   → NoteCommandOutputPort
│   │                                   → NoteCommandRepository
│   └── note_query_port.go          ← Query系のポート
│                                       → NoteQueryInputPort
│                                       → NoteQueryOutputPort
│                                       → NoteQueryRepository
│
├── usecase/
│   ├── note_command_interactor.go   ← Command系のユースケース
│   │                                   書く + Read Model同期
│   └── note_query_interactor.go     ← Query系のユースケース
│                                       キャッシュテーブルから取るだけ
│
├── adapter/
│   ├── http/
│   │   ├── controller/
│   │   │   └── note_controller.go   ← Command/Queryを使い分ける
│   │   └── presenter/
│   │       ├── note_command_presenter.go ← Command用プレゼンター
│   │       └── note_query_presenter.go   ← Query用プレゼンター
│   └── gateway/db/sqlc/
│       ├── note_command_repository.go ← Command用リポジトリ
│       ├── note_query_repository.go   ← Query用リポジトリ
│       └── queries/
│           ├── notes.sql              ← Command用SQL（既存）
│           └── note_read_models.sql   ← Query用SQL
│
├── domain/
│   └── note/
│       └── read_model.go            ← NoteReadModel（読む用の型）
│
└── driver/
    ├── factory/
    │   ├── repository_factory.go    ← Command/Query両方のFactory
    │   ├── usecase_factory.go       ← Command/Query両方のFactory
    │   └── http/
    │       └── presenter_factory.go ← Command/Query両方のFactory
    └── initializer/
        └── api/initializer.go       ← Command/Query両方を配線

migrations/
├── 20250210000000_add_note_read_models.up.sql   ← キャッシュテーブル作成
└── 20250210000001_seed_note_read_models.up.sql  ← 既存データの初回同期
```

---

## 📐 各レイヤーの実装を見ていこう

### 1️⃣ ドメイン層 - Read Modelの型を追加

> 📂 `internal/domain/note/read_model.go`

```go
// NoteReadModel はRead Model（読む専用）のデータ構造。
// キャッシュテーブル note_read_models に対応する。
// JOINの結果を事前にまとめたもの。
type NoteReadModel struct {
    ID             string
    Title          string
    Status         NoteStatus
    TemplateID     string
    TemplateName   string         // ← templates テーブルの情報が入ってる！
    OwnerID        string
    OwnerFirstName string         // ← accounts テーブルの情報が入ってる！
    OwnerLastName  string
    OwnerThumbnail *string
    Sections       []SectionReadModel  // ← sections + fields の情報が入ってる！
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type SectionReadModel struct {
    ID         string
    FieldID    string
    FieldLabel string
    FieldOrder int
    IsRequired bool
    Content    string
}
```

**ポイント：JOINしないと取れなかった情報が、1つの構造体にぜんぶ入ってる。**

ドメイン層で変更したのは **このファイル1つだけ**。
`entity.go`、`logic.go`、`aggregate.go` は一切触ってない。

---

### 2️⃣ ポート層 - CommandとQueryを分ける

#### Command側

> 📂 `internal/port/note_command_port.go`

```go
// NoteCommandInputPort は書く専用のユースケース。
type NoteCommandInputPort interface {
    Create(ctx context.Context, input NoteCreateInput) error
    Update(ctx context.Context, input NoteUpdateInput) error
    ChangeStatus(ctx context.Context, input NoteStatusChangeInput) error
    Delete(ctx context.Context, id, ownerID string) error
}

// NoteCommandOutputPort は書く専用のプレゼンター。
type NoteCommandOutputPort interface {
    PresentNote(ctx context.Context, note *note.WithMeta) error
    PresentNoteDeleted(ctx context.Context) error
}

// NoteCommandRepository は書く専用のリポジトリ。
type NoteCommandRepository interface {
    Get(ctx context.Context, id string) (*note.WithMeta, error)   // バリデーション用
    Create(ctx context.Context, n note.Note) (*note.Note, error)
    Update(ctx context.Context, n note.Note) (*note.Note, error)
    UpdateStatus(ctx context.Context, id string, status note.NoteStatus) (*note.Note, error)
    Delete(ctx context.Context, id string) error
    ReplaceSections(ctx context.Context, noteID string, sections []note.Section) error
}
```

#### Query側

> 📂 `internal/port/note_query_port.go`

```go
// NoteQueryInputPort は読む専用のユースケース。
type NoteQueryInputPort interface {
    List(ctx context.Context, filters note.Filters) error
    Get(ctx context.Context, id string) error
}

// NoteQueryOutputPort は読む専用のプレゼンター。
type NoteQueryOutputPort interface {
    PresentNoteList(ctx context.Context, notes []note.NoteReadModel) error
    PresentNote(ctx context.Context, note *note.NoteReadModel) error
}

// NoteQueryRepository は読む専用のリポジトリ。
// Upsert/Delete は Command 側から呼ばれる同期用メソッド。
type NoteQueryRepository interface {
    List(ctx context.Context, filters note.Filters) ([]note.NoteReadModel, error)
    Get(ctx context.Context, id string) (*note.NoteReadModel, error)
    Upsert(ctx context.Context, model note.NoteReadModel) error  // ← 同期用
    Delete(ctx context.Context, id string) error                  // ← 同期用
}
```

#### 図で見ると

```
【Before: CRUD】                  【After: CQRS】

NoteInputPort                     NoteQueryInputPort（読む）
  .List()    ← 読む                .List()
  .Get()     ← 読む                .Get()
  .Create()  ← 書く
  .Update()  ← 書く               NoteCommandInputPort（書く）
  .Delete()  ← 書く                .Create()
                                    .Update()
                                    .ChangeStatus()
                                    .Delete()

NoteRepository                    NoteQueryRepository（読む）
  .List()    ← 読む                .List()
  .Get()     ← 読む                .Get()
  .Create()  ← 書く                .Upsert()    ← 同期用
  .Update()  ← 書く                .Delete()    ← 同期用
  .Delete()  ← 書く
                                  NoteCommandRepository（書く）
                                    .Get()       ← バリデーション用
                                    .Create()
                                    .Update()
                                    .Delete()
                                    .ReplaceSections()
```

---

### 3️⃣ ユースケース層 - QueryとCommandの中身

#### Query側 — めっちゃシンプル

> 📂 `internal/usecase/note_query_interactor.go`

```go
type NoteQueryInteractor struct {
    queryRepo port.NoteQueryRepository
    output    port.NoteQueryOutputPort
}

// List はキャッシュテーブルから一覧を取得する。JOINなし！
func (u *NoteQueryInteractor) List(ctx context.Context, filters note.Filters) error {
    notes, err := u.queryRepo.List(ctx, filters)
    if err != nil {
        return err
    }
    return u.output.PresentNoteList(ctx, notes)
}

// Get はキャッシュテーブルから1件取得する。JOINなし！
func (u *NoteQueryInteractor) Get(ctx context.Context, id string) error {
    n, err := u.queryRepo.Get(ctx, id)
    if err != nil {
        return err
    }
    return u.output.PresentNote(ctx, n)
}
```

**バリデーションもドメインロジックもない。ただ取って返すだけ。これがQuery側の強み。**

#### Command側 — 書く＋Read Model同期

> 📂 `internal/usecase/note_command_interactor.go`

```go
type NoteCommandInteractor struct {
    commandRepo port.NoteCommandRepository   // ← Write用
    queryRepo   port.NoteQueryRepository     // ← Read Model同期用
    templates   port.TemplateRepository
    tx          port.TxManager
    output      port.NoteCommandOutputPort
}

// Create はノートを作成し、Read Modelも同期する。
func (u *NoteCommandInteractor) Create(ctx context.Context, input port.NoteCreateInput) error {
    // ... バリデーション ...

    err = u.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
        // 1. Write DBに書く
        nn, err := u.commandRepo.Create(txCtx, newNote)
        // ... sections も作成 ...

        // 2. Read Modelも同期する（同じトランザクション内！）
        created, err := u.commandRepo.Get(txCtx, noteID)
        if err != nil {
            return err
        }
        return u.queryRepo.Upsert(txCtx, toReadModel(created))
    })
    // ...
}

// Delete はノートを削除し、Read Modelからも消す。
func (u *NoteCommandInteractor) Delete(ctx context.Context, id, ownerID string) error {
    // ... バリデーション ...

    err = u.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
        if err := u.commandRepo.Delete(txCtx, id); err != nil {
            return err
        }
        return u.queryRepo.Delete(txCtx, id)  // ← キャッシュからも消す
    })
    // ...
}
```

**ポイント：書くときに同じトランザクションでRead Modelも更新する。だから整合性が保たれる。**

```
トランザクション開始
  │
  ├─ notesテーブルに書く          ✅
  ├─ sectionsテーブルに書く       ✅
  ├─ note_read_modelsに同期      ✅
  │
  └─ 全部成功 → COMMIT
     1つでも失敗 → ROLLBACK（全部元に戻る）
```

---

### 4️⃣ アダプター層 - キャッシュテーブルとリポジトリ

#### キャッシュテーブル（マイグレーション）

> 📂 `migrations/20250210000000_add_note_read_models.up.sql`

```sql
CREATE TABLE note_read_models (
    id UUID PRIMARY KEY,                -- notes.id と同じ
    title TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('Draft', 'Publish')),
    template_id UUID NOT NULL,
    template_name TEXT NOT NULL,         -- ← JOINしなくていい！
    owner_id UUID NOT NULL,
    owner_first_name TEXT NOT NULL,      -- ← JOINしなくていい！
    owner_last_name TEXT NOT NULL,
    owner_thumbnail TEXT,
    sections_json JSONB NOT NULL DEFAULT '[]',  -- ← 全セクション入り！
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_note_read_models_status ON note_read_models(status);
CREATE INDEX idx_note_read_models_owner_id ON note_read_models(owner_id);
CREATE INDEX idx_note_read_models_template_id ON note_read_models(template_id);
CREATE INDEX idx_note_read_models_updated_at ON note_read_models(updated_at DESC);
```

#### 初回同期（既存データの流し込み）

> 📂 `migrations/20250210000001_seed_note_read_models.up.sql`

CQRSを導入する前に作られたノートは `note_read_models` に入ってない。
だから **1回だけ** 既存データを流し込む必要がある。

```sql
INSERT INTO note_read_models (...)
SELECT
    n.id, n.title, n.status, n.template_id, t.name,
    n.owner_id, a.first_name, a.last_name, a.thumbnail,
    COALESCE(
        (SELECT jsonb_agg(...) FROM sections s JOIN fields f ...),
        '[]'::jsonb
    ),
    n.created_at, n.updated_at
FROM notes n
JOIN templates t ON t.id = n.template_id
JOIN accounts a ON a.id = n.owner_id
ON CONFLICT (id) DO NOTHING;
```

**これ以降はCommand経由で自動同期されるから、このSQLは2度と実行しなくていい。**

#### Query用SQLクエリ

> 📂 `internal/adapter/gateway/db/sqlc/queries/note_read_models.sql`

```sql
-- JOINなし！1テーブルから取るだけ！
-- name: ListNoteReadModels :many
SELECT * FROM note_read_models
WHERE (NULLIF($1::text, '') IS NULL OR status = $1)
  AND ($2::uuid IS NULL OR template_id = $2)
  AND ($3::uuid IS NULL OR owner_id = $3)
  AND (NULLIF($4::text, '') IS NULL OR title ILIKE '%' || $4 || '%')
ORDER BY updated_at DESC;

-- name: GetNoteReadModel :one
SELECT * FROM note_read_models WHERE id = $1;

-- name: UpsertNoteReadModel :exec
INSERT INTO note_read_models (...) VALUES (...)
ON CONFLICT (id) DO UPDATE SET ...;  -- ← 同じIDがあれば更新

-- name: DeleteNoteReadModel :exec
DELETE FROM note_read_models WHERE id = $1;
```

**Before（CRUD）のListNotesと比較してみよう：**

```sql
-- 【Before】 3テーブルJOIN + セクション別クエリ
SELECT n.*, t.name, a.first_name, a.last_name, a.thumbnail
FROM notes n
JOIN templates t ON t.id = n.template_id   -- JOIN!
JOIN accounts a ON a.id = n.owner_id       -- JOIN!
WHERE ...

-- 【After】 JOINなし！
SELECT * FROM note_read_models WHERE ...
```

#### Query用リポジトリ

> 📂 `internal/adapter/gateway/db/sqlc/note_query_repository.go`

```go
type NoteQueryRepository struct {
    pool    *pgxpool.Pool
    queries *generated.Queries
}

// List はキャッシュテーブルから取得。JOINなし。
func (r *NoteQueryRepository) List(ctx context.Context, filters note.Filters) ([]note.NoteReadModel, error) {
    rows, err := queriesForContext(ctx, r.queries).ListNoteReadModels(ctx, params)
    // ... rows を NoteReadModel に変換して返す
}

// Upsert はキャッシュテーブルを更新（Command側から呼ばれる）。
func (r *NoteQueryRepository) Upsert(ctx context.Context, model note.NoteReadModel) error {
    sectionsJSON, _ := json.Marshal(model.Sections)
    return queriesForContext(ctx, r.queries).UpsertNoteReadModel(ctx, &generated.UpsertNoteReadModelParams{
        // ... NoteReadModel → DB型に変換
    })
}
```

---

### 5️⃣ コントローラー層 - CommandとQueryの使い分け

> 📂 `internal/adapter/http/controller/note_controller.go`

```go
type NoteController struct {
    // Command用（書く）
    commandInputFactory  func(...) port.NoteCommandInputPort
    commandOutputFactory func() *presenter.NoteCommandPresenter

    // Query用（読む）
    queryInputFactory  func(...) port.NoteQueryInputPort
    queryOutputFactory func() *presenter.NoteQueryPresenter

    // 共通
    commandRepoFactory func() port.NoteCommandRepository
    queryRepoFactory   func() port.NoteQueryRepository
    tplRepoFactory     func() port.TemplateRepository
    txFactory          func() port.TxManager
}

// List は Query側を使う（読む）
func (c *NoteController) List(ctx echo.Context, params ...) error {
    queryInput, p := c.newQueryIO()  // ← Query用のUseCaseを生成
    if err := queryInput.List(...); err != nil {
        return handleError(ctx, err)
    }
    return ctx.JSON(http.StatusOK, p.Notes())
}

// Create は Command側を使う（書く）
func (c *NoteController) Create(ctx echo.Context) error {
    commandInput, p := c.newCommandIO()  // ← Command用のUseCaseを生成
    if err := commandInput.Create(...); err != nil {
        return handleError(ctx, err)
    }
    return ctx.JSON(http.StatusOK, p.Note())
}
```

**どのエンドポイントがどっちを使うか一覧：**

| エンドポイント | Query / Command | メソッド |
|--------------|----------------|---------|
| `GET /notes` | **Query** | `newQueryIO()` → `queryInput.List()` |
| `GET /notes/:id` | **Query** | `newQueryIO()` → `queryInput.Get()` |
| `POST /notes` | **Command** | `newCommandIO()` → `commandInput.Create()` |
| `PUT /notes/:id` | **Command** | `newCommandIO()` → `commandInput.Update()` |
| `DELETE /notes/:id` | **Command** | `newCommandIO()` → `commandInput.Delete()` |
| `POST /notes/:id/publish` | **Command** | `newCommandIO()` → `commandInput.ChangeStatus()` |
| `POST /notes/:id/unpublish` | **Command** | `newCommandIO()` → `commandInput.ChangeStatus()` |

---

### 6️⃣ ファクトリー＆初期化 - 配線

> 📂 `internal/driver/factory/repository_factory.go`

```go
func NewNoteCommandRepoFactory(pool *pgxpool.Pool) func() port.NoteCommandRepository {
    return func() port.NoteCommandRepository {
        return sqlc.NewNoteCommandRepository(pool)
    }
}

func NewNoteQueryRepoFactory(pool *pgxpool.Pool) func() port.NoteQueryRepository {
    return func() port.NoteQueryRepository {
        return sqlc.NewNoteQueryRepository(pool)
        // ↑ 将来Redisにしたいときはここを変えるだけ！
    }
}
```

> 📂 `internal/driver/initializer/api/initializer.go`

```go
noteCommandRepoFactory := factory.NewNoteCommandRepoFactory(pool)
noteQueryRepoFactory := factory.NewNoteQueryRepoFactory(pool)
noteCommandInputFactory := factory.NewNoteCommandInputFactory()
noteQueryInputFactory := factory.NewNoteQueryInputFactory()
noteCommandOutputFactory := httpfactory.NewNoteCommandOutputFactory()
noteQueryOutputFactory := httpfactory.NewNoteQueryOutputFactory()

nc := httpcontroller.NewNoteController(
    noteCommandInputFactory,   // Command用
    noteCommandOutputFactory,  // Command用
    noteQueryInputFactory,     // Query用
    noteQueryOutputFactory,    // Query用
    noteCommandRepoFactory,    // Command用
    noteQueryRepoFactory,      // Query用
    templateRepoFactory,
    txFactory,
)
```

---

## 🔄 データの流れを追ってみよう

### 📖 読む場合（GET /notes）

```
1. 📱 画面: GET /notes?status=Publish

2. 🎮 NoteController: List()
   → queryInput, p := c.newQueryIO()

3. 📋 NoteQueryInteractor: List()
   → queryRepo.List(ctx, filters)

4. 💾 NoteQueryRepository: List()
   → SELECT * FROM note_read_models WHERE ...
   → JOINなし！1テーブルから一発取得！

5. 🎨 NoteQueryPresenter: PresentNoteList()
   → NoteReadModel → APIレスポンスに変換

6. 📱 画面: ノート一覧が表示される（速い！）
```

### ✏️ 書く場合（POST /notes）

```
1. 📱 画面: POST /notes {title: "新しいノート", ...}

2. 🎮 NoteController: Create()
   → commandInput, p := c.newCommandIO()

3. 📋 NoteCommandInteractor: Create()
   → バリデーション実行
   → tx.WithinTransaction 開始

4.    💾 NoteCommandRepository:
      → notes テーブルに INSERT
      → sections テーブルに INSERT

5.    💾 NoteQueryRepository:
      → note_read_models テーブルに UPSERT（同期！）

      → トランザクション COMMIT
        （4と5が両方成功して初めて確定）

6. 🎨 NoteCommandPresenter: PresentNote()
   → ドメインモデル → APIレスポンスに変換

7. 📱 画面: 作成されたノートが表示される
```

---

## 🧱 変更しなかったところ（超重要）

CQRSにしても **変えなかったファイルがこんなにある**。
これがClean Architectureの強み。

```
┌─────────────────────────────────────────────────┐
│  ✅ 変更なし                                     │
│                                                  │
│  domain/note/entity.go      ← Note, Section     │
│  domain/note/logic.go       ← バリデーション     │
│  domain/note/aggregate.go   ← 集約操作           │
│  domain/note/types.go       ← Filters, WithMeta  │
│  domain/errors/             ← エラー定義         │
│  domain/service/            ← ドメインサービス    │
│  domain/template/           ← テンプレート全般    │
│  domain/account/            ← アカウント全般      │
│                                                  │
│  adapter/http/generated/    ← OpenAPI生成コード   │
│  adapter/grpc/              ← gRPC全般           │
│                                                  │
│  port/account_port.go       ← アカウント系ポート  │
│  port/template_port.go      ← テンプレート系ポート│
│  port/tx_port.go            ← トランザクション    │
│                                                  │
│  driver/config/             ← 設定               │
│  driver/db/                 ← DB接続/TxManager   │
│                                                  │
│  → ドメイン層はまったく触ってない！               │
│  → Account, Template も変更なし！                │
└─────────────────────────────────────────────────┘
```

---

## 🏫 学校で言うと

```
【Before: CRUD】

  先生（UseCase）が1人で
  「テストを作る」「テストを配る」「テストを採点する」を全部やってた

  保健室の先生も、音楽の先生も、この1人に頼んでた

【After: CQRS】

  「テストを作る先生」と「テストを配る先生」に分けた

  作る先生（Command）:
    → テスト問題を作る
    → 正式な記録に残す
    → ついでに配布用のコピーも作っておく ← 同期

  配る先生（Query）:
    → 事前にコピーされた配布物を配るだけ
    → 超速い！
    → 作る先生の仕事を邪魔しない！
```

---

## 🎯 まとめ

| 質問 | 答え |
|-----|------|
| 何を分けた？ | Port / UseCase / Repository / Presenter を Query（読む）と Command（書く）に分離 |
| ドメイン層は変わった？ | ほぼ変わらない。`read_model.go` を1つ追加しただけ |
| 同期はどうやる？ | Command側で書くとき、同じトランザクション内でRead Modelも更新 |
| 既存データは？ | 初回マイグレーションで既存ノートをキャッシュテーブルに流し込む |
| Controller は？ | 読むときは `newQueryIO()`、書くときは `newCommandIO()` で使い分け |
| 全員のデータが同期される？ | はい。誰がAPIを叩いても、Command経由で必ずキャッシュが更新される |
| Clean Architectureとの関係は？ | ポート（インターフェース）があるから、Read Modelの実装を自由に差し替えられる |
