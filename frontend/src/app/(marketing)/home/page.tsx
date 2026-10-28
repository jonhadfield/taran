import Image from "next/image";
import Link from "next/link";
import { DemoDigest } from "@/components/demo-digest";
import { GitHubIcon } from "@/components/github-icon";
import { PublicShell } from "@/components/public-shell";
import { Button } from "@/components/ui/button";
import { APP_NAME, GITHUB_REPO_URL, SITE_DESCRIPTION } from "@/lib/config";

const steps = [
  {
    title: "Sign in",
    body: "Google or GitHub. You get a managed inbox such as you@mailbrief.io.",
  },
  {
    title: "Forward newsletters",
    body: "Point subscriptions at that address. No IMAP, no mailbox password.",
  },
  {
    title: "Read one brief",
    body: "Each day, one AI-written digest with key points, topics, and sentiment.",
  },
];

export default function HomePage() {
  return (
    <PublicShell width="2xl" showBrand={false}>
      <div className="flex items-center gap-3">
        <Image
          src="/logo.svg"
          alt=""
          width={48}
          height={48}
          priority
          className="rounded-xl"
        />
        <p className="text-3xl font-bold tracking-tight sm:text-4xl">{APP_NAME}</p>
      </div>

      <div className="space-y-4 text-center">
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          Forward your newsletters. Read one brief.
        </h1>
        <p className="max-w-lg text-base text-muted-foreground">{SITE_DESCRIPTION}</p>
      </div>

      <div className="flex w-full flex-col gap-3 sm:flex-row sm:justify-center">
        <Button asChild className="w-full sm:w-auto">
          <Link href="/login">Get started</Link>
        </Button>
        <Button asChild variant="outline" className="w-full bg-background sm:w-auto">
          <a href={GITHUB_REPO_URL} target="_blank" rel="noopener noreferrer">
            <GitHubIcon className="size-4" />
            Source on GitHub
          </a>
        </Button>
      </div>

      <section aria-labelledby="how-it-works" className="w-full space-y-4">
        <h2 id="how-it-works" className="text-center text-sm font-medium text-muted-foreground">
          How it works
        </h2>
        <ol className="space-y-4">
          {steps.map((step, index) => (
            <li key={step.title} className="flex gap-3">
              <span
                className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium text-muted-foreground"
                aria-hidden
              >
                {index + 1}
              </span>
              <div className="space-y-1">
                <p className="font-medium leading-snug">{step.title}</p>
                <p className="text-sm text-muted-foreground">{step.body}</p>
              </div>
            </li>
          ))}
        </ol>
      </section>

      <DemoDigest priority />

      <p className="max-w-lg text-center text-sm text-muted-foreground">
        Open source and self-hostable. MailBrief is the hosted app; the Taran
        codebase runs the same digest pipeline on your own infrastructure.
      </p>

      <p className="text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link href="/login" className="underline underline-offset-4">
          Sign in
        </Link>
      </p>
    </PublicShell>
  );
}
