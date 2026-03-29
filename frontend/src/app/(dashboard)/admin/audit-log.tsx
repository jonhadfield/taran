"use client";

import { usePolling } from "@/hooks/use-polling";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatDateTime } from "@/lib/utils";

interface AuditEntry {
  ID: string;
  UserEmail: string;
  Action: string;
  Target: string;
  Detail: string;
  IPAddress: string;
  CreatedAt: string;
}

interface AuditResponse {
  data: AuditEntry[];
  total: number;
}

export function AuditLog() {
  const { data } = usePolling<AuditResponse>(
    "admin/audit-log",
    { data: [], total: 0 },
    30_000,
  );

  const entries = data.data || [];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit Log</CardTitle>
      </CardHeader>
      <CardContent>
        {entries.length === 0 ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            No admin actions recorded yet.
          </p>
        ) : (
          <div className="divide-y max-h-96 overflow-y-auto">
            {entries.map((entry) => (
              <div key={entry.ID} className="py-2.5 flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-3 text-sm">
                <span className="text-xs text-muted-foreground tabular-nums whitespace-nowrap">
                  {formatDateTime(entry.CreatedAt)}
                </span>
                <span className="text-muted-foreground truncate max-w-[180px]">
                  {entry.UserEmail}
                </span>
                <span className="font-mono text-xs bg-muted px-1.5 py-0.5 rounded whitespace-nowrap">
                  {entry.Action}
                </span>
                {entry.Target && (
                  <span className="text-muted-foreground truncate">
                    {entry.Target}
                  </span>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
