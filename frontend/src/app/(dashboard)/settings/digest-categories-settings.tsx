"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const ALL_CATEGORIES = [
  { value: "newsletter", label: "Newsletters", description: "Regular email newsletters and subscriptions" },
  { value: "personal", label: "Personal", description: "Personal messages and correspondence" },
  { value: "transactional", label: "Transactional", description: "Order confirmations, shipping updates, receipts" },
  { value: "marketing", label: "Marketing", description: "Promotional emails and offers" },
  { value: "notification", label: "Notifications", description: "Account alerts, password resets, system notices" },
  { value: "other", label: "Other", description: "Uncategorized emails" },
];

interface DigestCategoriesSettingsProps {
  excludedCategories: string[];
  prefLoading: boolean;
  prefSaving: boolean;
  onExcludedCategoriesChange: (categories: string[]) => void;
}

export function DigestCategoriesSettings({
  excludedCategories,
  prefLoading,
  prefSaving,
  onExcludedCategoriesChange,
}: DigestCategoriesSettingsProps) {
  const toggleCategory = (category: string) => {
    const isExcluded = excludedCategories.includes(category);
    if (isExcluded) {
      onExcludedCategoriesChange(excludedCategories.filter((c) => c !== category));
    } else {
      onExcludedCategoriesChange([...excludedCategories, category]);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Digest Categories</CardTitle>
        <CardDescription>
          Choose which types of emails to include in your digests. Unchecked
          categories will be excluded.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {ALL_CATEGORIES.map((cat) => {
          const isIncluded = !excludedCategories.includes(cat.value);
          return (
            <label
              key={cat.value}
              className="flex items-start gap-3 cursor-pointer"
            >
              <input
                type="checkbox"
                checked={isIncluded}
                onChange={() => toggleCategory(cat.value)}
                disabled={prefLoading || prefSaving}
                className="mt-0.5 size-4 rounded border-input accent-primary cursor-pointer"
              />
              <div>
                <span className="text-sm font-medium">{cat.label}</span>
                <p className="text-xs text-muted-foreground">
                  {cat.description}
                </p>
              </div>
            </label>
          );
        })}
      </CardContent>
    </Card>
  );
}
