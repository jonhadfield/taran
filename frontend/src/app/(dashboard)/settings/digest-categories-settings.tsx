"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CATEGORY_OPTIONS } from "@/lib/category-constants";

const CATEGORY_DESCRIPTIONS: Record<string, string> = {
  newsletter: "Regular email newsletters and subscriptions",
  personal: "Personal messages and correspondence",
  transactional: "Order confirmations, shipping updates, receipts",
  marketing: "Promotional emails and offers",
  notification: "Account alerts, password resets, system notices",
  other: "Uncategorized emails",
};

const ALL_CATEGORIES = CATEGORY_OPTIONS
  .filter((o) => o.value !== "")
  .map((o) => ({
    value: o.value,
    label: o.label === "Newsletter" ? "Newsletters" : o.label === "Notification" ? "Notifications" : o.label,
    description: CATEGORY_DESCRIPTIONS[o.value] || "",
  }));

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
