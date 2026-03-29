"use client";

import { useEffect } from "react";

interface FaviconBadgeProps {
  count: number;
}

export function FaviconBadge({ count }: FaviconBadgeProps) {
  useEffect(() => {
    const canvas = document.createElement("canvas");
    canvas.width = 32;
    canvas.height = 32;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const img = new Image();
    img.crossOrigin = "anonymous";
    img.src = "/logo.svg";

    img.onload = () => {
      ctx.drawImage(img, 0, 0, 32, 32);

      if (count > 0) {
        const text = count > 99 ? "99+" : String(count);

        // Badge circle
        ctx.beginPath();
        ctx.arc(24, 8, 9, 0, 2 * Math.PI);
        ctx.fillStyle = "#ef4444";
        ctx.fill();

        // Badge text
        ctx.fillStyle = "#ffffff";
        ctx.font = "bold 11px sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(text, 24, 9);
      }

      // Update favicon
      let link = document.querySelector<HTMLLinkElement>("link[rel='icon']");
      if (!link) {
        link = document.createElement("link");
        link.rel = "icon";
        document.head.appendChild(link);
      }
      link.href = canvas.toDataURL("image/png");
    };

    img.onerror = () => {
      // If logo.svg fails to load, just set the badge on a blank canvas
      if (count > 0) {
        ctx.beginPath();
        ctx.arc(16, 16, 14, 0, 2 * Math.PI);
        ctx.fillStyle = "#ef4444";
        ctx.fill();
        ctx.fillStyle = "#ffffff";
        ctx.font = "bold 14px sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(count > 99 ? "99+" : String(count), 16, 17);

        let link = document.querySelector<HTMLLinkElement>("link[rel='icon']");
        if (!link) {
          link = document.createElement("link");
          link.rel = "icon";
          document.head.appendChild(link);
        }
        link.href = canvas.toDataURL("image/png");
      }
    };

    return () => {
      // Restore original favicon when unmounted
      const link = document.querySelector<HTMLLinkElement>("link[rel='icon']");
      if (link) link.href = "/favicon.ico";
    };
  }, [count]);

  return null;
}
