create table repos (
    name varchar(255) primary key, -- e.g. someuser/somerepo/somebranch
    locs varchar(10000000), -- e.g. ["/somefile.go": 45, "/folder/otherfile.txt": 339]
    cached bigint -- unix time stamp
);
