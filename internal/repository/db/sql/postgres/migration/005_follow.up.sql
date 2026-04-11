-- user follows
CREATE TABLE user_follows (
    follower_uid uuid NOT NULL,
    followee_uid uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (follower_uid, followee_uid),
    CHECK (follower_uid <> followee_uid)
);

CREATE INDEX idx_user_follows_follower_created_at ON user_follows (follower_uid, created_at DESC, followee_uid DESC);
CREATE INDEX idx_user_follows_followee_created_at ON user_follows (followee_uid, created_at DESC, follower_uid DESC);
