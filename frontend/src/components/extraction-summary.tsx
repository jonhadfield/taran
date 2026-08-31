import type { Extraction } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2 } from "lucide-react";
import { isSafeURL } from "@/lib/utils";

interface ExtractionSummaryProps {
  extraction: Extraction;
  compact?: boolean;
}

export function ExtractionSummary({ extraction, compact = false }: ExtractionSummaryProps) {
  const ext = extraction;

  return (
    <div className={compact ? "space-y-3" : "space-y-4"}>
      <p className="font-reading text-[0.9375rem] leading-relaxed">{ext.Summary}</p>

      {ext.KeyPoints?.length > 0 && (
        <div>
          <h4 className={compact ? "text-xs font-medium text-muted-foreground mb-1.5" : "text-sm font-medium mb-2"}>
            Key Points
          </h4>
          <ul className={compact ? "space-y-1" : "space-y-1.5"}>
            {ext.KeyPoints.map((point, i) => (
              <li key={i} className="flex items-start gap-2 font-reading text-[0.9375rem] leading-relaxed text-muted-foreground">
                <CheckCircle2 className={`${compact ? "size-3.5" : "size-4"} mt-0.5 shrink-0 text-info`} />
                <span>{point}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {ext.Topics?.length > 0 && (
        <div className={compact ? "flex flex-wrap gap-1.5" : ""}>
          {!compact && <h4 className="text-sm font-medium mb-2">Topics</h4>}
          <div className={compact ? "flex flex-wrap gap-1.5" : "flex flex-wrap gap-2"}>
            {ext.Topics.map((topic) => (
              <Badge key={topic} variant="secondary" className={compact ? "text-xs" : ""}>
                {topic}
              </Badge>
            ))}
          </div>
        </div>
      )}

      {ext.Links?.length > 0 && (
        <div>
          <h4 className={compact ? "text-xs font-medium text-muted-foreground mb-1.5" : "text-sm font-medium mb-2"}>
            Links
          </h4>
          <ul className={`${compact ? "space-y-0.5" : "space-y-1"} text-sm`}>
            {ext.Links.map((link, i) => (
              <li key={i}>
                <a
                  href={isSafeURL(link.url) ? link.url : "#"}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {link.title || link.url}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {ext.Sentiment && (
        <div>
          {!compact && <h4 className="text-sm font-medium mb-1">Sentiment</h4>}
          <Badge variant="outline" className={compact ? "text-xs" : ""}>{ext.Sentiment}</Badge>
        </div>
      )}
    </div>
  );
}
