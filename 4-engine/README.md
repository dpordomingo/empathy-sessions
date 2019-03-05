# Empathy Sessions #4 source{d} Engine

_**note:** this log was originally stored at https://github.com/dpordomingo/empathy-sessions/tree/master/4-engine where you can also find some assets produced during the session._

## Goal

according to [empathy-sessions #4](https://github.com/src-d/empathy-sessions/issues/34)

1. On Ubuntu, trying `torvalds/linux` repo
1. Try source{d} Engine following its docs @ https://docs.sourced.tech/engine
1. Can I try different configurations? How difficult would it be?
1. Try LA use cases:
    1. try `srcd parse uast` over a 4MB file
    1. request `UAST` column for a large repo SQL query in gitbase

## TL;DR

What I did is to run and try Engine following the docs... and log the issues I found. Since this Empathy Session was not meant to reach another goal than just trying the tool, you won't find a clear log here, but a bunch of different things that I tried and recorded:

- [Some Things About Docs](#some-things-about-docs)
- [About Port Configuration](#about-port-configuration)
- [About Error Standardization](#about-error-standardization)
- [Running Engine Over a Big Repo](#running-engine-over-a-big-repo)
- [Getting Files Language](#getting-files-language)
- [Parsing Files With Babelfish](#parsing-files-with-babelfish)
- [Csharp Lang Is Not Working](#csharp-lang-is-not-working)
- [Clients And Connectors](#clients-and-connectors)
- [Using Gitbase Spark Connector](#using-gitbase-spark-connector)
- [Using MySQL Workbench](#using-mysql-workbench)

I could not try LA use cases, but I assume both will hit the same problem than the one described by _["Running Engine Over a Big Repo"](#running-engine-over-a-big-repo)_

# Log

## Some Things About Docs

Reading the intro I saw that it is said that:
  - It retrieves code
  - Let's you to "analyze your code through [...] REST"

But it's not correct, is it?

In the docs, the section [Guides & Examples using the source{d} Engine](https://docs.sourced.tech/engine#guides-and-examples) only lists [SonarSource Java Static Analysis Rules](https://github.com/bblfsh/sonar-checks) but it does not seem related to source{d} Engine.


What is the structure of the `gitbase` tables?
Is there any diagram that could explain it?

[Babelfish UAST](https://docs.sourced.tech/engine#babelfish-uast) section says that it can be installed a driver using `srcd parse drivers install`, but that command does not exist.


Reading architecture section, it is said:

>_This also allows us not to expose any unnecessary port, avoiding possible port conflicts. The only port that needs to be exposed is the one pointing to srcd-server which I randomly assigned to be the TCP port 4242_

But I saw that all services ports are publicly exposed.


## About Port Configuration

The docs for `--config` are not linked in the docs `README.md`; I think it could be useful.

### Init Could Check Ports

It can be ran `srcd init` with already used ports, and it won't complain:

```shell
srcd init /projects/src/github.com/torvalds --config conf/conf.default.yml 
INFO[0000] starting daemon with working directory: /projects/src/github.com/torvalds 
```

...until they're required for each component, that will cause Engine to fail

```shell
$ srcd sql "SHOW tables;"
FATA[0000] could not start gitbase: rpc error: code = Unknown desc = could not create srcd-cli-bblfshd:
could not start container: srcd-cli-bblfshd: Error response from daemon:
driver failed programming external connectivity on endpoint srcd-cli-bblfshd (8545...):
Bind for 0.0.0.0:9432 failed: port is already allocated 
```

I wonder if ports from config could be checked once Engine starts.

### Port Conflicts Could Be Less Verbose

Also, error logs are too verbose; I'd expect:

```shell
FATAL[0000] could not start gitbase. Port 9432 is already allocated.
```

### Port Error Message Doesn't Tell The Truth 

Also, the error says that the container could not be created, but it was:

```shell
$docker ps -a
CONTAINER  IMAGE                        COMMAND              CREATED  STATUS   PORTS  NAMES
fc2f408    bblfsh/bblfshd:v2.11-drivers "bblfshd -ctl-add…"  1s ago   Created         srcd-cli-bblfshd
```

but not listed as running when listing them with `srcd`

```shell
$ srcd components list
IMAGE                             INSTALLED    RUNNING    CONTAINER NAME
bblfsh/bblfshd:v2.11.8-drivers    yes          no         srcd-cli-bblfshd
```

### Passing a New Config

What's the purpose of having a `--config` param in all commands?
If I init with one config, and then I run a command with another conf, it does not take effect, even though it is logged that it is being used.

```shell
$ srcd init /path/to/scan --config conf/conf.default.yml --verbose
DEBU[0000] Using config file: /home/david/.srcd/conf.default.yml

$ srcd sql "SHOW tables;" --config conf/conf.yml --verbose
DEBU[0000] Using config file: /home/david/.srcd/conf.yml
FATA[0000] could not start gitbase: could not start container: listen tcp 0.0.0.0:3306:
bind: address already in use
```

Also, if it is passed a different conf, compared with the one used to initialize the daemon, it should raise an error, and maybe a hint recommending to `init` again, for example:

```
FATA[0000] Passed config differs with the used to initialize the daemon.
Run `srcd init` again with the new config in order to make it active.

srcd init /path/to/scan --config ~/.srcd/conf.yml --verbose
```

## About Error Standardization

Errors do not look like to follow a fixed format

### Errors From Port Conflicts

```shell
$ srcd sql "SHOW tables;"
FATA[0000] could not start gitbase: rpc error: code = Unknown desc = could not create srcd-cli-bblfshd:
could not start container: srcd-cli-bblfshd: Error response from daemon:
driver failed programming external connectivity on endpoint srcd-cli-bblfshd (3120...):
Bind for 0.0.0.0:9432 failed: port is already allocated 
```

```shell
$ srcd sql "SHOW tables;"
FATA[0000] could not start gitbase: rpc error: code = Unknown desc = could not create srcd-cli-gitbase:
could not start container: srcd-cli-gitbase: Error response from daemon:
driver failed programming external connectivity on endpoint srcd-cli-gitbase (5e3b...):
Error starting userland proxy: listen tcp 0.0.0.0:3306: bind: address already in use 
```

This one is funny: it outputs the same error log twice, but also the `srcd parse uast --help` message in the middle :)

```shell
$ srcd parse uast examples/example.php
INFO[0000] detected language: php             
Error: could not list drivers: rpc error: code = Unknown desc = could not create srcd-cli-bblfshd:
could not start container: srcd-cli-bblfshd: Error response from daemon:
driver failed programming external connectivity on endpoint srcd-cli-bblfshd (fc8a...):
Bind for 0.0.0.0:9432 failed: port is already allocated
Usage:
  srcd parse uast [file-path] [flags]

...

could not list drivers: rpc error: code = Unknown desc = could not create srcd-cli-bblfshd:
could not start container: srcd-cli-bblfshd: Error response from daemon:
driver failed programming external connectivity on endpoint srcd-cli-bblfshd (fc8a...):
Bind for 0.0.0.0:9432 failed: port is already allocated
```

### Error When RPC Message Is Too Big

```shell
$ srcd sql "SELECT cf.file_path, f.blob_content
FROM ref_commits r
NATURAL JOIN commit_files cf
NATURAL JOIN files f
WHERE r.ref_name = 'HEAD'
AND r.history_index = 0;"

2019/03/04 10:55:11 rpc error: code = ResourceExhausted desc = grpc:
                    received message larger than max (5175028 vs. 4194304)
```

### Strange WARN Message

This `WARN` is returned always. What does it mean? What can I do as a user?

```shell
$ srcd init /projects/src/github.com/torvalds
WARN[0001] unable to list the available daemon versions on Docker Hub: 
can't find compatible image in docker registry for srcd/cli-daemon 
```

If the last message after a `init` is a `WARN`, the user can think that something went wrong; why not adding a succeeds message?


## Running Engine Over a Big Repo

I ran first example from docs, and it took maaany time...
```shell
$ srcd sql "SELECT cf.file_path, f.blob_content
FROM ref_commits r
NATURAL JOIN commit_files cf
NATURAL JOIN files f
WHERE r.ref_name = 'HEAD'
AND r.history_index = 0;"

2019/03/04 10:55:11 rpc error: code = ResourceExhausted desc = grpc:
                    received message larger than max (5175028 vs. 4194304)
```

and after ~2mins with no output about what was happening, it returned an error. I wonder if we could let the user know that something is being processed.

The count over that files took ~1min, and it succeeded, returning a count of `63135` files in `HEAD`

```shell
$ srcd sql "SELECT count(*)
FROM ref_commits r
NATURAL JOIN commit_files cf
NATURAL JOIN files f
WHERE r.ref_name = 'HEAD'
AND r.history_index = 0;"

+----------+
| COUNT(*) |
+----------+
|    63135 |
+----------+
```

Removing the content column, it took ~1.5min, and returned the file names.

```shell
$ srcd sql "SELECT cf.file_path
FROM ref_commits r
NATURAL JOIN commit_files cf
NATURAL JOIN files f
WHERE r.ref_name = 'HEAD'
AND r.history_index = 0;"

+--------------------------+
|        FILE PATH         |
+--------------------------+
| .clang-format            |
| .cocciconfig             |
| .get_maintainer.ignore   |
| .gitattributes           |
| .gitignore               |
| .mailmap                 |
| COPYING                  |
| CREDITS                  |
| Documentation/.gitignore |
| Documentation/ABI/README | 
| ...                      |
+--------------------------+
```

## Getting Files Language

Trying to get files language from `HEAD` is a pain:

```shell
$ srcd sql "SELECT cf.file_path,
    LANGUAGE(f.file_path,  f.blob_content) as lang
FROM ref_commits r
NATURAL JOIN commit_files cf
NATURAL JOIN files f
WHERE r.ref_name = 'HEAD'
AND r.history_index = 0;"
```

Isn't there a way to get files from `HEAD` without adding `history_index = 0` everytime?

On big repos it is very slow: it took 4 mins on `linux` repo to get the languages at `HEAD`.

Could it be useful to cache or to precompute it?

What if many queries need lang? grouping, filtering... that if they're calculated everytime, it takes a lot.

```shell
$ srcd sql "SELECT count(*) as count,
    LANGUAGE(f.file_path,  f.blob_content) as lang
FROM ref_commits r
NATURAL JOIN commit_files cf
NATURAL JOIN files f
WHERE r.ref_name = 'HEAD'
AND r.history_index = 0
GROUP BY lang ORDER BY count desc;"

+-------+------------------+
| COUNT |       LANG       |
+-------+------------------+
| 45213 | C                |
|  6356 |                  |
|  4409 | Text             |
|  2607 | Makefile         |
|  1329 | Unix Assembly    |
|  1068 | reStructuredText |
|   625 | C++              |
+   ... | ...              +
+-------+------------------+
```

## Parsing Files With Babelfish

If you choose a nonexistent driver, it outputs the help message, but it should list the available languages instead.

```shell
$ srcd parse uast examples/example.php --lang unknown
Error: language unknown is not supported
Usage:
  srcd parse uast [file-path] [flags]
...
```

Also, why the message _"while we install it"_ if `srcd` uses a `bblfsh` image with already builtin drivers?

```shell
$ srcd parse uast examples/example.php
INFO[0003] if this is the first time using a driver for a language,
this might take a few more minutes while we install it 
```


## Csharp Lang Is Not Working

Running parse over a `.cs` file is not working

```shell
$ srcd parse uast examples/example.cs
INFO[0000] detected language: c#                        
Error: language c# is not supported
Usage:
  srcd parse uast [file-path] [flags]
...
```

Running parse over a `.cs` file, but asking for csharp lang takes forever:

```shell
$ srcd parse uast examples/example.cs --lang csharp
INFO[0003] if this is the first time using a driver for a language,
this might take a few more minutes while we install it 
```

but if you check the container's logs there is a panic that is not exported in Engine logs, what is wrong.

```shell
docker logs -f srcd-cli-bblfshd

panic: invalid tracer configuration: lookup localhost on 1.1.1.1:53: no such host

goroutine 1 [running]:
github.com/bblfsh/csharp-driver/vendor/gopkg.in/bblfsh/sdk.v2/driver/server.RunNative(0xba9260, 0xc4200be180, 0xb139e9, 0x6, 0xc420357ac0, 0x1, 0x1, 0xf0a260, 0x2, 0x2, ...)
	/go/src/github.com/bblfsh/csharp-driver/vendor/gopkg.in/bblfsh/sdk.v2/driver/server/common.go:35 +0x162
github.com/bblfsh/csharp-driver/vendor/gopkg.in/bblfsh/sdk.v2/driver/server.Run(0xb139e9, 0x6, 0xc420357ac0, 0x1, 0x1, 0xf0a260, 0x2, 0x2, 0xc4203bd5a0, 0x2, ...)
	/go/src/github.com/bblfsh/csharp-driver/vendor/gopkg.in/bblfsh/sdk.v2/driver/server/common.go:20 +0x74
main.main()
	/go/src/github.com/bblfsh/csharp-driver/driver/main.go:11 +0x5a
```

and if you `CTRL+C` on `srcd parse uast` tab, the container will log:

```
time="2019-03-04T12:34:37Z" level=error msg="error selecting pool:
                            unexpected error: context canceled"
time="2019-03-04T12:34:37Z" level=error msg="request processed content 348 bytes error:
                            unexpected error: context canceled" elapsed=13.66024993s
                            filename="/projects/src/github.com/src-d/lookout/doc.go"
                            language=csharp
```

it also happens using plain `bblfsh`, so I assume it's a problem with Babelfish driver. 

```shell
$ docker exec -it srcd-cli-bblfshd /bin/sh
# bblfshctl driver list
| csharp | docker://bblfsh/csharp-driver:latest | v1.4.0 | beta | 12 days |  | 1.10 | microsoft/dotnet:2.1-runtime |
# bblfshctl parse example.cs
```

And it does not answer. And if you `CTRL+C` the container seems to be broken, because then you're not even able to run `parse` again on different languages.

The curious thing is that parsing `csharp` from Engine does not break the `bblfsh` container; it does not parse that language, but it lets you to parse other languages after you stop the `csharp` query.



## Clients And Connectors

In [Clients and Connectors](https://docs.sourced.tech/engine#clients-and-connectors) section from docs, it is said that
>_to connect to the language parsing server (Babelfish) and analyzing the UAST, there are several language clients_

I think it should be mentioned that:
- it is feasible because Engine exposes `gitbase` and `bblfsh` containers under certain ports
- `gitbase` is also reachable with many [mysql clients](https://github.com/src-d/go-mysql-server/blob/master/SUPPORTED_CLIENTS.md)
  - it could be useful an example:
    ```shell
    $ mysql \
      --user=root \
      --host=127.0.0.1 \
      --port=3306 \
      --execute "SHOW tables"
    ```

But when using these kinds of alternatives to connect to engine services, it is [required that you run the service in advance](https://github.com/src-d/engine/blob/master/docs/commands.md#srcd-components-start), but that command does not exist (yet).

That problem can be workarounded running a `sql` command to let the daemon start the required services:

```shell
$ srcd ini
$ srcd sql "select *"
```

Doing so, I was able to run queries over `bblfsh` and `gitbase` containers using `go-client` and `python` commands.

I think it could be provided [such examples](https://github.com/dpordomingo/empathy-sessions/tree/master/4-engine/cmd).


## Using Gitbase Spark Connector

I was able to run https://github.com/src-d/gitbase-spark-connector against engine containers :tada:

```shell
$ srcd ini
$ srcd sql "select *"
$ docker run --rm
  --publish 8080:8080
  --env BBLFSH_HOST=bblfshd
  --env BBLFSH_PORT=9432
  --env GITBASE_SERVERS=gitbase:3306
  --link srcd-cli-bblfshd:bblfshd
  --link srcd-cli-gitbase:gitbase
  --name=spark
  srcd/gitbase-spark-connector-jupyter:latest
```

and then run from http://localhost:8080 the notebook [gitbase.example.ipynb](https://github.com/dpordomingo/empathy-sessions/tree/master/4-engine/pynotebook/gitbase.example.ipynb)

```scala
import tech.sourced.gitbase.spark.GitbaseSessionBuilder
val spark = SparkSession.builder().appName("example")
    .master("local[*]")
    .config("spark.driver.host", "localhost")
    .registerGitbaseSource("gitbase:3306")
    .getOrCreate()
spark.sql("SELECT * FROM repositories").show()
spark.sql("""SELECT
        file_path,
        LANGUAGE(file_path, blob_content) as lang
    FROM files
    WHERE LANGUAGE(file_path, blob_content)='Go'
    LIMIT 1""").show()
```

This last query raises an error, but I think it's out of the purpose of this empathy session

```scala
spark.sql("""SELECT
        file_path,
        LANGUAGE(file_path, blob_content) as lang,
        UAST(blob_content, LANGUAGE(file_path, blob_content)) as uast
    FROM files
    WHERE LANGUAGE(file_path, blob_content)='Go'
    LIMIT 1""").show()
```

<details>
<pre>
Name: java.lang.ClassCastException
Message: tech.sourced.gitbase.spark.udf.Uast$$anonfun$function$1 cannot be cast to scala.Function2
StackTrace:   at org.apache.spark.sql.catalyst.expressions.ScalaUDF.<init>(ScalaUDF.scala:107)
  at org.apache.spark.sql.expressions.UserDefinedFunction.apply(UserDefinedFunction.scala:71)
  at org.apache.spark.sql.UDFRegistration.org$apache$spark$sql$UDFRegistration$$builder$2(UDFRegistration.scala:104)
  at org.apache.spark.sql.UDFRegistration$$anonfun$register$2.apply(UDFRegistration.scala:105)
  at org.apache.spark.sql.UDFRegistration$$anonfun$register$2.apply(UDFRegistration.scala:105)
  at org.apache.spark.sql.catalyst.analysis.SimpleFunctionRegistry.lookupFunction(FunctionRegistry.scala:115)
  at org.apache.spark.sql.catalyst.catalog.SessionCatalog.lookupFunction(SessionCatalog.scala:1216)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$$anonfun$apply$16$$anonfun$applyOrElse$6$$anonfun$applyOrElse$53.apply(Analyzer.scala:1244)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$$anonfun$apply$16$$anonfun$applyOrElse$6$$anonfun$applyOrElse$53.apply(Analyzer.scala:1244)
  at org.apache.spark.sql.catalyst.analysis.package$.withPosition(package.scala:53)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$$anonfun$apply$16$$anonfun$applyOrElse$6.applyOrElse(Analyzer.scala:1243)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$$anonfun$apply$16$$anonfun$applyOrElse$6.applyOrElse(Analyzer.scala:1227)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$2.apply(TreeNode.scala:267)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$2.apply(TreeNode.scala:267)
  at org.apache.spark.sql.catalyst.trees.CurrentOrigin$.withOrigin(TreeNode.scala:70)
  at org.apache.spark.sql.catalyst.trees.TreeNode.transformDown(TreeNode.scala:266)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$transformDown$1.apply(TreeNode.scala:272)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$transformDown$1.apply(TreeNode.scala:272)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$4.apply(TreeNode.scala:306)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapProductIterator(TreeNode.scala:187)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapChildren(TreeNode.scala:304)
  at org.apache.spark.sql.catalyst.trees.TreeNode.transformDown(TreeNode.scala:272)
  at org.apache.spark.sql.catalyst.plans.QueryPlan$$anonfun$transformExpressionsDown$1.apply(QueryPlan.scala:85)
  at org.apache.spark.sql.catalyst.plans.QueryPlan$$anonfun$transformExpressionsDown$1.apply(QueryPlan.scala:85)
  at org.apache.spark.sql.catalyst.plans.QueryPlan$$anonfun$1.apply(QueryPlan.scala:107)
  at org.apache.spark.sql.catalyst.plans.QueryPlan$$anonfun$1.apply(QueryPlan.scala:107)
  at org.apache.spark.sql.catalyst.trees.CurrentOrigin$.withOrigin(TreeNode.scala:70)
  at org.apache.spark.sql.catalyst.plans.QueryPlan.transformExpression$1(QueryPlan.scala:106)
  at org.apache.spark.sql.catalyst.plans.QueryPlan.org$apache$spark$sql$catalyst$plans$QueryPlan$$recursiveTransform$1(QueryPlan.scala:118)
  at org.apache.spark.sql.catalyst.plans.QueryPlan$$anonfun$org$apache$spark$sql$catalyst$plans$QueryPlan$$recursiveTransform$1$1.apply(QueryPlan.scala:122)
  at scala.collection.TraversableLike$$anonfun$map$1.apply(TraversableLike.scala:234)
  at scala.collection.TraversableLike$$anonfun$map$1.apply(TraversableLike.scala:234)
  at scala.collection.immutable.List.foreach(List.scala:381)
  at scala.collection.TraversableLike$class.map(TraversableLike.scala:234)
  at scala.collection.immutable.List.map(List.scala:285)
  at org.apache.spark.sql.catalyst.plans.QueryPlan.org$apache$spark$sql$catalyst$plans$QueryPlan$$recursiveTransform$1(QueryPlan.scala:122)
  at org.apache.spark.sql.catalyst.plans.QueryPlan$$anonfun$2.apply(QueryPlan.scala:127)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapProductIterator(TreeNode.scala:187)
  at org.apache.spark.sql.catalyst.plans.QueryPlan.mapExpressions(QueryPlan.scala:127)
  at org.apache.spark.sql.catalyst.plans.QueryPlan.transformExpressionsDown(QueryPlan.scala:85)
  at org.apache.spark.sql.catalyst.plans.QueryPlan.transformExpressions(QueryPlan.scala:76)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$$anonfun$apply$16.applyOrElse(Analyzer.scala:1227)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$$anonfun$apply$16.applyOrElse(Analyzer.scala:1225)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$transformUp$1.apply(TreeNode.scala:289)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$transformUp$1.apply(TreeNode.scala:289)
  at org.apache.spark.sql.catalyst.trees.CurrentOrigin$.withOrigin(TreeNode.scala:70)
  at org.apache.spark.sql.catalyst.trees.TreeNode.transformUp(TreeNode.scala:288)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$3.apply(TreeNode.scala:286)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$3.apply(TreeNode.scala:286)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$4.apply(TreeNode.scala:306)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapProductIterator(TreeNode.scala:187)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapChildren(TreeNode.scala:304)
  at org.apache.spark.sql.catalyst.trees.TreeNode.transformUp(TreeNode.scala:286)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$3.apply(TreeNode.scala:286)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$3.apply(TreeNode.scala:286)
  at org.apache.spark.sql.catalyst.trees.TreeNode$$anonfun$4.apply(TreeNode.scala:306)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapProductIterator(TreeNode.scala:187)
  at org.apache.spark.sql.catalyst.trees.TreeNode.mapChildren(TreeNode.scala:304)
  at org.apache.spark.sql.catalyst.trees.TreeNode.transformUp(TreeNode.scala:286)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$.apply(Analyzer.scala:1225)
  at org.apache.spark.sql.catalyst.analysis.Analyzer$ResolveFunctions$.apply(Analyzer.scala:1224)
  at org.apache.spark.sql.catalyst.rules.RuleExecutor$$anonfun$execute$1$$anonfun$apply$1.apply(RuleExecutor.scala:87)
  at org.apache.spark.sql.catalyst.rules.RuleExecutor$$anonfun$execute$1$$anonfun$apply$1.apply(RuleExecutor.scala:84)
  at scala.collection.LinearSeqOptimized$class.foldLeft(LinearSeqOptimized.scala:124)
  at scala.collection.immutable.List.foldLeft(List.scala:84)
  at org.apache.spark.sql.catalyst.rules.RuleExecutor$$anonfun$execute$1.apply(RuleExecutor.scala:84)
  at org.apache.spark.sql.catalyst.rules.RuleExecutor$$anonfun$execute$1.apply(RuleExecutor.scala:76)
  at scala.collection.immutable.List.foreach(List.scala:381)
  at org.apache.spark.sql.catalyst.rules.RuleExecutor.execute(RuleExecutor.scala:76)
  at org.apache.spark.sql.catalyst.analysis.Analyzer.org$apache$spark$sql$catalyst$analysis$Analyzer$$executeSameContext(Analyzer.scala:124)
  at org.apache.spark.sql.catalyst.analysis.Analyzer.execute(Analyzer.scala:118)
  at org.apache.spark.sql.catalyst.analysis.Analyzer.executeAndCheck(Analyzer.scala:103)
  at org.apache.spark.sql.execution.QueryExecution.analyzed$lzycompute(QueryExecution.scala:57)
  at org.apache.spark.sql.execution.QueryExecution.analyzed(QueryExecution.scala:55)
  at org.apache.spark.sql.execution.QueryExecution.assertAnalyzed(QueryExecution.scala:47)
  at org.apache.spark.sql.Dataset$.ofRows(Dataset.scala:74)
  at org.apache.spark.sql.SparkSession.sql(SparkSession.scala:642)
</pre>
</details>


## Using MySQL Workbench

I tried to connect to `gitbase` using [MySQL Workbench 6.3](https://dev.mysql.com/downloads/workbench) (even it's not explicitly supported by `gitbase`) and I found some issues:

- `current_user()` is not supported
- `show status` is not supported
- `show engines` is not supported

I could mock data for all of them in `gitbase`, and doing so Workbench was able to start with a warning:

```
Incompatible/nonstandard server version or connection protocol detected ().

A connection to this database can be established but some MySQL Workbench features may not work
properly since the database is not fully compatible with the supported versions of MySQL.

MySQL Workbench is developed and tested for MySQL Server versions 5.1, 5.5, 5.6 and 5.7
```

And there is no tables shown in the left panel, plus a log:
```sql
# 15:32:31  Error loading schema content
#           Error Code: 0
#           MySQL_ResultSet::getString: invalid value of 'columnIndex'
```

Fetching commits also fails:
```sql
select * from commits
# 15:35:27  select * from commits LIMIT 0, 10
#           Fetching...
# 15:35:27  select * from commits LIMIT 0, 10
#           Error Code: 0
#           0) . Please reportn charsetnr (
```
