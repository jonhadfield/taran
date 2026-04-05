"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";

interface InboxDisplaySettingsProps {
  topicLimit: number;
  prefLoading: boolean;
  prefSaving: boolean;
  onTopicLimitChange: (value: number) => void;
}

export function InboxDisplaySettings({
  topicLimit,
  prefLoading,
  prefSaving,
  onTopicLimitChange,
}: InboxDisplaySettingsProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Inbox Display</CardTitle>
        <CardDescription>
          Customize how your inbox looks
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="topic-limit">Topic cloud size</Label>
          <p className="text-sm text-muted-foreground">
            Maximum number of topics shown in your inbox word cloud
          </p>
          <div className="flex items-center gap-3">
            <input
              id="topic-limit"
              type="range"
              min={5}
              max={50}
              step={5}
              value={topicLimit}
              onChange={(e) => onTopicLimitChange(Number(e.target.value))}
              disabled={prefLoading || prefSaving}
              className="w-full max-w-xs accent-primary"
            />
            <span className="text-sm font-medium w-8 text-right">{topicLimit}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
