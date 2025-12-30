-- V1__initial_schema.down.sql
-- Rollback migration for initial schema
-- This will drop all tables and return to empty database

-- Drop FTS triggers and table
DROP TRIGGER IF EXISTS content_items_au;
DROP TRIGGER IF EXISTS content_items_ad;
DROP TRIGGER IF EXISTS content_items_ai;
DROP TABLE IF EXISTS content_fts;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS export_archives;
DROP TABLE IF EXISTS sync_queue;
DROP TABLE IF EXISTS ai_config;
DROP TABLE IF EXISTS sync_credentials;
DROP TABLE IF EXISTS conflict_log;
DROP TABLE IF EXISTS change_log;
DROP TABLE IF EXISTS content_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS content_items;
DROP TABLE IF EXISTS schema_migrations;
