# Log Retention Configuration

## Tổng quan

Hệ thống đã được cập nhật để hỗ trợ **log retention** với thời gian lưu trữ tùy chỉnh. Bạn có thể thiết lập logs chỉ tồn tại trong 7 ngày, 14 ngày, 30 ngày, 90 ngày, 365 ngày hoặc bất kỳ thời gian nào bạn muốn.

## Cấu hình Log Retention

### 1. Các tham số cấu hình

```yaml
logger:
  # Log retention settings
  max_age_days: 7      # Số ngày giữ logs (1, 7, 14, 30, 90, 365, etc.)
  max_size_mb: 10      # Kích thước tối đa mỗi file log (MB)
  max_backups: 5       # Số lượng backup files tối đa
  compress: true       # Nén các file log đã rotate
  log_dir: "logs"      # Thư mục chứa logs (tùy chọn)
```

### 2. Ví dụ cấu hình cho các môi trường khác nhau

#### Development Environment (1 ngày)
```yaml
logger:
  max_age_days: 1
  max_size_mb: 5
  max_backups: 3
  compress: false
```

#### Production Environment (30 ngày)
```yaml
logger:
  max_age_days: 30
  max_size_mb: 100
  max_backups: 20
  compress: true
```

#### High-Volume Environment (7 ngày)
```yaml
logger:
  max_age_days: 7
  max_size_mb: 500
  max_backups: 50
  compress: true
```

## Cách hoạt động

### 1. **Time-based Rotation**
- Logs sẽ được tự động xóa sau `max_age_days` ngày
- Ví dụ: `max_age_days: 7` → logs sẽ bị xóa sau 7 ngày

### 2. **Size-based Rotation**
- Khi file log đạt `max_size_mb` MB, nó sẽ được rotate
- File cũ sẽ được đổi tên thành `log_YYYYMMDD_HHMMSS.log`

### 3. **Backup Management**
- Chỉ giữ lại `max_backups` file gần nhất
- Các file cũ hơn sẽ bị xóa tự động

### 4. **Compression**
- Khi `compress: true`, các file log đã rotate sẽ được nén
- Tiết kiệm disk space

## Sử dụng

### 1. Cấu hình trong file config

Tạo file `config/config_qa.yml`:
```yaml
# ... other configs ...

logger:
  mode: "development"
  level: "info"
  encoding: "console"
  
  # Log retention settings
  max_age_days: 7      # Keep logs for 7 days
  max_size_mb: 10      # Max 10MB per file
  max_backups: 5       # Keep 5 backup files
  compress: true       # Compress rotated logs
  log_dir: "logs"      # Log directory
```

### 2. Chạy ứng dụng

```bash
go run main.go
```

### 3. Kiểm tra logs

```bash
# Xem thư mục logs
ls -la logs/

# Xem kích thước logs
du -sh logs/

# Xem file logs mới nhất
tail -f logs/log_*.log
```

## Quản lý Log Files

### 1. Sử dụng Log Manager Tool

```bash
# Chạy log manager để xem thống kê
go run tools/log_manager.go
```

### 2. Cleanup thủ công

```bash
# Xóa logs cũ hơn 7 ngày
find logs/ -name "*.log" -mtime +7 -delete

# Xóa logs cũ hơn 30 ngày
find logs/ -name "*.log" -mtime +30 -delete
```

## Các tùy chọn thời gian phổ biến

| Thời gian | max_age_days | Mô tả |
|-----------|--------------|-------|
| 1 ngày | 1 | Development, testing |
| 1 tuần | 7 | Short-term monitoring |
| 2 tuần | 14 | Medium-term monitoring |
| 1 tháng | 30 | Production standard |
| 3 tháng | 90 | Long-term monitoring |
| 1 năm | 365 | Compliance, audit |

## Lợi ích

### 1. **Tiết kiệm disk space**
- Tự động xóa logs cũ
- Nén logs để tiết kiệm dung lượng

### 2. **Quản lý hiệu quả**
- Không cần manual cleanup
- Tự động rotate logs

### 3. **Linh hoạt**
- Cấu hình theo từng môi trường
- Dễ dàng thay đổi retention policy

### 4. **Performance**
- Giảm I/O khi đọc logs
- Tối ưu disk usage

## Troubleshooting

### 1. Logs không được xóa
- Kiểm tra `max_age_days` trong config
- Đảm bảo ứng dụng có quyền xóa files

### 2. Disk space vẫn đầy
- Giảm `max_size_mb`
- Giảm `max_backups`
- Bật `compress: true`

### 3. Logs bị mất quá sớm
- Tăng `max_age_days`
- Tăng `max_backups`
- Kiểm tra timezone settings

## Monitoring

### 1. Kiểm tra log directory size
```bash
du -sh logs/
```

### 2. Đếm số lượng log files
```bash
ls -1 logs/*.log | wc -l
```

### 3. Xem logs mới nhất
```bash
ls -la logs/ | tail -5
```

## Kết luận

Với tính năng log retention này, bạn có thể:
- **Tự động quản lý** logs theo thời gian
- **Tiết kiệm disk space** với compression
- **Linh hoạt cấu hình** theo từng môi trường
- **Dễ dàng monitoring** và troubleshooting

Chỉ cần cấu hình `max_age_days` trong file config là hệ thống sẽ tự động xóa logs cũ theo thời gian bạn mong muốn! 