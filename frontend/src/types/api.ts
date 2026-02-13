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
  IsRead: boolean;
  IsStarred: boolean;
  IsArchived: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export type EmailStatus = "pending" | "processing" | "processed" | "failed";

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
  Items: DigestItem[];
}

export interface DigestItem {
  ID: string;
  DigestID: string;
  EmailID: string;
  ExtractionID: string;
  SortOrder: number;
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

export interface ListResponse<T> {
  data: T[];
  total: number;
}
