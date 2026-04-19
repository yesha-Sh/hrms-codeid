# Test Accounts

Use these seeded accounts to test the main flows after running:

```powershell
cd C:\laragon\www\HRMS\backend
go run ./cmd\migrate up
go run ./cmd\seed-admin
go run ./cmd\seed-demo
```

## Primary Accounts
- Admin: `admin@northstar.id` / `ChangeMe123!`
- Team manager: `fajar.maulana@codeid.co.id` / `Manager123!`
- Subdepartment manager: `salma.nuraini@codeid.co.id` / `Manager123!`
- Division manager: `dimas.kusuma@codeid.co.id` / `Manager123!`
- Employee: `nabila.putri@codeid.co.id` / `Employee123!`

## Suggested QA Order
1. Admin for employee, department, holiday, and audit supervision.
2. Team manager for team membership, self attendance, and leave request.
3. Subdepartment manager for approval flow.
4. Employee for self-service attendance, leave, and profile.
