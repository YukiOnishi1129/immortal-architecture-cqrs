# 🏗️ アーキテクチャ変更ガイド - CRUDからCQRSへ

> 💡 **このドキュメントのゴール**
> 今のClean Architecture（CRUD）を、**どこをどう変えてCQRSにするか**が
> 具体的にイメージできるようになること。

---

## 🔄 Before / After を一枚絵で見よう

### Before: 今のClean Architecture（CRUD）

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

## 🧩 じゃあ具体的にどこを変える？

### 今のファイル構成（変更前）

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

### 変更後のファイル構成

```
internal/
├── port/
│   ├── note_command_port.go         ← 🆕 Command系（書く）のポート
│   └── note_query_port.go          ← 🆕 Query系（読む）のポート
│
├── usecase/
│   ├── note_command_interactor.go   ← 🆕 Command系（書く）のユースケース
│   │                                   ※ 書いた後にRead Modelも更新する
│   └── note_query_interactor.go     ← 🆕 Query系（読む）のユースケース
│
├── adapter/
│   ├── http/controller/
│   │   └── note_controller.go       ← Command/Queryを使い分ける
│   └── gateway/db/sqlc/
│       ├── note_command_repository.go ← 🆕 Command用（書く専用）
│       └── note_query_repository.go   ← 🆕 Query用（Read Model読み取り）
│
├── domain/
│   └── note/
│       └── read_model.go            ← 🆕 Read Model（読む用のデータ構造）
│
└── driver/
    ├── factory/
    │   ├── repository_factory.go    ← 🆕 Command/Query両方のFactory追加
    │   └── usecase_factory.go       ← 🆕 Command/Query両方のFactory追加
    └── initializer/
        └── api/initializer.go       ← 🆕 Command/Query両方を配線
```

**🆕がついてるのが新しく作るファイル。既存ファイルの変更は最小限。**

---

## 📐 各レイヤーの変更を詳しく見よう

### 1️⃣ ドメイン層 - Read Modelを定義する

**今：** `note.WithMeta` が読み取り用のデータ構造。でもこれはJOINの結果を表現してるだけ。

**変更後：** Read Model専用のデータ構造を追加。

```
domain/note/
├── entity.go       ← Note, Section（変更なし）
├── types.go        ← WithMeta, Filters（変更なし）
├── logic.go        ← バリデーション（変更なし）
├── aggregate.go    ← 集約操作（変更なし）
└── read_model.go   ← 🆕 NoteReadModel（読む専用）
```

```go
// 🆕 read_model.go

// NoteReadModel はRead Model（読む専用）のデータ構造。
// JOINの結果を事前にまとめたもの。
type NoteReadModel struct {
    ID             string
    Title          string
    Status         NoteStatus
    TemplateName   string
    OwnerID        string
    OwnerFirstName string
    OwnerLastName  string
    OwnerThumbnail *string
    Sections       []SectionReadModel
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// SectionReadModel はセクションのRead Model。
type SectionReadModel struct {
    ID         string
    FieldID    string
    FieldLabel string
    FieldOrder int
    IsRequired bool
    Content    string
}
```

**ポイント：WithMetaと似てるけど、「DBのキャッシュテーブルに対応する構造」として独立させる。**

---

### 2️⃣ ポート層 - QueryとCommandを分ける

**今のNoteInputPort（読み書き混在）：**

```go
// 今のport/note_port.go
type NoteInputPort interface {
    List(ctx, filters) error        // ← 読む
    Get(ctx, id) error              // ← 読む
    Create(ctx, input) error        // ← 書く
    Update(ctx, input) error        // ← 書く
    ChangeStatus(ctx, input) error  // ← 書く
    Delete(ctx, id, ownerID) error  // ← 書く
}
```

**変更後：2つに分ける**

