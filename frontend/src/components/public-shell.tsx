import Image from "next/image";
import { APP_NAME } from "@/lib/config";
import { cn } from "@/lib/utils";

const widthClass = {
  md: "max-w-md",
  "2xl": "max-w-2xl",
} as const;

export function PublicShell({
  children,
  width = "md",
  showBrand = true,
  className,
}: {
  children: React.ReactNode;
  width?: keyof typeof widthClass;
  /** Logo + APP_NAME header. Set false when the page supplies its own title. */
  showBrand?: boolean;
  className?: string;
}) {
  return (
    <div className="public-wash flex min-h-screen flex-col items-center px-4 py-10 sm:py-16">
      <div
        className={cn(
          "flex w-full flex-col items-center gap-8",
          widthClass[width],
          className
        )}
      >
        {showBrand && (
          <div className="flex items-center gap-3">
            <Image
              src="/logo.svg"
              alt=""
              width={48}
              height={48}
              priority
              className="rounded-xl"
            />
            <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
              {APP_NAME}
            </h1>
          </div>
        )}
        {children}
      </div>
    </div>
  );
}
