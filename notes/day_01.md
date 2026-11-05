# Ngày 1 — Môi trường phát triển và chương trình Go đầu tiên

> Những kiến thức mà tôi học được trong hành trình 30 ngày xây []()

- Dùng mise để quản lý phiên bản của các tools sử dụng trong dự án(phương án thay thế devcontainer).

Trả lời một số câu hỏi cuối bài:
1. go.mod và go.sum khác nhau thế nào? Vì sao phải commit cả hai?
- go.mod để khai báo module path, go version, các dependency(`direct` + `indirect`) kèm phiên bản.
- go.sum chứa hash(h1:, SHA-256) của nội dung từng module@version cần để build dự án.
 Đảm bảo code tải về của từng module phải đúng là code đã dùng trước đó(tránh tình trạng module đã bị sửa).
- Vì sao phải commit cả hai:
  - go.mod giúp mọi người/CI dùng cùng phiên bản.
  - go.sum giúp mọi người/CI nhận cùng nội dung, sai thì báo checksum mismatch.
  - Thiếu go.sum thì `go build` lỗi `missing go.sum entry`, vì Go mặc định không
    tự sửa go.mod/go.sum khi build (-mod=readonly).

2. Dòng `// indirect` nghĩa là gì? Khi nào nó tự mất?
- Module có trong go.mod nhưng KHÔNG package nào của mình import trực tiếp.
  Thường gặp khi:
  + nó là dependency của dependency (vd BurntSushi/toml do cleanenv kéo vào);
  + mình `go get` thư viện nhưng chưa import trong code.
- Nó không tự mất. Phải import trong code rồi chạy `go mod tidy` thì comment
  mới mất (go build không sửa go.mod). Không còn ai cần module đó thì
  `go mod tidy` xóa luôn cả dòng.

3. Khi chạy `go mod tidy` mà chưa có file nào import `cleanenv` thì chuyện gì xảy ra? 
- tidy xóa dòng require của cleanenv trong go.mod, và các hash của nó trong go.sum
  (kể cả các dependency chỉ cleanenv kéo vào).
- Vì tidy quét import của mọi file .go (gồm cả _test.go) để biết cần module nào:
  không ai import thì xóa, import mà go.mod chưa có thì thêm.
- Thứ tự đúng: viết `import` trước, rồi `go mod tidy` (tidy tự tải về).

4. `internal/` khác `pkg/` thế nào? Ai bắt buộc tuân theo quy tắc của `internal`?
- internal/: quy tắc do Go thực thi. Package trong X/internal/... chỉ được import
  bởi code nằm trong X/... Import từ ngoài sẽ lỗi "use of internal package not allowed".
- pkg/: chỉ là quy ước (golang-standards/project-layout, không chính thức),
  báo hiệu "code dùng chung, bên ngoài import được". Go không quan tâm cái tên này.
- Người bắt buộc tuân theo là công cụ Go (go build/test/vet), lúc biên dịch.
  Không phải linter hay review, nên mọi module khác import code mình đều bị chặn. 

5. Struct embedding có phải là kế thừa không? Promoted field hoạt động ra sao?
**Không phải kế thừa.** Embedding là **composition** (ghép struct này vào struct khác). Go chỉ cung cấp thêm cú pháp viết tắt để truy cập field và method của struct được nhúng. Cú pháp viết tắt đó gọi là **promotion**.

## 1. Embedding thực chất là gì

Lấy luôn code của bạn:

```go
type Config struct {
	App  `yaml:"app"`
	HTTP `yaml:"http"`
	Log  `yaml:"logger"`
}
```

Đây là một field **không đặt tên**. Go tự lấy tên kiểu làm tên field, nên đoạn trên tương đương với:

```go
type Config struct {
	App  App  `yaml:"app"`
	HTTP HTTP `yaml:"http"`
	Log  Log  `yaml:"logger"`
}
```

Vì vậy cả `cfg.App.Name` lẫn `cfg.Name` đều hợp lệ. Bạn đã thấy điều này khi lint: staticcheck gợi ý bỏ `.App` (`QF1008: could remove embedded field "App" from selector`).

## 2. Promoted field hoạt động ra sao

Khi bạn viết `cfg.Name`, compiler tìm `Name` theo **độ sâu**:

1. **Độ sâu 0:** `Config` có field nào tên `Name` không? Không có.
2. **Độ sâu 1:** các struct được nhúng trực tiếp là `App`, `HTTP`, `Log`. `App` có `Name`, nên compiler dịch `cfg.Name` thành `cfg.App.Name`.
3. Chưa thấy thì xuống tiếp độ sâu 2, 3...

