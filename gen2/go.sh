#!/usr/local/bin/rc

sqlite3 test.db <<'!!'
-- Create the manifests table. mid is the manifest id. This is 1-1 with the snapshot id
create table  if not exists manifests (
	id INTEGER PRIMARY KEY,
	mid TEXT unique,
	snid TEXT unique,
	hostname TEXT,
	path TEXT,
	type TEXT,
	username TEXT,
	mtime TEXT,
	state TEXT
);

-- Makes an index on hid.
-- Why can't I haz a unique index?
create index if not exists manifests_snid_idx on manifests(snid);

-- Populate the manifests table 
-- TODO(rjk): Don't update if it already exists (this is breaking my worlds)
-- TODO(rjk): Run a command to generate.
-- TODO(rjk): Track state
INSERT or ignore INTO manifests (
	mid,
	hostname,
	path,
	type,
	username,
	mtime
) SELECT 
  json_extract(value, '$.id'), 
  json_extract(value, '$.labels.hostname'),
  json_extract(value, '$.labels.path'),
  json_extract(value, '$.labels.type'),
  json_extract(value, '$.labels.username'),
  json_extract(value, '$.mtime')
FROM json_each(readfile('manifest.json'));

create table  if not exists files (
	id INTEGER PRIMARY KEY,
	fid TEXT ,
	snid TEXT,
	path TEXT
);

-- TODO(rjk): Add indicies as needed. Probably on fid?
-- TODO(rjk): Adjust the indexing of the path to address wanting to sub-string it
-- TODO(rjk): How do I figure out what are the right indicies
create  index if not exists files_fid_idx on files(fid);
create  index if not exists files_path_idx on files(path);
!!

# TODO(rjk): items will be removed from the manifest.
# How do I know what's nolonger in the manifest?

# Note that I have truncated.


# Note that I should add the snapshot id too because it's really handy.
# And I need to index on it. And it will let me do what I want with the right
# magic query.
for (i in `{sqlite3 test.db 'select mid from manifests where type == "snapshot" and state ISNULL limit 7;'}) {
	echo starting $i^.manifest && \
	kopia manifest show $i > $i^.manifest && \
	$KOPIAINDEXER/cmd/lister/lister $i^.manifest | sed 's/ /,/g' > $i^.index && \
	sqlite3 test.db 'UPDATE manifests SET state  = "fetched"
		WHERE mid == "'^$i^'";' && \
	echo done $i^.manifest &
}

# Wait for everything.
wait

sqlite3 test.db <<'!!'
UPDATE manifests
SET snid = json_extract(readfile(mid || ".manifest"), '$.rootEntry.obj')
WHERE state == "fetched";

-- TODO(rjk): There's might be a better way using extensions.
-- TODO(rjk): Can I build an extension for the Apple database?
!!

# Load freshly fetched.
for (i in `{sqlite3 test.db 'select mid from manifests where state == "fetched" limit 3;'}) {
	echo loading $i^.index
	# I should use a temporary table. I can't use a temporary table unless
	# I have an extension to import into the table?
	sqlite3 test.db 'create table if not exists tfiles (fid text, _p text, snid text, path text);'
	sqlite3 test.db '.import -csv "'^$i^'.index" tfiles'
	sqlite3 test.db 'INSERT or ignore INTO files (
		fid,
		snid,
		path
	) SELECT 
		fid, snid, path
	from tfiles;
	drop table tfiles;
	UPDATE manifests SET state  = "loaded" WHERE mid == "'^$i^'";'
	rm -f $i.manifest $i.index
	echo done $i^.index
}


# select fid, mid, hostname || ":" || manifests.path ||  files.path from files inner join manifests on manifests.snid == files.snid limit 10;