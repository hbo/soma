package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/client9/reopen"
	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/config"
	"github.com/mjolnir42/soma/internal/rest"
	"github.com/mjolnir42/soma/internal/soma"
	"github.com/mjolnir42/soma/internal/super"
	metrics "github.com/rcrowley/go-metrics"
)

// global variables
var (
	// main database connection pool
	conn *sql.DB
	// lookup table for go routine input channels
	handlerMap = make(map[string]interface{})
	// config file runtime configuration
	SomaCfg config.Config
	// Orderly shutdown of the system has been called. GrimReaper is active
	ShutdownInProgress = false
	// lookup table of logfile handles for logrotate reopen
	logFileMap = make(map[string]*reopen.FileWriter)
	// Global metrics registry
	Metrics = make(map[string]metrics.Registry)
	// version string set at compile time
	somaVersion string
)

const (
	// Format string for millisecond precision RFC3339
	rfc3339Milli string = "2006-01-02T15:04:05.000Z07:00"
)

// Logging format strings
const (
	LogStrReq = `Subsystem=%s, Request=%s, User=%s, Addr=%s`
	LogStrSRq = `Section=%s, Action=%s, User=%s, Addr=%s`
	LogStrArg = `Subsystem=%s, Request=%s, User=%s, Addr=%s, Arg=%s`
	LogStrOK  = `Section=%s, Action=%s, InternalCode=%d, ExternalCode=%d`
	LogStrErr = `Section=%s, Action=%s, InternalCode=%d, Error=%s`
)

func init() {
	log.SetOutput(os.Stderr)
}

