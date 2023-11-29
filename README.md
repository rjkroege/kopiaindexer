Tools for indexing a [Kopia](https://kopia.io/) repository

# HowTo

```shell
export KOPIAINDEXER=$HOME/kopiaindexer

# Build the index (be patient)
mk -f $KOPIAINDEXER/mkfile
```

or in `rc`:

```shell
KOPIAINDEXER=$_h/tools/kopiaindexer  mk -f $KOPIAINDEXER/mkfile
```



# Lister
Lister is a helper tool to create a well-structured listing of all
files in a Kopia repository. This listing has the following structure:

## Columns

* filehash
* prefix
* snapshot hash
* path

##  Data Format Constraints

* URL escape the textual fields (prefix, path)
* space separate the columns. A space character is
an adequate separator because of the URL escaping.
* terminate each row with a newline

