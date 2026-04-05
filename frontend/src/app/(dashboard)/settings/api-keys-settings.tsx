"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { apiGet, apiPut, apiPatch, apiDelete } from "@/lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import type { UserLLMKey, ListResponse } from "@/types/api";

const PROVIDERS = [
  { id: "anthropic", name: "Anthropic", placeholder: "sk-ant-..." },
  { id: "openai", name: "OpenAI", placeholder: "sk-..." },
] as const;

export function ApiKeysSettings() {
  const [keys, setKeys] = useState<UserLLMKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [addingProvider, setAddingProvider] = useState<string | null>(null);
  const [apiKeyInput, setApiKeyInput] = useState("");
  const [modelInput, setModelInput] = useState("");
  const [saving, setSaving] = useState(false);

  const fetchKeys = async () => {
    try {
      const res = await apiGet<ListResponse<UserLLMKey>>("llm-keys");
      setKeys(res.data || []);
    } catch (err) {
      if (err instanceof Error && err.message.includes("not configured")) {
        setKeys([]);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchKeys();
  }, []);

  const handleSave = async (providerId: string) => {
    if (!apiKeyInput.trim()) {
      toast.error("API key is required");
      return;
    }
    setSaving(true);
    try {
      await apiPut<UserLLMKey>(`llm-keys/${providerId}`, {
        api_key: apiKeyInput,
        model: modelInput || "",
      });
      toast.success("API key saved");
      setAddingProvider(null);
      setApiKeyInput("");
      setModelInput("");
      await fetchKeys();
    } catch {
      toast.error("Failed to save API key");
    } finally {
      setSaving(false);
    }
  };

  const handleToggle = async (providerId: string, active: boolean) => {
    try {
      await apiPatch<UserLLMKey>(`llm-keys/${providerId}`, {
        is_active: active,
      });
      toast.success(active ? "Key enabled" : "Key disabled");
      await fetchKeys();
    } catch {
      toast.error("Failed to update key");
    }
  };

  const handleDelete = async (providerId: string) => {
    try {
      await apiDelete(`llm-keys/${providerId}`);
      toast.success("API key removed");
      await fetchKeys();
    } catch {
      toast.error("Failed to remove key");
    }
  };

  const getKey = (providerId: string) =>
    keys.find((k) => k.Provider === providerId);

  return (
    <Card>
      <CardHeader>
        <CardTitle>API Keys (BYOK)</CardTitle>
        <CardDescription>
          Use your own API keys for email processing and digest generation.
          Your keys are encrypted at rest and never shared.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? (
          <p className="text-sm text-muted-foreground">Loading...</p>
        ) : (
          PROVIDERS.map((provider) => {
            const existing = getKey(provider.id);
            const isAdding = addingProvider === provider.id;

            return (
              <div
                key={provider.id}
                className="flex flex-col gap-3 rounded-lg border p-4"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{provider.name}</span>
                    {existing && (
                      <Badge variant={existing.IsActive ? "default" : "secondary"}>
                        {existing.IsActive ? "Active" : "Inactive"}
                      </Badge>
                    )}
                  </div>
                  {existing ? (
                    <div className="flex items-center gap-3">
                      <Switch
                        checked={existing.IsActive}
                        onCheckedChange={(checked) =>
                          handleToggle(provider.id, checked)
                        }
                      />
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setAddingProvider(provider.id)}
                      >
                        Replace
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-destructive hover:text-destructive"
                        onClick={() => handleDelete(provider.id)}
                      >
                        Remove
                      </Button>
                    </div>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setAddingProvider(provider.id);
                        setApiKeyInput("");
                        setModelInput("");
                      }}
                    >
                      Add Key
                    </Button>
                  )}
                </div>

                {existing && !isAdding && (
                  <p className="text-sm text-muted-foreground">
                    Key: {existing.KeyHint}
                    {existing.Model && (
                      <span className="ml-2">Model: {existing.Model}</span>
                    )}
                  </p>
                )}

                {isAdding && (
                  <div className="space-y-3 pt-2">
                    <div className="space-y-1">
                      <Label htmlFor={`key-${provider.id}`}>API Key</Label>
                      <Input
                        id={`key-${provider.id}`}
                        type="password"
                        placeholder={provider.placeholder}
                        value={apiKeyInput}
                        onChange={(e) => setApiKeyInput(e.target.value)}
                      />
                    </div>
                    <div className="space-y-1">
                      <Label htmlFor={`model-${provider.id}`}>
                        Model Override{" "}
                        <span className="text-muted-foreground font-normal">
                          (optional)
                        </span>
                      </Label>
                      <Input
                        id={`model-${provider.id}`}
                        placeholder="Leave blank for default"
                        value={modelInput}
                        onChange={(e) => setModelInput(e.target.value)}
                      />
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        onClick={() => handleSave(provider.id)}
                        disabled={saving}
                      >
                        {saving ? "Saving..." : "Save"}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setAddingProvider(null);
                          setApiKeyInput("");
                          setModelInput("");
                        }}
                      >
                        Cancel
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            );
          })
        )}

        <p className="text-xs text-muted-foreground">
          When set, your key is used instead of the platform key. Processing
          errors are logged but won&apos;t be retried with the platform key.
        </p>
      </CardContent>
    </Card>
  );
}
