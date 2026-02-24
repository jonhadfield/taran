export interface Email {
  ID: string;
  UserID: string;
  AccountID: string;
  MessageID: string;
  FromAddress: string;
  FromName: string;
  ToAddress: string;
  Subject: string;
  TextBody: string;
  HTMLBody: string;
  ReceivedAt: string;
  DateHeader: string;
  Status: EmailStatus;
  StatusReason: string;
  IsRead: boolean;
  IsStarred: boolean;
  IsArchived: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export type EmailStatus = "pending" | "processing" | "processed" | "failed" | "skipped";

export interface EmailState {
  IsRead?: boolean;
  IsStarred?: boolean;
  IsArchived?: boolean;
}

export interface Extraction {
  ID: string;
  EmailID: string;
  Summary: string;
  KeyPoints: string[];
  Topics: string[];
  Links: Link[];
  ActionItems: string[];
  Sentiment: string;
  SourceCategory: string;
  Provider: string;
  Model: string;
  TokensUsed: number;
  ProcessedAt: string;
  CreatedAt: string;
}

export interface Link {
  url: string;
  title: string;
}

export interface EmailResponse extends Email {
  extraction?: Extraction;
}

export interface Digest {
  ID: string;
  UserID: string;
  Title: string;
  Summary: string;
  Highlights: string[];
  TopTopics: string[];
  PeriodStart: string;
  PeriodEnd: string;
  PeriodType: string;
  EmailCount: number;
  Provider: string;
  Model: string;
  GeneratedAt: string;
  SentAt: string | null;
  CreatedAt: string;
  ShareToken: string | null;
  Items: DigestItem[];
}

export interface DigestItem {
  ID: string;
  DigestID: string;
  EmailID: string;
  ExtractionID: string;
  SortOrder: number;
  Subject?: string;
  FromName?: string;
  FromAddress?: string;
  Summary?: string;
}

export interface EmailAccount {
  ID: string;
  UserID: string;
  EmailAddress: string;
  DisplayName: string;
  IsActive: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface UserPreference {
  UserID: string;
  DigestEmail: boolean;
  DigestFrequency: string;
  DigestHour: number;
  DigestDay: number;
  DigestTimezone: string;
  TopicLimit: number;
  DigestStyle: string;
  InterestKeywords: string[];
  ExclusionKeywords: string[];
  CreatedAt: string;
  UpdatedAt: string;
}

export interface SenderInfo {
  FromAddress: string;
  FromName: string;
  EmailCount: number;
  Status: string;
}

export interface SenderCount {
  FromAddress: string;
  FromName: string;
  Count: number;
}

export interface UserStats {
  EmailsThisWeek: number;
  EmailsLastWeek: number;
  TotalEmails: number;
  TopSenders: SenderCount[];
}

export interface EmailFeedback {
  ID: string;
  UserID: string;
  EmailID: string;
  Rating: "useful" | "not_useful" | null;
  CreatedAt: string;
}

export interface AdminStats {
  TotalUsers: number;
  ActiveUsersWeek: number;
  TotalEmails: number;
  EmailsThisWeek: number;
  TotalDigests: number;
  DigestsThisWeek: number;
  TopGlobalSenders: SenderCount[];
}

export interface Invite {
  ID: string;
  Email: string;
  InvitedBy: string;
  CreatedAt: string;
  AcceptedAt: string | null;
}

export interface AccessCheck {
  hasAccess: boolean;
  reason: "admin" | "invited" | "not_invited";
}

export interface SenderSuggestion {
  FromAddress: string;
  FromName: string;
  NotUsefulCount: number;
  TotalCount: number;
}

export interface ListResponse<T> {
  data: T[];
  total: number;
}

export interface ProcessingStats {
  PendingCount: number;
  ProcessingCount: number;
  FailedCount: number;
}

export interface WeekCount {
  Week: string;
  Count: number;
}

export interface WaitlistRequest {
  ID: string;
  Email: string;
  CreatedAt: string;
}

export interface DigestFeedback {
  ID: string;
  UserID: string;
  DigestID: string;
  Rating: "useful" | "not_useful" | null;
  CreatedAt: string;
}

export interface DashboardData {
  emails: Email[];
  emailTotal: number;
  digests: Digest[];
  unreadCount: number;
  stats: UserStats;
  processing: ProcessingStats;
}
