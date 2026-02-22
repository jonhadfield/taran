"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import sanitizeHtml from "sanitize-html";

interface EmailContentCardProps {
  htmlBody: string;
  textBody: string;
}

export function EmailContentCard({ htmlBody, textBody }: EmailContentCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Email Content</CardTitle>
      </CardHeader>
      <CardContent>
        {htmlBody ? (
          <div
            className="prose prose-sm max-w-none rounded-md bg-background p-4 text-foreground"
            dangerouslySetInnerHTML={{
              __html: sanitizeHtml(htmlBody, {
                allowedTags: sanitizeHtml.defaults.allowedTags.concat(["img"]),
                allowedAttributes: {
                  ...sanitizeHtml.defaults.allowedAttributes,
                  "*": ["style", "class"],
                },
              }),
            }}
          />
        ) : (
          <pre className="whitespace-pre-wrap text-sm font-mono">
            {textBody}
          </pre>
        )}
      </CardContent>
    </Card>
  );
}
