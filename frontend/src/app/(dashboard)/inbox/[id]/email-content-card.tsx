import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import sanitizeHtml from "sanitize-html";

const sanitizeConfig: sanitizeHtml.IOptions = {
  allowedTags: sanitizeHtml.defaults.allowedTags.concat(["img"]),
  allowedAttributes: {
    ...sanitizeHtml.defaults.allowedAttributes,
  },
  allowedStyles: {
    "*": {
      "color": [/^(#[0-9a-fA-F]{3,8}|rgb\(|rgba\(|[a-zA-Z]+)$/],
      "background-color": [/^(#[0-9a-fA-F]{3,8}|rgb\(|rgba\(|transparent|[a-zA-Z]+)$/],
      "font-size": [/^[\d.]+(px|em|rem|pt|%)$/],
      "font-weight": [/^(normal|bold|bolder|lighter|\d{3})$/],
      "font-style": [/^(normal|italic|oblique)$/],
      "text-align": [/^(left|center|right|justify)$/],
      "text-decoration": [/^(none|underline|overline|line-through)$/],
      "line-height": [/^[\d.]+(px|em|rem|%)?$/],
      "margin": [/^[\d.]+(px|em|rem|%)(\s+[\d.]+(px|em|rem|%))*$/],
      "margin-top": [/^[\d.]+(px|em|rem|%)$/],
      "margin-bottom": [/^[\d.]+(px|em|rem|%)$/],
      "margin-left": [/^[\d.]+(px|em|rem|%|auto)$/],
      "margin-right": [/^[\d.]+(px|em|rem|%|auto)$/],
      "padding": [/^[\d.]+(px|em|rem|%)(\s+[\d.]+(px|em|rem|%))*$/],
      "padding-top": [/^[\d.]+(px|em|rem|%)$/],
      "padding-bottom": [/^[\d.]+(px|em|rem|%)$/],
      "padding-left": [/^[\d.]+(px|em|rem|%)$/],
      "padding-right": [/^[\d.]+(px|em|rem|%)$/],
      "border": [/^[\d.]+(px|em)\s+(solid|dashed|dotted|double|none)\s+/],
      "border-radius": [/^[\d.]+(px|em|rem|%)$/],
      "width": [/^[\d.]+(px|em|rem|%|auto)$/],
      "max-width": [/^[\d.]+(px|em|rem|%|none)$/],
      "height": [/^[\d.]+(px|em|rem|%|auto)$/],
      "display": [/^(block|inline|inline-block|none|flex|table|table-row|table-cell)$/],
      "vertical-align": [/^(top|middle|bottom|baseline|sub|super)$/],
      "list-style-type": [/^(disc|circle|square|decimal|lower-alpha|upper-alpha|lower-roman|upper-roman|none)$/],
    },
  },
};

interface EmailContentCardProps {
  htmlBody: string;
  textBody: string;
}

export function EmailContentCard({ htmlBody, textBody }: EmailContentCardProps) {
  const sanitizedHtml = htmlBody ? sanitizeHtml(htmlBody, sanitizeConfig) : "";

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Email Content</CardTitle>
      </CardHeader>
      <CardContent>
        {sanitizedHtml ? (
          <div
            className="prose prose-sm max-w-none rounded-md bg-background p-4 text-foreground overflow-x-auto [&_img]:max-w-full [&_img]:h-auto [&_table]:max-w-full [&_table]:overflow-x-auto [&_pre]:overflow-x-auto"
            dangerouslySetInnerHTML={{ __html: sanitizedHtml }}
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
