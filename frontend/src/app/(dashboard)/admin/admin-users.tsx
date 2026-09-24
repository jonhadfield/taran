"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { usePolling } from "@/hooks/use-polling";
import { apiPatch } from "@/lib/api";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import type { AdminUser, ListResponse } from "@/types/api";
import { formatTokens, pluralize } from "@/lib/utils";
import { Pencil, Check, X, ChevronLeft, ChevronRight } from "lucide-react";

// Matches the server's default page size.
const PAGE_SIZE = 25;

export function AdminUsers() {
  const [page, setPage] = useState(0);
  // The path carries the page, so changing page refetches.
  const { data, refresh } = usePolling<ListResponse<AdminUser>>(
    `admin/users?limit=${PAGE_SIZE}&offset=${page * PAGE_SIZE}`,
    { data: [], total: 0 },
    60_000
  );
  const users = data.data ?? [];
  const total = data.total ?? 0;
  const firstOnPage = page * PAGE_SIZE + 1;
  const lastOnPage = page * PAGE_SIZE + users.length;
  const hasPrevious = page > 0;
  const hasNext = lastOnPage < total;

  // Users can be removed while you are on the last page; step back rather than
  // stranding the viewer on an empty one.
  useEffect(() => {
    if (total > 0 && page > 0 && page * PAGE_SIZE >= total) {
      setPage(Math.max(0, Math.ceil(total / PAGE_SIZE) - 1));
    }
  }, [total, page]);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [limitValue, setLimitValue] = useState("");
  const [saving, setSaving] = useState(false);

  const startEdit = (user: AdminUser) => {
    setEditingId(user.ID);
    setLimitValue(String(user.MonthlyTokenLimit));
  };

  const cancelEdit = () => {
    setEditingId(null);
    setLimitValue("");
  };

  const saveLimit = async (userID: string) => {
    const parsed = parseInt(limitValue, 10);
    if (isNaN(parsed) || parsed < 0) {
      toast.error("Enter a valid number (0 for default)");
      return;
    }
    setSaving(true);
    try {
      await apiPatch(`admin/users/${userID}/token-limit`, {
        MonthlyTokenLimit: parsed,
      });
      toast.success("Token limit updated");
      setEditingId(null);
      refresh();
    } catch {
      toast.error("Failed to update token limit");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">
          User token limits
          {total > 0 && (
            <span className="ml-2 text-sm font-normal text-muted-foreground">
              {total} {pluralize(total, "user")}
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="pb-2 font-medium">User</th>
                <th className="pb-2 font-medium text-right">Emails</th>
                <th className="pb-2 font-medium text-right">Monthly usage</th>
                <th className="pb-2 font-medium text-right">Limit</th>
                <th className="pb-2 font-medium w-10"></th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => {
                const pct =
                  user.MonthlyTokenLimit > 0
                    ? Math.min(
                        100,
                        (user.MonthlyTokensUsed / user.MonthlyTokenLimit) * 100
                      )
                    : 0;
                const isEditing = editingId === user.ID;

                return (
                  <tr key={user.ID} className="border-b last:border-0">
                    <td className="py-2.5">
                      <div className="font-medium truncate max-w-[200px]">
                        {user.Name || user.Email}
                      </div>
                      {user.Name && (
                        <div className="text-xs text-muted-foreground truncate max-w-[200px]">
                          {user.Email}
                        </div>
                      )}
                    </td>
                    <td className="py-2.5 text-right tabular-nums">
                      {user.EmailCount.toLocaleString()}
                    </td>
                    <td className="py-2.5 text-right">
                      <div className="tabular-nums">
                        {formatTokens(user.MonthlyTokensUsed)}
                      </div>
                      {user.MonthlyTokenLimit > 0 && (
                        <div className="w-16 ml-auto mt-0.5 h-1.5 rounded-full bg-muted overflow-hidden">
                          <div
                            className={`h-full rounded-full ${
                              pct >= 90
                                ? "bg-destructive"
                                : pct >= 70
                                ? "bg-warning"
                                : "bg-primary"
                            }`}
                            style={{ width: `${pct}%` }}
                          />
                        </div>
                      )}
                    </td>
                    <td className="py-2.5 text-right">
                      {isEditing ? (
                        <div className="flex items-center justify-end gap-1">
                          <Input
                            type="number"
                            min={0}
                            value={limitValue}
                            onChange={(e) => setLimitValue(e.target.value)}
                            className="h-7 w-24 text-sm text-right"
                            placeholder="0 = default"
                            onKeyDown={(e) => {
                              if (e.key === "Enter") saveLimit(user.ID);
                              if (e.key === "Escape") cancelEdit();
                            }}
                            autoFocus
                          />
                          <Button
                            size="icon"
                            variant="ghost"
                            className="size-7"
                            onClick={() => saveLimit(user.ID)}
                            disabled={saving}
                          >
                            <Check className="size-3.5" />
                          </Button>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="size-7"
                            onClick={cancelEdit}
                          >
                            <X className="size-3.5" />
                          </Button>
                        </div>
                      ) : (
                        <span className="tabular-nums">
                          {formatTokens(user.MonthlyTokenLimit)}
                        </span>
                      )}
                    </td>
                    <td className="py-2.5 text-right">
                      {!isEditing && (
                        <Button
                          size="icon"
                          variant="ghost"
                          className="size-7"
                          onClick={() => startEdit(user)}
                        >
                          <Pencil className="size-3.5" />
                        </Button>
                      )}
                    </td>
                  </tr>
                );
              })}
              {users.length === 0 && (
                <tr>
                  <td
                    colSpan={5}
                    className="py-6 text-center text-muted-foreground"
                  >
                    No users yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {total > PAGE_SIZE && (
          <div className="mt-4 flex items-center justify-between gap-4 text-sm">
            <p className="text-muted-foreground">
              Showing {firstOnPage}&ndash;{lastOnPage} of {total}
            </p>
            <div className="flex items-center gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={!hasPrevious}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setPage((p) => p + 1)}
                disabled={!hasNext}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