func main() {
	var (
		configFlag, configFile, obsRepoFlag         string
		noPokeFlag, forcedCorruption, versionFlag   bool
		err                                         error
		appLog, reqLog, errLog, auditLog            *log.Logger
		lfhGlobal, lfhApp, lfhReq, lfhErr, lfhAudit *reopen.FileWriter
		app                                         *soma.Soma
		hm                                          soma.HandlerMap
		lm                                          soma.LogHandleMap
		rst                                         *rest.Rest
	)

	// Daemon command line flags
	flag.StringVar(&configFlag, "config", "/srv/soma/huxley/conf/soma.conf", "Configuration file location")
	flag.StringVar(&obsRepoFlag, "repo", "", "Single-repository mode target repository")
	flag.BoolVar(&noPokeFlag, "nopoke", false, "Disable lifecycle pokes")
	flag.BoolVar(&forcedCorruption, `allowdatacorruption`, false, `Allow single-repo mode on production`)
	flag.BoolVar(&versionFlag, `version`, false, `Print version information`)
	flag.Parse()

	if versionFlag {
		version() // exit(0)
	}

	log.Printf("Starting runtime config initialization, SOMA v%s", somaVersion)
	/*
	 * Read configuration file
	 */
	if configFile, err = filepath.Abs(configFlag); err != nil {
		log.Fatal(err)
	}
	if configFile, err = filepath.EvalSymlinks(configFile); err != nil {
		log.Fatal(err)
	}
	err = SomaCfg.ReadConfigFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Open logfiles
	if lfhGlobal, err = reopen.NewFileWriter(
		filepath.Join(SomaCfg.LogPath, `global.log`),
	); err != nil {
		log.Fatalf("Unable to open global output log: %s", err)
	}
	log.SetOutput(lfhGlobal)
	logFileMap[`global`] = lfhGlobal

	appLog = log.New()
	if lfhApp, err = reopen.NewFileWriter(
		filepath.Join(SomaCfg.LogPath, `application.log`),
	); err != nil {
		log.Fatalf("Unable to open application log: %s", err)
	}
	appLog.Out = lfhApp
	logFileMap[`application`] = lfhApp

	reqLog = log.New()
	if lfhReq, err = reopen.NewFileWriter(
		filepath.Join(SomaCfg.LogPath, `request.log`),
	); err != nil {
		log.Fatalf("Unable to open request log: %s", err)
	}
	reqLog.Out = lfhReq
	logFileMap[`request`] = lfhReq

	errLog = log.New()
	if lfhErr, err = reopen.NewFileWriter(
		filepath.Join(SomaCfg.LogPath, `error.log`),
	); err != nil {
		log.Fatalf("Unable to open error log: %s", err)
	}
	errLog.Out = lfhErr
	logFileMap[`error`] = lfhErr

	auditLog = log.New()
	if lfhAudit, err = reopen.NewFileWriter(
		filepath.Join(SomaCfg.LogPath, `audit.log`),
	); err != nil {
		log.Fatalf("Unable to open audit log: %s", err)
	}
	auditLog.Out = lfhAudit
	logFileMap[`audit`] = lfhAudit

	// signal handler will reopen all logfiles on USR2
	sigChanLogRotate := make(chan os.Signal, 1)
	signal.Notify(sigChanLogRotate, syscall.SIGUSR2)
	go logrotate(sigChanLogRotate)

	// print selected runtime mode
	if SomaCfg.ReadOnly {
		appLog.Println(`Instance has been configured as: read-only mode`)
	} else if SomaCfg.Observer {
		appLog.Println(`Instance has been configured as: observer mode`)
	} else {
		appLog.Println(`Instance has been configured as: normal mode`)
	}

	// single-repo cli argument overwrites config file
	if obsRepoFlag != `` {
		SomaCfg.ObserverRepo = obsRepoFlag
	}
	if SomaCfg.ObserverRepo != `` {
		appLog.Printf("Single-repository mode active for: %s", SomaCfg.ObserverRepo)
	}

	// disallow single-repository mode on production r/w instances
	if !SomaCfg.ReadOnly && !SomaCfg.Observer &&
		SomaCfg.ObserverRepo != `` && SomaCfg.Environment == `production` &&
		!forcedCorruption {
		errLog.Fatal(`Single-repository r/w mode disallowed for production environments. ` +
			`Use the -allowdatacorruption flag if you are sure this will be the only ` +
			`running SOMA instance.`)
	}

	if noPokeFlag {
		SomaCfg.NoPoke = true
		appLog.Println(`Instance has disabled outgoing pokes by lifeCycle manager`)
	}

	/*
	 * Register metrics collections
	 */
	Metrics[`golang`] = metrics.NewPrefixedRegistry(`golang.`)
	metrics.RegisterRuntimeMemStats(Metrics[`golang`])
	go metrics.CaptureRuntimeMemStats(Metrics[`golang`], time.Second*60)

	Metrics[`soma`] = metrics.NewPrefixedRegistry(`soma`)
	Metrics[`soma`].Register(`requests.latency`,
		// TODO NewCustomTimer(Histogram, Meter) so there is access
		// to Histogram.Clear()
		metrics.NewTimer())
	soma.Metrics = Metrics
	rest.Metrics = Metrics

	/*
	 * Construct listen address
	 */
	SomaCfg.Daemon.URL = &url.URL{}
	SomaCfg.Daemon.URL.Host = fmt.Sprintf("%s:%s", SomaCfg.Daemon.Listen, SomaCfg.Daemon.Port)
	if SomaCfg.Daemon.TLS {
		SomaCfg.Daemon.URL.Scheme = "https"
		if ok, pt := govalidator.IsFilePath(SomaCfg.Daemon.Cert); !ok {
			errLog.Fatal("Missing required certificate configuration config/daemon/cert-file")
		} else {
			if pt != govalidator.Unix {
				errLog.Fatal("config/daemon/cert-File: valid Windows paths are not helpful")
			}
		}
		if ok, pt := govalidator.IsFilePath(SomaCfg.Daemon.Key); !ok {
			errLog.Fatal("Missing required key configuration config/daemon/key-file")
		} else {
			if pt != govalidator.Unix {
				errLog.Fatal("config/daemon/key-file: valid Windows paths are not helpful")
			}
		}
	} else {
		SomaCfg.Daemon.URL.Scheme = "http"
	}

	connectToDatabase(appLog, errLog)
	go pingDatabase(errLog)

	hm = soma.HandlerMap{}
	lm = soma.LogHandleMap{}

	app = soma.New(&hm, &lm, conn, &SomaCfg, appLog, reqLog, errLog, auditLog)
	app.Start()

	rst = rest.New(super.IsAuthorized, &hm, &SomaCfg)

	//XXX compilefix
	router := httprouter.New()

	router.HEAD(`/`, Check(Ping))

	router.GET(`/buckets/:bucket/instances/`, Check(BasicAuth(InstanceList)))
	router.GET(`/buckets/:bucket/tree/:tree`, Check(BasicAuth(OutputTree)))
	router.GET(`/category/:category/permissions/:permission`, Check(BasicAuth(PermissionShow)))
	router.GET(`/category/:category/permissions/`, Check(BasicAuth(PermissionList)))
	router.GET(`/cluster/:clusterID/instances/`, Check(BasicAuth(InstanceList)))
	router.GET(`/cluster/:clusterID/tree/:tree`, Check(BasicAuth(OutputTree)))
	//TODO router.GET(`/category/:category/permissions/:permission/grant/`)
	//TODO router.GET(`/category/:category/permissions/:permission/grant/:grant`)
	router.GET(`/groups/:group/instances/`, Check(BasicAuth(InstanceList)))
	router.GET(`/groups/:group/members/`, Check(BasicAuth(GroupListMember)))
	router.GET(`/groups/:group/tree/:tree`, Check(BasicAuth(OutputTree)))
	router.GET(`/groups/:group`, Check(BasicAuth(GroupShow)))
	router.GET(`/groups/`, Check(BasicAuth(GroupList)))
	router.GET(`/hostdeployment/:system/:assetid`, Check(HostDeploymentFetch))
	router.GET(`/instances/:instance/versions`, Check(BasicAuth(InstanceVersions)))
	router.GET(`/instances/:instance`, Check(BasicAuth(InstanceShow)))
	router.GET(`/instances/`, Check(BasicAuth(InstanceListAll)))
	router.GET(`/jobs/id/:jobid`, Check(BasicAuth(JobShow)))
	router.GET(`/jobs/all`, Check(BasicAuth(JobListAll)))
	router.GET(`/jobs/`, Check(BasicAuth(JobList)))
	router.GET(`/levels/:level`, Check(BasicAuth(LevelShow)))
	router.GET(`/levels/`, Check(BasicAuth(LevelList)))
	router.GET(`/metrics/:metric`, Check(BasicAuth(MetricShow)))
	router.GET(`/metrics/`, Check(BasicAuth(MetricList)))
	router.GET(`/modes/:mode`, Check(BasicAuth(ModeShow)))
	router.GET(`/modes/`, Check(BasicAuth(ModeList)))
	router.GET(`/monitoring/:monitoring`, Check(BasicAuth(MonitoringShow)))
	router.GET(`/monitoring/`, Check(BasicAuth(MonitoringList)))
	router.GET(`/nodes/:node/config`, Check(BasicAuth(NodeShowConfig)))
	router.GET(`/nodes/:node/instances/`, Check(BasicAuth(InstanceList)))
	router.GET(`/nodes/:node`, Check(BasicAuth(NodeShow)))
	router.GET(`/nodes/`, Check(BasicAuth(NodeList)))
	router.GET(`/oncall/:oncall`, Check(BasicAuth(OncallShow)))
	router.GET(`/oncall/`, Check(BasicAuth(OncallList)))
	router.GET(`/predicates/:predicate`, Check(BasicAuth(PredicateShow)))
	router.GET(`/predicates/`, Check(BasicAuth(PredicateList)))
	router.GET(`/property/custom/:repository/:custom`, Check(BasicAuth(PropertyShow)))
	router.GET(`/property/custom/:repository/`, Check(BasicAuth(PropertyList)))
	router.GET(`/property/native/:native`, Check(BasicAuth(PropertyShow)))
	router.GET(`/property/native/`, Check(BasicAuth(PropertyList)))
	router.GET(`/property/service/global/:service`, Check(BasicAuth(PropertyShow)))
	router.GET(`/property/service/global/`, Check(BasicAuth(PropertyList)))
	router.GET(`/property/service/team/:team/:service`, Check(BasicAuth(PropertyShow)))
	router.GET(`/property/service/team/:team/`, Check(BasicAuth(PropertyList)))
	router.GET(`/property/system/:system`, Check(BasicAuth(PropertyShow)))
	router.GET(`/property/system/`, Check(BasicAuth(PropertyList)))
	router.GET(`/providers/:provider`, Check(BasicAuth(ProviderShow)))
	router.GET(`/providers/`, Check(BasicAuth(ProviderList)))
	router.GET(`/repository/:repository/instances/`, Check(BasicAuth(InstanceList)))
	router.GET(`/repository/:repository/tree/:tree`, Check(BasicAuth(OutputTree)))
	router.GET(`/repository/:repository`, Check(BasicAuth(RepositoryShow)))
	router.GET(`/repository/`, Check(BasicAuth(RepositoryList)))
	router.GET(`/sections/:section`, Check(BasicAuth(SectionShow)))
	router.GET(`/sections/`, Check(BasicAuth(SectionList)))
	router.GET(`/servers/:server`, Check(BasicAuth(ServerShow)))
	router.GET(`/servers/`, Check(BasicAuth(ServerList)))
	router.GET(`/states/:state`, Check(BasicAuth(StateShow)))
	router.GET(`/states/`, Check(BasicAuth(StateList)))
	router.GET(`/status/:status`, Check(BasicAuth(StatusShow)))
	router.GET(`/status/`, Check(BasicAuth(StatusList)))
	router.GET(`/sync/servers/`, Check(BasicAuth(ServerSync)))
	router.GET(`/sync/teams/`, Check(BasicAuth(TeamSync)))
	router.GET(`/sync/users/`, Check(BasicAuth(UserSync)))
	router.GET(`/teams/:team`, Check(BasicAuth(TeamShow)))
	router.GET(`/teams/`, Check(BasicAuth(TeamList)))
	router.GET(`/units/:unit`, Check(BasicAuth(UnitShow)))
	router.GET(`/units/`, Check(BasicAuth(UnitList)))
	router.GET(`/users/:user`, Check(BasicAuth(UserShow)))
	router.GET(`/users/`, Check(BasicAuth(UserList)))
	router.GET(`/views/:view`, Check(BasicAuth(ViewShow)))
	router.GET(`/views/`, Check(BasicAuth(ViewList)))
	router.GET(`/workflow/summary`, Check(BasicAuth(WorkflowSummary)))
	router.POST(`/filter/grant/`, Check(BasicAuth(RightSearch)))
	router.POST(`/filter/groups/`, Check(BasicAuth(GroupList)))
	router.POST(`/filter/jobs/`, Check(BasicAuth(JobSearch)))
	router.POST(`/filter/levels/`, Check(BasicAuth(LevelList)))
	router.POST(`/filter/monitoring/`, Check(BasicAuth(MonitoringSearch)))
	router.POST(`/filter/nodes/`, Check(BasicAuth(NodeList)))
	router.POST(`/filter/oncall/`, Check(BasicAuth(OncallList)))
	router.POST(`/filter/permission/`, Check(BasicAuth(PermissionSearch)))
	router.POST(`/filter/property/custom/:repository/`, Check(BasicAuth(PropertyList)))
	router.POST(`/filter/property/service/global/`, Check(BasicAuth(PropertyList)))
	router.POST(`/filter/property/service/team/:team/`, Check(BasicAuth(PropertyList)))
	router.POST(`/filter/property/system/`, Check(BasicAuth(PropertyList)))
	router.POST(`/filter/repository/`, Check(BasicAuth(RepositoryList)))
	router.POST(`/filter/sections/`, Check(BasicAuth(SectionSearch)))
	router.POST(`/filter/servers/`, Check(BasicAuth(ServerSearch)))
	router.POST(`/filter/teams/`, Check(BasicAuth(TeamList)))
	router.POST(`/filter/users/`, Check(BasicAuth(UserList)))
	router.POST(`/filter/workflow/`, Check(BasicAuth(WorkflowList)))
	router.POST(`/hostdeployment/:system/:assetid`, Check(HostDeploymentAssemble))

	if !SomaCfg.ReadOnly {

		if !SomaCfg.Observer {
			router.DELETE(`/category/:category/permissions/:permission/grant/:grant`, Check(BasicAuth(RightRevoke)))
			router.DELETE(`/category/:category/permissions/:permission`, Check(BasicAuth(PermissionRemove)))
			router.DELETE(`/groups/:group/property/:type/:source`, Check(BasicAuth(GroupRemoveProperty)))
			router.DELETE(`/levels/:level`, Check(BasicAuth(LevelRemove)))
			router.DELETE(`/metrics/:metric`, Check(BasicAuth(MetricRemove)))
			router.DELETE(`/modes/:mode`, Check(BasicAuth(ModeRemove)))
			router.DELETE(`/monitoring/:monitoring`, Check(BasicAuth(MonitoringRemove)))
			router.DELETE(`/nodes/:node/property/:type/:source`, Check(BasicAuth(NodeRemoveProperty)))
			router.DELETE(`/oncall/:oncall`, Check(BasicAuth(OncallRemove)))
			router.DELETE(`/predicates/:predicate`, Check(BasicAuth(PredicateRemove)))
			router.DELETE(`/property/custom/:repository/:custom`, Check(BasicAuth(PropertyRemove)))
			router.DELETE(`/property/native/:native`, Check(BasicAuth(PropertyRemove)))
			router.DELETE(`/property/service/global/:service`, Check(BasicAuth(PropertyRemove)))
			router.DELETE(`/property/service/team/:team/:service`, Check(BasicAuth(PropertyRemove)))
			router.DELETE(`/property/system/:system`, Check(BasicAuth(PropertyRemove)))
			router.DELETE(`/providers/:provider`, Check(BasicAuth(ProviderRemove)))
			router.DELETE(`/repository/:repository/property/:type/:source`, Check(BasicAuth(RepositoryRemoveProperty)))
			router.DELETE(`/sections/:section`, Check(BasicAuth(SectionRemove)))
			router.DELETE(`/servers/:server`, Check(BasicAuth(ServerRemove)))
			router.DELETE(`/states/:state`, Check(BasicAuth(StateRemove)))
			router.DELETE(`/status/:status`, Check(BasicAuth(StatusRemove)))
			router.DELETE(`/teams/:team`, Check(BasicAuth(TeamRemove)))
			router.DELETE(`/units/:unit`, Check(BasicAuth(UnitRemove)))
			router.DELETE(`/users/:user`, Check(BasicAuth(UserRemove)))
			router.DELETE(`/views/:view`, Check(BasicAuth(ViewRemove)))
			router.GET(`/deployments/id/:uuid`, Check(DeploymentDetailsInstance))
			router.GET(`/deployments/monitoring/:uuid/:all`, Check(DeploymentDetailsMonitoring))
			router.GET(`/deployments/monitoring/:uuid`, Check(DeploymentDetailsMonitoring))
			router.PATCH(`/category/:category/permissions/:permission`, Check(BasicAuth(PermissionEdit)))
			router.PATCH(`/deployments/id/:uuid/:result`, Check(DeploymentDetailsUpdate))
			router.PATCH(`/oncall/:oncall`, Check(BasicAuth(OncallUpdate)))
			router.PATCH(`/views/:view`, Check(BasicAuth(ViewRename)))
			router.PATCH(`/workflow/instanceconfig/:instanceconfig`, Check(BasicAuth(WorkflowSet)))
			router.PATCH(`/workflow/retry`, Check(BasicAuth(WorkflowRetry)))
			router.POST(`/category/:category/permissions/:permission/grant/`, Check(BasicAuth(RightGrant)))
			router.POST(`/category/:category/permissions/`, Check(BasicAuth(PermissionAdd)))
			router.POST(`/groups/:group/members/`, Check(BasicAuth(GroupAddMember)))
			router.POST(`/groups/:group/property/:type/`, Check(BasicAuth(GroupAddProperty)))
			router.POST(`/groups/`, Check(BasicAuth(GroupCreate)))
			router.POST(`/levels/`, Check(BasicAuth(LevelAdd)))
			router.POST(`/metrics/`, Check(BasicAuth(MetricAdd)))
			router.POST(`/modes/`, Check(BasicAuth(ModeAdd)))
			router.POST(`/monitoring/`, Check(BasicAuth(MonitoringAdd)))
			router.POST(`/nodes/:node/property/:type/`, Check(BasicAuth(NodeAddProperty)))
			router.POST(`/oncall/`, Check(BasicAuth(OncallAdd)))
			router.POST(`/predicates/`, Check(BasicAuth(PredicateAdd)))
			router.POST(`/property/custom/:repository/`, Check(BasicAuth(PropertyAdd)))
			router.POST(`/property/native/`, Check(BasicAuth(PropertyAdd)))
			router.POST(`/property/service/global/`, Check(BasicAuth(PropertyAdd)))
			router.POST(`/property/service/team/:team/`, Check(BasicAuth(PropertyAdd)))
			router.POST(`/property/system/`, Check(BasicAuth(PropertyAdd)))
			router.POST(`/providers/`, Check(BasicAuth(ProviderAdd)))
			router.POST(`/repository/:repository/property/:type/`, Check(BasicAuth(RepositoryAddProperty)))
			router.POST(`/repository/`, Check(BasicAuth(RepositoryCreate)))
			router.POST(`/sections/`, Check(BasicAuth(SectionAdd)))
			router.POST(`/servers/:server`, Check(BasicAuth(ServerAddNull)))
			router.POST(`/servers/`, Check(BasicAuth(ServerAdd)))
			router.POST(`/states/`, Check(BasicAuth(StateAdd)))
			router.POST(`/status/`, Check(BasicAuth(StatusAdd)))
			router.POST(`/system/`, Check(BasicAuth(SystemOperation)))
			router.POST(`/teams/`, Check(BasicAuth(TeamAdd)))
			router.POST(`/units/`, Check(BasicAuth(UnitAdd)))
			router.POST(`/users/`, Check(BasicAuth(UserAdd)))
			router.POST(`/views/`, Check(BasicAuth(ViewAdd)))
			router.PUT(`/jobs/id/:jobid`, Check(BasicAuth(JobDelay)))
			router.PUT(`/nodes/:node/config`, Check(BasicAuth(NodeAssign)))
			router.PUT(`/servers/:server`, Check(BasicAuth(ServerUpdate)))
			router.PUT(`/states/:state`, Check(BasicAuth(StateRename)))
			router.PUT(`/teams/:team`, Check(BasicAuth(TeamUpdate)))
			router.PUT(`/users/:user`, Check(BasicAuth(UserUpdate)))
		}
	}

	go rst.Run()

	//XXX wait for shutdown
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