```go
// 🆕 port/note_command_port.go

// NoteCommandInputPort はCommand（書く）専用のユースケースインターフェース。
type NoteCommandInputPort interface {
    Create(ctx context.Context, input NoteCreateInput) error
    Update(ctx context.Context, input NoteUpdateInput) error
    ChangeStatus(ctx context.Context, input NoteStatusChangeInput) error
    Delete(ctx context.Context, id, ownerID string) error
}

// NoteCommandOutputPort はCommand（書く）専用のプレゼンターインターフェース。
type NoteCommandOutputPort interface {
    PresentNote(ctx context.Context, note *note.WithMeta) error
    PresentNoteDeleted(ctx context.Context) error
}

// NoteCommandRepository はCommand（書く）専用のリポジトリインターフェース。
type NoteCommandRepository interface {
    Get(ctx context.Context, id string) (*note.WithMeta, error)   // ← バリデーション用
    Create(ctx context.Context, n note.Note) (*note.Note, error)
    Update(ctx context.Context, n note.Note) (*note.Note, error)
    UpdateStatus(ctx context.Context, id string, status note.NoteStatus) (*note.Note, error)
    Delete(ctx context.Context, id string) error
    ReplaceSections(ctx context.Context, noteID string, sections []note.Section) error
}
```

```go
// 🆕 port/note_query_port.go

// NoteQueryInputPort はQuery（読む）専用のユースケースインターフェース。
type NoteQueryInputPort interface {
    List(ctx context.Context, filters note.Filters) error
    Get(ctx context.Context, id string) error
}

// NoteQueryOutputPort はQuery（読む）専用のプレゼンターインターフェース。
type NoteQueryOutputPort interface {
    PresentNoteList(ctx context.Context, notes []note.NoteReadModel) error
    PresentNote(ctx context.Context, note *note.NoteReadModel) error
}

// NoteQueryRepository はQuery（読む）専用のリポジトリインターフェース。
type NoteQueryRepository interface {
    List(ctx context.Context, filters note.Filters) ([]note.NoteReadModel, error)
    Get(ctx context.Context, id string) (*note.NoteReadModel, error)
    Upsert(ctx context.Context, model note.NoteReadModel) error  // ← 同期用
    Delete(ctx context.Context, id string) error                  // ← 同期用
}
```

**図で見ると：**

```
【Before】                      【After】

NoteInputPort                   NoteQueryInputPort（読む）
  .List()    ← 読む              .List()
  .Get()     ← 読む              .Get()
  .Create()  ← 書く
  .Update()  ← 書く             NoteCommandInputPort（書く）
  .Delete()  ← 書く              .Create()
                                  .Update()
                                  .ChangeStatus()
                                  .Delete()

NoteRepository                  NoteQueryRepository（読む）
  .List()    ← 読む              .List()
  .Get()     ← 読む              .Get()
  .Create()  ← 書く              .Upsert()    ← 同期用
  .Update()  ← 書く              .Delete()    ← 同期用
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

#### Query側（新規作成）

```go
// 🆕 usecase/note_query_interactor.go

// NoteQueryInteractor は読み取り専用のユースケース。
type NoteQueryInteractor struct {
    queryRepo port.NoteQueryRepository    // ← Read Model用のリポジトリ
    output    port.NoteQueryOutputPort
}

// List はRead Modelから一覧を取得する。JOINなし！
func (u *NoteQueryInteractor) List(ctx context.Context, filters note.Filters) error {
    notes, err := u.queryRepo.List(ctx, filters)  // ← キャッシュテーブルから取得
    if err != nil {
        return err
    }
    return u.output.PresentNoteList(ctx, notes)
}

// Get はRead Modelから1件取得する。JOINなし！
func (u *NoteQueryInteractor) Get(ctx context.Context, id string) error {
    n, err := u.queryRepo.Get(ctx, id)  // ← キャッシュテーブルから取得
    if err != nil {
        return err
    }
    return u.output.PresentNote(ctx, n)
}
```

**めっちゃシンプル。バリデーションもドメインロジックもない。ただ取って返すだけ。**

#### Command側（新規作成）

```go
// 🆕 usecase/note_command_interactor.go

type NoteCommandInteractor struct {
    commandRepo port.NoteCommandRepository   // ← Write用
    queryRepo   port.NoteQueryRepository     // ← Read Model同期用
    templates   port.TemplateRepository
    tx          port.TxManager
    output      port.NoteCommandOutputPort
}

