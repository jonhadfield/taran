import Image from "next/image";

export function DemoDigest({ priority = false }: { priority?: boolean }) {
  return (
    <>
      <figure className="w-full overflow-hidden rounded-xl border border-border/60 bg-background/80">
        <Image
          src="/digest-flow.png"
          alt="Forward newsletters, digest with AI, read one brief"
          width={1280}
          height={720}
          className="h-auto w-full"
          priority={priority}
        />
      </figure>

      <figure className="digest-sheet w-full space-y-4 p-5 sm:p-6">
        <figcaption className="text-center text-sm text-muted-foreground">
          What a digest looks like
        </figcaption>
        <div className="space-y-1">
          <p className="font-reading text-lg font-medium leading-snug">
            Your Daily Newsletter Digest
          </p>
          <p className="text-xs text-muted-foreground">
            14 February – 15 February · 8 emails
          </p>
        </div>
        <p className="font-reading text-[0.9375rem] leading-relaxed text-foreground/90">
          AI breakthroughs dominated today&apos;s newsletters with major
          announcements from leading labs. Markets reacted positively to strong
          earnings, while new open-source tools gained traction in the developer
          community.
        </p>
        <div>
          <p className="mb-2 text-xs font-medium text-muted-foreground">
            Highlights
          </p>
          <ul className="space-y-2">
            {[
              "New reasoning model achieves state-of-the-art benchmarks",
              "Tech earnings beat expectations across the board",
              "Open-source framework hits 50k GitHub stars",
            ].map((item) => (
              <li
                key={item}
                className="flex items-start gap-2 font-reading text-[0.9375rem] leading-relaxed text-muted-foreground"
              >
                <span
                  className="mt-2 size-1 shrink-0 rounded-full bg-primary/50"
                  aria-hidden
                />
                {item}
              </li>
            ))}
          </ul>
        </div>
        <ul className="flex flex-wrap gap-2 pt-1">
          {["AI", "Markets", "Open Source", "Startups"].map((topic) => (
            <li
              key={topic}
              className="rounded-md bg-muted px-2 py-0.5 text-xs text-muted-foreground"
            >
              {topic}
            </li>
          ))}
        </ul>
      </figure>
    </>
  );
}
