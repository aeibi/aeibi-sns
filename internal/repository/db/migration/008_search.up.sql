CREATE EXTENSION IF NOT EXISTS pgroonga;
-- post: like search
CREATE INDEX IF NOT EXISTS idx_posts_text_pgroonga ON posts USING pgroonga (text);
-- user: nickname + description like search
CREATE INDEX IF NOT EXISTS idx_users_search_pgroonga ON users USING pgroonga ((ARRAY [nickname, description]));
-- user: nickname prefix search
CREATE INDEX IF NOT EXISTS idx_users_nickname_prefix_pgroonga ON users USING pgroonga (nickname pgroonga_text_term_search_ops_v2);
-- tag: like search
CREATE INDEX IF NOT EXISTS idx_tags_name_search_pgroonga ON tags USING pgroonga (name);
-- tag: name prefix search
CREATE INDEX IF NOT EXISTS idx_tags_name_prefix_pgroonga ON tags USING pgroonga (name pgroonga_text_term_search_ops_v2);