
# khi nào mux := http.NewServeMux() trả về 404 khi nào trả về 405
Mux quyết định theo **hai bước, kiểm tra path trước rồi mới tới method**:

```
Request đến
   │
   ├─ Có pattern nào khớp PATH không?
   │     └─ Không → 404 Not Found      (method là gì cũng không quan trọng)
   │
   └─ Có. Trong các pattern khớp path, có cái nào khớp METHOD không?
         ├─ Có    → gọi handler đó (200...)
         └─ Không → 405 Method Not Allowed
                     + header Allow: liệt kê các method của những pattern khớp path
```

Nói gọn: **404 nghĩa là "không có đường này", 405 nghĩa là "có đường này nhưng không đi bằng phương thức đó".**

## Kết quả chạy thử

| Request | Kết quả | Lý do |
|---|---|---|
| `GET /v1/api/item-types` | 200 | Khớp cả path lẫn method |
| `POST /v1/api/item-types` | **405** `Allow: GET, HEAD` | Path khớp, method không khớp |
| `HEAD /v1/api/item-types` | 200 | Pattern `GET` **tự động khớp cả `HEAD`** |
| `GET /v1/api/item-typez` | 404 | Path sai |
| `POST /v1/api/item-typez` | **404**, không phải 405 | Path sai thì mux dừng luôn, không xét method |
| `GET /v1/api/item-types/` | 404 | Pattern không có `/` ở cuối chỉ khớp **đúng** path đó |
| `PUT /v1/api/items-by-types/LATTE` | 405 `Allow: DELETE, GET, HEAD` | `Allow` gom method từ **mọi** pattern khớp path |
| `GET /v1/api/items-by-types/` | 404 | `{itemTypes}` không khớp segment rỗng |
| `GET /v1/api/items-by-types/A/B` | 404 | `{x}` chỉ khớp **một** segment. Muốn khớp phần còn lại thì dùng `{x...}` |
| `PATCH /any-method` | 200 | Pattern không ghi method thì khớp **mọi** method, nên path này không bao giờ trả 405 |

## Ngoài 404 và 405, mux còn trả 307 (redirect)

| Request | Kết quả | Lý do |
|---|---|---|
| `GET /docs` (route đăng ký là `GET /docs/`) | **307** → `/docs/` | Pattern kết thúc bằng `/` khớp cả cây con. Thiếu `/` thì mux redirect sang bản có `/` |
| `GET /v1//api/item-types` | **307** → `/v1/api/item-types` | Mux tự chuẩn hóa path, xử lý `//`, `.` và `..` |

Từ Go 1.22 đến nay, mã redirect là 307. Một số bài viết cũ ghi 301, vì vậy mình chạy thử thay vì trả lời theo trí nhớ.



1. http.Handler và http.HandlerFunc là gì? Vì sao một kiểu hàm lại có method được?


http.Handler là interface có đúng một method ServeHTTP(ResponseWriter, *Request). http.HandlerFunc là kiểu được định nghĩa từ kiểu hàm func(ResponseWriter, *Request), có method ServeHTTP chỉ gọi lại f(w, r), nên nó thỏa mãn http.Handler.

Kiểu hàm có method được vì trong Go, mọi kiểu tự đặt tên trong cùng package đều gắn được method, bất kể kiểu gốc là int, struct hay hàm. Kiểu hàm chưa đặt tên thì không gắn được.

http.HandlerFunc(f) là phép chuyển kiểu (cùng kiểu gốc, không tốn gì lúc chạy). Nhờ nó, một hàm thường dùng được ở chỗ cần http.Handler (adapter pattern). mux.HandleFunc và middleware logRequests đều dựa vào cơ chế này.

2. Thứ tự gọi Header().Set, WriteHeader, Write. Gọi sai thứ tự thì sao?

Thứ tự đúng: w.Header().Set(...) → w.WriteHeader(status) → w.Write(body). Response HTTP gửi theo thứ tự status, header, body. WriteHeader chốt status và header, sau đó không sửa được nữa.

Gọi sai:

