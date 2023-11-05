package perf

type Tselector string

func (t Tselector) String() string {
	return string(t)
}

// Suffixes
const (
	PPROF       Tselector = "_PPROF"
	PPROF_MEM             = "_PPROF_MEM"
	PPROF_MUTEX           = "_PPROF_MUTEX"
	PPROF_BLOCK           = "_PPROF_BLOCK"
	CPU                   = "_CPU"
	TPT                   = "_TPT"
)

// Tests & benchmarking
const (
	TEST  Tselector = "TEST"
	BENCH           = "BENCH"
)

// kernel procs
const (
	NAMED   Tselector = "NAMED"
	KNAMED            = "KNAMED"
	PROCD             = "PROCD"
	S3                = "S3"
	UX                = "UX"
	SCHEDD            = "SCHEDD"
	UPROCD            = "UPROCD"
	PROCQ             = "PROCQ"
	LCSCHED           = "LCSCHED"
)

// libs
const (
	GROUP Tselector = "GROUP"
)

// mr
const (
	MRMAPPER  Tselector = "MRMAPPER"
	MRREDUCER           = "MRREDUCER"
	SEQGREP             = "SEQGREP"
	SEQWC               = "SEQWC"
)

// mr
const (
	THUMBNAIL Tselector = "THUMBNAIL"
)

// kv
const (
	KVCLERK Tselector = "KVCLERK"
)

// hotel
const (
	HOTEL_WWW     Tselector = "HOTEL_WWW"
	HOTEL_GEO               = "HOTEL_GEO"
	HOTEL_RESERVE           = "HOTEL_RESERVE"
	HOTEL_SEARCH            = "HOTEL_SEARCH"
	HOTEL_RATE              = "HOTEL_RATE"
)

// socialnetwork
const (
	SOCIAL_NETWORK_FRONTEND Tselector = "SOCIAL_NETWORK_FRONTEND"
	SOCIAL_NETWORK_USER               = "SOCIAL_NETWORK_USER"
	SOCIAL_NETWORK_GRAPH              = "SOCIAL_NETWORK_GRAPH"
	SOCIAL_NETWORK_POST               = "SOCIAL_NETWORK_POST"
	SOCIAL_NETWORK_TIMELINE           = "SOCIAL_NETWORK_TIMELINE"
	SOCIAL_NETWORK_HOME               = "SOCIAL_NETWORK_HOME"
	SOCIAL_NETWORK_COMPOSE            = "SOCIAL_NETWORK_COMPOSE"
)

// cache
const (
	CACHECLERK Tselector = "CACHECLERK"
	CACHESRV             = "CACHESRV"
)

// microbenchmarks
const (
	WRITER         Tselector = "writer"
	BUFWRITER                = "bufwriter"
	ABUFWRITER               = "abufwriter"
	READER                   = "reader"
	BUFREADER                = "bufreader"
	ABUFREADER               = "abufreader"
	RPC_BENCH_SRV            = "RPC_BENCH_SRV"
	RPC_BENCH_CLNT           = "RPC_BENCH_CLNT"
)
