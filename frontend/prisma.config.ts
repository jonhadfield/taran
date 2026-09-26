import "dotenv/config";
import { defineConfig } from "prisma/config";

// Prisma 7 reads CLI configuration from here rather than from the schema's
// datasource block. The app itself connects through the pg driver adapter in
// src/lib/auth.ts; this only covers CLI commands such as `prisma db push`.
//
// Deliberately not prisma/config's env(): that throws when the variable is
// missing, and this file is loaded by every CLI invocation including
// `prisma generate`, which needs no connection at all. Generate runs in
// postinstall and in preview builds, where DATABASE_URL is absent — it is set
// only for production. Falling back to an empty string keeps generation
// working; the commands that do connect still fail loudly without a real URL.
export default defineConfig({
  schema: "prisma/schema.prisma",
  datasource: {
    url: process.env.DATABASE_URL ?? "",
  },
});
