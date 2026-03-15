"use client";

import { useState, useEffect } from "react";
import { toast } from "sonner";
import { apiGet, apiPost, apiPatch, apiDelete } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import type { Label } from "@/types/api";
import { Plus, Pencil, Trash2, Tag, X, Check } from "lucide-react";

const LABEL_COLORS = [
  { value: "blue", bg: "bg-blue-500" },
  { value: "green", bg: "bg-green-500" },
  { value: "red", bg: "bg-red-500" },
  { value: "purple", bg: "bg-purple-500" },
  { value: "orange", bg: "bg-orange-500" },
  { value: "yellow", bg: "bg-yellow-500" },
  { value: "pink", bg: "bg-pink-500" },
  { value: "cyan", bg: "bg-cyan-500" },
  { value: "", bg: "bg-gray-400" },
];

export function LabelSettings() {
  const [labels, setLabels] = useState<Label[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");
  const [newColor, setNewColor] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [editColor, setEditColor] = useState("");
  const [saving, setSaving] = useState(false);

  const fetchLabels = async () => {
    try {
      const data = await apiGet<Label[]>("labels");
      setLabels(data || []);
    } catch {
      // keep existing
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLabels();
  }, []);

  const handleCreate = async () => {
    if (!newName.trim()) return;
    setSaving(true);
    try {
      const updated = await apiPost<Label[]>("labels", {
        Name: newName.trim(),
        Color: newColor,
      });
      setLabels(updated);
      setNewName("");
      setNewColor("");
      setCreating(false);
      toast.success("Label created");
    } catch {
      toast.error("Failed to create label");
    } finally {
      setSaving(false);
    }
  };

  const handleUpdate = async (id: string) => {
    if (!editName.trim()) return;
    setSaving(true);
    try {
      const updated = await apiPatch<Label[]>(`labels/${id}`, {
        Name: editName.trim(),
        Color: editColor,
      });
      setLabels(updated);
      setEditingId(null);
      toast.success("Label updated");
    } catch {
      toast.error("Failed to update label");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await apiDelete(`labels/${id}`);
      setLabels((prev) => prev.filter((l) => l.ID !== id));
      toast.success("Label deleted");
    } catch {
      toast.error("Failed to delete label");
    }
  };

  const startEdit = (label: Label) => {
    setEditingId(label.ID);
    setEditName(label.Name);
    setEditColor(label.Color);
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Tag className="size-5" />
              Labels
            </CardTitle>
            <CardDescription>
              Organize your emails with custom labels
            </CardDescription>
          </div>
          {!creating && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCreating(true)}
            >
              <Plus className="size-4 mr-1.5" />
              New Label
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {creating && (
          <div className="flex items-center gap-2 mb-4 p-3 rounded-lg border bg-muted/30">
            <ColorPicker value={newColor} onChange={setNewColor} />
            <input
              autoFocus
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleCreate();
                if (e.key === "Escape") setCreating(false);
              }}
              placeholder="Label name"
              maxLength={50}
              className="flex-1 h-8 rounded-md border border-input bg-transparent px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
            <Button
              size="icon"
              variant="ghost"
              className="size-8"
              disabled={saving || !newName.trim()}
              onClick={handleCreate}
            >
              <Check className="size-4" />
            </Button>
            <Button
              size="icon"
              variant="ghost"
              className="size-8"
              onClick={() => { setCreating(false); setNewName(""); setNewColor(""); }}
            >
              <X className="size-4" />
            </Button>
          </div>
        )}

        {loading ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            Loading labels...
          </p>
        ) : labels.length === 0 && !creating ? (
          <p className="text-sm text-muted-foreground text-center py-4">
            No labels yet. Create one to start organizing your emails.
          </p>
        ) : (
          <div className="space-y-1">
            {labels.map((label) => (
              <div
                key={label.ID}
                className="flex items-center gap-2 rounded-md px-3 py-2 hover:bg-muted/50 transition-colors"
              >
                {editingId === label.ID ? (
                  <>
                    <ColorPicker value={editColor} onChange={setEditColor} />
                    <input
                      autoFocus
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleUpdate(label.ID);
                        if (e.key === "Escape") setEditingId(null);
                      }}
                      maxLength={50}
                      className="flex-1 h-8 rounded-md border border-input bg-transparent px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    />
                    <Button
                      size="icon"
                      variant="ghost"
                      className="size-8"
                      disabled={saving || !editName.trim()}
                      onClick={() => handleUpdate(label.ID)}
                    >
                      <Check className="size-4" />
                    </Button>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="size-8"
                      onClick={() => setEditingId(null)}
                    >
                      <X className="size-4" />
                    </Button>
                  </>
                ) : (
                  <>
                    <LabelDot color={label.Color} />
                    <span className="text-sm font-medium flex-1">
                      {label.Name}
                    </span>
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {label.EmailCount}
                    </span>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="size-7"
                      onClick={() => startEdit(label)}
                    >
                      <Pencil className="size-3.5" />
                    </Button>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="size-7 text-destructive hover:text-destructive"
                      onClick={() => handleDelete(label.ID)}
                    >
                      <Trash2 className="size-3.5" />
                    </Button>
                  </>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function LabelDot({ color, className }: { color: string; className?: string }) {
  const colorClass = LABEL_COLORS.find((c) => c.value === color)?.bg || "bg-gray-400";
  return (
    <span
      className={`inline-block size-3 rounded-full shrink-0 ${colorClass} ${className || ""}`}
    />
  );
}

export const LABEL_COLOR_MAP: Record<string, string> = Object.fromEntries(
  LABEL_COLORS.map((c) => [c.value, c.bg])
);

function ColorPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex items-center gap-1">
      {LABEL_COLORS.map((c) => (
        <button
          key={c.value || "none"}
          type="button"
          onClick={() => onChange(c.value)}
          className={`size-5 rounded-full transition-all ${c.bg} ${
            value === c.value
              ? "ring-2 ring-offset-2 ring-primary"
              : "hover:scale-110"
          }`}
          aria-label={c.value || "gray"}
        />
      ))}
    </div>
  );
}
