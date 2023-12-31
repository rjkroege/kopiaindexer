# Set KOPIAINDEXER to the location of the root of this repository.
# e.g. KOPIAINDEXER = $HOME/kopiaindexer

_kibin = $KOPIAINDEXER/bin

# Builds allmanifests: a variable defining the ids of all of the kopia manifests.
# `{head -8 manifest-listing |  awk 'BEGIN { printf("allmanifests =")} /type:snapshot/ {printf(" %s", $1)} END { printf("\n") }'}
`{kopia manifest list  |  awk 'BEGIN { printf("allmanifests =")} /type:snapshot/ {printf(" %s", $1)} END { printf("\n") }'}

# Builds the specific targets.
allfilenames = ${allmanifests:%=%.filenames}
manifestsfiles = ${allmanifests:%=%.manifest}
indexfiles = ${allmanifests:%=%.index}

%.filenames: %.index
	$_kibin/filenamearray < $stem.index > $stem.filenames

%.manifest:
	kopia manifest show $stem > $target

%.index: %.manifest
	$KOPIAINDEXER/cmd/lister/lister $stem.manifest | sort > $stem.index

# Consider making the mkfile permit keeping this in a different directory.
all:V: \
	kopia.filenames \
	kopia.fullindex

kopia.filenames: $allfilenames
	sort --merge $prereq  > $target
	
kopia.fullindex: $indexfiles
	sort --merge $prereq  > $target
