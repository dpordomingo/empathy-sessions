Goal
====

To use PGA as a dataset for repositories analytics. Calculate size repositories, amount of blobs per repository and their sizes, rank of users activity...


TL;DR
====

To perform the analysis I followed https://docs.sourced.tech/intro
- **PGA** was properly documented and I was able to fetch filtered repositories from there,
- **gitbase** is the recommended way to analyze git repository using SQL queries, so I used it,
Then I came across **source{d}Engine** and I discovered a more flexible way to analyze the data.

My main problems were understanding the gitbase schema and the mapping between git objects and gitbase tables and references.


## TODO

- Try it large scale. I was not able to run over all PGA locally; I just tried with a tiny dataset.
- Think about how to update PGA running borges and rovers.


Log
====

## reading intro docs

I started reading what PGA is from [docs.sourced.tech/intro/#datasets](https://docs.sourced.tech/intro/#datasets) where I found @vmarkovtsev and @warenlg [PGA paper presentation](http://vmarkovtsev.github.io/msr-2018-gothenburg/#using) what helped me as a quick recap.
PGA is huge 180k repos, 3Tb in disc, 450 languages, 300 licenses, 90% repos takes 50% disc
From the [PGA blog announcement](https://blog.sourced.tech/post/announcing-pga/) I learned the difference between PGA, GHTorrent, and GHData
- [GHTorrent](http://ghtorrent.org) contains metadata; PGA contains data
- [GitHub Data](https://cloud.google.com/bigquery/public-data) contains only HEAD; PGA contains the history

**issue:** https://github.com/src-d/intro/issues/4 Links in intro doc are not visible


## pga docs

From the intro docs I navigate to https://docs.sourced.tech/datasets/publicgitarchive and to start using pga I installed it following the recommended way to do so:
```bash
$ go get github.com/src-d/datasets/PublicGitArchive/pga
```
And then download the dataset.

Since PGA dataset is too big, lets use a tity subset:
```bash
$ pga list -u /src-d/ -l Go -f json | jq -r ".sivaFilenames[]" | pga get -i -o /var/gitbase/dataset
```

**issue:** https://github.com/src-d/datasets/issues/91 Too many readme in docs
**issue:** https://github.com/src-d/datasets/issues/92 Four similar doc pages with the same/outdated content #92
**issue:** https://github.com/src-d/datasets/issues/93 Properties descriptions at web doc are outdated
**issue:** https://github.com/src-d/datasets/issues/94 There are no stats for files without lang
**issue:** https://github.com/src-d/datasets/issues/95 COMMITS_HEAD_COUNT does not appear in 'pga list'
**suggestion:** https://github.com/src-d/datasets/issues/96 Could we pipe 'list' into 'get -i'?

**notice:** Filtering by language, it will consider only languages from HEAD, so old disappeared languages (no longer in HEAD) will be ignored from the repositories.


## gitbase

As recommended at the intro docs, it will be needed gitbase to analyze the Git Repositories, so lets follow https://docs.sourced.tech/gitbase/using-gitbase/getting-started guide:
- Installing bblfshd
```bash
$ docker run -d --name bblfshd --privileged -p 9432:9432 -v /var/lib/bblfshd:/var/lib/bblfshd bblfsh/bblfshd
$ docker exec -it bblfshd bblfshctl driver install --all
```
- Installing gitbase
 ```bash
$ docker run --rm --name gitbase -p 3306:3306 --link bblfshd:bblfshd -e BBLFSH_ENDPOINT=bblfshd:9432 -v /var/gitbase/dataset:/opt/repos srcd/gitbase:latest
```

## understanding the gitbase data schema

From https://github.com/src-d/gitbase/blob/master/docs/using-gitbase/schema.md and the schema diagram provided
https://raw.githubusercontent.com/src-d/gitbase/master/docs/assets/gitbase-db-diagram.png

I'm not used to the notation of the DB scheme diagram

Since data from gitbase seems to be tied to git objects, I'll read some [docs about internal at Pro Git Book](https://git-scm.com/book/en/v2/Git-Internals-Git-Objects)

I realized that gitbase does not map 1:1 the git data structure, but it offers some tables quite similar to git objects plus extra fields (like `repository_id` everywhere), other tables that are relations between objects (like `commit_blobs`, `commit_trees`, `commit_files` or `ref_commits`) and some other tables that joins the data from other tables (`files`)...

Due to this mapping, the table/properties names can be sometimes misleading, for example:
- `files` is a table that lists regular files, but `tree_hash` points to the root `tree_entry` instead of pointing to the `tree object` containing the `blob object`.
- `tree_entries` is a table that _"contains objects that are tree objects"_; but a repository with only 2 files under the root level will be represented with 2 `tree_entries` in gitbase instead of only one git `tree object`.

**issue** https://github.com/src-d/gitbase/issues/608 Scheme diagram image is outdated


## source{d}Engine

I discovered **source{d}Engine** thanks to the video [source{d} Engine in five minutes](https://www.youtube.com/watch?v=YLxvK027Npo) from @campoy when I was finishing the day.

It's a pity that it's not documented in http://docs.sourced.tech/intro because it is a great product to do what I'm trying to do; let's give a try.

From https://docs.sourced.tech/engine/ docs it is recommended to install it from releases; I was expecting a docker image os something cool like when installing `gitbase` or `bblfsh`

I had to run twice `srcd init`
<details>
<pre>
$ srcd init /var/gitbase/dataset
INFO[0000] starting daemon with working directory: /projects/src/github.com/dpordomingo/empathy-sessions/3-pga/dataset3 
INFO[0000] installing "srcd/cli-daemon:latest"          
INFO[0005] installed "srcd/cli-daemon:latest"           
INFO[0006] couldn't find network srcd-cli-network: Error: No such network: srcd-cli-network 
INFO[0006] creating it now

$ srcd init /var/gitbase/dataset
INFO[0000] daemon already running, killing it first     
INFO[0000] starting daemon with working directory: 
</pre>
</details>

Running `srcd sql` failed too
<details>
<pre>
$ srcd sql "SHOW tables;"
2018/11/06 06:13:35 rpc error: code = Unknown desc = could not create srcd-cli-bblfshd: could not create container srcd-cli-bblfshd: Error response from daemon: invalid volume spec "srcd-cli-bblfsh-storage": invalid volume specification: 'srcd-cli-bblfsh-storage': invalid mount config for type "volume": invalid mount path: 'srcd-cli-bblfsh-storage' mount path must be absolute
</pre>
</details>

I tried many times, dropping all previous docker images, docker containers, docker volumes, `~/.srcd` content (I do not understand why there is data in my home)...
And I finally realized that I needed to install `docker-ce`, what in Ubuntu is done running:
```bash
$ curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
$ sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic test"
$ sudo apt update
$ sudo apt install docker-ce
```

**complain:** If you want to change the dataset loaded by Engine it is needed to kill the daemon and run it again because running `srcd init` with the new datasets does not take effect till you kill the daemon (what is a pain).


## analyzing the data

Are PGA contents somehow updated? where its data come from? the list was built from starred repositories at the end of Feb 2018, but is the repository content being downloaded from GitHub?
```sql
SELECT remote_fetch_url AS url,
       Max(commit_author_when) AS date
FROM   commits c
       NATURAL JOIN remotes r
GROUP  BY c.repository_id
ORDER  BY date DESC
```
The data itself is also outdated since the end of January


Final queries
====

### Repositories max/avg/sum blobs size, and count of blobs in HEAD per repo

```sql
SELECT remote_fetch_url,
       Sum(blob_size),
       Max(blob_size),
       Avg(blob_size),
       Count(*) 
FROM   files
       NATURAL JOIN commit_files
       NATURAL JOIN refs
       NATURAL JOIN remotes
WHERE  ref_name = Concat('refs/heads/master/', remote_name)
GROUP  BY remote_fetch_url
```

**complain:** It's a pain not to have a better way to access `HEAD` reference

### Repositories avg/max/min/sum size and count of repos in the whole dataset

```sql
SELECT Avg(size),
       Max(size),
       Min(size),
       Sum(size),
       Count(*)
FROM   (SELECT remote_fetch_url,
               Sum(blob_size) AS size,
               Count(*)
        FROM   files
               NATURAL JOIN commit_files
               NATURAL JOIN refs
               NATURAL JOIN remotes
        WHERE  ref_name = Concat('refs/heads/master/', remote_name)
        GROUP  BY remote_fetch_url
    ) repositories
```

### Contributors per rooted repo: unique, total commits, avg commits/contributor

```sql
SELECT repo,
       Count(1)     AS 'unique',
       Sum(commits) AS 'commits',
       Avg(commits) AS 'avg'
FROM   (SELECT c.repository_id AS repo,
               commit_author_email,
               Count(*) AS commits
        FROM   commits c
               NATURAL JOIN refs
               NATURAL JOIN remotes
        GROUP  BY c.repository_id,
                  commit_author_email
        ORDER  BY c.repository_id,
                  commits DESC
    ) stats
GROUP  BY stats.repo
```

### Contributors with more than 10 commits

```sql
SELECT *
FROM   (SELECT c.repository_id,
               commit_author_email,
               Count(*) AS commits
        FROM   commits c
               NATURAL JOIN refs
               NATURAL JOIN remotes
        GROUP  BY c.repository_id,
                  commit_author_email
        ORDER  BY c.repository_id,
                  commits DESC
    ) stats
WHERE  stats.commits > 10
```

It would be great to obtain at once the top 10 contributors for each repository, but it's not possible with SQL

**complain:** There is no `HAVING` to make the query simpler
