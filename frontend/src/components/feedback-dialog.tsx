"use client";

import { useState } from "react";
import { toast } from "sonner";
import { apiPost, ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";

export const MAX_FEEDBACK_LENGTH = 2000;

/**
 * Sends a message straight to the people who run MailBrief.
 *
 * Separate from the thumbs up/down on emails and digests: those tune the AI,
 * this reaches a person.
 */
export function FeedbackDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [message, setMessage] = useState("");
  const [sending, setSending] = useState(false);

  const handleSend = async () => {
    setSending(true);
    try {
      await apiPost("feedback", { Message: message.trim() });
      onOpenChange(false);
      setMessage("");
      toast.success("Thanks — your feedback is on its way");
    } catch (err) {
      toast.error(
        err instanceof ApiError && err.status !== 500
          ? err.message
          : "Couldn't send your feedback, please try again",
      );
    } finally {
      setSending(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Send feedback</DialogTitle>
          <DialogDescription>
            Tell us what is working, what isn&apos;t, or what you would like to
            see. It goes straight to us, and we can reply to your email address.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-1">
          <Textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            maxLength={MAX_FEEDBACK_LENGTH}
            rows={6}
            aria-label="Your feedback"
            placeholder="What's on your mind?"
            autoFocus
          />
          <p className="text-right text-xs text-muted-foreground">
            {message.length}/{MAX_FEEDBACK_LENGTH}
          </p>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSend} disabled={sending || !message.trim()}>
            {sending ? "Sending..." : "Send feedback"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
