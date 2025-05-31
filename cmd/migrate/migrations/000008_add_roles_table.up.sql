CREATE TABLE IF NOT EXISTS roles(
    id SERIAL PRIMARY KEY,
    roleid INT DEFAULT(0),
    name VARCHAR(10) NOT NULL UNIQUE,
    descript VARCHAR(100)
);

INSERT INTO roles (roleid, name, descript) 
VALUES (
    1,
    'user',
    'create and comment'
),(
    2,
    'moderator',
    'update other posts/users'
),(
    3,
    'admin',
    'all access'
);

