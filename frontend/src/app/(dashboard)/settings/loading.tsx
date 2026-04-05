import { Skeleton } from "@/components/ui/skeleton";

export default function SettingsLoading() {
  return (
    <div className="space-y-6 max-w-2xl">
      <Skeleton className="h-8 w-32" />
      <Skeleton className="h-40" />
      <Skeleton className="h-32" />
    </div>
  );
}