Đây là cú pháp viết tắt **lúc biên dịch**. Lúc chạy không có tra cứu nào, nên không tốn chi phí.

Method cũng được promote theo đúng cách đó:

```go
func (a App) Banner() string { return a.Name + " " + a.Version }

cfg.Banner()   // compiler dịch thành cfg.App.Banner()
```

## 3. Vì sao không phải kế thừa: 4 khác biệt quan trọng

**a) Không có quan hệ "là một" (is-a).** `Config` không phải là một `App`:

```go
func printApp(a App) { fmt.Println(a.Name) }

printApp(cfg)       // ❌ compile error: cannot use cfg (type Config) as type App
printApp(cfg.App)   // ✅ phải truyền rõ phần được nhúng
```

Trong Java/C#, `class Config extends App` thì truyền `cfg` vào chỗ cần `App` là được. Go thì không cho.

**b) Receiver vẫn là struct được nhúng.** Method được promote chỉ "nhìn thấy" `App`, không biết gì về `Config` bao ngoài:

```go
func (a App) Banner() string { return a.Name } // a là App, không truy cập được a.Port
```

Không có `super`, cũng không có `this` trỏ tới object bên ngoài.

**c) Không có override hay đa hình.** Nếu `Config` tự định nghĩa method trùng tên, method đó **che** method của `App` (shadowing). Nhưng code bên trong `App` vẫn luôn gọi bản của `App`:

```go
type Base struct{}
func (Base) Name() string  { return "base" }
func (b Base) Hello() string { return "hello " + b.Name() }

type Child struct{ Base }
func (Child) Name() string { return "child" }

c := Child{}
c.Name()   // "child"       ← method của Child che method của Base
c.Hello()  // "hello base"  ← KHÔNG phải "hello child": Hello chạy trên Base nên gọi Base.Name
```

Trong Java, `Hello()` sẽ gọi `Name()` đã override và in ra `"hello child"`. Đây là khác biệt cốt lõi giữa embedding và kế thừa.

**d) Tên trùng ở cùng độ sâu thì báo lỗi khi dùng.** Đây là đáp án cho câu hỏi ở khối 4.1:

```go
type App  struct{ Name string }
type HTTP struct{ Name string }
type Config struct{ App; HTTP }

var cfg Config
cfg.Name      // ❌ compile error: ambiguous selector cfg.Name
cfg.App.Name  // ✅ phải chỉ rõ
```

Khai báo struct thì **không lỗi**, chỉ lỗi khi bạn **dùng** `cfg.Name`. Còn nếu `Config` tự có field `Name` (độ sâu 0) thì nó thắng, không còn nhập nhằng.

## 4. Embedding và interface: chỗ nó thực sự mạnh

Method được promote **được tính** khi kiểm tra struct có thỏa interface hay không:

```go
type Stringer interface{ String() string }

func (a App) String() string { return a.Name }

var s fmt.Stringer = cfg   // ✅ Config có String() (promote từ App) nên thỏa Stringer
```

Đa hình trong Go đến từ **interface**, không đến từ embedding. Embedding chỉ giúp "mượn" method để thỏa interface mà không phải viết lại.

Bạn sẽ gặp mẫu này nhiều trong dự án:
- **Ngày 4:** gRPC server nhúng `UnimplementedProductServiceServer`, nhờ đó tự có sẵn các method chưa viết, nên struct vẫn thỏa interface của gRPC.
- **Ngày 6:** `pkg/config` gom các struct config dùng chung, rồi nhúng vào `Config` của từng service.

## Tóm tắt để ghi vào notes

