#!/usr/local/bin/rc

idxdir = $HOME/.kopiaindex
db = $idxdir/kopiaindex.db

# Fetch an updated manifest list.
kopia manifest list --json > $idxdir/manifest.json

sqlite3 $db 'create table  if not exists manifests (
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
create index if not exists manifests_snid_idx on manifests(snid);
-- now there''s a manifests table

-- this can be a temporary table
create table  if not exists tmp_manifests (
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
-- and a tmp manifests table that I will populate from from manifests list

INSERT or ignore INTO tmp_manifests (
	mid,
	hostname,
	path,
	type,
	username,
	mtime
) SELECT 
  json_extract(value, "$.id"), 
  json_extract(value, "$.labels.hostname"),
  json_extract(value, "$.labels.path"),
  json_extract(value, "$.labels.type"),
  json_extract(value, "$.labels.username"),
  json_extract(value, "$.mtime")
FROM json_each(readfile("'^$idxdir^'/manifest.json"));

-- mark the deleted manifest entries
UPDATE manifests SET state  = "deleted" WHERE mid not in (SELECT mid FROM tmp_manifests);

-- add the new items
INSERT or ignore INTO manifests (
	mid,
	hostname,
	path,
	type,
	username,
	mtime
) SELECT mid, hostname, path, type, username, mtime
FROM tmp_manifests;

-- create the files table
create table  if not exists files (
	id INTEGER PRIMARY KEY,
	fid TEXT ,
	snid TEXT,
	path TEXT
);

-- TODO(rjk): Add indicies as needed. Probably on fid?
-- TODO(rjk): Adjust the indexing of the path to address wanting to sub-string it
-- TODO(rjk): How do I figure out what are the right indicies
-- TODO(rjk): explore creating a suffix tree
create  index if not exists files_fid_idx on files(fid);
create  index if not exists files_path_idx on files(path);

-- drop the temp table
drop table tmp_manifests;

-- Remove all of the items in files corresponding to deleted manifests
delete  from files where snid in (select snid from manifests where state == "deleted");
delete from manifests where state == "deleted";
'

# Note that I should add the snapshot id too because it's really handy.
# And I need to index on it. And it will let me do what I want with the right
# magic query.
for (i in `{sqlite3 $db 'select mid from manifests where type == "snapshot" and state ISNULL;'}) {
	echo starting $idxdir/$i^.manifest && \
	kopia manifest show $i > $idxdir/$i^.manifest && \
	$KOPIAINDEXER/cmd/lister/lister $idxdir/$i^.manifest | sed 's/ /,/g' > $idxdir/$i^.index && \
	sqlite3 $db 'UPDATE manifests SET state  = "fetched"
		WHERE mid == "'^$i^'";' && \
	echo done $idxdir/$i^.manifest &
}

# Wait for everything.
wait

sqlite3 $db  'UPDATE manifests
SET snid = json_extract(readfile("'^$idxdir^'/" || mid || ".manifest"), "$.rootEntry.obj")
WHERE state == "fetched";
'

# Load freshly fetched.
for (i in `{sqlite3 $db 'select mid from manifests where state == "fetched";'}) {
	echo loading $idxdir/$i^.index
	# TODO(rjk): There is a better way with a Sqlite extension and temporary tables.
	sqlite3 $db 'create table if not exists tfiles (fid text, _p text, snid text, path text);'
	sqlite3 $db '.import -csv "'^$idxdir/$i^'.index" tfiles'
	sqlite3 $db 'INSERT or ignore INTO files (
		fid,
		snid,
		path
	) SELECT 
		fid, snid, path
	from tfiles;
	drop table tfiles;
	UPDATE manifests SET state  = "loaded" WHERE mid == "'^$i^'";'
	rm -f $idxdir/$i^.manifest $idxdir/$i^.index
	echo done $idxdir/$i^.index
}

# select fid, mid, hostname || ":" || manifests.path ||  files.path from files inner join manifests on manifests.snid == files.snid limit 10;