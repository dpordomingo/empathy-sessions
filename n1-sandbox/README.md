# Empathy Sessions source{d} Sandbox CE

## Goal

Since this session was not part of an official empathy session, I just tried to use source{d} UI a bit from a user point of view.


## TL;DR

What I did is to run and try source{d} Sandbox CE following its docs...
It worked pretty well, but when trying some edge cases, or advanced use cases related with users and dashboards I found some issues;
here is the list of things that I feel could be improved:

- [About Docs](#about-docs)
- [About Sandbox Commands and Workflow](#about-sandbox-commands-and-workflow)
  - [Failures When Installing Sandbox](#failures-when-installing-sandbox)
  - [Running UI With Broken Sandbox](#running-ui-with-broken-sandbox)
- [About the UI (superset)](#about-the-ui-superset)
  - [Users](#users)
  - [SQL Lab. SQL Editor](#sql-lab-sql-editor)
  - [Dashboards and Charts](#dashboards-and-charts)
- [About Logs](#about-logs)

I'm not sure if some of those issues are already scheduled to be fixed, so I'll review them in a while to check if they were, or they were not.


# Log

## About Docs

- [ ] I miss proper docs, hosted in gitbook, as usually done for other company projects.

- [ ] Project `README.md` says that the default user/pass is `admin`, but you can not use default pass: you need to create one; I'd also avoid referring to `admin` as a default user, because when you install the sandbox that's the first thing you can change.
  In another part of the docs it is also said something wrong:
  >_Use login: admin and password admin to access it_

- [ ] According to the project `README.md`, the `sandbox-ce start` command expects `GITBASE_REPOS_DIR`, but it is ignored and uses the previous one. Also in `stop` command, that would not expect it, warns you if you miss it:
  ```bash
  $ sandbox-ce stop
  WARNING: The GITBASE_REPOS_DIR variable is not set. Defaulting to a blank string.
  ```

- [ ] Also `sandbox-ce status` and `sandbox-ce web` does not exist even though project `README.md` says so.

- [ ] The `sandbox prune --images` flag is not available even though project `README.md` says so.

- [ ] Project `README.md` also seems outdated when asserting:
  >_After the initialization [...] It will automatically open WebUI_


## About Sandbox Commands and Workflow

- [ ] There is no valuable info in `--help` messages: `install` _Installs_, `stop` _Stops_ and so on.

- [ ] I miss `sandbox prune --images` or `--all`, to delete all images.

- [ ] The `sandbox-ce install --help` message should say that it expects the DB working dir under `GITBASE_REPOS_DIR` env var.


### Failures When Installing Sandbox

- [ ] If you try to install superset without specifying a working directory, it should default in `.`, but it does not, and it fails.
  ```bash
  $ sandbox-ce install
  WARNING: The GITBASE_REPOS_DIR variable is not set. Defaulting to a blank string.
  Creating network "src_default" with the default driver
  Creating volume "src_postgres" with default driver
  Creating volume "src_redis" with default driver
  Creating src_bblfsh_1   ... done
  Creating src_postgres_1 ... done
  Creating src_redis_1    ... done
  Creating src_gitbase_1  ... error
  ERROR: for src_gitbase_1  Cannot create container for service gitbase: create .: volume name is too short, names should be at least two alphanumeric characters
  
  ERROR: for gitbase  Cannot create container for service gitbase: create .: volume name is too short, names should be at least two alphanumeric characters
  ERROR: Encountered errors while bringing up the project.
  exit status 1
  ```

- [ ] Also, even with a non-running sandbox, some services are still running
  ```bash
  $ docker ps -a
  CONTAINER ID  IMAGE                           COMMAND                 CREATED     STATUS     PORTS       NAMES
  50452a1831bd  redis:3.2                       "docker-entrypoint.s…"  24 sec ago  Up 22 sec  6379->6379  src_redis_1
  da8aa8d5e40d  bblfsh/bblfshd:v2.13.0-drivers  "bblfshd"               24 sec ago  Up 23 sec  9432->9432  src_bblfsh_1
  be5813774c50  postgres:10                     "docker-entrypoint.s…"  24 sec ago  Up 23 sec  5432->5432  src_postgres_1
  ```


### Running UI With Broken Sandbox

- [ ] In one circumstance, stopping sandbox, and then starting it again, caused a broken one with a no usable bblfsh service.
(I could not reproduce, maybe LA team will know in which circumstances bblfsh can be stuck when starting its container after being stopped or recreated or deleted)
  ```bash
  $ sandbox-ce stop
  Stopping src_superset_1   ... done
  Stopping src_bblfsh-web_1 ... done
  Stopping src_gitbase_1    ... done
  Stopping src_postgres_1   ... done
  Stopping src_redis_1      ... done
  
  $ docker ps -a
  CONTAINER ID  IMAGE                           COMMAND                 CREATED     STATUS             NAMES
  d1fbf9bb0d5a  srcd/superset:latest            "/entrypoint.sh"        42 min ago  Exited 47 sec ago  src_superset_1
  f2144b82912d  bblfsh/web:v0.11.0              "/bin/bblfsh-web -ad…"  42 min ago  Exited 57 sec ago  src_bblfsh-web_1
  c79f5a7ccd03  srcd/gitbase:v0.20.0-beta4      "./init.sh"             48 min ago  Exited 36 sec ago  src_gitbase_1
  eae6ccdc6f40  redis:3.2                       "docker-entrypoint.s…"  1 hour ago  Exited 46 sec ago  src_redis_1
  116bbcdf78db  postgres:10                     "docker-entrypoint.s…"  1 hour ago  Exited 46 sec ago  src_postgres_1
  9a6351fd85af  bblfsh/bblfshd:v2.13.0-drivers  "bblfshd"               1 hour ago  Exited 50 sec ago  src_bblfsh_1
  
  $ GITBASE_REPOS_DIR=./github.com/dpordomingo sandbox-ce start
  Starting bblfsh     ... done
  Starting gitbase    ... done
  Starting bblfsh-web ... done
  Starting redis      ... done
  Starting postgres   ... done
  Starting superset   ... done
  
  $ docker ps -a
  CONTAINER ID  IMAGE                           COMMAND                 CREATED     STATUS          PORTS       NAMES
  d1fbf9bb0d5a  srcd/superset:latest            "/entrypoint.sh"        43 min ago  Up 3 seconds    8088->8088  src_superset_1
  f2144b82912d  bblfsh/web:v0.11.0              "/bin/bblfsh-web -ad…"  43 min ago  Up 3 seconds    9999->8080  src_bblfsh-web_1
  c79f5a7ccd03  srcd/gitbase:v0.20.0-beta4      "./init.sh"             48 min ago  Up 3 seconds    3306->3306  src_gitbase_1
  eae6ccdc6f40  redis:3.2                       "docker-entrypoint.s…"  1 hour ago  Up 4 seconds    6379->6379  src_redis_1
  116bbcdf78db  postgres:10                     "docker-entrypoint.s…"  1 hour ago  Up 4 seconds    5432->5432  src_postgres_1
  9a6351fd85af  bblfsh/bblfshd:v2.13.0-drivers  "bblfshd"               1 hour ago  Exit 4 sec ago              src_bblfsh_1
  ```

  I realized that bblfsh was crashing silently:
  ```bash
  $ docker logs src_bblfsh_1
  time="2019-05-30T13:38:47Z" level=info msg="bblfshd version: v2.13.0 (build: 2019-05-03T14:12:51+0000)"
  time="2019-05-30T13:38:47Z" level=info msg="running metrics on :2112"
  time="2019-05-30T13:38:47Z" level=info msg="initializing runtime at /var/lib/bblfshd"
  time="2019-05-30T13:38:47Z" level=info msg="server listening in 0.0.0.0:9432 (tcp)"
  time="2019-05-30T13:38:47Z" level=fatal msg="error creating control listener: listen unix /var/run/bblfshctl.sock: bind: address already in use"
  ```

  IMO, sandbox should panic if any of its components fails, instead of letting the user run the UI, and getting errors when it tries to use bblfsh features.


## About the UI (superset)

- [ ] When you install the sandbox for the very first time, and then you go to `http://0.0.0.0:8088`, the initial page (`http://0.0.0.0:8088/superset/welcome`) shows an empty list and a "loading animation" that keeps loading forever.
An alternative could be to make the user profile (`http://0.0.0.0:8088/superset/profile/david/`) the landing page, that contains proper data.

- [ ] I'd expect a fixed header (logo, manage, SQL lab, uast...); when you're in a long page (in a dashboard with many charts, for example).

- [ ] While you're browsing, you end up with many superset tabs opened at once.


### Users

- [ ] When you install a sandbox, what it does is to change the working dir and to add a new admin user, keeping previous users and dashboards.
I find a bit strange that what `install` command does, is to add a new admin use.

- [ ] Admin users cannot add charts nor modify other users dashboards, which could be useful in some circumstances.

- [ ] Non-admin users can create their own dashboards, or copy existent ones, but they can not import dashboards from files, which could be useful in regular usage.


### SQL Lab. SQL Editor

- [ ] If I open many table schemas (from schema combo on the aside), and uncollapse all their properties, I can not scroll there to see all schemas properties.

- [ ] It would be great if clicking in one schema property, it would be pasted in SQL editor. Also if there were a direct way to run `SELECT * FROM table` instead of using `copy` from the icon, and then replacing current SQL query with the content of the clipboard.


### Dashboards and Charts

- [ ] Dashboards could not be exported in Chrome because they were blocked by the default Chrome popup blocker.

- [ ] When using a different dashboard, created from other working dir, its data is loaded from the previous state; you need to "force refresh" to use data from current working dir".

- [ ] The `word cloud` chart, when is in a dashboard, it depends on chart width to work properly: big words disappear if they do not fit in the chart, what is wrong.

- [ ] I could not import any dashboard:  
  I created a simple chart (a piechart grouping remotes by name), then I saved it in a new dashboard, and then I exported it into a file from dashboard `export` action. Then I tried to import it from `Manage > Import` menu in different sandboxes:
  - stopping current sandbox, and installing a brand new one with the same user info and a different working directory. 
  - pruning, and installing a new sandbox with the same user info and working directory.

  I always got a 500 error:
  ```
  Traceback (most recent call last):
    File "/usr/local/lib/python3.6/site-packages/flask/app.py", line 2292, in wsgi_app
      response = self.full_dispatch_request()
    File "/usr/local/lib/python3.6/site-packages/flask/app.py", line 1815, in full_dispatch_request
      rv = self.handle_user_exception(e)
    File "/usr/local/lib/python3.6/site-packages/flask/app.py", line 1718, in handle_user_exception
      reraise(exc_type, exc_value, tb)
    File "/usr/local/lib/python3.6/site-packages/flask/_compat.py", line 35, in reraise
      raise value
    File "/usr/local/lib/python3.6/site-packages/flask/app.py", line 1813, in full_dispatch_request
      rv = self.dispatch_request()
    File "/usr/local/lib/python3.6/site-packages/flask/app.py", line 1799, in dispatch_request
      return self.view_functions[rule.endpoint](**req.view_args)
    File "/home/superset/superset/models/core.py", line 1158, in wrapper
      value = f(*args, **kwargs)
    File "/usr/local/lib/python3.6/site-packages/flask_appbuilder/security/decorators.py", line 26, in wraps
      return f(self, *args, **kwargs)
    File "/home/superset/superset/views/core.py", line 1239, in import_dashboards
      dashboard_import_export.import_dashboards(db.session, f.stream)
    File "/home/superset/superset/utils/dashboard_import_export.py", line 37, in import_dashboards
      dashboard, import_time=import_time)
    File "/home/superset/superset/models/core.py", line 586, in import_obj
      alter_positions(dashboard_to_import, old_to_new_slc_id_dict)
    File "/home/superset/superset/models/core.py", line 525, in alter_positions
      position_data = json.loads(dashboard.position_json)
    File "/usr/local/lib/python3.6/json/__init__.py", line 348, in loads
      'not {!r}'.format(s.__class__.__name__))
  TypeError: the JSON object must be str, bytes or bytearray, not 'NoneType'
  ```


## About Logs

- [ ] When using `sandbox-ce start`, there is no compose logs as when started with `sandbox-ce install`

- [ ] When using `sandbox-ce install`, using the same user data (I think the `unique` key is the email), it is logged `No user created an error occured` error message, but it does not impact the initialization. IMO, if the user is using defaults, this error should not be raised. 
  ```bash
  $ sandbox-ce install
  ...
  2019-05-30 13:42:41,482:ERROR:flask_appbuilder.security.sqla.manager:Error adding new user to database. (psycopg2.IntegrityError) duplicate key value violates unique constraint "ab_user_email_key"
  DETAIL:  Key (email)=(admin@fab.org) already exists.
  
  [SQL: INSERT INTO ab_user (id, first_name, last_name, username, password, active, email, last_login, login_count, fail_login_count, created_on, changed_on, created_by_fk, changed_by_fk) VALUES (nextval('ab_user_id_seq'), %(first_name)s, %(last_name)s, %(username)s, %(password)s, %(active)s, %(email)s, %(last_login)s, %(login_count)s, %(fail_login_count)s, %(created_on)s, %(changed_on)s, %(created_by_fk)s, %(changed_by_fk)s) RETURNING ab_user.id]
  [parameters: {'first_name': 'david', 'last_name': 'p', 'username': 'admin', 'password': '...', 'active': True, 'email': 'admin@fab.org', 'last_login': None, 'login_count': None, 'fail_login_count': None, 'created_on': datetime.datetime(2019, 5, 30, 13, 42, 41, 480550), 'changed_on': datetime.datetime(2019, 5, 30, 13, 42, 41, 480562), 'created_by_fk': None, 'changed_by_fk': None}]
  (Background on this error at: http://sqlalche.me/e/gkpj)
  No user created an error occured
  ```

- [ ] I got many python errors when installing sandbox;
  Docs says that it will disappear when https://github.com/src-d/gitbase/issues/808 is closed, but it was already merged, but the python errors are still there
  ```
  ...
  + python add_gitbase.py
  Loaded your LOCAL configuration at [/home/superset/superset/superset_config.py]
  2019-05-31 06:47:06,253:INFO:root:Database.get_sqla_engine(). Masked URL: mysql://root:XXXXXXXXXX@gitbase:3306/gitbase
  /usr/local/lib/python3.6/site-packages/sqlalchemy/dialects/mysql/reflection.py:193: SAWarning: Did not recognize type 'int64' of column 'blob_size'
    "Did not recognize type '%s' of column '%s'" % (type_, name)
  2019-05-31 06:47:06,372:ERROR:root:Unrecognized data type in blobs.blob_size
  2019-05-31 06:47:06,372:ERROR:root:Can't generate DDL for NullType(); did you forget to specify a type on this Column?
  Traceback (most recent call last):
    File "/home/superset/superset/connectors/sqla/models.py", line 870, in fetch_metadata
      datatype = col.type.compile(dialect=db_dialect).upper()
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/type_api.py", line 580, in compile
      return dialect.type_compiler.process(self)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/compiler.py", line 400, in process
      return type_._compiler_dispatch(self, **kw)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/visitors.py", line 91, in _compiler_dispatch
      return meth(self, **kw)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/compiler.py", line 3350, in visit_null
      "type on this Column?" % type_
  sqlalchemy.exc.CompileError: Can't generate DDL for NullType(); did you forget to specify a type on this Column?
  2019-05-31 06:47:06,699:ERROR:root:Unrecognized data type in files.blob_size
  2019-05-31 06:47:06,700:ERROR:root:Can't generate DDL for NullType(); did you forget to specify a type on this Column?
  Traceback (most recent call last):
    File "/home/superset/superset/connectors/sqla/models.py", line 870, in fetch_metadata
      datatype = col.type.compile(dialect=db_dialect).upper()
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/type_api.py", line 580, in compile
      return dialect.type_compiler.process(self)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/compiler.py", line 400, in process
      return type_._compiler_dispatch(self, **kw)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/visitors.py", line 91, in _compiler_dispatch
      return meth(self, **kw)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/compiler.py", line 3350, in visit_null
      "type on this Column?" % type_
  sqlalchemy.exc.CompileError: Can't generate DDL for NullType(); did you forget to specify a type on this Column?
  /usr/local/lib/python3.6/site-packages/sqlalchemy/dialects/mysql/reflection.py:193: SAWarning: Did not recognize type 'int64' of column 'history_index'
    "Did not recognize type '%s' of column '%s'" % (type_, name)
  2019-05-31 06:47:06,790:ERROR:root:Unrecognized data type in ref_commits.history_index
  2019-05-31 06:47:06,790:ERROR:root:Can't generate DDL for NullType(); did you forget to specify a type on this Column?
  Traceback (most recent call last):
    File "/home/superset/superset/connectors/sqla/models.py", line 870, in fetch_metadata
      datatype = col.type.compile(dialect=db_dialect).upper()
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/type_api.py", line 580, in compile
      return dialect.type_compiler.process(self)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/compiler.py", line 400, in process
      return type_._compiler_dispatch(self, **kw)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/visitors.py", line 91, in _compiler_dispatch
      return meth(self, **kw)
    File "/usr/local/lib/python3.6/site-packages/sqlalchemy/sql/compiler.py", line 3350, in visit_null
      "type on this Column?" % type_
  sqlalchemy.exc.CompileError: Can't generate DDL for NullType(); did you forget to specify a type on this Column?
  ```

- [ ] Also, if a port is not available, the error is logged twice:
  ```bash
  $ GITBASE_REPOS_DIR=./github.com/dpordomingo sandbox-ce install 
  Starting src_redis_1  ... done
  Starting src_bblfsh_1   ... done
  Starting src_postgres_1 ... done
  Creating src_gitbase_1  ... error
  
  ERROR: for src_gitbase_1  Cannot start service gitbase: driver failed programming external connectivity on endpoint src_gitbase_1 (b07e89d1e44884a1abbb0cbe25c5b842bf00641a61af4afa7269779762d1611f): Error starting userland proxy: listen tcp 0.0.0.0:3306: bind: address already in use
  
  ERROR: for gitbase  Cannot start service gitbase: driver failed programming external connectivity on endpoint src_gitbase_1 (b07e89d1e44884a1abbb0cbe25c5b842bf00641a61af4afa7269779762d1611f): Error starting userland proxy: listen tcp 0.0.0.0:3306: bind: address already in use
  ERROR: Encountered errors while bringing up the project.
  exit status 1
  ```