// Create はノートを作成し、Read Modelも同期する。
func (u *NoteCommandInteractor) Create(ctx context.Context, input port.NoteCreateInput) error {
    // ... バリデーション＆書き込み ...

    err = u.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
        // 1. Write DBに書く
        nn, err := u.commandRepo.Create(txCtx, newNote)
        // ...sectionsも作成...

        // 2. 🆕 Read Modelも更新する（同じトランザクション内！）
        readModel := buildReadModel(nn, tpl, sections)
        return u.queryRepo.Upsert(txCtx, readModel)
    })
    // ...
}
```

**ポイント：書くときに同じトランザクションでRead Modelも更新する。だから整合性が保たれる。**

---

### 4️⃣ アダプター層 - Read Model用リポジトリ

#### キャッシュテーブル（マイグレーション）

```sql
-- 🆕 新しいマイグレーション
CREATE TABLE note_read_models (
    id UUID PRIMARY KEY,                -- notes.id と同じ
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    template_name TEXT NOT NULL,
    owner_id UUID NOT NULL,
    owner_first_name TEXT NOT NULL,
    owner_last_name TEXT NOT NULL,
    owner_thumbnail TEXT,
    sections_json JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- 検索用インデックス
CREATE INDEX idx_note_read_models_status ON note_read_models(status);
CREATE INDEX idx_note_read_models_owner ON note_read_models(owner_id);
CREATE INDEX idx_note_read_models_updated ON note_read_models(updated_at DESC);
CREATE INDEX idx_note_read_models_title ON note_read_models USING gin(title gin_trgm_ops);
```

#### Query用リポジトリ（新規作成）

```go
// 🆕 adapter/gateway/db/sqlc/note_query_repository.go

type NoteQueryRepository struct {
    pool    *pgxpool.Pool
    queries *generated.Queries
}

// List はnote_read_modelsテーブルから一覧を取得する。JOINなし。
func (r *NoteQueryRepository) List(ctx context.Context, filters note.Filters) ([]note.NoteReadModel, error) {
    // SELECT * FROM note_read_models WHERE ... ORDER BY updated_at DESC
    // → JOINなし！1テーブルから取るだけ！
}

// Upsert はRead Modelを挿入or更新する（Command側から呼ばれる）。
func (r *NoteQueryRepository) Upsert(ctx context.Context, model note.NoteReadModel) error {
    // INSERT INTO note_read_models (...) VALUES (...)
    // ON CONFLICT (id) DO UPDATE SET ...
}
```

---

### 5️⃣ コントローラー層 - QueryとCommandの使い分け

```go
// adapter/http/controller/note_controller.go（変更）

type NoteController struct {
    // Command用（書く）
    commandInputFactory func(...) port.NoteCommandInputPort
    commandOutputFactory func() *presenter.NotePresenter

    // 🆕 Query用（読む）
    queryInputFactory   func(...) port.NoteQueryInputPort
    queryOutputFactory  func() *presenter.NoteQueryPresenter

    // 共通
    commandRepoFactory  func() port.NoteCommandRepository
    queryRepoFactory    func() port.NoteQueryRepository
    tplRepoFactory      func() port.TemplateRepository
    txFactory           func() port.TxManager
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

---

### 6️⃣ ファクトリー＆初期化 - 配線を追加

```go
// driver/factory/repository_factory.go（変更）

// 🆕 追加
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

```go
// driver/initializer/api/initializer.go（変更）

// 🆕 Command/Query両方のFactoryを作成
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

2. 🎮 Controller: List()
   → queryInput, p := c.newQueryIO()
   → queryInput.List(ctx, filters)

3. 📋 NoteQueryInteractor: List()
   → readRepo.List(ctx, filters)
   → JOINなし！SELECT * FROM note_read_models WHERE ...

4. 💾 note_read_models テーブル
   → 一発で取得！

5. 🎨 NoteQueryPresenter: PresentNoteList()
   → Read Model → APIレスポンスに変換

6. 📱 画面: ノート一覧が表示される（速い！）
```

### ✏️ 書く場合（POST /notes）

```
1. 📱 画面: POST /notes {title: "新しいノート", ...}

2. 🎮 Controller: Create()
   → commandInput, p := c.newCommandIO()
   → commandInput.Create(ctx, input)

3. 📋 NoteCommandInteractor: Create()
   → バリデーション実行
   → tx.WithinTransaction(ctx, func(txCtx) {

4.    💾 Write DB:
      → commandRepo.Create(txCtx, note)        // notesテーブルに書く
      → commandRepo.ReplaceSections(txCtx, ...) // sectionsテーブルに書く

5.    💾 Read Model同期:                          // 🆕
      → queryRepo.Upsert(txCtx, model)          // note_read_modelsも更新

      }) // トランザクション終了
         // → 両方成功 or 両方ロールバック

6. 🎨 NotePresenter: PresentNote()
   → ドメインモデル → APIレスポンスに変換

7. 📱 画面: 作成されたノートが表示される
```

---

## 🧱 変更しないところ（超重要）

CQRSにしても **変えなくていい場所がたくさんある**。
これがClean Architectureの強み。

```
┌─────────────────────────────────────────────────┐
│  ✅ 変更なし                                     │
│                                                  │
│  domain/note/entity.go      ← Note, Section     │
│  domain/note/logic.go       ← バリデーション     │
│  domain/note/aggregate.go   ← 集約操作           │
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
│  → ドメイン層はまったく触らない！                 │
│  → 他のエンティティ（Account, Template）も       │
│    変更なし！                                    │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│  ✏️ 変更あり（小さい変更）                       │
│                                                  │
│  adapter/http/controller/   ← Query/Command分岐  │
│    note_controller.go                            │
│  driver/factory/            ← Factory追加        │
│  driver/initializer/        ← 配線追加           │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│  🆕 新規作成                                     │
│                                                  │
│  domain/note/read_model.go        ← Read Modelの型│
│  port/note_command_port.go        ← Command系ポート│
│  port/note_query_port.go          ← Query系ポート │
│  usecase/note_command_             ← Command用     │
│    interactor.go                    UseCase       │
│  usecase/note_query_              ← Query用       │
│    interactor.go                    UseCase       │
│  adapter/gateway/db/sqlc/         ← Command用     │
│    note_command_repository.go       リポジトリ     │
│  adapter/gateway/db/sqlc/         ← Query用       │
│    note_query_repository.go         リポジトリ     │
│  adapter/http/presenter/          ← Command用     │
│    note_command_presenter.go        プレゼンター   │
│  adapter/http/presenter/          ← Query用       │
│    note_query_presenter.go          プレゼンター   │
│  migrations/XXXXXX_add_           ← キャッシュ     │
│    note_read_models.sql             テーブル       │
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

## 📊 変更量の見積もり

| 種類 | ファイル数 | 変更の大きさ |
|-----|----------|------------|
| **新規作成** | 9〜10ファイル | Read Model型、Command/Queryポート、Command/Query UseCase、Command/Query Repo、Command/Query Presenter、マイグレーション |
| **既存変更（小）** | 2〜3ファイル | Factory追加、Initializer修正、NoteController（分岐追加） |
| **変更なし** | 大半 | ドメイン層、他のエンティティ、gRPC、設定、DB接続 |

**→ 全体の8割は変更なし。CQRSは「追加」が中心で「破壊」は少ない。**

---

## 🎯 まとめ

| 質問 | 答え |
|-----|------|
| 何を分ける？ | InputPort / UseCase / Repository を Query（読む）と Command（書く）に分離 |
| ドメイン層は変わる？ | ほぼ変わらない。Read Model用の型を1つ追加するだけ |
| 同期はどうやる？ | Command側で書くときに、同じトランザクション内でRead Modelも更新 |
| Controller は？ | 読むときはQuery用、書くときはCommand用のUseCaseを呼び分ける |
| 変更量は？ | 新規9〜10ファイル + 既存2〜3ファイルの小変更。大半は変更なし |
| Clean Architectureとの関係は？ | ポート（インターフェース）があるから、Read Modelの実装を自由に差し替えられる |

---

> 📘 **次のステップ**: 実際にコードを書いていきましょう！
