
## Go test

Cờ -covermode quyết định cách Go ghi nhận độ bao phủ (coverage) khi chạy test, tức là với mỗi câu lệnh trong code, Go sẽ lưu thông tin gì. Có 3 chế độ:

1. set (mặc định)
Chỉ ghi nhận câu lệnh có được chạy hay không (có/không). Đây là chế độ nhanh nhất và đủ dùng nếu bạn chỉ quan tâm đến tỷ lệ phần trăm coverage.

2. count
Ghi nhận số lần mỗi câu lệnh được thực thi. Hữu ích khi bạn muốn biết đoạn code nào là “điểm nóng” (chạy rất nhiều lần) hoặc đoạn nào chỉ chạy đúng một lần. Tuy nhiên, bộ đếm này không an toàn khi có nhiều goroutine cùng chạy, nên số liệu có thể bị sai trong code đồng thời.

3. atomic
Giống count, nhưng dùng thao tác nguyên tử (sync/atomic) để tăng bộ đếm, nên chính xác khi có nhiều goroutine. Đổi lại, nó chậm hơn đáng kể. Khi bạn bật -race, Go sẽ tự động dùng atomic làm mặc định.
``` bash
# Chỉ xem phần trăm coverage
go test -covermode=set -coverprofile=cover.out ./...

# Xem số lần chạy của từng câu lệnh
go test -covermode=count -coverprofile=cover.out ./...

# Code có goroutine, cần đếm chính xác
go test -covermode=atomic -coverprofile=cover.out ./...
```


Với `count` hoặc `atomic`, báo cáo HTML sẽ tô màu đậm nhạt theo số lần chạy, giúp bạn thấy rõ đoạn nào chạy nhiều, đoạn nào ít. Với set thì chỉ có hai màu: xanh (đã chạy) và đỏ (chưa chạy).

Nếu mở file `cover.out`, dòng đầu sẽ cho biết chế độ (ví dụ `mode: count`), các dòng sau có dạng:

example.com/pkg/file.go:10.2,12.16 2 5

Nghĩa là khối code từ dòng 10 cột 2 đến dòng 12 cột 16, gồm 2 câu lệnh, đã chạy 5 lần (với set thì số cuối chỉ là 0 hoặc 1).

Tóm lại cách chọn: dùng set cho nhu cầu thông thường và CI, count khi muốn phân tích tần suất trong code tuần tự, và atomic khi code có goroutine hoặc khi chạy cùng -race.


## Docker build 


## Trả lời một số câu hỏi
1. Go biết một type thỏa mãn interface bằng cách nào? Vì sao var _ Interface = (*Logger)(nil) hữu ích?

- Type thỏa mãn interface khi có ĐỦ các method của interface, khớp chính xác chữ ký
  (tên, tham số, kiểu trả về). Ngầm định: không có từ khóa `implements`.
- Tính theo method set: method có receiver *T thì chỉ *T thỏa mãn, T thì không.
- Compiler kiểm tra lúc biên dịch, tại chỗ GÁN giá trị vào biến/tham số kiểu interface.
- `var _ Interface = (*Logger)(nil)`: gán một giá trị nil kiểu *Logger vào biến kiểu
  Interface, nên compiler buộc phải kiểm tra. Hữu ích vì:
  + lỗi báo ngay cạnh type, không phải ở chỗ dùng ở xa;
  + bắt lỗi kể cả khi chưa ai dùng type như interface;
  + làm tài liệu ("type này cài đặt Interface");
  + không tốn chi phí khi chạy (`_` vứt đi, nil không cấp phát).
  
2. Method receiver T khác *T thế nào khi xét chuyện thỏa mãn interface?
- Method set: T chỉ gồm các method có receiver T; *T gồm cả method có receiver T và *T.
- Method có receiver *T → chỉ *T thỏa mãn interface.
  Method có receiver T   → cả T và *T đều thỏa mãn.
  Trộn lẫn (có ít nhất một method receiver *T) → chỉ *T thỏa mãn.
