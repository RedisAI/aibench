//

// This program has no knowledge of the internals of the endpoint.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/RedisAI/aibench/cmd/aibench_generate_data/fraud"
	"github.com/RedisAI/aibench/inference"
	"github.com/RedisAI/redisai-go/redisai"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
)

// Program option vars:
var (
	host               string
	mysqlHost          string
	pipelineSize       uint
	setBlob            bool
	setTensor          bool
	useRedis           bool
	useMysql           bool
	runner             *inference.LoadRunner
	cpool              *redis.Pool
	sqldb              *sql.DB
	rowBenchmarkNBytes = 8 + 120 + 1024
)

// Parse args:
func init() {
	runner = inference.NewLoadRunner()
	flag.StringVar(&host, "redis-host", "redis://localhost:6379", "Redis host address and port")
	flag.StringVar(&mysqlHost, "mysql-host", "perf:perf@tcp(127.0.0.1:3306)", "MySql host address and port")
	flag.UintVar(&pipelineSize, "pipeline", 10, "Redis pipeline size")
	flag.BoolVar(&setBlob, "set-blob", true, "Set reference data in plain binary safe Redis string format")
	flag.BoolVar(&setTensor, "set-tensor", true, "Set reference data in AI.TENSOR format")
	flag.BoolVar(&useRedis, "use-redis", false, "Use Redis as the reference data holder")
	flag.BoolVar(&useMysql, "use-mysql", false, "Use MySql as the reference data holder")
	flag.Parse()
	if useRedis {
		cpool = &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.DialURL(host) },
		}
	}
	if useMysql {
		var err error
		sqldb, err = sql.Open("mysql", mysqlHost)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Error connection to MySql %v", err))
		}
		db := initDB(sqldb,
			"create database if not exists test",
			"DROP TABLE IF EXISTS test.tbltensorblobs",
			"CREATE TABLE test.tbltensorblobs (id varchar(50) PRIMARY KEY, blobtensor LONGBLOB)",
		)
		db.SetMaxIdleConns(500)
	}
}

func main() {
	runner.RunLoad(&inference.RedisAIPool, newProcessor, 0)
}

type Loader struct {
	Wg      *sync.WaitGroup
	pclient *redisai.Client
	psqldb  *sql.DB
}

func (p *Loader) Close() {
	if useRedis {
		p.pclient.Close()
	}
	if useMysql {
		p.psqldb.Close()
	}
}

func newProcessor() inference.Loader { return &Loader{} }

func (p *Loader) Init(numWorker int, wg *sync.WaitGroup) {
	p.Wg = wg
	if useRedis {
		p.pclient = redisai.Connect(host, cpool)
		p.pclient.Pipeline(uint32(pipelineSize))
	}
	if useMysql {
		var err error
		psqldb, err := sql.Open("mysql", mysqlHost)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Error connection to MySql %v", err))
		}
		p.psqldb = psqldb
	}

}

func initDB(db *sql.DB, queries ...string) *sql.DB {
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Fatalf("error on %q: %v", query, err)
		}
	}
	return db
}

func (p *Loader) ProcessLoadQuery(q []byte, debug int) ([]*inference.Stat, uint64, error) {
	if len(q) != (1024 + 8 + 120) {
		log.Fatalf("wrong Row lenght. Expected Set:%d got %d\n", 1024+8+120, len(q))
	}
	tmp := make([]byte, 8)
	referenceValues := make([]byte, 1024)
	copy(tmp, q[0:8])
	copy(referenceValues, q[128:1152])

	idF := fraud.Uint64frombytes(tmp)
	id := "referenceTensor:{" + fmt.Sprintf("%d", int(idF)) + "}"
	idBlob := "referenceBLOB:{" + fmt.Sprintf("%d", int(idF)) + "}"
	issuedCommands := 0
	if useRedis {
		p.pclient.ActiveConnNX()
		if setBlob {
			errSet := p.pclient.ActiveConn.Send("SET", idBlob, referenceValues)
			if errSet != nil {
				log.Fatal(errSet)
			}
			issuedCommands++
		}
		if setTensor {
			err := p.pclient.TensorSet(id, redisai.TypeFloat, []int64{1, 256}, referenceValues)
			if err != nil {
				log.Fatal(err)
			}
			issuedCommands++
		}
	}
	if useMysql {
		_, err := p.psqldb.Exec("INSERT INTO test.tbltensorblobs VALUES (?, ?)", idBlob, referenceValues)
		if err != nil {
			log.Fatal(err)
		}
		issuedCommands++
	}

	return nil, uint64(issuedCommands), nil
}
