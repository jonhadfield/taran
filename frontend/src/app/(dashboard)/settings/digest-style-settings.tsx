"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface DigestStyleSettingsProps {
  digestStyle: string;
  prefLoading: boolean;
  prefSaving: boolean;
  onDigestStyleChange: (value: string) => void;
}

export function DigestStyleSettings({
  digestStyle,
  prefLoading,
  prefSaving,
  onDigestStyleChange,
}: DigestStyleSettingsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Digest Style</CardTitle>
        <CardDescription>
          Choose how detailed your digest summaries are
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex gap-2">
          <Button
            type="button"
            size="sm"
            variant={digestStyle === "detailed" ? "default" : "outline"}
            onClick={() => onDigestStyleChange("detailed")}
            disabled={prefLoading || prefSaving}
          >
            Detailed
          </Button>
          <Button
            type="button"
            size="sm"
            variant={digestStyle === "concise" ? "default" : "outline"}
            onClick={() => onDigestStyleChange("concise")}
            disabled={prefLoading || prefSaving}
          >
            Concise
          </Button>
        </div>
        <p className="text-sm text-muted-foreground">
          {digestStyle === "concise"
            ? "Shorter summaries with fewer highlights for a quick overview."
            : "Full summaries with detailed highlights and more context."}
        </p>
      </CardContent>
    </Card>
  );
}