- Lý do: interface giữ BẢN SAO của giá trị, và bản sao đó không addressable.
  Nếu cho gọi method receiver *T, method sẽ sửa bản sao chứ không sửa object gốc,
  nên Go chặn lúc biên dịch.
  Lỗi: "T does not implement I (method M has pointer receiver)".
- Khác với gọi trực tiếp trên biến: t.PtrMethod() vẫn được vì Go tự viết thành (&t).PtrMethod().
- Thực tế: phân vân thì dùng receiver *T cho tất cả method, rồi gán &T{} vào interface.

3. "Accept interfaces, return structs" nghĩa là gì?
- Nhận interface: hàm chỉ đòi đúng hành vi nó cần, nên người gọi linh hoạt,
  dễ test (truyền bản giả), ít phụ thuộc.
- Trả struct: người gọi dùng được mọi method/field. Cần thì tự gán vào interface.
  Trả interface thì giấu bớt khả năng, thêm method là phải sửa interface.
- Trong code của mình:
  + NewSlogLogger(w io.Writer, ...) nhận io.Writer → test truyền &bytes.Buffer{}.
  + heartbeat(ctx, log logger.Logger, ...) nhận interface Logger (và context.Context).
  + NewSlogLogger trả *SlogLogger (bản đầu trả Logger, đã sửa); NewConfig trả *Config.
- Ngoại lệ: luôn trả `error` (không trả *MyErr); factory chọn cài đặt lúc chạy;
  cài đặt unexported (vd sha256.New() trả hash.Hash).

4. Bẫy nil interface là gì?
- Interface gồm 2 phần (type, value). Chỉ bằng nil khi CẢ HAI đều nil.
- Gán con trỏ nil có kiểu vào interface → (type=*T, value=nil) → != nil.
- Ví dụ:
    func validate(name string) error {
        var verr *ValidationError      // nil
        if name == "" { verr = &ValidationError{"name"} }
        return verr                    // name hợp lệ vẫn trả về err != nil!
    }
- Cách tránh: hàm trả về `error` và `return nil` trực tiếp; không trả *MyErr.
- Compiler và go vet không bắt được; staticcheck SA4023 bắt được một số trường hợp.
- Áp dụng cho mọi interface, không riêng error.
5. Vì sao bỏ Fatal khỏi logger interface? os.Exit và defer có liên quan gì?
- Fatal = log + os.Exit(1). os.Exit kết thúc process NGAY: không chạy defer
  (ở mọi hàm, mọi goroutine), không flush buffer, không chờ goroutine.
  (Khác panic: panic vẫn chạy defer và recover được.)
- Bỏ Fatal vì:
  + logger chỉ nên ghi log; việc thoát hay không là quyết định của người gọi;
  + Fatal ở tầng sâu sẽ bỏ qua graceful shutdown (chờ worker, đóng kết nối, stop());
  + chỉ main được quyết định thoát, còn lại thì `return err` lên trên. slog cố ý không có Fatal.
- Gặp thực tế: gocritic exitAfterDefer báo log.Fatalf làm defer stop() không chạy.
  Cách sửa: main() { if err := run(); err != nil { ...; os.Exit(1) } },
  mọi defer đặt trong run().

6. signal.Notify và signal.NotifyContext khác nhau thế nào? Vì sao channel của Notify cần buffer?
- Notify(c, sigs...): mỗi tín hiệu được GỬI vào c, nên chỉ 1 người nhận. Biết chính
  xác tín hiệu nào. Nhận mọi tín hiệu liên tiếp. Hủy bằng signal.Stop(c).
- NotifyContext(parent, sigs...): tín hiệu đầu tiên HỦY ctx (đóng ctx.Done()), nên mọi
  goroutine nghe ctx đều biết; lan xuống cả cây context. Lý do xem qua context.Cause(ctx).
  Tín hiệu sau bị nuốt cho tới khi gọi stop() (trả về hành vi mặc định).
