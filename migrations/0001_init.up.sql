create table repos (
    id varchar(255) primary key, -- e.g. someuser/somerepo/somebranch
    locs varchar(10000000), -- e.g. ["/somefile.go": 45, "/folder/otherfile.txt": 339]
    too_large boolean, -- `locs` is not filled in this case
    created_at bigint -- unix time stamp
);
