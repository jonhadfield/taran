export interface Email {
  ID: string;
  UserID: string;
  AccountID: string;
  MessageID: string;
  InReplyTo: string;
  ThreadID: string;
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
  UnsubscribeURL: string;
  UnsubscribeMailto: string;
  RetryCount: number;
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

export interface EmailAttachment {
  ID: string;
  EmailID: string;
  Filename: string;
  ContentType: string;
  SizeBytes: number;
  CreatedAt: string;
}

export interface EmailResponse extends Email {
  extraction?: Extraction;
  attachments?: EmailAttachment[];
  ThreadCount?: number;
}

export interface ThreadEmail {
  ID: string;
  Subject: string;
  FromName: string;
  FromAddress: string;
  ReceivedAt: string;
  IsRead: boolean;
  Summary?: string;
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
  DigestFrequency: "daily" | "weekly";
  DigestHour: number;
  DigestDay: number;
  DigestTimezone: string;
  TopicLimit: number;
  DigestStyle: "detailed" | "concise";
  InterestKeywords: string[];
  ExclusionKeywords: string[];
  ColorTheme: string;
  ExcludedCategories: string[];
  DailyTokenLimit: number;
  QuietHoursEnabled: boolean;
  QuietHoursStart: number;
  QuietHoursEnd: number;
  WeeklySummary: boolean;
  DigestWebhook: boolean;
  WebhookURL: string;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface SenderInfo {
  FromAddress: string;
  FromName: string;
  EmailCount: number;
  Status: string;
  Category: string;
  AutoCategory: string;
}

export interface SenderDetail {
  FromAddress: string;
  FromName: string;
  EmailCount: number;
  Status: string;
  Category: string;
  AutoCategory: string;
  FirstSeen: string;
  LastSeen: string;
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
  LLMProvider: string;
  LLMModel: string;
  MonthlyTokensUsed: number;
  DefaultMonthlyTokenLimit: number;
  ProcessedCount: number;
  FailedCount: number;
  SkippedCount: number;
  PendingCount: number;
  FeedbackUseful: number;
  FeedbackNotUseful: number;
  WeeklyEmails: WeekCount[];
  WeeklyDigests: WeekCount[];
  WeeklyTokens: WeekCount[];
}

export interface AdminUser {
  ID: string;
  Email: string;
  Name: string;
  EmailCount: number;
  MonthlyTokensUsed: number;
  MonthlyTokenLimit: number;
}

export interface DailyTokenCount {
  Date: string;
  Tokens: number;
}

export interface UsageStats {
  MonthlyTokensUsed: number;
  MonthlyTokenLimit: number;
  DailyTokensUsed: number;
  DailyTokenLimit: number;
  TriageTokens: number;
  ExtractTokens: number;
  DigestTokens: number;
  DailyHistory: DailyTokenCount[] | null;
  PeriodStart: string;
  PeriodEnd: string;
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

export interface UserLLMKey {
  Provider: string;
  KeyHint: string;
  Model: string;
  IsActive: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface WeeklySummary {
  ID: string;
  UserID: string;
  PeriodStart: string;
  PeriodEnd: string;
  EmailCount: number;
  TopSenders: SenderCount[];
  Categories: Record<string, number>;
  ActionItems: number;
  SentAt: string | null;
  CreatedAt: string;
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

export interface DigestPreviewItem {
  EmailID: string;
  Subject: string;
  FromName: string;
  FromAddress: string;
  Summary: string;
  ReceivedAt: string;
}

export interface DigestPreview {
  PeriodStart: string;
  PeriodEnd: string;
  PeriodType: string;
  EmailCount: number;
  Items: DigestPreviewItem[];
}

export interface Label {
  ID: string;
  UserID: string;
  Name: string;
  Color: string;
  EmailCount: number;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface SubscriptionInfo {
  FromAddress: string;
  FromName: string;
  EmailCount: number;
  LastSeen: string;
  UnsubscribeURL: string;
  UnsubscribeMailto: string;
  UnsubscribedAt: string | null;
}

export interface SavedSearch {
  ID: string;
  Name: string;
  Filters: {
    filter?: string;
    search?: string;
    topic?: string;
    category?: string;
    sort?: string;
    labelId?: string;
    hasAttachment?: boolean;
    since?: string;
    before?: string;
  };
  CreatedAt: string;
  UpdatedAt: string;
}

export interface AutoArchiveRule {
  ID: string;
  RuleType: "category" | "sender";
  RuleValue: string;
  ArchiveAfterDays: number;
  IsActive: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface FailedEmail extends Email {
  UserEmail: string;
}

export interface PipelineHealth {
  pending: number;
  processing: number;
  processed: number;
  failed: number;
  skipped: number;
  payloads: number;
}

export interface TopicCount {
  Topic: string;
  Count: number;
}

export interface CategoryCount {
  Category: string;
  Count: number;
}

export interface HeatmapCell {
  Hour: number;
  Day: number;
  Count: number;
}

export interface DashboardData {
  emails: Email[];
  emailTotal: number;
  digests: Digest[];
  unreadCount: number;
  stats: UserStats;
  processing: ProcessingStats;
  weeklyHistory: WeekCount[];
  topicsWithCount: TopicCount[];
  categories: CategoryCount[];
  actionItems: number;
  heatmap: HeatmapCell[];
}
