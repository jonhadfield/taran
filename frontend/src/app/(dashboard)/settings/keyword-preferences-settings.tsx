"use client";

import { useState, type KeyboardEvent } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";

const KEYWORD_PATTERN = /^[a-zA-Z0-9 \-./+]+$/;
const MAX_KEYWORD_LENGTH = 30;
const MAX_KEYWORDS = 20;

interface KeywordPreferencesSettingsProps {
  interestKeywords: string[];
  exclusionKeywords: string[];
  prefLoading: boolean;
  prefSaving: boolean;
  onInterestKeywordsChange: (keywords: string[]) => void;
  onExclusionKeywordsChange: (keywords: string[]) => void;
}

function KeywordInput({
  keywords,
  onChange,
  disabled,
  placeholder,
}: {
  keywords: string[];
  onChange: (keywords: string[]) => void;
  disabled: boolean;
  placeholder: string;
}) {
  const [input, setInput] = useState("");
  const [error, setError] = useState("");

  const addKeyword = (raw: string) => {
    const kw = raw.trim();
    if (!kw) return;

    if (kw.length > MAX_KEYWORD_LENGTH) {
      setError(`Keyword must be ${MAX_KEYWORD_LENGTH} characters or less`);
      return;
    }
    if (!KEYWORD_PATTERN.test(kw)) {
      setError("Only letters, numbers, spaces, hyphens, dots, slashes, and + allowed");
      return;
    }
    if (keywords.length >= MAX_KEYWORDS) {
      setError(`Maximum ${MAX_KEYWORDS} keywords allowed`);
      return;
    }
    if (keywords.some((k) => k.toLowerCase() === kw.toLowerCase())) {
      setError("Keyword already added");
      return;
    }

    setError("");
    setInput("");
    onChange([...keywords, kw]);
  };

  const removeKeyword = (index: number) => {
    onChange(keywords.filter((_, i) => i !== index));
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addKeyword(input);
    }
    if (e.key === "Backspace" && !input && keywords.length > 0) {
      removeKeyword(keywords.length - 1);
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-1.5 min-h-[32px]">
        {keywords.map((kw, i) => (
          <Badge key={i} variant="secondary" className="gap-1 pr-1">
            {kw}
            <button
              type="button"
              onClick={() => removeKeyword(i)}
              disabled={disabled}
              className="ml-0.5 rounded-full p-0.5 hover:bg-muted-foreground/20 disabled:opacity-50"
            >
              <X className="h-3 w-3" />
            </button>
          </Badge>
        ))}
      </div>
      <Input
        value={input}
        onChange={(e) => {
          setInput(e.target.value);
          setError("");
        }}
        onKeyDown={handleKeyDown}
        onBlur={() => { if (input.trim()) addKeyword(input); }}
        placeholder={disabled ? "Loading..." : placeholder}
        disabled={disabled}
        className="text-sm"
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}

export function KeywordPreferencesSettings({
  interestKeywords,
  exclusionKeywords,
  prefLoading,
  prefSaving,
  onInterestKeywordsChange,
  onExclusionKeywordsChange,
}: KeywordPreferencesSettingsProps) {
  const disabled = prefLoading || prefSaving;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Content Keywords</CardTitle>
        <CardDescription>
          Boost or exclude content in your digests based on keywords
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-2">
          <label className="text-sm font-medium">Interest Keywords</label>
          <p className="text-xs text-muted-foreground">
            Content matching these keywords will be given more prominence in your digests.
          </p>
          <KeywordInput
            keywords={interestKeywords}
            onChange={onInterestKeywordsChange}
            disabled={disabled}
            placeholder="Type a keyword and press Enter..."
          />
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium">Exclusion Keywords</label>
          <p className="text-xs text-muted-foreground">
            Content matching these keywords will be omitted from your digests.
          </p>
          <KeywordInput
            keywords={exclusionKeywords}
            onChange={onExclusionKeywordsChange}
            disabled={disabled}
            placeholder="Type a keyword and press Enter..."
          />
        </div>
      </CardContent>
    </Card>
  );
}
