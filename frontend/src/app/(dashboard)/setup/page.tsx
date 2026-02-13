"use client";

import { useRouter } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { UsernameForm } from "@/components/username-form";
import Link from "next/link";

export default function SetupPage() {
  const router = useRouter();

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Choose your inbox address</CardTitle>
          <CardDescription>
            Pick a unique username to create your newsletter inbox.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <UsernameForm onSuccess={() => router.push("/")} />
          <div className="text-center">
            <Link
              href="/inbox"
              className="text-sm text-muted-foreground hover:underline"
            >
              Skip for now
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
