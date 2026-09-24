"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiDelete, apiGet } from "@/lib/api";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ForwardingGuide } from "@/components/forwarding-guide";
import { NewsletterSuggestions } from "@/components/newsletter-suggestions";
import { AccountSettings } from "../account-settings";
import { SettingsHeader } from "../settings-panel";
import { settingsGroupBySlug } from "../settings-groups";
import type { EmailAccount, ListResponse } from "@/types/api";

const group = settingsGroupBySlug("inbox")!;

export default function InboxSettingsPage() {
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<EmailAccount | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchAccounts = useCallback(async () => {
    try {
      const res = await apiGet<ListResponse<EmailAccount>>("accounts");
      setAccounts(res.data || []);
      setError("");
    } catch {
      setError("Failed to load accounts");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAccounts();
  }, [fetchAccounts]);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await apiDelete(`accounts/${deleteTarget.ID}`);
      setDeleteTarget(null);
      toast.success("Inbox deleted");
      await fetchAccounts();
    } catch {
      toast.error("Failed to delete inbox");
      setDeleteTarget(null);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <>
      <SettingsHeader title={group.label} description={group.description} />

      <section id="accounts" className="scroll-mt-24">
        <AccountSettings
          accounts={accounts}
          loading={loading}
          error={error}
          deleteTarget={deleteTarget}
          deleting={deleting}
          onFetchAccounts={fetchAccounts}
          onSetDeleteTarget={setDeleteTarget}
          onDelete={handleDelete}
        />
      </section>

      {accounts.length > 0 && (
        <>
          <section id="forwarding" className="scroll-mt-24">
            <Card>
              <CardHeader>
                <CardTitle>Forwarding</CardTitle>
                <CardDescription>
                  How to forward newsletters from your email provider
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ForwardingGuide emailAddress={accounts[0]?.EmailAddress} />
              </CardContent>
            </Card>
          </section>

          <section id="newsletters" className="scroll-mt-24">
            <NewsletterSuggestions emailAddress={accounts[0]?.EmailAddress} />
          </section>
        </>
      )}
    </>
  );
}
