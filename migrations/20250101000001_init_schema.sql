-- +goose Up
-- (ใส่โค้ด CREATE TABLE ที่ผมให้ไปในข้อความที่แล้วตรงนี้ทั้งหมด)
CREATE TYPE user_role AS ENUM ('patient', 'doctor', 'admin');
-- ... ฯลฯ ...
-- +goose Down
-- (สำหรับเวลาต้องการ Rollback หรือย้อนกลับ ให้ DROP ทิ้งให้หมด)
DROP TABLE IF EXISTS redemption_history;
DROP TABLE IF EXISTS rewards;
DROP TABLE IF EXISTS patient_progress;
DROP TABLE IF EXISTS stages;
DROP TABLE IF EXISTS animals;
DROP TABLE IF EXISTS animal_categories;
DROP TABLE IF EXISTS assessments;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS progress_status;
DROP TYPE IF EXISTS media_type_enum;
DROP TYPE IF EXISTS fear_level;
DROP TYPE IF EXISTS user_role;