Header().Set sau WriteHeader: bị bỏ qua trong im lặng, không lỗi, không log.
WriteHeader sau Write: Write đầu tiên đã tự gửi 200. Lời gọi sau bị bỏ qua kèm log superfluous response.WriteHeader call, nên client nhận 200.
Write khi chưa set Content-Type: Go tự đoán từ nội dung, JSON thành text/plain; charset=utf-8

3. Pattern của ServeMux từ Go 1.22: method, {name}, {name...}, 405, quy tắc ưu tiên.

> Cú pháp pattern: `[METHOD ][HOST]/PATH`.
> - **Method:** `GET /x` chỉ khớp GET, và **tự khớp cả HEAD**. Pattern không ghi method thì khớp mọi method.
> - **`{name}`** khớp đúng **một** segment, không khớp segment rỗng. Đọc giá trị bằng `r.PathValue("name")`, và chỉ có giá trị khi request đi qua mux.
> - **`{name...}`** khớp **phần còn lại** của path (`/files/{rest...}` cho ra `rest="a/b/c.png"`). Chỉ được đặt ở cuối pattern.
> - **`{$}`** khớp đúng path kết thúc tại đó: `/{$}` chỉ khớp `/`. Pattern kết thúc bằng `/` (như `/` hay `/docs/`) thì khớp cả cây con.
>
> Cách mux chọn: **xét path trước, rồi mới xét method.** Không pattern nào khớp path thì trả **404**. Có pattern khớp path nhưng sai method thì trả **405**, kèm header `Allow` liệt kê các method hợp lệ (ví dụ `GET, HEAD`). Ngoài ra mux tự redirect **307** khi path thiếu `/` cuối (`/docs` → `/docs/`) hoặc chưa sạch (`//`, `..`).
>
> **Ưu tiên:** pattern cụ thể hơn thắng, không phụ thuộc thứ tự đăng ký. Path cố định thắng wildcard (`/items/LATTE` thắng `/items/{name}`); có method thắng không có method. Hai pattern xung đột mà không cái nào cụ thể hơn (`/a/{x}` và `/{y}/b`) thì **panic lúc đăng ký**.
>
> Nguồn: `go doc net/http.ServeMux`, blog *Routing Enhancements for Go 1.22*, hàm `findHandler` trong `net/http/server.go`.

4. Vì sao field viết thường không xuất hiện trong JSON? omitempty và json:"-" làm gì? Slice nil và slice rỗng khác nhau thế nào khi encode?

> - **Field viết thường** không được export. `encoding/json` dùng `reflect`, mà `reflect` không được đọc giá trị field chưa export từ package khác, nên field đó bị bỏ qua cả khi encode lẫn decode.
> - **`json:"-"`** luôn bỏ qua field. (`json:"-,"` thì ngược lại: đặt tên field là `-`.)
> - **`omitempty`** bỏ field nếu là `false`, `0`, `""`, con trỏ/interface `nil`, hoặc slice/map/array có `len == 0`. Nó **không** bỏ struct: `time.Time` zero vẫn ra `"0001-01-01T00:00:00Z"`. Go 1.24 có thêm **`omitzero`** cho trường hợp đó.
> - **Slice nil → `null`, slice rỗng → `[]`.** Có `omitempty` thì cả hai đều bị bỏ. API nên trả `[]` để client khỏi phải xử lý `null`, vì vậy khởi tạo `items := []T{}` thay vì `var items []T`.

5. Vì sao writeJSON nên Marshal trước rồi mới WriteHeader?

> `writeJSON` gọi `json.Marshal` trước để có `[]byte` trong tay **trước khi chốt status**: lỗi thì trả 500 (kèm log lỗi), ổn thì gọi `Header().Set` → `WriteHeader(status)` → `Write(b)`.
>
> Nếu `WriteHeader(status)` trước rồi mới encode thẳng vào `w`, encode lỗi thì không sửa được status nữa (lần `WriteHeader` sau bị bỏ qua, chỉ có log `superfluous`). Client nhận 200 với body hỏng, tưởng là thành công.
>
> Marshal lỗi khi gặp `NaN` hoặc `±Inf` (`float64`), kiểu `chan` hoặc `func`, hoặc `MarshalJSON` trả lỗi.

