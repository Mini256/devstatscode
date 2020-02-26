package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	lib "github.com/cncf/devstatscode"
	yaml "gopkg.in/yaml.v2"
)

var (
	gNameToDB map[string]string
	gMtx      *sync.RWMutex
)

type apiPayload struct {
	API     string                 `json:"api"`
	Payload map[string]interface{} `json:"payload"`
}

type errorPayload struct {
	Error string `json:"error"`
}

type healthPayload struct {
	Project string `json:"project"`
	DB      string `json:"db_name"`
	Events  int    `json:"events"`
}

func returnError(w http.ResponseWriter, err error) {
	epl := errorPayload{Error: err.Error()}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(epl)
}

func nameToDB(name string) (db string, err error) {
	gMtx.RLock()
	db, ok := gNameToDB[name]
	gMtx.RUnlock()
	if !ok {
		err = fmt.Errorf("database not found for project '%s'", name)
	}
	return
}

func handleSharedPayload(apiName string, w http.ResponseWriter, payload map[string]interface{}) (project, db string, passed bool) {
	if len(payload) == 0 {
		returnError(w, fmt.Errorf("API '%s' 'payload' section empty or missing", apiName))
		return
	}
	iproject, ok := payload["project"]
	if !ok {
		returnError(w, fmt.Errorf("API '%s' missing 'project' field in 'payload' section", apiName))
		return
	}
	project, ok = iproject.(string)
	if !ok {
		returnError(w, fmt.Errorf("API '%s' 'payload' 'project' field '%+v' is not a string", apiName, iproject))
		return
	}
	db, err := nameToDB(project)
	if err != nil {
		returnError(w, err)
		return
	}
	passed = true
	return
}

func getPayloadStringParam(apiName, paramName string, w http.ResponseWriter, payload map[string]interface{}) (param string, passed bool) {
	iparam, ok := payload[paramName]
	if !ok {
		returnError(w, fmt.Errorf("API '%s' missing '%s' field in 'payload' section", apiName, paramName))
		return
	}
	param, ok = iparam.(string)
	if !ok {
		returnError(w, fmt.Errorf("API '%s' 'payload' '%s' field '%+v' is not a string", apiName, paramName, iparam))
		return
	}
	passed = true
	return
}

func apiDevActCntRepoGrp(w http.ResponseWriter, payload map[string]interface{}) {
	apiName := lib.DevActCntRepoGrp
	project, db, ok := handleSharedPayload(apiName, w, payload)
	if !ok {
		return
	}
	params := map[string]string{"range": "", "metric": "", "repository_group": "", "country": "", "github_id": ""}
	for paramName := range params {
		paramValue, ok := getPayloadStringParam(apiName, paramName, w, payload)
		if !ok {
			return
		}
		params[paramName] = paramValue
	}
	fmt.Printf("project:%s db:%s params:%+v\n", project, db, params)
}

func apiHealth(w http.ResponseWriter, payload map[string]interface{}) {
	apiName := lib.Health
	project, db, ok := handleSharedPayload(apiName, w, payload)
	if !ok {
		return
	}
	var ctx lib.Ctx
	ctx.Init()
	ctx.PgHost = os.Getenv("PG_HOST_RO")
	ctx.PgUser = os.Getenv("PG_USER_RO")
	ctx.PgPass = os.Getenv("PG_PASS_RO")
	ctx.PgDB = db
	c, err := lib.PgConnErr(&ctx)
	if err != nil {
		returnError(w, err)
		return
	}
	defer func() { _ = c.Close() }()
	rows, err := lib.QuerySQL(c, &ctx, "select count(*) from gha_events")
	if err != nil {
		returnError(w, err)
		return
	}
	defer func() { _ = rows.Close() }()
	events := 0
	for rows.Next() {
		err = rows.Scan(&events)
		if err != nil {
			returnError(w, err)
			return
		}
	}
	err = rows.Err()
	if err != nil {
		returnError(w, err)
		return
	}
	hpl := healthPayload{Project: project, DB: db, Events: events}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hpl)
}

func handleAPI(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var pl apiPayload
	err := json.NewDecoder(req.Body).Decode(&pl)
	if err != nil {
		returnError(w, err)
		return
	}
	switch pl.API {
	case lib.Health:
		apiHealth(w, pl.Payload)
	case lib.DevActCntRepoGrp:
		apiDevActCntRepoGrp(w, pl.Payload)
	default:
		returnError(w, fmt.Errorf("unknown API '%s'", pl.API))
	}
}

func checkEnv() {
	requiredEnv := []string{"PG_PASS", "PG_PASS_RO", "PG_USER_RO", "PG_HOST_RO"}
	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			lib.Fatalf("%s env variable must be set", env)
		}
	}
}

func readProjects(ctx *lib.Ctx) {
	dataPrefix := ctx.DataDir
	if ctx.Local {
		dataPrefix = "./"
	}
	data, err := ioutil.ReadFile(dataPrefix + ctx.ProjectsYaml)
	lib.FatalOnError(err)
	var projects lib.AllProjects
	lib.FatalOnError(yaml.Unmarshal(data, &projects))
	gNameToDB = make(map[string]string)
	for projName, projData := range projects.Projects {
		disabled := projData.Disabled
		if disabled {
			continue
		}
		db := projData.PDB
		gNameToDB[projName] = db
		gNameToDB[projData.FullName] = db
	}
	gMtx = &sync.RWMutex{}
}

func serveAPI() {
	var ctx lib.Ctx
	ctx.Init()
	lib.Printf("Starting API serve\n")
	checkEnv()
	readProjects(&ctx)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGALRM)
	go func() {
		for {
			sig := <-sigs
			lib.Fatalf("Exiting due to signal %v\n", sig)
		}
	}()
	http.HandleFunc("/api/v1", handleAPI)
	lib.FatalOnError(http.ListenAndServe("0.0.0.0:8080", nil))
}

func main() {
	serveAPI()
	lib.Fatalf("serveAPI exited without error, returning error state anyway")
}
