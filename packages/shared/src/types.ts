/**
 * MemoNexus Shared TypeScript Types
 *
 * This file contains type definitions shared across the monorepo.
 * Used primarily for desktop wrapper TypeScript integration.
 */

/**
 * Media type enumeration for content items
 */
export enum MediaType {
  Web = 'web',
  Image = 'image',
  Video = 'video',
  PDF = 'pdf',
  Markdown = 'markdown',
}

/**
 * Content item model
 */
export interface ContentItem {
  id: string; // UUID v4
  title: string;
  content_text: string;
  source_url?: string;
  media_type: MediaType;
  tags: string; // Comma-separated
  summary?: string;
  is_deleted: boolean;
  created_at: number; // Unix timestamp
  updated_at: number; // Unix timestamp
  version: number;
  content_hash?: string; // SHA-256
}

/**
 * Tag model
 */
export interface Tag {
  id: string; // UUID v4
  name: string;
  color: string; // Hex color code
  is_deleted: boolean;
  created_at: number;
  updated_at: number;
}

/**
 * Search result with relevance score
 */
export interface SearchResult {
  item: ContentItem;
  relevance: number;
  matched_terms: string[];
}

/**
 * AI configuration
 */
export interface AIConfig {
  id: string;
  provider: 'openai' | 'claude' | 'ollama';
  api_endpoint: string;
  api_key_encrypted: string; // Never expose in JSON responses
  model_name: string;
  max_tokens: number;
  is_enabled: boolean;
}

/**
 * Sync credentials
 */
export interface SyncCredential {
  id: string;
  endpoint: string;
  bucket_name: string;
  region?: string;
  access_key_encrypted: string; // Never expose in JSON responses
  secret_key_encrypted: string; // Never expose in JSON responses
  is_enabled: boolean;
}

/**
 * API error response
 */
export interface APIError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}