6. http.ErrServerClosed là gì? Vì sao phải kiểm tra bằng errors.Is?

**1. Đây là sentinel error.** Trong `net/http` nó được khai báo là một biến cố định:
```go
var ErrServerClosed = errors.New("http: Server closed")
```
`Serve` và `ListenAndServe` trả về lỗi này sau khi gọi **`Shutdown` hoặc `Close`** (cả hai, không chỉ `Shutdown`). Nó là **tín hiệu bình thường** cho biết server đã tắt theo yêu cầu, không phải lỗi thật. Vì vậy phải lọc nó ra, nếu không mỗi lần tắt êm app lại báo lỗi.

**2. Vì sao dùng `errors.Is` mà không dùng `==`.** Hiện tại `Serve` trả lỗi này ở dạng trần, nên `err != http.ErrServerClosed` vẫn chạy đúng. Nhưng chỉ cần có một lớp nào đó bọc lỗi lại bằng `fmt.Errorf("serve: %w", err)` là `==` sai ngay, vì giá trị lúc đó là lỗi bọc ngoài, không còn là biến gốc. `errors.Is` thì lần theo cả chuỗi `Unwrap()` (kể cả lỗi được gộp bằng `errors.Join`) nên vẫn tìm thấy. Linter `errorlint` trong `.golangci.yml` của bạn sẽ báo lỗi khi so sánh error bằng `==`.

**3. Liên hệ ngày 4:** gRPC không có sentinel này. Sau `GracefulStop` hoặc `Stop`, `grpc.Server.Serve` trả về **`nil`**, nên không cần lọc.

**Bản gợi ý để dán vào note:**
> `http.ErrServerClosed` là sentinel error (`var ErrServerClosed = errors.New("http: Server closed")`). `Serve`/`ListenAndServe` trả về nó sau khi gọi `Shutdown` hoặc `Close`. Đây là tín hiệu tắt bình thường, phải lọc ra: `if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) { ... }`.
>
> Dùng `errors.Is` vì nó lần theo chuỗi lỗi bọc bằng `%w` (và `errors.Join`). `==` chỉ đúng khi lỗi chưa bị bọc lớp nào. Linter `errorlint` bắt lỗi dùng `==`.
>
> So sánh: gRPC `Serve` trả `nil` sau `GracefulStop`, không có sentinel.

7. Shutdown làm gì với kết nối mới, kết nối rảnh, request đang chạy? Khác Close thế nào?

> `Shutdown(ctx)`:
> 1. Đóng listener, nên kết nối mới bị từ chối (`connection refused`). `Serve` trả về `ErrServerClosed` **ngay**.
> 2. Đóng các kết nối idle (keep-alive đang chờ request).
> 3. **Chờ** request đang chạy xong. Sau khi xong, kết nối đó cũng bị đóng, không giữ keep-alive. Tất cả xong thì trả `nil`; hết `ctx` thì trả `ctx.Err()` và **không** cắt request còn lại, nên phải tự gọi `Close()`.
>
> `Close()`: đóng listener và mọi kết nối ngay, kể cả kết nối đang có request (client nhận empty reply). Handler vẫn chạy tiếp, chỉ là ghi ra kết nối đã đóng.
>
> Vì `Serve` trả về ngay ở bước 1, chương trình phải chờ **`Shutdown` return**, không phải chờ `Serve`. `Shutdown` không chờ kết nối bị hijack (WebSocket); dùng `RegisterOnShutdown` cho trường hợp đó.

8. Kể 4 loại timeout của http.Server. Slowloris là gì?

Bạn mới gửi câu hỏi, chưa có câu trả lời. Câu này đã được thảo luận kỹ trong buổi hỏi về timeout, nên bản gợi ý dưới đây dựa trên chính các thí nghiệm hôm đó. Bạn nên đọc lại, rồi tự viết bằng lời của mình.

