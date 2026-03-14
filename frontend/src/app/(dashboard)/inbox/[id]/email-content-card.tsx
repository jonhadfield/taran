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
            className="prose prose-sm max-w-none rounded-md bg-background p-4 text-foreground overflow-x-auto [&_img]:max-w-full [&_img]:h-auto [&_table]:max-w-full [&_table]:overflow-x-auto [&_pre]:overflow-x-auto"
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
          <pre className="whitespace-pre-wrap text-sm font-mono overflow-x-auto">
            {textBody}
          </pre>
        )}
      </CardContent>
    </Card>
  );
}
