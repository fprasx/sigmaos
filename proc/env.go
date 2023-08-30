package proc

import (
	"os"
	"strings"

	sp "sigmaos/sigmap"
)

// Environment variables which every proc expects to have.
const (
	SIGMAPRIVILEGEDPROC = "SIGMAKERNELPROC"
	SIGMADEBUGPID       = "SIGMADEBUGPID"
	//	SIGMAPID            = "SIGMAPID"
	//	SIGMAPROGRAM = "SIGMAPROGRAM"
	//	SIGMAPROCDIR   = "SIGMAPROCDIR"
	//	SIGMAPARENTDIR = "SIGMAPARENTDIR"
	SIGMACRASH     = "SIGMACRASH"
	SIGMAPARTITION = "SIGMAPARTITION"
	SIGMANETFAIL   = "SIGMANETFAIL"
	SIGMAPERF      = "SIGMAPERF"
	SIGMADEBUG     = "SIGMADEBUG"
	//	SIGMANAMED          = "SIGMANAMED"
	SIGMALOCAL    = "SIGMALOCAL"
	SIGMATAG      = "SIGMATAG"
	SIGMAROOTFS   = "SIGMAROOTFS"
	SIGMAREALM    = "SIGMAREALM"
	SIGMANET      = "SIGMANET"
	SIGMAJAEGERIP = "SIGMAJAEGERIP"
	SIGMAKERNEL   = "SIGMAKERNEL"
	SIGMAUPROCD   = "SIGMAUPROCD"
)

func SetSigmaDebugPid(pid string) {
	os.Setenv(SIGMADEBUGPID, pid)
}

func GetSigmaDebugPid() string {
	return os.Getenv(SIGMADEBUGPID)
}

func GetIsPrivilegedProc() bool {
	return os.Getenv(SIGMAPRIVILEGEDPROC) == "true"
}

func GetSigmaPerf() string {
	return os.Getenv(SIGMAPERF)
}

func GetSigmaDebug() string {
	return os.Getenv(SIGMADEBUG)
}

func GetSigmaLocal() string {
	return os.Getenv(SIGMALOCAL)
}

func SetSigmaLocal(ip string) {
	os.Setenv(SIGMALOCAL, ip)
}

func GetSigmaRootFs() string {
	return os.Getenv(SIGMAROOTFS)
}

func SetSigmaJaegerIP(ip string) {
	os.Setenv(SIGMAJAEGERIP, ip)
}

func GetSigmaJaegerIP() string {
	return os.Getenv(SIGMAJAEGERIP)
}

func GetRealm() sp.Trealm {
	return sp.Trealm(os.Getenv(SIGMAREALM))
}

func GetKernelId() string {
	return os.Getenv(SIGMAKERNEL)
}

func GetNet() string {
	return os.Getenv(SIGMANET)
}

func GetUprocdPid() sp.Tpid {
	return sp.Tpid(os.Getenv(SIGMAUPROCD))
}

func GetBuildTag() string {
	return os.Getenv(SIGMATAG)
}

func GetLabels(envvar string) map[string]bool {
	m := make(map[string]bool)
	s := os.Getenv(envvar)
	if s == "" {
		return m
	}
	labels := strings.Split(s, ";")
	for _, l := range labels {
		m[l] = true
	}
	return m
}
