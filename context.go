package devstatscode

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Ctx - environment context packed in structure
type Ctx struct {
	DataDir                  string                       // From GHA2DB_DATADIR, default /etc/gha2db/
	Debug                    int                          // From GHA2DB_DEBUG Debug level: 0-no, 1-info, 2-verbose, including SQLs, default 0
	CmdDebug                 int                          // From GHA2DB_CMDDEBUG Commands execution Debug level: 0-no, 1-only output commands, 2-output commands and their output, 3-output full environment as well, default 0
	GitHubDebug              int                          // From GHA2DB_GITHUB_DEBUG debug GitHub rate limits
	DryRun                   bool                         // From GHA2DB_DRY_RUN, import_affs tool - stop before doing any updates
	JSONOut                  bool                         // From GHA2DB_JSON gha2db: write JSON files? default false
	DBOut                    bool                         // From GHA2DB_NODB gha2db: write to SQL database, default true
	ST                       bool                         // From GHA2DB_ST true: use single threaded version, false: use multi threaded version, default false
	NCPUs                    int                          // From GHA2DB_NCPUS, set to override number of CPUs to run, this overwrites GHA2DB_ST, default 0 (which means do not use it)
	PgHost                   string                       // From PG_HOST, default "localhost"
	PgPort                   string                       // From PG_PORT, default "5432"
	PgDB                     string                       // From PG_DB, default "gha"
	PgUser                   string                       // From PG_USER, default "gha_admin"
	PgPass                   string                       // From PG_PASS, default "password"
	PgSSL                    string                       // From PG_SSL, default "disable"
	Index                    bool                         // From GHA2DB_INDEX Create DB index? default false
	Table                    bool                         // From GHA2DB_SKIPTABLE Create table structure? default true
	Tools                    bool                         // From GHA2DB_SKIPTOOLS Create DB tools (like views, summary tables, materialized views etc)? default true
	Mgetc                    string                       // From GHA2DB_MGETC Character returned by mgetc (if non empty), default ""
	QOut                     bool                         // From GHA2DB_QOUT output all SQL queries?, default false
	CtxOut                   bool                         // From GHA2DB_CTXOUT output all context data (this struct), default false
	LogTime                  bool                         // From GHA2DB_SKIPTIME, output time with all lib.Printf(...) calls, default true, use GHA2DB_SKIPTIME to disable
	DefaultStartDate         time.Time                    // From GHA2DB_STARTDT, default `2012-07-01 00:00 UTC`, expects format "YYYY-MM-DD HH:MI:SS", can be set in `projects.yaml` via `start_date:`, value from projects.yaml (if set) has the highest priority.
	ForceStartDate           bool                         // From GHA2DB_STARTDT_FORCE, default false
	LastSeries               string                       // From GHA2DB_LASTSERIES, use this TSDB series to determine last timestamp date, default "events_h"
	SkipTSDB                 bool                         // From GHA2DB_SKIPTSDB gha2db_sync tool, skip TS DB processing? for calc_metric it skips final series write, default false
	SkipPDB                  bool                         // From GHA2DB_SKIPPDB gha2db_sync tool, skip Postgres DB processing (gha2db part) default false
	ResetTSDB                bool                         // From GHA2DB_RESETTSDB sync tool, regenerate all TS points? default false
	ResetRanges              bool                         // From GHA2DB_RESETRANGES sync tool, regenerate all past quick ranges? default false
	Explain                  bool                         // From GHA2DB_EXPLAIN runq tool, prefix query with "explain " - it will display query plan instead of executing real query, default false
	OldFormat                bool                         // From GHA2DB_OLDFMT gha2db tool, if set then use pre 2015 GHA JSONs format
	Exact                    bool                         // From GHA2DB_EXACT gha2db tool, if set then orgs list provided from commandline is used as a list of exact repository full names, like "a/b,c/d,e", if not only full names "a/b,x/y" can be treated like this, names without "/" are either orgs or repos.
	LogToDB                  bool                         // From GHA2DB_SKIPLOG all tools, if set, DB logging into Postgres table `gha_logs` in `devstats` database will be disabled
	Local                    bool                         // From GHA2DB_LOCAL many tools, if set it will use data files prefixed with "./" to use local ones. Otherwise it will search for data files in /etc/gha2db.
	LocalCmd                 bool                         // From GHA2DB_LOCAL_CMD many tools, if set it will call other tools prefixed with "./" to use locally compiled ones. Otherwise it will call binaries without prefix (so it will use those in $PATH).
	MetricsYaml              string                       // From GHA2DB_METRICS_YAML gha2db_sync tool, set other metrics.yaml file, default is "metrics/{{project}}metrics.yaml"
	TagsYaml                 string                       // From GHA2DB_TAGS_YAML tags tool, set other tags.yaml file, default is "metrics/{{project}}/tags.yaml"
	ColumnsYaml              string                       // From GHA2DB_COLUMNS_YAML tags tool, set other columns.yaml file, default is "metrics/{{project}}/columns.yaml"
	VarsYaml                 string                       // From GHA2DB_VARS_YAML db_vars tool, set other vars.yaml file (full path), default is "metrics/{{project}}/vars.yaml"
	VarsFnYaml               string                       // From GHA2DB_VARS_FN_YAML db_vars tool, set other vars.yaml file (final file name without path), default is "vars.yaml"
	SkipDatesYaml            string                       // From GHA2DB_SKIP_DATES_YAML gha2db tool, set other skip_dates.yaml file, default is "skip_dates.yaml"
	GitHubOAuth              string                       // From GHA2DB_GITHUB_OAUTH ghapi2db tool, if not set reads from /etc/github/oauth file, set to "-" to force public access.
	ClearDBPeriod            string                       // From GHA2DB_MAXLOGAGE gha2db_sync tool, maximum age of devstats.gha_logs entries, default "1 week"
	ClearAffsLockPeriod      string                       // From GHA2DB_MAX_AFFS_LOCK_AGE devstats tool, maximum age of devstats.gha_metrics "affs_lock" age, default "30 hours"
	ClearGiantLockPeriod     string                       // From GHA2DB_MAX_GIANT_LOCK_AGE devstats tool, maximum age of devstats.gha_metrics "giant_lock" age, default "60 hours"
	Trials                   []int                        // From GHA2DB_TRIALS, all Postgres related tools, retry periods for some retryable errors
	WebHookRoot              string                       // From GHA2DB_WHROOT, webhook tool, default "/hook", must match .travis.yml notifications webhooks
	WebHookPort              string                       // From GHA2DB_WHPORT, webhook tool, default ":1982", note that webhook listens using http:1982, but we use apache on https:2982 (to enable https protocol and proxy requests to http:1982)
	WebHookHost              string                       // From GHA2DB_WHHOST, webhook tool, default "127.0.0.1" (this can be localhost to disable access by IP, we use Apache proxy to enable https and then apache only need 127.0.0.1)
	CheckPayload             bool                         // From GHA2DB_SKIP_VERIFY_PAYLOAD, webhook tool, default true, use GHA2DB_SKIP_VERIFY_PAYLOAD=1 to manually test payloads
	FullDeploy               bool                         // From GHA2DB_SKIP_FULL_DEPLOY, webhook tool, default true, use GHA2DB_SKIP_FULL_DEPLOY=1 to ignore "[deploy]" requests that call `./devel/deploy_all.sh`.
	DeployBranches           []string                     // From GHA2DB_DEPLOY_BRANCHES, webhook tool, default "master" - comma separated list
	DeployStatuses           []string                     // From GHA2DB_DEPLOY_STATUSES, webhook tool, default "Passed,Fixed", - comma separated list
	DeployResults            []int                        // From GHA2DB_DEPLOY_RESULTS, webhook tool, default "0", - comma separated list
	DeployTypes              []string                     // From GHA2DB_DEPLOY_TYPES, webhook tool, default "push", - comma separated list
	ProjectRoot              string                       // From GHA2DB_PROJECT_ROOT, webhook tool, no default, must be specified to run webhook tool
	ExecFatal                bool                         // default true, set this manually to false to avoid lib.ExecCommand calling os.Exit() on failure and return error instead
	ExecQuiet                bool                         // default false, set this manually to true to have quite exec failures (for example `get_repos` git-clones or git-pulls on errors).
	ExecOutput               bool                         // default false, set to true to capture commands STDOUT
	Project                  string                       // From GHA2DB_PROJECT, gha2db_sync default "", You should set it to something like "kubernetes", "prometheus" etc.
	TestsYaml                string                       // From GHA2DB_TESTS_YAML ./dbtest.sh tool, set other tests.yaml file, default is "tests.yaml"
	ReposDir                 string                       // From GHA2DB_REPOS_DIR get_repos tool, default "~/devstats_repos/"
	ProcessRepos             bool                         // From GHA2DB_PROCESS_REPOS get_repos tool, enable processing (cloning/pulling) all devstats repos, default false
	ProcessCommits           bool                         // From GHA2DB_PROCESS_COMMITS get_repos tool, enable update/create mapping table: commit - list of file that commit refers to, default false
	ExternalInfo             bool                         // From GHA2DB_EXTERNAL_INFO get_repos tool, enable outputing data needed by external tools (cncf/gitdm), default false
	ProjectsCommits          string                       // From GHA2DB_PROJECTS_COMMITS get_repos tool, set list of projects for commits analysis instead of analysing all, default "" - means all
	PropagateOnlyVar         bool                         // From GHA2DB_PROPAGATE_ONLY_VAR, if set the it will check ONLY="a b c" env variable and propagate it into other project filter variables if they're not set, for example GHA2DB_PROJECTS_COMMITS
	ProjectsYaml             string                       // From GHA2DB_PROJECTS_YAML, many tools - set main projects file, default "projects.yaml"
	CompanyAcqYaml           string                       // From GHA2DB_COMPANY_ACQ_YAML, import_affs tool, set non-standard "companies.yaml" file
	ProjectsOverride         map[string]bool              // From GHA2DB_PROJECTS_OVERRIDE, get_repos and ./devstats tools - for example "-pro1,+pro2" means never sync pro1 and always sync pro2 (even if disabled in `projects.yaml`).
	AffiliationsJSON         string                       // From GHA2DB_AFFILIATIONS_JSON, import_affs tool - set main affiliations file, default "github_users.json"
	ExcludeRepos             map[string]bool              // From GHA2DB_EXCLUDE_REPOS, gha2db tool, default "" - comma separated list of repos to exclude, example: "theupdateframework/notary,theupdateframework/other"
	InputDBs                 []string                     // From GHA2DB_INPUT_DBS, merge_dbs tool - list of input databases to merge, order matters - first one will insert on a clean DB, next will do insert ignore (to avoid constraints failure due to common data)
	OutputDB                 string                       // From GHA2DB_OUTPUT_DB, merge_dbs tool - output database to merge into
	TmOffset                 int                          // From GHA2DB_TMOFFSET, gha2db_sync tool - uses time offset to decide when to calculate various metrics, default offset is 0 which means UTC, good offset for USA is -6, and for Poland is 1 or 2
	DefaultHostname          string                       // "devstats.cncf.io"
	RecentRange              string                       // From GHA2DB_RECENT_RANGE, ghapi2db tool, default '2 hours'. This is a recent period to check open issues/PR to fix their labels and milestones.
	RecentReposRange         string                       // From GHA2DB_RECENT_REPOS_RANGE, ghapi2db tool, default '1 day'. This is a recent period to check modified repositories.
	MinGHAPIPoints           int                          // From GHA2DB_MIN_GHAPI_POINTS, ghapi2db tool, minimum GitHub API points, before waiting for reset.
	MaxGHAPIWaitSeconds      int                          // From GHA2DB_MAX_GHAPI_WAIT, ghapi2db tool, maximum wait time for GitHub API points reset (in seconds).
	MaxGHAPIRetry            int                          // From GHA2DB_MAX_GHAPI_RETRY, ghapi2db tool, maximum wait retries
	GHAPIErrorIsFatal        bool                         // From GHA2DB_GHAPI_ERROR_FATAL, ghapi2db tool, make any GH API error fatal, default false
	SkipGHAPI                bool                         // From GHA2DB_GHAPISKIP, ghapi2db tool, if set then tool is skipping GH API calls (all: events (artificial events to make sure we are in sync with GH) and commits (enriches obfuscated GHA commits data)
	SkipAPIEvents            bool                         // From GHA2DB_GHAPISKIPEVENTS, ghapi2db tool, if set then tool is skipping GH API events sync
	SkipAPICommits           bool                         // From GHA2DB_GHAPISKIPCOMMITS, ghapi2db tool, if set then tool is skipping GH API commits enrichment
	SkipAPILicenses          bool                         // From GHA2DB_GHAPISKIPLICENSES, ghapi2db tool, if set then tool is skipping GH API licenses enrichment
	ForceAPILicenses         bool                         // From GHA2DB_GHAPIFORCELICENSES, ghapi2db tool, if set, recheck licenses on repos that already have licenses fetched
	SkipAPILangs             bool                         // From GHA2DB_GHAPISKIPLANGS, ghapi2db tool, if set then tool is skipping GH API repos programming languages enrichment
	ForceAPILangs            bool                         // From GHA2DB_GHAPIFORCELANGS, ghapi2db tool, if set, recheck programming languages on repos that already have them fetched
	SkipGetRepos             bool                         // From GHA2DB_GETREPOSSKIP, get_repos tool, if set then tool does nothing
	CSVFile                  string                       // From GHA2DB_CSVOUT, runq tool, if set, saves result in this file
	ComputeAll               bool                         // From GHA2DB_COMPUTE_ALL, all tools, if set then no period decisions are taken based on time, but all possible periods are recalculated
	ActorsFilter             bool                         // From GHA2DB_ACTORS_FILTER gha2db tool, if enabled then actor filterning will be added, default false
	ActorsAllow              *regexp.Regexp               // From GHA2DB_ACTORS_ALLOW, gha2db tool, process JSON if actor matches this regexp, default "" which means skip this check
	ActorsForbid             *regexp.Regexp               // From GHA2DB_ACTORS_FORBID, gha2db tool, process JSON if actor doesn't match this regexp, default "" which means skip this check
	SkipMetrics              map[string]bool              // From GHA2DB_SKIP_METRICS, gha2db_sync tool, default "" - comma separated list of metrics to skip, as given by "sql: name" in the "metrics.yaml" file. Those metrics will be skipped.
	OnlyMetrics              map[string]bool              // From GHA2DB_ONLY_METRICS, gha2db_sync tool, default "" - comma separated list of metrics to process, as given by "sql: name" in the "metrics.yaml" file. Only those metrics will be calculated.
	AllowBrokenJSON          bool                         // From GHA2DB_ALLOW_BROKEN_JSON, gha2db tool, default false. If set then gha2db skips broken jsons and saves them as jsons/error_YYYY-MM-DD-h-n-m.json (n is the JSON number (1-m) of m JSONS array)
	JSONsDir                 string                       // From GHA2DB_JSONS_DIR, website_data tool, default "./jsons/"
	WebsiteData              bool                         // From GHA2DB_WEBSITEDATA, devstats tool, run website_data just after sync is complete, default false.
	SkipUpdateEvents         bool                         // From GHA2DB_SKIP_UPDATE_EVENTS, ghapi2db tool, drop and recreate artificial events if their state differs, default false
	ComputePeriods           map[string]map[bool]struct{} // From GHA2DB_FORCE_PERIODS, gha2db_sync tool, force recompute only given periods, "y10:t,m:f,...", default ""
	AutoFetchCommits         bool                         // From GHA2DB_NO_AUTOFETCHCOMMITS, ghapi2db, disable fetching from last enriched commit data, it will fetch from 'RecentRange instead, AutoFetchCommits is enabled by default
	SkipTags                 bool                         // From GHA2DB_SKIP_TAGS, gha2db_sync tool, skip calling tags tool, default false
	SkipAnnotations          bool                         // From GHA2DB_SKIP_ANNOTATIONS, gha2db_sync tool, skip calling annotations tool, default false
	SkipColumns              bool                         // From GHA2DB_SKIP_COLUMNS, gha2db_sync tool, skip calling columns tool, default false
	RunColumns               bool                         // From GHA2DB_RUN_COLUMNS, gha2db_sync tool, force calling columns tool, default false
	SkipVars                 bool                         // From GHA2DB_SKIP_VARS, gha2db_sync tool, skip calling vars tool, default false
	SkipRand                 bool                         // From GHA2DB_SKIP_RAND, gha2db_sync tool, skip randomizing metrics calculation, default false
	ExcludeVars              map[string]bool              // From GHA2DB_EXCLUDE_VARS, vars tool, default "" - comma separated list of variable names to exclude, example: "hostname,projects_health_partial_html"
	OnlyVars                 map[string]bool              // From GHA2DB_ONLY_VARS, vars tool, default "" - comma separated list of variable names to write (and skip all others): "hostname,projects_health_partial_html", not used if empty
	ElasticURL               string                       // From GHA2DB_ES_URL, calc_metric, tags, annotations tools - ElasticSearch URL (if used), default http://127.0.0.1:9200
	UseES                    bool                         // From GHA2DB_USE_ES, calc_metric, tags, annotations tools - enable ElasticSearch, default false
	UseESOnly                bool                         // From GHA2DB_USE_ES_ONLY, calc_metric, annotations tools - enable ElasticSearch and do not write PSQL TSDB, default false
	UseESRaw                 bool                         // From GHA2DB_USE_ES_RAW, gha2es, gha2db_sync tools - enable generating RAW ElasticSearch data (directly from gha_tables instead of aggregated data from TSDB)
	ResetESRaw               bool                         // From GHA2DB_RESET_ES_RAW, gha2db_sync tools - generate RAW ES index from project start date
	SkipSharedDB             bool                         // From GHA2DB_SKIP_SHAREDDB, annotations tool, default false, will skip writing to shared_db (from projects.yaml) if set
	SkipPIDFile              bool                         // From GHA2DB_SKIP_PIDFILE, devstats tool, skip creating, checking and removing PID file
	SkipCompanyAcq           bool                         // From GHA2DB_SKIP_COMPANY_ACQ, import_affs tool, skip processing company acquisitions from companies.yaml file
	CheckProvisionFlag       bool                         // From GHA2DB_CHECK_PROVISION_FLAG, devstats tool - check if there is a 'provision' metric saved in 'gha_computed' table - if not, abort
	CheckRunningFlag         bool                         // From GHA2DB_CHECK_RUNNING_FLAG, devstats tool - check if there is a 'devstats_running' metric saved in 'gha_computed' table - if yes, abort
	SetRunningFlag           bool                         // From GHA2DB_SET_RUNNING_FLAG, devstats tool - set 'devstats_running' flag on 'gha_computed' table while devstats cronjob is running
	MaxRunningFlagAge        time.Duration                // From GHA2DB_MAX_RUNNING_FLAG_AGE, how log "running_flag" can be present for next devstats sync to treat it as orphan, default "6h"
	CheckImportedSHA         bool                         // From GHA2DB_CHECK_IMPORTED_SHA, import_affs tool - check if given JSON was already imported using 'gha_imported_shas' table
	OnlyCheckImportedSHA     bool                         // From GHA2DB_ONLY_CHECK_IMPORTED_SHA, import_affs tool - check if given JSON was already imported using 'gha_imported_shas' table, do not attempt to import, only return status: 3=imported, 0=not imported
	EnableMetricsDrop        bool                         // From GHA2DB_ENABLE_METRICS_DROP, if enabled will process each metric's 'drop:' property if present - use when regenerating affiliations data or reinitializing entire TSDB data
	ESBulkSize               int                          // From GHA2DB_ES_BULK_SIZE, calc_metric and gha2es tools, default 10000
	HTTPTimeout              int                          // From GHA2DB_HTTP_TIMEOUT, gha2db - data.gharchive.org timeout value in minutes, default 2
	HTTPRetry                int                          // From GHA2DB_HTTP_RETRY, gha2db - data.gharchive.org data fetch retries, default 4 (each retry takes 1*timeout*N), so in default config it will try timeouts: 1min, 2min, 3min, but if timeout is 3 and retry is 2, it will try 3min, 6min
	ProjectScale             float64                      // From GHA2DB_PROJECT_SCALE, calc_metric tool, project scale (default 1), some metrics can use this to adapt their SQLs to bigger/smaller projects
	PidFileRoot              string                       // From GHA2DB_PID_FILE_ROOT, devstats tool, use '/tmp/PidFileRoot.pid' as PID file, default 'devstats' -> '/tmp/devstats.pid'
	SharedDB                 string                       // Currently annotations tool read this from projects.yaml:shared_db and if set, outputs annotations data to the sharded DB in addition to the current DB
	ProjectMainRepo          string                       // Used by annotations tool to store project's main repo name
	TestMode                 bool                         // True when running tests
	CanReconnect             bool                         // True, unless connecting to a custom database, in this case there can be multiple threads sharing context and we don't want to write to a random database
	CommitsFilesStatsEnabled bool                         // True, can be disabled by GHA2DB_SKIP_COMMITS_FILES, get_repos tool
	CommitsLOCStatsEnabled   bool                         // True, can be disabled by GHA2DB_SKIP_COMMITS_LOC, get_repos tool
	RecalcReciprocal         int                          // From GHA2DB_RECALC_RECIPROCAL: 1/RecalcReciprocal of recalc metric at given datetime, even if it should be calculated at this datetime, default 24 (means 4.1(6)%, or about once/day)
	MaxHistograms            int                          // From GHA2DB_MAX_HIST: maximum histogram concurrency, default: 0 - means unlimited
}

