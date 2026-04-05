"use client";

import { useState } from "react";
import { toast } from "sonner";
import { usePolling } from "@/hooks/use-polling";
import { apiPost } from "@/lib/api";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { FailedEmail, ListResponse, PipelineHealth } from "@/types/api";
import { formatDateTime } from "@/lib/utils";
import {
  AlertTriangle,
  RefreshCw,
  RotateCcw,
  Database,
  ChevronDown,
  ChevronUp,
} from "lucide-react";

export function AdminFailedEmails() {
  const { data: failedRes, refresh } = usePolling<ListResponse<FailedEmail>>(
    "admin/emails/failed?limit=100",
    { data: [], total: 0 },
    30_000,
  );
  const { data: pipeline } = usePolling<PipelineHealth>(
    "admin/pipeline",
    { pending: 0, processing: 0, processed: 0, failed: 0, skipped: 0, payloads: 0 },
    30_000,
  );
  const [retrying, setRetrying] = useState<string | null>(null);
  const [batchRetrying, setBatchRetrying] = useState(false);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const emails = failedRes.data || [];
  const total = failedRes.total;

  const retryOne = async (id: string) => {
    setRetrying(id);
    try {
      await apiPost(`admin/emails/${id}/retry`, {});
      toast.success("Email queued for retry");
      refresh();
    } catch {
      toast.error("Failed to retry email");
    } finally {
      setRetrying(null);
    }
  };

  const retryAll = async () => {
    setBatchRetrying(true);
    try {
      const res = await apiPost<{ count: number }>("admin/emails/batch-retry", {
        all_failed: true,
      });
      toast.success(`${res.count} emails queued for retry`);
      refresh();
    } catch {
      toast.error("Failed to batch retry");
    } finally {
      setBatchRetrying(false);
    }
  };

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  return (
    <div className="space-y-4">
      {/* Pipeline health overview */}
      <div className="grid grid-cols-3 sm:grid-cols-6 gap-3">
        {[
          { label: "Failed", value: pipeline.failed, color: "text-red-600 dark:text-red-400" },
          { label: "Pending", value: pipeline.pending, color: "text-blue-600 dark:text-blue-400" },
          { label: "Processing", value: pipeline.processing, color: "text-yellow-600 dark:text-yellow-400" },
          { label: "Processed", value: pipeline.processed, color: "text-green-600 dark:text-green-400" },
          { label: "Skipped", value: pipeline.skipped, color: "text-muted-foreground" },
          { label: "Payloads", value: pipeline.payloads, color: "text-muted-foreground" },
        ].map((item) => (
          <div key={item.label} className="text-center">
            <div className={`text-lg font-bold tabular-nums ${item.color}`}>
              {item.value.toLocaleString()}
            </div>
            <div className="text-xs text-muted-foreground">{item.label}</div>
          </div>
        ))}
      </div>

      {/* Failed emails card */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <AlertTriangle className="size-5 text-red-500" />
            Failed Emails ({total})
          </CardTitle>
          {total > 0 && (
            <Button
              variant="outline"
              size="sm"
              disabled={batchRetrying}
              onClick={retryAll}
            >
              <RotateCcw className={`size-4 mr-1.5 ${batchRetrying ? "animate-spin" : ""}`} />
              Retry All
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {emails.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-6">
              No failed emails
            </p>
          ) : (
            <div className="space-y-2">
              {emails.map((email) => (
                <div
                  key={email.ID}
                  className="rounded-lg border p-3 space-y-2"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <p className="text-sm font-medium truncate">
                          {email.Subject || "(no subject)"}
                        </p>
                        <Badge
                          variant={email.Status === "failed" ? "destructive" : "secondary"}
                          className="text-xs shrink-0"
                        >
                          {email.Status}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground mt-0.5">
                        From: {email.FromName || email.FromAddress}
                        {email.UserEmail && (
                          <span className="ml-2">User: {email.UserEmail}</span>
                        )}
                      </p>
                      {email.StatusReason && (
                        <p className="text-xs text-red-600 dark:text-red-400 mt-1">
                          {email.StatusReason}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-8"
                        onClick={() => toggleExpand(email.ID)}
                        aria-label="Toggle details"
                      >
                        {expanded.has(email.ID) ? (
                          <ChevronUp className="size-4" />
                        ) : (
                          <ChevronDown className="size-4" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-8"
                        disabled={retrying === email.ID}
                        onClick={() => retryOne(email.ID)}
                        aria-label="Retry email"
                      >
                        <RefreshCw
                          className={`size-4 ${retrying === email.ID ? "animate-spin" : ""}`}
                        />
                      </Button>
                    </div>
                  </div>

                  {expanded.has(email.ID) && (
                    <div className="text-xs text-muted-foreground space-y-1 pt-2 border-t">
                      <p>ID: <span className="font-mono">{email.ID}</span></p>
                      <p>To: {email.ToAddress}</p>
                      <p>Received: {formatDateTime(email.ReceivedAt)}</p>
                      <p>Retries: {email.RetryCount}</p>
                      <p>Updated: {formatDateTime(email.UpdatedAt)}</p>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {pipeline.payloads > 0 && (
            <div className="flex items-center gap-2 mt-4 pt-4 border-t text-sm text-muted-foreground">
              <Database className="size-4" />
              {pipeline.payloads} stored webhook payload{pipeline.payloads !== 1 ? "s" : ""} available for replay
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
