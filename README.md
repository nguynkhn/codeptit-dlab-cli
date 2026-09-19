# CodePTIT DLab CLI

Bản cập nhật của [CodePTIT CLI](https://github.com/nguynkhn/codeptit-cli) được
viết bằng Go.

> Công cụ này không dính dáng gì tới DLab và không nhằm mục đích khai thác, lạm
dụng hoặc truy cập trái phép. Người dùng tự chịu trách nhiệm về việc sử dụng
công cụ và mọi hậu quả phát sinh từ việc sử dụng sai mục đích.

## Hướng dẫn sử dụng

Tải mã nguồn về và biên dịch (yêu cầu có **Go 1.21+** trở lên).

```sh
# Đăng nhập lần đầu
codeptit-dlab-cli login

# Xem danh sách các môn học
# Đầu ra ví dụ: 1. 14 - INT1306 — Cấu trúc dữ liệu và giải thuật
codeptit-dlab-cli course

# Chọn môn học ở chỉ số 1
codeptit-dlab-cli course 1

# Gửi lời giải trong tệp CHELLO.cpp lên
codeptit-dlab-cli submit CHELLO.cpp
```

Khi gửi lời giải lên thì mặc định sẽ ưu tiên lấy mã bài tập ở dòng comment đầu
tiên trong tệp bài làm, nếu không có thì sẽ lấy tên tệp bỏ đuôi mở rộng đi.

```c++
// CHELLO
#include <iostream>
int main() {
    std::cout << "Hello PTIT.\n";
}
```

```go
/* CHELLO */
package main
func main() {
    println("Hello PTIT.")
}
```
