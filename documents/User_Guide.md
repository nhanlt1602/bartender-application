# User Guide: Print RFID & RFID Mapping UID Flow

---

## 1. Print RFID Flow (Luồng In RFID)

### 1.1. Mục đích (Purpose)
In tem RFID cho sản phẩm từ màn hình chi tiết lệnh sản xuất (Manufacture Order Detail).

### 1.2. Quy trình thực hiện (Step-by-step Guide)

**Bước 1:** Nhấn nút **"In"** trên màn hình **Manufacture Order Detail**
- Người dùng chọn lệnh sản xuất cần in tem RFID và nhấn nút "In".

**Bước 2:** Hệ thống backend xử lý yêu cầu
- Backend nhận yêu cầu và xử lý tại API `print_manf_order`.
- Trong quá trình này, hệ thống sẽ:
  1. **Kiểm tra dữ liệu (Validate):** Xác thực thông tin lệnh sản xuất, sản phẩm, số lượng, v.v.
  2. **Gọi dịch vụ RFID bên WMS:** Lấy thông tin RFID cần in từ hệ thống quản lý kho (WMS).
  3. **Đăng ký cổng từ:** Đăng ký thông tin RFID với hệ thống kiểm soát cổng từ (nếu có).
  4. **Gửi dữ liệu xuống Application Consumer:** Dữ liệu in được gửi xuống Application Consumer (Kafka Consumer) dưới server để thực hiện in thực tế qua Bartender.

**Bước 3:** In tem RFID
- Ứng dụng Bartender nhận dữ liệu và thực hiện in tem RFID cho từng sản phẩm.

### 1.3. Lưu ý (Notes)
- Đảm bảo máy in Bartender đã được kết nối và cấu hình đúng.
- Kiểm tra lại số lượng tem in và thông tin sản phẩm trước khi xác nhận in.

### 1.4. Hình minh họa (Illustrations)

#### 1.4.1. Print RFID Flow Diagram
![Print RFID Flow](documents/Print_Rfid_flow.png)

#### 1.4.2. Print RFID Sequence Diagram
![Print RFID Sequence](documents/SequenceDiagramRfidPrint.png)

---

## 2. RFID Mapping UID Flow (Luồng Mapping RFID UID)

### 2.1. Mục đích (Purpose)
Mapping (gắn) UID thực tế của RFID vào hệ thống WMS sau khi sản phẩm đã hoàn thành.

### 2.2. Quy trình thực hiện (Step-by-step Guide)

**Bước 1:** Hoàn thành công đoạn cuối cùng ở **Work Order Product**
- Sau khi sản phẩm đã hoàn thành tất cả các công đoạn sản xuất, chuyển sang bước mapping RFID.

**Bước 2:** Đưa sản phẩm thành phẩm đã dán tem RFID vào thiết bị đọc RFID
- Sản phẩm đã được dán tem RFID (in từ bước trước) sẽ được đưa qua thiết bị đọc RFID.

**Bước 3:** Tiến hành mapping trên hệ thống WMS
- Hệ thống sẽ đọc UID thực tế từ tem RFID.
- Gửi thông tin này lên backend để thực hiện mapping giữa mã sản phẩm, mã lệnh sản xuất và UID RFID thực tế vào hệ thống WMS.
- Trạng thái mapping sẽ được cập nhật, đảm bảo sản phẩm đã được gắn đúng UID RFID.

### 2.3. Lưu ý (Notes)
- Đảm bảo sản phẩm đã dán đúng tem RFID trước khi mapping.
- Kiểm tra lại thông tin mapping trên hệ thống WMS sau khi hoàn thành.

### 2.4. Hình minh họa (Illustrations)

#### 2.4.1. RFID Mapping UID Flow Diagram
![RFID Mapping UID Flow](documents/Mapping_Rfid_Uid_flow.png)

#### 2.4.2. RFID Mapping UID Sequence Diagram
![RFID Mapping UID Sequence](documents/SequenceDiagramMappingRfidUid.png)

---

## 3. Tổng quan hệ thống (System Overview)

### 3.1. Các thành phần chính (Main Components)
- **Manufacture Order Detail Screen:** Màn hình chi tiết lệnh sản xuất
- **Backend API:** Xử lý logic nghiệp vụ
- **WMS System:** Hệ thống quản lý kho
- **Kafka Consumer:** Xử lý message queue
- **Bartender Application:** Ứng dụng in tem
- **RFID Reader Device:** Thiết bị đọc RFID

### 3.2. Luồng dữ liệu (Data Flow)
1. **Print Flow:** UI → Backend → WMS → Kafka → Bartender → Printer
2. **Mapping Flow:** RFID Device → Backend → WMS → Database Update

### 3.3. Trạng thái RFID (RFID Status)
- **Created:** RFID đã được tạo
- **Printed:** RFID đã được in
- **Mapped:** RFID đã được mapping với UID thực tế

---

## 4. Troubleshooting (Xử lý sự cố)

### 4.1. Lỗi thường gặp khi in RFID
- **Máy in không phản hồi:** Kiểm tra kết nối và cấu hình Bartender
- **Dữ liệu không đúng:** Kiểm tra thông tin lệnh sản xuất
- **Lỗi kết nối WMS:** Kiểm tra kết nối mạng và API WMS

### 4.2. Lỗi thường gặp khi mapping RFID
- **Thiết bị không đọc được RFID:** Kiểm tra tem RFID và thiết bị đọc
- **Mapping thất bại:** Kiểm tra thông tin sản phẩm và lệnh sản xuất
- **Lỗi cập nhật WMS:** Kiểm tra quyền truy cập và kết nối database

---

## 5. Liên hệ hỗ trợ (Support Contact)
Nếu gặp vấn đề trong quá trình sử dụng, vui lòng liên hệ:
- **Email:** support@hasaki.com
- **Hotline:** 1900-xxxx
- **Documentation:** [Link tài liệu kỹ thuật] 