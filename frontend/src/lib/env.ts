/**
 * Reads a required environment variable, throwing at module-load time if it is
 * missing so misconfiguration surfaces at startup rather than as a confusing
 * runtime failure part-way through a request.
 */
export function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} environment variable is required`);
  }
  return value;
}