// Init - get context from environment variables
func (ctx *Ctx) Init() {
	ctx.ExecFatal = true
	ctx.ExecQuiet = false
	ctx.ExecOutput = false
	ctx.CanReconnect = true

	/// Commits analysis
	ctx.CommitsFilesStatsEnabled = os.Getenv("GHA2DB_SKIP_COMMITS_FILES") == ""
	ctx.CommitsLOCStatsEnabled = os.Getenv("GHA2DB_SKIP_COMMITS_LOC") == ""

	// Data directory
	ctx.DataDir = os.Getenv("GHA2DB_DATADIR")
	if ctx.DataDir == "" {
		ctx.DataDir = DefaultDataDir
	}
	if ctx.DataDir[len(ctx.DataDir)-1:] != "/" {
		ctx.DataDir += "/"
	}

	// Outputs
	ctx.JSONOut = os.Getenv("GHA2DB_JSON") != ""
	ctx.DBOut = os.Getenv("GHA2DB_NODB") == ""

	// Dry run
	ctx.DryRun = os.Getenv("GHA2DB_DRY_RUN") != ""

	// GitHub API points and waiting for reset
	ctx.MinGHAPIPoints = 1
	if os.Getenv("GHA2DB_MIN_GHAPI_POINTS") != "" {
		pts, err := strconv.Atoi(os.Getenv("GHA2DB_MIN_GHAPI_POINTS"))
		FatalNoLog(err)
		if pts >= 0 {
			ctx.MinGHAPIPoints = pts
		}
	}
	ctx.MaxGHAPIWaitSeconds = 10
	if os.Getenv("GHA2DB_MAX_GHAPI_WAIT") != "" {
		secs, err := strconv.Atoi(os.Getenv("GHA2DB_MAX_GHAPI_WAIT"))
		FatalNoLog(err)
		if secs >= 0 {
			ctx.MaxGHAPIWaitSeconds = secs
		}
	}
	ctx.MaxGHAPIRetry = 6
	if os.Getenv("GHA2DB_MAX_GHAPI_RETRY") != "" {
		tr, err := strconv.Atoi(os.Getenv("GHA2DB_MAX_GHAPI_RETRY"))
		FatalNoLog(err)
		if tr >= 1 {
			ctx.MaxGHAPIRetry = tr
		}
	}

	// Debug
	if os.Getenv("GHA2DB_DEBUG") == "" {
		ctx.Debug = 0
	} else {
		debugLevel, err := strconv.Atoi(os.Getenv("GHA2DB_DEBUG"))
		FatalNoLog(err)
		if debugLevel != 0 {
			ctx.Debug = debugLevel
		}
	}
	// CmdDebug
	if os.Getenv("GHA2DB_CMDDEBUG") == "" {
		ctx.CmdDebug = 0
	} else {
		debugLevel, err := strconv.Atoi(os.Getenv("GHA2DB_CMDDEBUG"))
		FatalNoLog(err)
		ctx.CmdDebug = debugLevel
	}
	// GitHubDebug
	if os.Getenv("GHA2DB_GITHUB_DEBUG") == "" {
		ctx.GitHubDebug = 0
	} else {
		debugLevel, err := strconv.Atoi(os.Getenv("GHA2DB_GITHUB_DEBUG"))
		FatalNoLog(err)
		ctx.GitHubDebug = debugLevel
	}
	ctx.QOut = os.Getenv("GHA2DB_QOUT") != ""
	ctx.CtxOut = os.Getenv("GHA2DB_CTXOUT") != ""

	// Threading
	ctx.ST = os.Getenv("GHA2DB_ST") != ""
	// NCPUs
	if os.Getenv("GHA2DB_NCPUS") == "" {
		ctx.NCPUs = 0
	} else {
		nCPUs, err := strconv.Atoi(os.Getenv("GHA2DB_NCPUS"))
		FatalNoLog(err)
		if nCPUs > 0 {
			ctx.NCPUs = nCPUs
			if ctx.NCPUs == 1 {
				ctx.ST = true
			}
		}
	}

	// Postgres DB
	ctx.PgHost = os.Getenv("PG_HOST")
	ctx.PgPort = os.Getenv("PG_PORT")
	ctx.PgDB = os.Getenv("PG_DB")
	ctx.PgUser = os.Getenv("PG_USER")
	ctx.PgPass = os.Getenv("PG_PASS")
	ctx.PgSSL = os.Getenv("PG_SSL")
	if ctx.PgHost == "" {
		ctx.PgHost = Localhost
	}
	if ctx.PgPort == "" {
		ctx.PgPort = "5432"
	}
	if ctx.PgDB == "" {
		ctx.PgDB = GHA
	}
	if ctx.PgUser == "" {
		ctx.PgUser = GHAAdmin
	}
	if ctx.PgPass == "" {
		ctx.PgPass = Password
	}
	if ctx.PgSSL == "" {
		ctx.PgSSL = "disable"
	}

	// PID file
	ctx.PidFileRoot = os.Getenv("GHA2DB_PID_FILE_ROOT")
	if ctx.PidFileRoot == "" {
		ctx.PidFileRoot = Devstats
	}

	// Environment controlling index creation, table & tools
	ctx.Index = os.Getenv("GHA2DB_INDEX") != ""
	ctx.Table = os.Getenv("GHA2DB_SKIPTABLE") == ""
	ctx.Tools = os.Getenv("GHA2DB_SKIPTOOLS") == ""
	ctx.Mgetc = os.Getenv("GHA2DB_MGETC")
	if len(ctx.Mgetc) > 1 {
		ctx.Mgetc = ctx.Mgetc[:1]
	}

	// Log Time
	ctx.LogTime = os.Getenv("GHA2DB_SKIPTIME") == ""

	// Time offset for gha2db_sync
	if os.Getenv("GHA2DB_TMOFFSET") == "" {
		ctx.TmOffset = 0
	} else {
		off, err := strconv.Atoi(os.Getenv("GHA2DB_TMOFFSET"))
		FatalNoLog(err)
		ctx.TmOffset = off
	}

	// Default start date
	if os.Getenv("GHA2DB_STARTDT") != "" {
		ctx.DefaultStartDate = TimeParseAny(os.Getenv("GHA2DB_STARTDT"))
	} else {
		ctx.DefaultStartDate = time.Date(2012, 7, 1, 0, 0, 0, 0, time.UTC)
	}
	ctx.ForceStartDate = false
	if os.Getenv("GHA2DB_STARTDT_FORCE") != "" {
		ctx.ForceStartDate = true
	}

	// Skip ghapi2db and/or get_repos
	ctx.SkipGetRepos = os.Getenv("GHA2DB_GETREPOSSKIP") != ""
	ctx.SkipGHAPI = os.Getenv("GHA2DB_GHAPISKIP") != ""
	ctx.SkipAPIEvents = os.Getenv("GHA2DB_GHAPISKIPEVENTS") != ""
	ctx.SkipAPICommits = os.Getenv("GHA2DB_GHAPISKIPCOMMITS") != ""
	ctx.SkipAPILicenses = os.Getenv("GHA2DB_GHAPISKIPLICENSES") != ""
	ctx.ForceAPILicenses = os.Getenv("GHA2DB_GHAPIFORCELICENSES") != ""
	ctx.SkipAPILangs = os.Getenv("GHA2DB_GHAPISKIPLANGS") != ""
	ctx.ForceAPILangs = os.Getenv("GHA2DB_GHAPIFORCELANGS") != ""
	ctx.GHAPIErrorIsFatal = os.Getenv("GHA2DB_GHAPI_ERROR_FATAL") != ""
	ctx.AutoFetchCommits = os.Getenv("GHA2DB_NO_AUTOFETCHCOMMITS") == ""

	// Last TS series
	ctx.LastSeries = os.Getenv("GHA2DB_LASTSERIES")
	if ctx.LastSeries == "" {
		ctx.LastSeries = "events_h"
	}

	// Skip some tools
	ctx.SkipTags = os.Getenv("GHA2DB_SKIP_TAGS") != ""
	ctx.SkipAnnotations = os.Getenv("GHA2DB_SKIP_ANNOTATIONS") != ""
	ctx.SkipColumns = os.Getenv("GHA2DB_SKIP_COLUMNS") != ""
	ctx.RunColumns = os.Getenv("GHA2DB_RUN_COLUMNS") != ""
	ctx.SkipVars = os.Getenv("GHA2DB_SKIP_VARS") != ""

	// Skip randomizing task order
	ctx.SkipRand = os.Getenv("GHA2DB_SKIP_RAND") != ""

	// TS variables
	ctx.SkipTSDB = os.Getenv("GHA2DB_SKIPTSDB") != ""
	ctx.ResetTSDB = os.Getenv("GHA2DB_RESETTSDB") != ""
	ctx.ResetRanges = os.Getenv("GHA2DB_RESETRANGES") != ""

	// Allow broken JSON
	ctx.AllowBrokenJSON = os.Getenv("GHA2DB_ALLOW_BROKEN_JSON") != ""

	// Run website_data tool after sync
	ctx.WebsiteData = os.Getenv("GHA2DB_WEBSITEDATA") != ""

	// Disable delete & recreate past events
	ctx.SkipUpdateEvents = os.Getenv("GHA2DB_SKIP_UPDATE_EVENTS") != ""

	// Postgres DB variables
	ctx.SkipPDB = os.Getenv("GHA2DB_SKIPPDB") != ""

	// Explain
	ctx.Explain = os.Getenv("GHA2DB_EXPLAIN") != ""

	// Old (pre 2015) GHA JSONs format
	ctx.OldFormat = os.Getenv("GHA2DB_OLDFMT") != ""

	// Exact repository full names to match
	ctx.Exact = os.Getenv("GHA2DB_EXACT") != ""

	// Log to Postgres DB, table `devstats`.`gha_logs`
	ctx.LogToDB = os.Getenv("GHA2DB_SKIPLOG") == ""

	// Local data files mode
	ctx.Local = os.Getenv("GHA2DB_LOCAL") != ""

	// Local binary/shell files mode
	ctx.LocalCmd = os.Getenv("GHA2DB_LOCAL_CMD") != ""

	// Project
	ctx.Project = os.Getenv("GHA2DB_PROJECT")
	proj := ""
	if ctx.Project != "" {
		proj = ctx.Project + "/"
	}

	// YAML config files
	ctx.MetricsYaml = os.Getenv("GHA2DB_METRICS_YAML")
	ctx.TagsYaml = os.Getenv("GHA2DB_TAGS_YAML")
	ctx.ColumnsYaml = os.Getenv("GHA2DB_COLUMNS_YAML")
	ctx.VarsYaml = os.Getenv("GHA2DB_VARS_YAML")
	ctx.VarsFnYaml = os.Getenv("GHA2DB_VARS_FN_YAML")
	if ctx.VarsFnYaml == "" {
		ctx.VarsFnYaml = "vars.yaml"
	}
	if ctx.MetricsYaml == "" {
		ctx.MetricsYaml = "metrics/" + proj + "metrics.yaml"
	}
	if ctx.TagsYaml == "" {
		ctx.TagsYaml = "metrics/" + proj + "tags.yaml"
	}
	if ctx.ColumnsYaml == "" {
		ctx.ColumnsYaml = "metrics/" + proj + "columns.yaml"
	}
	if ctx.VarsYaml == "" {
		ctx.VarsYaml = "metrics/" + proj + ctx.VarsFnYaml
	}

	// GitHub OAuth
	ctx.GitHubOAuth = os.Getenv("GHA2DB_GITHUB_OAUTH")
	if ctx.GitHubOAuth == "" {
		fn := "/etc/github/oauths"
		_, err := os.Stat(fn)
		if err == nil {
			ctx.GitHubOAuth = fn
		} else if os.IsNotExist(err) {
			fn = "/etc/github/oauth"
			_, err := os.Stat(fn)
			if err == nil {
				ctx.GitHubOAuth = fn
			} else if os.IsNotExist(err) {
				ctx.GitHubOAuth = "-"
			} else {
				FatalNoLog(err)
			}
		} else {
			FatalNoLog(err)
		}
	}

	// Max DB logs age
	ctx.ClearDBPeriod = os.Getenv("GHA2DB_MAXLOGAGE")
	if ctx.ClearDBPeriod == "" {
		ctx.ClearDBPeriod = "1 week"
	}

	// Max locks ages
	ctx.ClearAffsLockPeriod = os.Getenv("GHA2DB_MAX_AFFS_LOCK_AGE")
	if ctx.ClearAffsLockPeriod == "" {
		ctx.ClearAffsLockPeriod = "30 hours"
	}
	ctx.ClearGiantLockPeriod = os.Getenv("GHA2DB_MAX_GIANT_LOCK_AGE")
	if ctx.ClearGiantLockPeriod == "" {
		ctx.ClearGiantLockPeriod = "60 hours"
	}

	// Trials
	trials := os.Getenv("GHA2DB_TRIALS")
	if trials == "" {
		ctx.Trials = []int{10, 30, 60, 120, 300, 600, 1200, 3600}
	} else {
		trialsArr := strings.Split(trials, ",")
		for _, try := range trialsArr {
			iTry, err := strconv.Atoi(try)
			FatalNoLog(err)
			ctx.Trials = append(ctx.Trials, iTry)
		}
	}

	// Deploy statuses and branches
	branches := os.Getenv("GHA2DB_DEPLOY_BRANCHES")
	if branches == "" {
		ctx.DeployBranches = []string{"master"}
	} else {
		ctx.DeployBranches = strings.Split(branches, ",")
	}
	statuses := os.Getenv("GHA2DB_DEPLOY_STATUSES")
	if statuses == "" {
		ctx.DeployStatuses = []string{"Passed", "Fixed"}
	} else {
		ctx.DeployStatuses = strings.Split(statuses, ",")
	}
	types := os.Getenv("GHA2DB_DEPLOY_TYPES")
	if types == "" {
		ctx.DeployTypes = []string{"push"}
	} else {
		ctx.DeployTypes = strings.Split(types, ",")
	}
	results := os.Getenv("GHA2DB_DEPLOY_RESULTS")
	if results == "" {
		ctx.DeployResults = []int{0}
	} else {
		resultsArr := strings.Split(results, ",")
		for _, result := range resultsArr {
			iResult, err := strconv.Atoi(result)
			FatalNoLog(err)
			ctx.DeployResults = append(ctx.DeployResults, iResult)
		}
	}
	ctx.ProjectRoot = os.Getenv("GHA2DB_PROJECT_ROOT")

	// Projects sync override
	ctx.ProjectsOverride = make(map[string]bool)
	overrides := os.Getenv("GHA2DB_PROJECTS_OVERRIDE")
	if overrides != "" {
		ary := strings.Split(overrides, ",")
		for _, override := range ary {
			if override == "" {
				continue
			}
			project := override[1:]
			if project == "" {
				continue
			}
			mode := override[:1]
			if mode == "-" {
				ctx.ProjectsOverride[project] = false
			} else if mode == "+" {
				ctx.ProjectsOverride[project] = true
			}
		}
	}

	// Exclude repos
	excludes := os.Getenv("GHA2DB_EXCLUDE_REPOS")
	ctx.ExcludeRepos = make(map[string]bool)
	if excludes != "" {
		excludeArray := strings.Split(excludes, ",")
		for _, exclude := range excludeArray {
			if exclude != "" {
				ctx.ExcludeRepos[exclude] = true
			}
		}
	}

	// Exclude vars
	excludes = os.Getenv("GHA2DB_EXCLUDE_VARS")
	ctx.ExcludeVars = make(map[string]bool)
	if excludes != "" {
		excludeArray := strings.Split(excludes, ",")
		for _, exclude := range excludeArray {
			if exclude != "" {
				ctx.ExcludeVars[exclude] = true
			}
		}
	}

	// Only vars
	only := os.Getenv("GHA2DB_ONLY_VARS")
	ctx.OnlyVars = make(map[string]bool)
	if only != "" {
		onlyArray := strings.Split(only, ",")
		for _, onl := range onlyArray {
			if onl != "" {
				ctx.OnlyVars[onl] = true
			}
		}
	}

	// Only metrics
	onlyMetrics := os.Getenv("GHA2DB_ONLY_METRICS")
	ctx.OnlyMetrics = make(map[string]bool)
	if onlyMetrics != "" {
		ary := strings.Split(onlyMetrics, ",")
		for _, metric := range ary {
			if metric != "" {
				ctx.OnlyMetrics[metric] = true
			}
		}
	}

	// Exclude metrics
	excludes = os.Getenv("GHA2DB_SKIP_METRICS")
	ctx.SkipMetrics = make(map[string]bool)
	if excludes != "" {
		excludeArray := strings.Split(excludes, ",")
		for _, exclude := range excludeArray {
			if exclude != "" {
				ctx.SkipMetrics[exclude] = true
			}
		}
	}

	// WebHook Host, Port, Root
	ctx.WebHookHost = os.Getenv("GHA2DB_WHHOST")
	if ctx.WebHookHost == "" {
		ctx.WebHookHost = "127.0.0.1"
	}
	ctx.WebHookPort = os.Getenv("GHA2DB_WHPORT")
	if ctx.WebHookPort == "" {
		ctx.WebHookPort = ":1982"
	} else {
		if ctx.WebHookPort[0:1] != ":" {
			ctx.WebHookPort = ":" + ctx.WebHookPort
		}
	}
	ctx.WebHookRoot = os.Getenv("GHA2DB_WHROOT")
	if ctx.WebHookRoot == "" {
		ctx.WebHookRoot = "/hook"
	}
	ctx.CheckPayload = os.Getenv("GHA2DB_SKIP_VERIFY_PAYLOAD") == ""
	ctx.FullDeploy = os.Getenv("GHA2DB_SKIP_FULL_DEPLOY") == ""

	// Tests
	ctx.TestsYaml = os.Getenv("GHA2DB_TESTS_YAML")
	if ctx.TestsYaml == "" {
		ctx.TestsYaml = "tests.yaml"
	}

	// Skip dates
	ctx.SkipDatesYaml = os.Getenv("GHA2DB_SKIP_DATES_YAML")
	if ctx.SkipDatesYaml == "" {
		ctx.SkipDatesYaml = "skip_dates.yaml"
	}

	// Main projects file
	ctx.ProjectsYaml = os.Getenv("GHA2DB_PROJECTS_YAML")
	if ctx.ProjectsYaml == "" {
		ctx.ProjectsYaml = "projects.yaml"
	}

	// Main affiliations file
	ctx.AffiliationsJSON = os.Getenv("GHA2DB_AFFILIATIONS_JSON")
	if ctx.AffiliationsJSON == "" {
		ctx.AffiliationsJSON = "github_users.json"
	}

	// Company acquisitions file
	ctx.CompanyAcqYaml = os.Getenv("GHA2DB_COMPANY_ACQ_YAML")
	if ctx.CompanyAcqYaml == "" {
		ctx.CompanyAcqYaml = "companies.yaml"
	}

	// `get_repos` repositories dir
	ctx.ReposDir = os.Getenv("GHA2DB_REPOS_DIR")
	if ctx.ReposDir == "" {
		ctx.ReposDir = os.Getenv("HOME") + "/devstats_repos/"
	}
	if ctx.ReposDir[len(ctx.ReposDir)-1:] != "/" {
		ctx.ReposDir += "/"
	}
	// `get_repos`: process repos, process commits, external info
	ctx.ProcessRepos = os.Getenv("GHA2DB_PROCESS_REPOS") != ""
	ctx.ProcessCommits = os.Getenv("GHA2DB_PROCESS_COMMITS") != ""
	ctx.ExternalInfo = os.Getenv("GHA2DB_EXTERNAL_INFO") != ""
	ctx.ProjectsCommits = os.Getenv("GHA2DB_PROJECTS_COMMITS")

	// PropagateOnlyVar
	ctx.PropagateOnlyVar = os.Getenv("GHA2DB_PROPAGATE_ONLY_VAR") != ""
	if ctx.PropagateOnlyVar {
		only = os.Getenv("ONLY")
		if only != "" {
			if ctx.ProjectsCommits == "" {
				ctx.ProjectsCommits = strings.Replace(only, " ", ",", -1)
			}
		}
	}

	// `website_data` JSONs dir
	ctx.JSONsDir = os.Getenv("GHA2DB_JSONS_DIR")
	if ctx.JSONsDir == "" {
		ctx.JSONsDir = "./jsons/"
	}
	if ctx.JSONsDir[len(ctx.JSONsDir)-1:] != "/" {
		ctx.JSONsDir += "/"
	}

	// ElasticSearch
	ctx.UseES = os.Getenv("GHA2DB_USE_ES") != ""
	ctx.UseESOnly = os.Getenv("GHA2DB_USE_ES_ONLY") != ""
	ctx.UseESRaw = os.Getenv("GHA2DB_USE_ES_RAW") != ""
	ctx.ResetESRaw = os.Getenv("GHA2DB_RESET_ES_RAW") != ""
	ctx.ElasticURL = os.Getenv("GHA2DB_ES_URL")
	if ctx.ElasticURL == "" {
		ctx.ElasticURL = "http://127.0.0.1:9200"
	}
	if os.Getenv("GHA2DB_ES_BULK_SIZE") == "" {
		ctx.ESBulkSize = 10000
	} else {
		size, err := strconv.Atoi(os.Getenv("GHA2DB_ES_BULK_SIZE"))
		FatalNoLog(err)
		ctx.ESBulkSize = size
	}

	// HTTP Timeout
	if os.Getenv("GHA2DB_HTTP_TIMEOUT") == "" {
		ctx.HTTPTimeout = 3
	} else {
		size, err := strconv.Atoi(os.Getenv("GHA2DB_HTTP_TIMEOUT"))
		FatalNoLog(err)
		ctx.HTTPTimeout = size
	}
	// HTTP RETRY
	if os.Getenv("GHA2DB_HTTP_RETRY") == "" {
		ctx.HTTPRetry = 5
	} else {
		retry, err := strconv.Atoi(os.Getenv("GHA2DB_HTTP_RETRY"))
		FatalNoLog(err)
		ctx.HTTPRetry = retry
	}

	// Skip writing to shared_db from projects.yaml
	ctx.SkipSharedDB = os.Getenv("GHA2DB_SKIP_SHAREDDB") != ""

	// Skip PID file
	ctx.SkipPIDFile = os.Getenv("GHA2DB_SKIP_PIDFILE") != ""

	// Skip company acquisitions file
	ctx.SkipCompanyAcq = os.Getenv("GHA2DB_SKIP_COMPANY_ACQ") != ""

	// Check provision flag
	ctx.CheckProvisionFlag = os.Getenv("GHA2DB_CHECK_PROVISION_FLAG") != ""

	// Set provision flag
	ctx.CheckRunningFlag = os.Getenv("GHA2DB_CHECK_RUNNING_FLAG") != ""

	// Set running flag
	ctx.SetRunningFlag = os.Getenv("GHA2DB_SET_RUNNING_FLAG") != ""

	mrfaS := os.Getenv("GHA2DB_MAX_RUNNING_FLAG_AGE")
	if mrfaS == "" {
		d, _ := time.ParseDuration("6h")
		ctx.MaxRunningFlagAge = d
	} else {
		d, err := time.ParseDuration(mrfaS)
		FatalNoLog(err)
		ctx.MaxRunningFlagAge = d
	}

	// Check Imported SHAs
	ctx.CheckImportedSHA = os.Getenv("GHA2DB_CHECK_IMPORTED_SHA") != ""
	ctx.OnlyCheckImportedSHA = os.Getenv("GHA2DB_ONLY_CHECK_IMPORTED_SHA") != ""

	// Calculate all periods?
	ctx.ComputeAll = os.Getenv("GHA2DB_COMPUTE_ALL") != ""

	// Forece compute periods
	periods := os.Getenv("GHA2DB_FORCE_PERIODS")
	if periods != "" {
		ary := strings.Split(periods, ",")
		for _, data := range ary {
			ary2 := strings.Split(data, ":")
			if len(ary2) != 2 {
				continue
			}
			period := ary2[0]
			shist := strings.TrimSpace(ary2[1])
			if shist != "t" && shist != "f" {
				continue
			}
			hist := false
			if shist == "t" {
				hist = true
			}
			if ctx.ComputePeriods == nil {
				ctx.ComputePeriods = make(map[string]map[bool]struct{})
			}
			_, ok := ctx.ComputePeriods[period]
			if !ok {
				ctx.ComputePeriods[period] = make(map[bool]struct{})
			}
			ctx.ComputePeriods[period][hist] = struct{}{}
		}
	}

	// Actor filtering?
	ctx.ActorsFilter = os.Getenv("GHA2DB_ACTORS_FILTER") != ""
	if ctx.ActorsFilter {
		actorsAllow := os.Getenv("GHA2DB_ACTORS_ALLOW")
		if actorsAllow != "" {
			ctx.ActorsAllow = regexp.MustCompile(actorsAllow)
		}
		actorsForbid := os.Getenv("GHA2DB_ACTORS_FORBID")
		if actorsForbid != "" {
			ctx.ActorsForbid = regexp.MustCompile(actorsForbid)
		}
	}

	// `merge_dbs` tool - input DBs and output DB
	dbs := os.Getenv("GHA2DB_INPUT_DBS")
	if dbs != "" {
		ctx.InputDBs = strings.Split(dbs, ",")
	}
	ctx.OutputDB = os.Getenv("GHA2DB_OUTPUT_DB")

	// RecentRange - ghapi2db will check issues/PRs from now() - this range to now()
	ctx.RecentRange = os.Getenv("GHA2DB_RECENT_RANGE")
	if ctx.RecentRange == "" {
		ctx.RecentRange = "2 hours"
	}
	ctx.RecentReposRange = os.Getenv("GHA2DB_RECENT_REPOS_RANGE")
	if ctx.RecentReposRange == "" {
		ctx.RecentReposRange = "1 day"
	}

	// Enable drop metrics support
	ctx.EnableMetricsDrop = os.Getenv("GHA2DB_ENABLE_METRICS_DROP") != ""

	// Project Scale
	if os.Getenv("GHA2DB_PROJECT_SCALE") == "" {
		ctx.ProjectScale = 1.0
	} else {
		projectScale, err := strconv.ParseFloat(os.Getenv("GHA2DB_PROJECT_SCALE"), 64)
		FatalNoLog(err)
		ctx.ProjectScale = projectScale
	}

	// CSV file
	ctx.CSVFile = os.Getenv("GHA2DB_CSVOUT")

	// RecalcReciprocal
	if os.Getenv("GHA2DB_RECALC_RECIPROCAL") == "" {
		ctx.RecalcReciprocal = 24
	} else {
		rr, err := strconv.Atoi(os.Getenv("GHA2DB_RECALC_RECIPROCAL"))
		FatalNoLog(err)
		if rr > 0 {
			ctx.RecalcReciprocal = rr
		} else {
			ctx.RecalcReciprocal = 24
		}
	}

	// MaxHistograms
	if os.Getenv("GHA2DB_MAX_HIST") != "" {
		mh, err := strconv.Atoi(os.Getenv("GHA2DB_MAX_HIST"))
		FatalNoLog(err)
		if mh > 0 {
			ctx.MaxHistograms = mh
		}
	}

	// Context out if requested
	if ctx.CtxOut {
		ctx.Print()
	}
}

// Print context contents
func (ctx *Ctx) Print() {
	fmt.Printf("Environment Context Dump\n%+v\n", ctx)
}
