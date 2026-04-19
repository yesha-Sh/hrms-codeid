DROP TABLE IF EXISTS team_memberships;
DROP TABLE IF EXISTS teams;

ALTER TABLE employees
  DROP COLUMN IF EXISTS management_scope,
  DROP COLUMN IF EXISTS work_mode;
