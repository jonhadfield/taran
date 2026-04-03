"use client";

import { authClient } from "@/lib/auth-client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { APP_NAME } from "@/lib/config";
import { Mail, Sparkles, BookOpen } from "lucide-react";
import Image from "next/image";

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-muted/40 p-4">
      <div className="flex w-full max-w-md flex-col items-center gap-10">
        {/* Hero */}
        <div className="flex flex-col items-center gap-3 text-center">
          <Image
            src="/logo.svg"
            alt={`${APP_NAME} logo`}
            width={80}
            height={80}
            priority
          />
          <h1 className="text-4xl font-bold tracking-tight">{APP_NAME}</h1>
          <p className="text-lg text-muted-foreground">
            AI-powered digests of your newsletters, delivered daily.
          </p>
          <p className="text-sm font-medium text-muted-foreground/80">
            Currently invite-only &mdash; sign in below to join the waitlist
          </p>
        </div>

        {/* How it works */}
        <div className="grid w-full grid-cols-3 gap-4 text-center">
          <div className="flex flex-col items-center gap-2">
            <Mail className="h-5 w-5 text-muted-foreground" />
            <p className="text-xs font-medium">Forward</p>
            <p className="text-xs text-muted-foreground">
              Send newsletters to your @mailbrief.io inbox
            </p>
          </div>
          <div className="flex flex-col items-center gap-2">
            <Sparkles className="h-5 w-5 text-muted-foreground" />
            <p className="text-xs font-medium">Digest</p>
            <p className="text-xs text-muted-foreground">
              AI reads everything and creates a daily summary
            </p>
          </div>
          <div className="flex flex-col items-center gap-2">
            <BookOpen className="h-5 w-5 text-muted-foreground" />
            <p className="text-xs font-medium">Read</p>
            <p className="text-xs text-muted-foreground">
              Get one concise digest instead of dozens of emails
            </p>
          </div>
        </div>

        {/* Sign in */}
        <div className="w-full space-y-3">
          <Button
            variant="outline"
            className="w-full"
            onClick={() =>
              authClient.signIn.social({ provider: "google", callbackURL: "/" })
            }
          >
            <GoogleIcon />
            Continue with Google
          </Button>
          <Button
            variant="outline"
            className="w-full"
            onClick={() =>
              authClient.signIn.social({ provider: "github", callbackURL: "/" })
            }
          >
            <GitHubIcon />
            Continue with GitHub
          </Button>
        </div>

        {/* Demo digest preview */}
        <div className="w-full space-y-3">
          <p className="text-center text-sm text-muted-foreground">
            Here&apos;s what a digest looks like
          </p>
          <Card className="opacity-90">
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Your Daily Newsletter Digest</CardTitle>
              <p className="text-xs text-muted-foreground">Feb 14 &ndash; Feb 15, 2026 &middot; 8 emails</p>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">
                AI breakthroughs dominated today&apos;s newsletters with major announcements from leading labs. Markets reacted positively to strong earnings, while new open-source tools gained traction in the developer community.
              </p>
              <div>
                <p className="text-xs font-medium mb-2">Highlights</p>
                <ul className="list-disc list-inside space-y-1 text-xs text-muted-foreground">
                  <li>New reasoning model achieves state-of-the-art benchmarks</li>
                  <li>Tech earnings beat expectations across the board</li>
                  <li>Open-source framework hits 50k GitHub stars</li>
                </ul>
              </div>
              <div className="flex flex-wrap gap-1.5">
                {["AI", "Markets", "Open Source", "Startups"].map((topic) => (
                  <Badge key={topic} variant="secondary" className="text-xs">
                    {topic}
                  </Badge>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function GoogleIcon() {
  return (
    <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24">
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  );
}

function GitHubIcon() {
  return (
    <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
    </svg>
  );
}
