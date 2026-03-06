"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { apiPost } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Loader2, Sparkles, Eye, X, ArrowRight } from "lucide-react";
import type { Digest, DigestPreview } from "@/types/api";

export function GenerateDigestButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<DigestPreview | null>(null);

  const handlePreview = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiPost<DigestPreview>("digests/preview", {});
      setPreview(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load preview");
    } finally {
      setLoading(false);
    }
  };

  const handleGenerate = async () => {
    setGenerating(true);
    setError(null);
    try {
      const digest = await apiPost<Digest>("digests/generate", {});
      router.push(`/digests/${digest.ID}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to generate digest");
    } finally {
      setGenerating(false);
    }
  };

  const handleCancel = () => {
    setPreview(null);
    setError(null);
  };

  if (preview) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg flex items-center gap-2">
              <Eye className="size-5" />
              Digest Preview
            </CardTitle>
            <Button variant="ghost" size="icon" className="size-8" onClick={handleCancel}>
              <X className="size-4" />
            </Button>
          </div>
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <span>
              {new Date(preview.PeriodStart).toLocaleDateString()} &ndash;{" "}
              {new Date(preview.PeriodEnd).toLocaleDateString()}
            </span>
            <Badge variant="secondary">{preview.EmailCount} emails</Badge>
            <Badge variant="outline">{preview.PeriodType}</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="divide-y rounded-lg border max-h-80 overflow-y-auto">
            {preview.Items.map((item) => (
              <div key={item.EmailID} className="p-3">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium truncate">{item.Subject}</p>
                    <p className="text-xs text-muted-foreground truncate">
                      {item.FromName || item.FromAddress}
                    </p>
                    {item.Summary && (
                      <p className="mt-1 text-xs text-muted-foreground line-clamp-2">
                        {item.Summary}
                      </p>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground whitespace-nowrap">
                    {new Date(item.ReceivedAt).toLocaleDateString()}
                  </span>
                </div>
              </div>
            ))}
          </div>

          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              These {preview.EmailCount} emails will be summarized by AI.
            </p>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={handleCancel} disabled={generating}>
                Cancel
              </Button>
              <Button size="sm" onClick={handleGenerate} disabled={generating}>
                {generating ? (
                  <Loader2 className="size-4 animate-spin mr-1.5" />
                ) : (
                  <ArrowRight className="size-4 mr-1.5" />
                )}
                Generate Digest
              </Button>
            </div>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Button onClick={handlePreview} disabled={loading} size="sm">
        {loading ? (
          <Loader2 className="size-4 animate-spin mr-1.5" />
        ) : (
          <Sparkles className="size-4 mr-1.5" />
        )}
        Generate Digest
      </Button>
      {error && <span className="text-sm text-destructive">{error}</span>}
    </div>
  );
}
