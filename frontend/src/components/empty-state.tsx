import { cn } from "@/lib/utils";

export function InboxIllustration({ className }: { className?: string }) {
  return (
    <svg className={cn("size-20", className)} viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="12" y="20" width="56" height="40" rx="4" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <path d="M12 28L40 44L68 28" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <circle cx="40" cy="36" r="8" className="stroke-muted-foreground/20" strokeWidth="2" strokeDasharray="4 3" fill="none" />
      <path d="M37 36L39 38L43 34" className="stroke-muted-foreground/30" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
    </svg>
  );
}

export function DigestIllustration({ className }: { className?: string }) {
  return (
    <svg className={cn("size-20", className)} viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="18" y="14" width="44" height="56" rx="3" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <line x1="26" y1="26" x2="54" y2="26" className="stroke-muted-foreground/20" strokeWidth="2" strokeLinecap="round" />
      <line x1="26" y1="34" x2="50" y2="34" className="stroke-muted-foreground/20" strokeWidth="2" strokeLinecap="round" />
      <line x1="26" y1="42" x2="46" y2="42" className="stroke-muted-foreground/20" strokeWidth="2" strokeLinecap="round" />
      <circle cx="54" cy="54" r="10" className="fill-primary/10 stroke-primary/40" strokeWidth="2" />
      <path d="M51 54L53 56L57 52" className="stroke-primary/50" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
    </svg>
  );
}

export function SendersIllustration({ className }: { className?: string }) {
  return (
    <svg className={cn("size-20", className)} viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="30" cy="32" r="8" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <path d="M18 52C18 46 23 42 30 42C37 42 42 46 42 52" className="stroke-muted-foreground/40" strokeWidth="2" strokeLinecap="round" fill="none" />
      <circle cx="50" cy="28" r="6" className="stroke-muted-foreground/25" strokeWidth="2" fill="none" />
      <path d="M42 44C42 40 45 37 50 37C55 37 58 40 58 44" className="stroke-muted-foreground/25" strokeWidth="2" strokeLinecap="round" fill="none" />
      <circle cx="56" cy="50" r="5" className="stroke-muted-foreground/15" strokeWidth="2" strokeDasharray="3 2" fill="none" />
    </svg>
  );
}

export function SearchIllustration({ className }: { className?: string }) {
  return (
    <svg className={cn("size-20", className)} viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="34" cy="34" r="14" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <line x1="44" y1="44" x2="58" y2="58" className="stroke-muted-foreground/40" strokeWidth="2.5" strokeLinecap="round" />
      <line x1="28" y1="30" x2="40" y2="30" className="stroke-muted-foreground/20" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="28" y1="36" x2="36" y2="36" className="stroke-muted-foreground/20" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

export function SubscriptionIllustration({ className }: { className?: string }) {
  return (
    <svg className={cn("size-20", className)} viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="14" y="24" width="52" height="32" rx="3" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <path d="M14 32L40 46L66 32" className="stroke-muted-foreground/30" strokeWidth="1.5" fill="none" />
      <line x1="50" y1="18" x2="50" y2="28" className="stroke-muted-foreground/25" strokeWidth="2" strokeLinecap="round" />
      <line x1="45" y1="23" x2="55" y2="23" className="stroke-muted-foreground/25" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

export function DashboardIllustration({ className }: { className?: string }) {
  return (
    <svg className={cn("size-20", className)} viewBox="0 0 80 80" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="12" y="22" width="56" height="38" rx="3" className="stroke-muted-foreground/40" strokeWidth="2" fill="none" />
      <rect x="18" y="36" width="8" height="18" rx="1" className="fill-muted-foreground/10 stroke-muted-foreground/20" strokeWidth="1" />
      <rect x="30" y="30" width="8" height="24" rx="1" className="fill-muted-foreground/15 stroke-muted-foreground/20" strokeWidth="1" />
      <rect x="42" y="38" width="8" height="16" rx="1" className="fill-muted-foreground/10 stroke-muted-foreground/20" strokeWidth="1" />
      <rect x="54" y="32" width="8" height="22" rx="1" className="fill-muted-foreground/15 stroke-muted-foreground/20" strokeWidth="1" />
      <circle cx="22" cy="18" r="3" className="stroke-muted-foreground/20" strokeWidth="1.5" fill="none" />
    </svg>
  );
}
