
export type User = {
  id: number;
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  email_verified: boolean;
  username: string;
  avatar_id: string;
  created_at: number;
  last_updated_at: number;
};

export type Collection = {
  id: string;
  name: string;
  slug: string;
  workspace_id: number;
  description: string;
  avatar_id: string;
  created_at: string;
  owner_id: number;
  members_count: number;
  entries_count: number;
};

export type Workspace = {
  id: string;
  name: string;
  slug: string;
  owner_id: number;
  description: string;
  avatar_id: string;
  enable_public_indexing: boolean;
  invite_only: boolean;
  created_at: string;
  collections: Array<Collection>;
};

export type MemberWithUserID = {
  member_id: number;
  user_id: number;
};

export type EntryAddedBy = {
  first_name: string;
  last_name: string;
  username: string;
};

export type EntryRelation = {
  id: string;
  name: string;
  slug: string;
};

export type Entry = {
  id: string;
  origin: string;
  name: string;
  content: string;
  text_content: string;
  version: number;
  type: 'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other';
  parent_id: number;
  file_id?: string | null;
  filesize_bytes: number;
  status: 'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'paused';
  queued_at: string;
  created_at: string;
  updated_at: string;
  archived_at: string;
  added_by: EntryAddedBy;
  collection: EntryRelation;
  workspace: EntryRelation;
  metadata: import('./types').FileMetadata | import('./types').Metadata;
};

export type PluginSource = {
  id: string;
  workspace_id: number;
  name: string;
  description: string;
  author: string;
  disabled_at: string;
  versioning_strategy: 'commit' | 'tag';
  source_url: string;
  version_id: string;
  sync_status: string;
  last_sync_error: string;
  last_synced_at: string;
  added_at: string;
  updated_at: string;
};

export type Metadata = {
  title: string;
  description: string;
  favicon: string;
  author: string;
  thumbnail: string;
  site_type: string;
  domain: string;
  link: string;
};

export type FileMetadata = {
  original_filename: string;
  mime_type: string;
  extension: string;
  extra_metadata?: Array<number> | null;
};

export type Member = {
  id: number;
  invite_id?: string | null;
  first_name: string;
  last_name: string;
  email: string;
  role: 'user' | 'admin' | 'guest' | 'owner';
  user: MemberUser;
  status: 'accepted' | 'pending' | 'declined' | 'revoked' | 'expired';
  created_at: string;
};

export type MembershipStatus = {
  member_id: number;
  user_id: number;
  role: 'user' | 'admin' | 'guest' | 'owner';
  is_member: boolean;
};

export type MemberUser = {
  id: string;
};

export type InstalledPlugin = {
  id: string;
  identifier: string;
  workspace_id: number;
  source_id: string;
  name: string;
  description: string;
  scope: 'global' | 'workspace';
  modes: Array<'on_create' | 'background'>;
  targets: Array<'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other' | '*'>;
  version_sha: string;
  last_updated_at: string;
  added_at: string;
  updated_at: string;
  metadata: Array<number>;
  tags: Array<string>;
  privileges: Array<Privilege>;
};

export type SearchResultChunkMetadata = {
  id: number;
  index: number;
};

export type SearchResult = {
  id: string;
  rank: number;
  name: string;
  preview: string;
  type: 'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other';
  search_type: 'full_text' | 'semantic';
  status: 'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'paused';
  matched_by: Array<'full_text' | 'semantic' | 'keyword'>;
  text_score: number;
  semantic_score: number;
  hybrid_score: number;
  chunk: SearchResultChunkMetadata;
  metadata: import('./types').FileMetadata | import('./types').Metadata;
  file_id?: string | null;
  filesize_bytes: number;
  collection: EntryRelation;
  workspace: EntryRelation;
  created_at: string;
  updated_at: string;
  archived_at: string;
};

export type MatchedChunk = {
  id: number;
  index: number;
  text: string;
  rank: number;
  text_score: number;
  semantic_score: number;
  hybrid_score: number;
};

export type CollapsedSearchResult = {
  id: string;
  name: string;
  type: 'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other';
  matches: Array<MatchedChunk>;
  status: 'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'paused';
  file_id?: string | null;
  filesize_bytes: number;
  relevance_percent: number;
  collection: EntryRelation;
  workspace: EntryRelation;
  created_at: string;
  updated_at: string;
  archived_at: string;
  metadata: import('./types').FileMetadata | import('./types').Metadata;
};

export type HybridSearchResults = {
  results: Array<CollapsedSearchResult>;
  min_hybrid_score: number;
  max_hybrid_score: number;
};

export type PluginListItemSource = {
  id: string;
  name: string;
  url: string;
};

export type PluginListItem = {
  identifier: string;
  name: string;
  description: string;
  author: string;
  source: PluginListItemSource;
  privileges: Array<Privilege>;
  targets: Array<string>;
  installed: boolean;
  updatable: boolean;
};

export type Privilege = {
  identifier: string;
  description: string;
};

export type PluginV1 = {
  name: string;
  description: string;
  modes: Array<'on_create' | 'background'>;
  targets: 'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other' | '*';
  privileges: Array<Privilege>;
};

export type SourceV1 = {
  name: string;
  author: string;
  description: string;
  plugins: Array<string>;
  versioning_strategy: 'on_create' | 'background';
  url: string;
};

export type SourceWithPlugins = {
  source: SourceV1;
  plugins: Array<PluginV1>;
  version: string;
};