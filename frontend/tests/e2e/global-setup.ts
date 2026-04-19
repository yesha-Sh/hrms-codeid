import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

export default async function globalSetup() {
  const currentDir = path.dirname(fileURLToPath(import.meta.url));
  const backendDir = path.resolve(currentDir, "..", "..", "..", "backend");
  const env = {
    ...process.env,
    ADMIN_SEED_EMAIL: process.env.ADMIN_SEED_EMAIL ?? "admin@northstar.id",
    ADMIN_SEED_PASSWORD: process.env.ADMIN_SEED_PASSWORD ?? "ChangeMe123!",
  };

  execFileSync("go", ["run", "./cmd/migrate", "up"], {
    cwd: backendDir,
    env,
    stdio: "inherit",
  });
  execFileSync("go", ["run", "./cmd/seed-admin"], {
    cwd: backendDir,
    env,
    stdio: "inherit",
  });
  execFileSync("go", ["run", "./cmd/seed-demo"], {
    cwd: backendDir,
    env,
    stdio: "inherit",
  });
}
