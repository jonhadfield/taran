import { Skeleton } from "@/components/ui/skeleton";

export default function EmailDetailLoading() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-5 w-32" />
      <div className="space-y-2">
        <Skeleton className="h-8 w-96" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-4 w-48" />
      </div>
      <Skeleton className="h-48" />
      <Skeleton className="h-96" />
    </div>
  );
}
