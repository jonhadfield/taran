"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { UsernameForm } from "@/components/username-form";
import { CopyButton } from "../settings/copy-button";
import { Mail, ArrowRight, Check } from "lucide-react";
import { APP_NAME } from "@/lib/config";
import { ForwardingGuide } from "@/components/forwarding-guide";
import { NewsletterSuggestions } from "@/components/newsletter-suggestions";

const EMAIL_DOMAIN = process.env.NEXT_PUBLIC_EMAIL_DOMAIN || "mailbrief.io";

export default function OnboardingPage() {
  const router = useRouter();
  const [step, setStep] = useState(1);
  const [emailAddress, setEmailAddress] = useState("");

  const handleInboxCreated = (createdEmailAddress: string) => {
    setEmailAddress(createdEmailAddress);
    setStep(3);
  };

  return (
    <div className="flex items-center justify-center min-h-[70vh]">
      <Card className="w-full max-w-lg">
        {step === 1 && (
          <>
            <CardHeader className="text-center">
              <div className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-primary">
                <Mail className="h-7 w-7 text-primary-foreground" />
              </div>
              <CardTitle className="text-2xl">Welcome to {APP_NAME}!</CardTitle>
              <CardDescription className="text-base">
                Let&apos;s set up your inbox in a few quick steps.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3 text-sm text-muted-foreground">
                <div className="flex items-start gap-3">
                  <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-medium text-primary-foreground">1</div>
                  <p>Choose a username for your inbox</p>
                </div>
                <div className="flex items-start gap-3">
                  <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">2</div>
                  <p>Subscribe newsletters to your new address</p>
                </div>
                <div className="flex items-start gap-3">
                  <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">3</div>
                  <p>Get AI-powered digests automatically</p>
                </div>
              </div>
              <Button className="w-full" onClick={() => setStep(2)}>
                Get Started
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </CardContent>
          </>
        )}

        {step === 2 && (
          <>
            <CardHeader>
              <CardTitle>Create your inbox</CardTitle>
              <CardDescription>
                Pick a username — newsletters sent to this address will appear in your dashboard.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <UsernameForm onSuccess={handleInboxCreated} />
            </CardContent>
          </>
        )}

        {step === 3 && (
          <>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Check className="h-5 w-5 text-success" />
                Your inbox is ready!
              </CardTitle>
              <CardDescription>
                Forward newsletters to your new address to get started.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center gap-2 rounded-lg border bg-muted/50 p-4">
                <code className="flex-1 text-sm font-medium">{emailAddress || `you@${EMAIL_DOMAIN}`}</code>
                {emailAddress && <CopyButton text={emailAddress} />}
              </div>

              <NewsletterSuggestions emailAddress={emailAddress} inline />

              <ForwardingGuide emailAddress={emailAddress} />

              <Button className="w-full" onClick={() => router.push("/inbox")}>
                Go to Dashboard
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </CardContent>
          </>
        )}


      </Card>
    </div>
  );
}
