/**
 * Identity palette for sender avatars — deliberately NOT the semantic status
 * tokens. These hues carry no meaning beyond "distinct from the last one", and
 * need more distinct values than the six categorical slots provide.
 */
const avatarColors = [
  "bg-blue-500",
  "bg-emerald-500",
  "bg-amber-500",
  "bg-rose-500",
  "bg-purple-500",
  "bg-cyan-500",
  "bg-indigo-500",
  "bg-pink-500",
];

export function getAvatarColor(name: string) {
  const charCode = name.charCodeAt(0) || 0;
  return avatarColors[charCode % avatarColors.length];
}
