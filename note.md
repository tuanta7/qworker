# Scaling

- Job: Cronjob 
- Scheduler: Quản lý danh sách CronJob
- Message: Thông điệp Job gửi sang Queue để Worker thực hiện Task
- Task: Công việc mà Worker thực hiện

## Scheduler

- Chỉ dùng 1 scheduler để tránh trùng lặp Task
- Tránh gửi Message trùng với Message chưa được xử lý ở vòng lặp trước
- 

## Workers

- Cấu hình Asynq Concurrency là số lượng routine sẽ chạy
- Routine chạy xong thì được fetch task mới (tỉ lệ được cân bằng theo cấu hình Priority)
- Các task trong queue thì FIFO, khác queue thì không xác định
- Queue ưu tiên cao thì được làm trước (Strict) hoặc làm nhiều hơn (theo tỉ lệ)
- Các queue được lấy theo Round Robin
- Task failed bị chuyển qua archived, không tự động xóa
- Task thành công không set Retention tự xóa