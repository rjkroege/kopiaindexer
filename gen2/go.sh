#!/bin/sh


# TODO(rjk): figure out how to do this.

sqlite3 test.db <<EOF
create table manifests 
	(id INTEGER PRIMARY KEY,
	hid TEXT,
	hostname TEXT,
	path TEXT,
	type TEXT,
	username TEXT,
	mtime TEXT);
create unique index manifests_idx on manifests(hid);

EOF


# How do I wait for all of the tasks? wait by itself will wait for all of the child processes
# wait

# slurp all the indexes into the database