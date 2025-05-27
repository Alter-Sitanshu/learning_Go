CREATE TABLE IF NOT EXISTS followers(
    userid BIGINT NOT NULL,
    follower_id BIGINT NOT NULL,

    PRIMARY KEY(userid, follower_id),
    FOREIGN KEY (userid) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE
);
