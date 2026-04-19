ALTER TABLE employees
  ADD COLUMN IF NOT EXISTS work_mode VARCHAR(50) NOT NULL DEFAULT 'onsite',
  ADD COLUMN IF NOT EXISTS management_scope VARCHAR(50) NOT NULL DEFAULT 'individual_contributor';

CREATE TABLE IF NOT EXISTS teams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  name VARCHAR(150) NOT NULL,
  department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
  manager_employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  is_cross_department BOOLEAN NOT NULL DEFAULT FALSE,
  focus_area TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (organization_id, name)
);

CREATE INDEX IF NOT EXISTS idx_teams_department_id ON teams(department_id);
CREATE INDEX IF NOT EXISTS idx_teams_manager_employee_id ON teams(manager_employee_id);

CREATE TABLE IF NOT EXISTS team_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  role_name VARCHAR(100) NOT NULL DEFAULT 'Member',
  start_date DATE NOT NULL DEFAULT CURRENT_DATE,
  end_date DATE,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (team_id, employee_id)
);

CREATE INDEX IF NOT EXISTS idx_team_memberships_team_id ON team_memberships(team_id);
CREATE INDEX IF NOT EXISTS idx_team_memberships_employee_id ON team_memberships(employee_id);
