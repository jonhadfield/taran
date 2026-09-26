import "dotenv/config";
import { defineConfig, env } from "prisma/config";

// Prisma 7 reads CLI configuration from here rather than from the schema's
// datasource block. The app itself connects through the pg driver adapter in
// src/lib/auth.ts; this only covers CLI commands such as `prisma db push`.
export default defineConfig({
  schema: "prisma/schema.prisma",
  datasource: {
    url: env("DATABASE_URL"),
  },
});