Nếu tự viết, hãy kiểm tra xem bản của bạn đã có đủ 4 ý này chưa:
1. Mỗi timeout bắt đầu đếm **từ lúc nào** và đếm **tới lúc nào**.
2. `WriteTimeout` **không dừng handler**.
3. Có **tiêu chí** để chọn giá trị, không chỉ đưa ra con số.
4. Slowloris là **gì**, và **vì sao** nó làm server chết.

**Bản gợi ý để dán vào note:**
> Vòng đời một kết nối và 4 timeout:
```
accept → đọc header → đọc body → handler chạy + ghi response → chờ request kế
         └────────┘ ReadHeaderTimeout
         └───────────────────┘ ReadTimeout
                   └───────────────────────────────────────┘ WriteTimeout (bắt đầu ngay sau header)
                                                               └────────────┘ IdleTimeout
```
> | Timeout | Giới hạn | Tiêu chí chọn | App mình |
> |---|---|---|---|
> | `ReadHeaderTimeout` | đọc xong header | Header nhỏ, client thật gửi trong vài ms. Chặn slowloris. Thiếu thì gosec báo G112 | 5s |
> | `ReadTimeout` | đọc xong header + body | Body lớn nhất ÷ băng thông chậm nhất chấp nhận được. Endpoint upload thì nới riêng bằng `http.NewResponseController(w).SetReadDeadline` | 5s |
> | `WriteTimeout` | tính từ lúc đọc xong header tới lúc ghi xong response (gồm cả thời gian handler chạy) | Lớn hơn thời gian xử lý p99.9. Giữ chuỗi: deadline trong handler < `WriteTimeout` ≤ `shutdownTimeout` < `stop_grace_period` | 5s |
> | `IdleTimeout` | khoảng chờ request kế trên kết nối keep-alive | Lớn hơn idle timeout của load balancer phía trước (ALB 60s), nếu không sẽ gặp 502 ngẫu nhiên | 120s |
>
> **Bẫy của `WriteTimeout`:** nó **không dừng handler**. Thí nghiệm: handler sleep 3s, `WriteTimeout` 2s. Handler vẫn chạy hết, `w.Write` trả `nil`, client chờ 3s rồi nhận `EOF`, không có status nào. Muốn trả 503 đàng hoàng thì đặt deadline bằng context trong handler, nhỏ hơn `WriteTimeout`.
>
> Bỏ trống `ReadHeaderTimeout` hoặc `IdleTimeout` thì Go dùng giá trị của `ReadTimeout`. Bỏ trống cả hai thì không có timeout nào, giống `http.ListenAndServe`.
>
> **Slowloris:** tấn công từ chối dịch vụ bằng cách mở rất nhiều kết nối, gửi header **cực chậm**, mỗi lúc vài byte, nhưng không bao giờ gửi xong. Mỗi kết nối giữ một goroutine và một file descriptor. Không có `ReadHeaderTimeout` thì server giữ chúng mãi, tới khi cạn file descriptor và không nhận được client thật. Kẻ tấn công chỉ cần rất ít băng thông. Thí nghiệm ngày 3: không có timeout thì sau 12s kết nối dở dang vẫn mở; có `ReadHeaderTimeout` 5s thì server đóng nó sau đúng 5.0s.

9. Vì sao net.Listen trước rồi mới srv.Serve(ln)? So sánh với cách của repo gốc.

> Gọi `net.Listen` đồng bộ trước, rồi mới chạy `srv.Serve(ln)` trong goroutine:
> 1. Bind lỗi (port bận, port < 1024 khi không phải root, host sai) thì `Run` trả lỗi ngay, app thoát với exit 1.
> 2. Log "listening" chỉ được in khi port đã thật sự bind xong.
> 3. `ln.Addr()` cho biết port thật khi dùng `:0` (cần cho test).
>
> Repo gốc gọi `ListenAndServe` **trong goroutine**, lỗi chỉ được log, `initHTTPServer` luôn trả `nil`. Port bận thì app vẫn sống, chờ SIGTERM, nhưng không phục vụ gì. Ngoài ra repo gốc không lọc `ErrServerClosed`, nên mỗi lần tắt êm đều log một lỗi.
