CREATE DATABASE IF NOT EXISTS `gorest`;

CREATE TABLE IF NOT EXISTS build (
    id int unsigned auto_increment primary key,
    uuid varchar(255) not null,
    label varchar(255) not null,
    commit_sha varchar(255) default '',
    build_status_id int unsigned not null,
    created_on datetime not null default current_timestamp,
    updated_on datetime not null default current_timestamp,
    deleted_on datetime
) engine = innodb;

CREATE TABLE IF NOT EXISTS build_status  (
    id int unsigned auto_increment primary key,
    alias varchar(255) not null,
    name varchar(255) not null,
    created_on datetime not null default current_timestamp,
    updated_on datetime not null default current_timestamp,
    deleted_on datetime
) engine = innodb;

INSERT INTO build (uuid, label, commit_sha, build_status_id) VALUES 
    ("84f29c10-fcd5-4057-a0f5-aa7778ecf3d4", "Build 1", "1111111", 1),
    ("9455c109-858a-4c3a-8ebf-8d1eb1f93ae9", "Build 2", "2222222", 2),
    ("c4bee891-4024-4afc-ab08-99389dc53131", "Build 3", "3333333", 3)
;

INSERT INTO build_status (alias, name) VALUES 
    ("processing", "Processing"),
    ("success", "Success"),
    ("failed", "Failed")
;