- Notify cần buffer: package signal gửi KHÔNG CHẶN (select + default). Channel không
  buffer mà lúc đó không ai đang nhận thì tín hiệu bị bỏ mất. Buffer 1 là đủ.
  go vet (sigchanyzer) bắt lỗi này.

7. ctx.Done() hoạt động thế nào để báo cho nhiều goroutine cùng lúc?
- ctx.Done() trả về channel chỉ-nhận (<-chan struct{}). Context tự ĐÓNG nó khi bị hủy
  (gọi cancel, hết deadline, có tín hiệu với NotifyContext). Code mình không close được.
- Đóng channel = broadcast: nhận từ channel đã đóng không bao giờ chặn, trả về ngay
  cho MỌI người nhận và mọi lần nhận. (Gửi giá trị thì chỉ 1 người nhận.)
- Goroutine không polling: chặn ở <-ctx.Done() hoặc select (ngủ, không tốn CPU),
  runtime đánh thức tất cả khi channel đóng.
- Lan theo cây: hủy cha thì mọi ctx con cũng Done (ngược lại thì không).
- Sau Done: ctx.Err() cho Canceled/DeadlineExceeded; context.Cause(ctx) cho lý do chi tiết.
- Background().Done() == nil → chặn mãi = không bao giờ bị hủy.


8. Vì sao COPY go.mod go.sum phải đặt trước COPY . . trong Dockerfile?

Cơ chế cache của Docker
1. Mỗi lệnh (COPY, RUN...) tạo ra một layer.
2. Khi build lại, với mỗi lệnh Docker hỏi: "lệnh này có giống hệt lần trước không?"
  - Với RUN: câu lệnh có giống không, và layer trước nó có giống không.
  - Với COPY: nội dung các file được chép (Docker tính checksum) có giống không.
3. Giống thì dùng lại layer cũ (CACHED). Khác thì build lại layer đó và TẤT CẢ layer phía sau, dù các layer sau không đổi gì.


- Mỗi lệnh = 1 layer. Build lại: lệnh và input (checksum file với COPY) giống lần trước
  thì dùng cache. Một layer đổi thì nó VÀ MỌI layer phía sau đều build lại.
- go.mod/go.sum hiếm khi đổi, code đổi liên tục. Nên:
  COPY go.mod go.sum → RUN go mod download (được cache) → COPY . . → go build.
- Sai thứ tự (COPY . . trước): sửa 1 dòng code cũng phải tải lại toàn bộ thư viện.
- Quy tắc chung: thứ ít thay đổi lên trên, thứ hay thay đổi xuống dưới.
- Liên quan: .dockerignore loại file không cần thiết để COPY . . không bị đổi vô ích;
  cache mount giữ module cache/build cache giữa các lần build.


9. Exit code 137 nghĩa là gì? Docker tắt container theo trình tự nào?
- Chết vì tín hiệu N → exit code 128 + N. 137 = 128 + 9 = SIGKILL
  (130 = SIGINT, 143 = SIGTERM không xử lý).
- SIGKILL không bắt được: chết ngay, không defer, không log.
- 137 có 2 nguyên nhân: OOM (docker inspect .State.OOMKilled = true),
  hoặc không tự thoát kịp khi bị stop.
- Trình tự docker stop: gửi SIGTERM (STOPSIGNAL) tới PID 1 → chờ grace period
  (mặc định 10s, stop_grace_period) → còn chạy thì SIGKILL → 137.
  `docker kill` thì SIGKILL ngay.
- Cần exec form CMD ["/app"] để app là PID 1 và nhận được tín hiệu.
- Quy tắc: shutdownTimeout của app (5s) < stop_grace_period (10s), để app tự kết thúc
  và ghi log trước khi bị kill. K8s tương tự (terminationGracePeriodSeconds = 30s).

