
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS gender VARCHAR(50) DEFAULT '',
  ADD COLUMN IF NOT EXISTS birth_date DATE DEFAULT CURRENT_DATE,
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE TABLE IF NOT EXISTS user_friends (
  user_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  friend_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, friend_id),
  CONSTRAINT chk_no_self_friend CHECK (user_id != friend_id),
  CONSTRAINT uq_friend_pair UNIQUE (user_id, friend_id)
);

CREATE INDEX IF NOT EXISTS idx_user_friends_friend_id ON user_friends(friend_id);
