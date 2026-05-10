"use client";

import { useState, useEffect } from "react";
import { apiGet, apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { WaitlistRequest, ListResponse } from "@/types/api";
import { CheckCircle2 } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import { toast } from "sonner";
import { formatShortDate } from "@/lib/utils";

export function WaitlistPanel() {
  const [requests, setRequests] = useState<WaitlistRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [approvingId, setApprovingId] = useState<string | null>(null);

  const fetchRequests = async () => {
    try {
      const res = await apiGet<ListResponse<WaitlistRequest>>(
        "admin/waitlist"
      );
      setRequests(res.data ?? []);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRequests();
  }, []);

  const handleApprove = async (id: string, email: string) => {
    setApprovingId(id);
    try {
      await apiPost(`admin/waitlist/${id}/approve`, {});
      setRequests((prev) => prev.filter((r) => r.ID !== id));
      toast.success(`${email} has been invited`);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to approve request"
      );
    } finally {
      setApprovingId(null);
    }
  };

  if (loading) return null;
  if (requests.length === 0) return null;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">
          Waitlist Requests ({requests.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {requests.map((req) => (
            <div
              key={req.ID}
              className="flex items-center justify-between"
            >
              <div className="min-w-0">
                <p className="text-sm font-medium truncate">{req.Email}</p>
                <p className="text-xs text-muted-foreground">
                  Requested on{" "}
                  {formatShortDate(req.CreatedAt)}
                </p>
              </div>
              <Button
                size="sm"
                onClick={() => handleApprove(req.ID, req.Email)}
                disabled={approvingId === req.ID}
              >
                {approvingId === req.ID ? (
                  <Spinner className="mr-1 size-3" />
                ) : (
                  <CheckCircle2 className="mr-1 size-3" />
                )}
                Approve
              </Button>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
