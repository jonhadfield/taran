import { getAvatarColor } from "@/lib/avatar-colors";

export function SenderAvatar({ name }: { name: string }) {
  return (
    <div
      className={`hidden sm:flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-medium text-white ${getAvatarColor(name)}`}
    >
      {(name || "?")[0].toUpperCase()}
    </div>
  );
}
