create table users (
    id int not null AUTO_INCREMENT,
    email varchar(320) not null UNIQUE,
    pass_hash varchar(320),
    user_name varchar(256) not null UNIQUE,
    first_name varchar(64),
    last_name varchar(128),
    photo_URL varchar(128) not null,
    PRIMARY KEY (id)
);