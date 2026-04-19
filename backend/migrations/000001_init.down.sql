-- Drop leaf/dependent tables first
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS holidays;
DROP TABLE IF EXISTS attendances;
DROP TABLE IF EXISTS employee_role_assignments;
DROP TABLE IF EXISTS refresh_tokens;

-- Drop the deferred cross-referencing FK constraints before dropping base tables
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS fk_users_employee;
ALTER TABLE IF EXISTS departments DROP CONSTRAINT IF EXISTS fk_departments_manager;
ALTER TABLE IF EXISTS departments DROP CONSTRAINT IF EXISTS fk_departments_parent;
ALTER TABLE IF EXISTS employees DROP CONSTRAINT IF EXISTS fk_employees_user;
ALTER TABLE IF EXISTS employees DROP CONSTRAINT IF EXISTS fk_employees_manager;

-- Now drop the base tables in safe reverse-creation order
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS employment_statuses;
DROP TABLE IF EXISTS leave_types;
DROP TABLE IF EXISTS employee_types;
DROP TABLE IF EXISTS job_levels;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS roles;
