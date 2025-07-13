# Kafka Consumer Auto-Start Setup

Hướng dẫn thiết lập tự động khởi động Kafka Consumer khi Windows restart.

## Các file script

### 1. `start.bat`
- Script chính để khởi động ứng dụng Kafka Consumer
- Tự động tạo log files với timestamp
- Tự động restart nếu ứng dụng crash
- Kiểm tra sự tồn tại của file executable

### 2. `setup-auto-start.bat`
- Script để đăng ký auto-start với Windows Task Scheduler
- **Phải chạy với quyền Administrator**
- Tạo scheduled task để tự động chạy khi Windows restart

### 3. `remove-auto-start.bat`
- Script để gỡ bỏ auto-start task
- **Phải chạy với quyền Administrator**
- Xóa scheduled task đã tạo trước đó

### 4. `check-auto-start.bat`
- Script để kiểm tra trạng thái auto-start
- Hiển thị thông tin chi tiết về scheduled task
- Kiểm tra các file và thư mục cần thiết

## Cách sử dụng

### Bước 1: Thiết lập auto-start
1. Đảm bảo file `kafka-consumer.exe` đã có trong thư mục `app-launch/`
2. Chuột phải vào file `setup-auto-start.bat`
3. Chọn "Run as administrator"
4. Đợi thông báo thành công

### Bước 2: Kiểm tra thiết lập
Có nhiều cách để kiểm tra:

#### Cách 1: Sử dụng script kiểm tra
```cmd
check-auto-start.bat
```

#### Cách 2: Kiểm tra bằng command line
```cmd
schtasks /query /tn "KafkaConsumerAutoStart"
```

#### Cách 3: Kiểm tra trong Task Scheduler GUI
1. Nhấn `Win + R`, gõ `taskschd.msc` và nhấn Enter
2. Tìm task có tên "KafkaConsumerAutoStart"
3. Kiểm tra trạng thái và cấu hình

#### Cách 4: Kiểm tra bằng PowerShell
```powershell
Get-ScheduledTask -TaskName "KafkaConsumerAutoStart"
```

### Bước 3: Test thủ công
Để test script start:
```cmd
start.bat
```

## Kiểm tra trạng thái Auto-Start

### Sử dụng script tự động
Chạy `check-auto-start.bat` để:
- ✅ Kiểm tra task có tồn tại không
- ✅ Hiển thị thông tin chi tiết về task
- ✅ Kiểm tra các file cần thiết
- ✅ Hiển thị log files gần đây

### Kiểm tra thủ công

#### 1. Task Scheduler (GUI)
```
Win + R → taskschd.msc → Task Scheduler Library → KafkaConsumerAutoStart
```

#### 2. Command Line
```cmd
# Kiểm tra task tồn tại
schtasks /query /tn "KafkaConsumerAutoStart"

# Xem thông tin chi tiết
schtasks /query /tn "KafkaConsumerAutoStart" /fo list

# Xem dạng bảng
schtasks /query /tn "KafkaConsumerAutoStart" /fo table
```

#### 3. PowerShell
```powershell
# Kiểm tra task
Get-ScheduledTask -TaskName "KafkaConsumerAutoStart"

# Xem thông tin chi tiết
Get-ScheduledTask -TaskName "KafkaConsumerAutoStart" | Get-ScheduledTaskInfo
```

### Các trạng thái có thể gặp

#### ✅ Task tồn tại và hoạt động
```
TaskName                                 Status
--------                                 ------
KafkaConsumerAutoStart                   Ready
```

#### ❌ Task không tồn tại
```
ERROR: The system cannot find the file specified.
```
→ Cần chạy `setup-auto-start.bat` với quyền Administrator

#### ⚠️ Task bị vô hiệu hóa
```
TaskName                                 Status
--------                                 ------
KafkaConsumerAutoStart                   Disabled
```
→ Chuột phải vào task → Enable

## Gỡ bỏ auto-start

Nếu muốn gỡ bỏ auto-start:
1. Chuột phải vào file `remove-auto-start.bat`
2. Chọn "Run as administrator"
3. Đợi thông báo thành công

## Log files

- Log files được tạo trong thư mục `logs/`
- Format tên file: `startup_YYYYMMDD_HHMMSS.log`
- Chứa thông tin về quá trình khởi động và lỗi (nếu có)

## Troubleshooting

### Lỗi "Access Denied"
- Đảm bảo chạy script với quyền Administrator
- Kiểm tra Windows Task Scheduler service đang chạy

### Lỗi "kafka-consumer.exe not found"
- Kiểm tra file `kafka-consumer.exe` có trong thư mục `app-launch/`
- Đảm bảo đường dẫn đúng

### Task không chạy khi restart
- Kiểm tra task trong Task Scheduler (taskschd.msc)
- Xem log files trong thư mục `logs/`
- Kiểm tra Windows Event Viewer

### Task tồn tại nhưng không hoạt động
- Kiểm tra task có bị Disabled không
- Kiểm tra trigger có đúng "At system startup" không
- Kiểm tra user account có quyền chạy task không

## Cấu trúc thư mục yêu cầu

```
kafka-consumer/
├── start.bat                    # Script khởi động chính
├── setup-auto-start.bat         # Script thiết lập auto-start
├── remove-auto-start.bat        # Script gỡ bỏ auto-start
├── check-auto-start.bat         # Script kiểm tra trạng thái
├── app-launch/
│   └── kafka-consumer.exe       # File executable chính
├── logs/                        # Thư mục chứa log files
└── config/                      # Thư mục cấu hình
```

## Lưu ý

- Script sẽ tự động restart ứng dụng nếu nó crash
- Log files sẽ được tạo với timestamp để dễ theo dõi
- Task được chạy với quyền SYSTEM để đảm bảo hoạt động ổn định
- Nếu thay đổi vị trí thư mục, cần chạy lại `setup-auto-start.bat`
- Luôn kiểm tra trạng thái sau khi thiết lập bằng `check-auto-start.bat` 