| | Kế thừa (Java, C#) | Embedding (Go) |
|---|---|---|
| Quan hệ | is-a: `Config` **là** `App` | has-a: `Config` **có** `App` |
| Truyền vào chỗ cần kiểu cha | Được | Không, phải truyền `cfg.App` |
| Override method | Có, và gọi được bản override từ lớp cha | Chỉ che tên ở bên ngoài, code bên trong không biết |
| `super` / `this` trỏ ra ngoài | Có | Không |
| Đa hình | Qua lớp cha | Qua **interface** |
| Cơ chế | Lúc chạy (vtable) | Viết tắt lúc biên dịch |

Câu nên nhớ: *"Embedding trong Go là composition kèm cú pháp viết tắt (promotion). Nó không phải kế thừa: không có is-a, không có override thật, đa hình đến từ interface."*

Để nhớ lâu hơn, hãy gõ thử đoạn `Base`/`Child` ở trên vào một file tạm, đoán trước rồi chạy để xem `c.Hello()` in ra gì.

6. Thứ tự ưu tiên của cleanenv: env, YAML hay default?
- env > YAML > env-default.
- Cách làm: đọc YAML trước, env (theo tag `env:"..."`) ghi đè, field vẫn là
  zero value thì mới dùng env-default. Sau đó field env-required mà vẫn zero thì lỗi.
- Bẫy: YAML ghi zero value (false, 0, "") thì vẫn bị env-default ghi đè.
- Bẫy: không có tag `env` thì biến môi trường không có tác dụng.
- Khác với thứ tự ưu tiên của ĐƯỜNG DẪN config trong main.go: flag > CONFIG_PATH > default.

7. %w khác %v trong fmt.Errorf thế nào?

Điểm cốt lõi: **chuỗi in ra giống hệt nhau. Khác biệt là `%w` giữ lại lỗi gốc để code kiểm tra được, còn `%v` chỉ giữ lại phần chữ.**

### So sánh

```go
_, err := os.Open("missing.yml")   // err là *fs.PathError, bọc fs.ErrNotExist

e1 := fmt.Errorf("read config: %v", err)
e2 := fmt.Errorf("read config: %w", err)

fmt.Println(e1)  // read config: open missing.yml: no such file or directory
fmt.Println(e2)  // read config: open missing.yml: no such file or directory  ← giống nhau

errors.Is(e1, fs.ErrNotExist)  // false: chỉ còn chuỗi, chuỗi lỗi đã bị cắt
errors.Is(e2, fs.ErrNotExist)  // true:  e2 có Unwrap() trả về err
```

| | `%v` | `%w` |
|---|---|---|
| Message | giống nhau | giống nhau |
| Lỗi trả về có `Unwrap()` | không | có |
| `errors.Is` / `errors.As` thấy lỗi gốc | không | có |
| Dùng được ở đâu | mọi hàm `fmt` | **chỉ trong `fmt.Errorf`** |

### Khi nào dùng cái nào

- **`%w` (mặc định nên dùng):** khi người gọi có thể cần phân biệt loại lỗi. Ví dụ: không tìm thấy file config thì dùng default, `sql.ErrNoRows` thì trả về HTTP 404.
- **`%v` (cố ý dùng):** khi lỗi gốc là chi tiết cài đặt mà bạn không muốn người gọi phụ thuộc vào. Wrap bằng `%w` thì lỗi gốc **trở thành một phần API của bạn**. Ví dụ: sau này đổi từ Postgres sang MySQL, code bên ngoài đang `errors.Is(err, pgx.ErrNoRows)` sẽ hỏng.

### Chi tiết thêm

- Từ Go 1.20, một lời gọi được dùng **nhiều `%w`**: `fmt.Errorf("%w; %w", err1, err2)`. `errors.Is` thấy được cả hai lỗi.
- Truyền cho `%w` một giá trị không phải error thì kết quả in ra `%!w(...)`. `go vet` bắt được lỗi này.
- Linter `errorlint` trong `.golangci.yml` của bạn **sẽ báo** khi dùng `%v` với một error trong `fmt.Errorf`. Nếu cố ý dùng `%v` thì thêm `//nolint:errorlint // lý do`.
- So sánh lỗi bằng `errors.Is(err, X)`, không dùng `err == X`, vì `==` không nhìn xuyên qua các lớp wrap.

**Bản gợi ý để dán vào note:**

```markdown
7. %w khác %v trong fmt.Errorf thế nào?
- Message in ra giống nhau.
- %w: lỗi trả về có Unwrap(), nên errors.Is/errors.As thấy được lỗi gốc.
  %v: chỉ còn chuỗi, chuỗi lỗi bị cắt.
- %w chỉ dùng được trong fmt.Errorf; từ Go 1.20 cho phép nhiều %w.
- Mặc định dùng %w. Dùng %v khi không muốn lỗi gốc (chi tiết cài đặt) thành một phần API.
- So sánh bằng errors.Is(err, X), không dùng ==.
```

**Thử ngay trong project:** `NewConfig` của bạn đang wrap bằng `%w`. Trong `main.go`, tạm thêm đoạn sau:

```go
if errors.Is(err, fs.ErrNotExist) {
    log.Fatal("không tìm thấy file config: ", *configPath)
}
```

Chạy `go run ./cmd/product -config nope.yml` và xem nhánh này có chạy không. Sau đó đổi `%w` thành `%v` trong config.go, chạy lại, rồi chạy thêm `make lint` để xem errorlint báo gì.

8. Vì sao Makefile cần .PHONY?
- Make mặc định coi target là tên FILE. Nếu có file hoặc thư mục trùng tên
  (vd docs/, test/, build/) thì make báo "is up to date" và không chạy recipe,
  cũng không báo lỗi.
- .PHONY khai báo target là lệnh, không phải file, nên lần nào gọi cũng chạy.
- Phụ: make bỏ qua tìm implicit rule (nhanh hơn), và đọc vào biết ngay target
  nào là lệnh.
- Mẹo: đặt .PHONY ngay trên từng target cho dễ bảo trì.

----------------------------

## Bug mount volume với Postgres 18

**Bối cảnh:** dùng image `postgres:alpine`. Tag này không ghim phiên bản, nên âm thầm lên **Postgres 18**. Từ bản 18, image đổi chỗ lưu dữ liệu:

| | Postgres ≤ 17 | Postgres 18+ |
|---|---|---|
| `PGDATA` (dữ liệu thật) | `/var/lib/postgresql/data` | `/var/lib/postgresql/18/docker` |
| `VOLUME` khai báo trong image | `/var/lib/postgresql/data` | `/var/lib/postgresql` |
| Nên mount vào | `/var/lib/postgresql/data` | **`/var/lib/postgresql`** |

### Lần 1: mount theo kiểu cũ, dữ liệu âm thầm bị mất

```yaml
volumes:
  - ./postgres_data:/var/lib/postgresql/data
```

- Postgres 18 không ghi vào `.../data` mà ghi vào `.../18/docker`.
- Thư mục `/var/lib/postgresql` không được mount, nên Docker tạo một **anonymous volume** cho nó (vì image có khai báo `VOLUME`).
- Kết quả: container vẫn chạy bình thường, **không báo lỗi gì**, nhưng dữ liệu nằm trong anonymous volume chứ không nằm ở `./postgres_data`. Sau `down` rồi `up`, container mới nhận một anonymous volume mới, coi như mất hết dữ liệu.
- Phát hiện bằng: `docker inspect <container> --format '{{json .Mounts}}'`

### Lần 2: mount thẳng vào `PGDATA`, container thoát ngay

```yaml
volumes:
  - postgres_data:/var/lib/postgresql/18/docker
```

- Named volume `postgres_data` đã có sẵn dữ liệu cũ, với cấu trúc của thư mục **cha** (`18/docker/...` và symlink `data -> .`).
- Gắn volume đó vào `PGDATA`, Postgres thấy thư mục không trống nhưng lại không phải một database hợp lệ.
- Log báo: `initdb: error: directory "/var/lib/postgresql/18/docker" exists but is not empty` → `Exited (1)`.
- Phát hiện bằng: `docker compose logs postgres` và `docker run --rm -v <volume>:/v alpine ls -la /v`

### Cách sửa

```yaml
services:
  postgres:
    image: postgres:18-alpine                 # ghim phiên bản
    volumes:
      - postgres_data:/var/lib/postgresql     # mount thư mục cha
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
volumes:
  postgres_data:
```

Sau đó chạy `docker compose down -v` để xóa volume hỏng, rồi `make compose-up`, rồi `docker compose ps` để thấy `(healthy)`.

Vì sao mount thư mục cha: đường dẫn `PGDATA` có số phiên bản. Mount ở thư mục cha thì các phiên bản 18, 19... nằm chung một volume, nhờ vậy `pg_upgrade` nâng cấp được dữ liệu.

### Bài học

1. **Ghim phiên bản image** (`postgres:18-alpine`), không dùng tag trôi (`alpine`, `latest`). Cùng lý do với việc ghim phiên bản trong `mise.toml`.
2. Đổi phiên bản lớn (major) thì **đọc release notes của image**, vì cách bố trí thư mục có thể đã thay đổi.
3. "Container chạy được" không có nghĩa là "dữ liệu được lưu đúng chỗ". Hãy kiểm tra mount thật bằng `docker inspect`.
4. Debug container bị thoát theo thứ tự: `ps -a` → `logs` → `config` → xem bên trong volume.