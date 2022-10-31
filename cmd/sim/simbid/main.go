package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"time"

	"gonum.org/v1/gonum/stat/distuv"
)

const (
	NTENANT = 100

	NTICK  = 100
	DEBUG  = false
	NTRIAL = 1

	NNODE             = 30
	NODES_PER_MACHINE = 1

	AVG_ARRIVAL_RATE float64 = 0.1 // per tick
	MAX_SERVICE_TIME         = 5   // in ticks
	MAX_LOAD         Load    = 0.8 // max load per node

	PRICE_ONDEMAND Price = 0.00000001155555555555 // per ms for 1h on AWS
	PRICE_SPOT     Price = 0.00000000347222222222 // per ms
	BID_INCREMENT  Price = 0.000000000001
	MAX_BID        Price = 3 * PRICE_SPOT
)

var nodes_per_machine = NODES_PER_MACHINE
var tick = Tick(0)

func zipf(r *rand.Rand) uint64 {
	z := rand.NewZipf(r, 2.0, 1.0, MAX_SERVICE_TIME-1)
	return z.Uint64() + 1
}

func uniform(r *rand.Rand) uint64 {
	return (rand.Uint64() % MAX_SERVICE_TIME) + 1
}

type Tpolicy func(*Tenant, Price) *Bid
type Tmid int

// Tick
type Tick uint64

func (t Tick) String() string {
	return fmt.Sprintf("%dT", t)
}

//
// Fractional tick
//

type FTick float64

func (f FTick) String() string {
	return fmt.Sprintf("%.1fT", f)
}

//
// Load
//

type Load float64

func (l Load) String() string {
	return fmt.Sprintf("%.2f%%", l)
}

//
// Price
//

type Price float64

func (p Price) String() string {
	return fmt.Sprintf("$%.12f", p)
}

//
// Bid
//

type Bid struct {
	tenant *Tenant
	bids   []Price // one bid per node
}

func (b *Bid) String() string {
	return fmt.Sprintf("{t: %p %v %d}", b.tenant, b.bids, len(b.bids))
}

func mkBid(t *Tenant, bs []Price) *Bid {
	return &Bid{t, bs}
}

type Bids []*Bid

func (bs *Bids) PopHighest(rand *rand.Rand) (*Tenant, Price) {
	bid := Price(0.0)

	if len(*bs) == 0 {
		return nil, bid
	}

	// find highest bid
	for _, b := range *bs {
		if b.bids[0] > bid {
			bid = b.bids[0]
		}
	}

	// find bidders for highest
	bidders := make([]int, 0)
	for i, b := range *bs {
		if b.bids[0] == bid {
			bidders = append(bidders, i)
		}
	}

	// pick a random higest
	n := bidders[int(rand.Uint64()%uint64(len(bidders)))]
	t := (*bs)[n].tenant

	// remove higest bid
	(*bs)[n].bids = (*bs)[n].bids[1:]
	if len((*bs)[n].bids) == 0 {
		*bs = append((*bs)[:n], (*bs)[n+1:]...)
	}
	return t, bid
}

//
// Tenants runs procs.  At each tick, each tenant creates new procs
// based AVG_ARRIVAL_RATE.  Each proc runs for nTick, following either
// uniform or zipfian distribution.
//

type Proc struct {
	nLength Tick  //
	nTick   FTick // # fractional ticks remaining
	time    Tick  // # ticks on a node
	cost    Price // cost for this proc
	// compute intensivity: 1.0 computes only, while 0.4 is doing i/o
	// for 0.6T.
	computeT FTick
}

func (p *Proc) String() string {
	return fmt.Sprintf("{n %v c %v t %v l %v}", p.nTick, p.computeT, p.time, p.nLength)
}

func mkProc(rand *rand.Rand) *Proc {
	p := &Proc{}
	// p.nTick = zipf(rand)
	t := Tick(uniform(rand))
	p.nTick = FTick(t)
	p.nLength = t
	p.computeT = 0.5
	return p
}

type Procs []*Proc

// Run procs until we used 1 tick of cpu power or we run out of procs.
// The last selected proc may run only for a fraction of tick.  Return
// how much we used of the 1 tick, and for procs that finished how
// much delay was incurred (because the nodes was overloaded).
func (ps *Procs) run(c Price) (Load, Tick) {
	load := Load(0.0)
	last := FTick(0.0)
	delay := Tick(0.0)

	// compute number of procs we can run until we hit 1 tick
	n := 0
	for _, p := range *ps {
		n += 1
		work := p.computeT
		if p.nTick < 1 { // only fraction of tick left?
			work = p.computeT * (1 - p.nTick)
		}
		f := FTick(1.0 - load)
		if work < f {
			last = p.computeT
			load += Load(work)
		} else {
			last = f
			load += Load(last)
			break
		}
	}

	for _, p := range *ps {
		p.time += 1
	}

	// run the first n
	qs := (*ps)[0:n]
	*ps = (*ps)[n:]
	for i, p := range qs {
		if i < len(qs)-1 {
			p.nTick--
		} else {
			p.nTick -= last / p.computeT
		}

		// charge every proc equally, even though last proc may not
		// get to run for p.computeT.
		p.cost += c / Price(len(qs))

		if p.nTick <= 0 { // p is done
			delay += p.time - p.nLength
		} else {
			// not done; put it at the end of procq so that procs run
			// round robin
			*ps = append(*ps, p)
		}
	}

	return load, delay
}

func (ps *Procs) wasted() (FTick, Price) {
	w := FTick(0)
	c := Price(0.0)
	for _, p := range *ps {
		w += FTick(p.nLength) - p.nTick // wasted ticks
		if w > 0 {
			c += p.cost
		}
	}
	return w, c
}

type Machine struct {
	mid     Tmid
	nodes   Nodes
	ntenant int
}

func (m *Machine) String() string {
	return fmt.Sprintf("{%d %d}", m.ntenant, len(m.nodes))
}

type Machines map[Tmid]*Machine

// Returns the m's (and their nodes) in ms that are also present in
// ms1 (which may have fewer nodes)
func (ms Machines) intersect(ms1 Machines) Machines {
	r := make(Machines)
	for k, m := range ms {
		if _, ok := ms1[k]; ok {
			r[k] = m
		}
	}
	return r
}

// Find the machine mostly heavily used
func (ms Machines) mostUsed() *Machine {
	var most *Machine
	high := 0
	for _, m := range ms {
		if m.ntenant > high && m.ntenant < nodes_per_machine {
			most = m
			high = m.ntenant
		}
	}
	return most
}

func (ms Machines) leastUsed() *Machine {
	var least *Machine
	low := NTENANT
	for _, m := range ms {
		if m.ntenant < low {
			least = m
			low = m.ntenant
		}
	}
	return least
}

func (ms Machines) nodeOnMachine(mid Tmid) *Node {
	load := Load(100)
	var r *Node
	for _, m := range ms {
		for _, n := range m.nodes {
			if n.mid == mid && n.load < load {
				r = n
				load = n.load
			}
		}
	}
	return r
}

//
// Computing nodes that the manager allocates to tenants.  Each node
// runs one proc or is idle.
//

type Node struct {
	procs  Procs
	price  Price // the price for a tick
	load   Load
	tenant *Tenant
	mid    Tmid
}

func (n *Node) String() string {
	return fmt.Sprintf("{%p: proc %v price %v l %v t %p m %d}", n, n.procs, n.price, n.load, n.tenant, n.mid)
}

func (n *Node) takeProcs(ps Procs) Procs {
	if n.load > MAX_LOAD {
		return ps
	}
	n.procs = append(n.procs, ps[0])
	ps = ps[1:]
	return ps
}

type Nodes []*Node

func (ns *Nodes) remove(n1 *Node) *Node {
	for i, n := range *ns {
		if n == n1 {
			*ns = append((*ns)[:i], (*ns)[i+1:]...)
			return n
		}
	}
	return nil
}

func (ns *Nodes) machines() Machines {
	ms := make(map[Tmid]*Machine)
	for _, n := range *ns {
		if _, ok := ms[n.mid]; !ok {
			m := &Machine{}
			m.nodes = make(Nodes, 0)
			ms[n.mid] = m
			m.mid = n.mid
		}
		m := ms[n.mid]
		m.nodes = append(m.nodes, n)
		if n.tenant != nil {
			m.ntenant += 1
		}
		if m.ntenant > nodes_per_machine {
			fmt.Printf("nodes %v\n", ns)
			panic("machines")
		}
	}
	return ms
}

func (ns *Nodes) findFree(tms Machines) *Node {
	if len(*ns) == 0 {
		return nil
	}
	fms := ns.machines()
	//ms1 := tms.intersect(fms)
	//m := ms1.mostUsed()
	//if m == nil {
	m := fms.leastUsed()
	//}
	for _, n := range m.nodes {
		if n.tenant == nil {
			ns.remove(n)
			return n
		}
	}
	return nil
}

func (ns *Nodes) findVictim(t *Tenant, bid Price) *Node {
	for _, n := range *ns {
		if n.tenant != t && bid > n.price {
			return n
		}
	}
	return nil
}

func (ns *Nodes) isPresent(nn *Node) bool {
	for _, n := range *ns {
		if n == nn {
			return true
		}
	}
	return false
}

func (ns *Nodes) check() {
	m := make(map[*Node]bool)
	for _, n := range *ns {
		_, ok := m[n]
		if !ok {
			m[n] = true
		} else {
			fmt.Printf("double %v\n", ns)
			panic("check")
		}

	}
}

// Schedule procs in ps on the nodes in ns
func (ns Nodes) schedule(ps Procs) Procs {
	for _, n := range ns {
		if len(ps) == 0 { // no procs left to schedule?
			break
		}
		ps = n.takeProcs(ps)
	}
	return ps
}

//
// Tenants run procs on the nodes allocated to them by the mgr. If
// they have more procs to run than available nodes, tenant bids for
// more nodes.
//

type Tenant struct {
	poisson  *distuv.Poisson
	procs    []*Proc
	nodes    Nodes
	nbid     int
	ngrant   int
	sim      *Sim
	nproc    int  // sum of # procs
	ntick    Tick // sum of # ticks
	maxnode  int
	nwork    Tick   // sum of # ticks running a proc
	cost     Price  // cost for nwork ticks
	nwait    Tick   // sum of # ticks waiting to be run
	ndelay   Tick   // sum of # extra ticks that proc was on node
	nmigrate uint64 // # procs migrated
	nevict   uint64 // # evicted procs
	nwasted  FTick  // sum # fticks wasted because of eviction
	sunkCost Price  // the cost of the wasted ticks
	policy   Tpolicy
}

func (t *Tenant) String() string {
	s := fmt.Sprintf("{nproc %d ntick %d procq (%d): [", t.nproc, t.ntick, len(t.procs))
	for _, p := range t.procs {
		s += fmt.Sprintf("{%v} ", p)
	}
	s += fmt.Sprintf("] nodes (%d): [", len(t.nodes))
	for _, n := range t.nodes {
		s += fmt.Sprintf("%v ", n)
	}
	ms := t.nodes.machines()
	s += fmt.Sprintf("ms (%d) %v", len(ms), ms)
	return s + "]}"
}

// New procs "arrive" based on Poisson distribution. Schedule queued
// procs on the available nodes, and release nodes we don't use.
func (t *Tenant) genProcs() (int, Tick) {
	nproc := int(t.poisson.Rand())
	len := Tick(0)
	for i := 0; i < nproc; i++ {
		p := mkProc(t.sim.rand)
		len += p.nLength
		t.procs = append(t.procs, p)
	}
	t.nproc += nproc
	t.procs = t.nodes.schedule(t.procs)
	t.yieldIdle()
	return nproc, len
}

func policyBigMore(t *Tenant, last Price) *Bid {
	bids := make([]Price, 0)
	if t == &t.sim.tenants[0] && len(t.nodes) == 0 {
		// very first bid for tenant 0, which has a higher load grab
		// one high-priced node to sustain the expected load of 1.
		bids = append(bids, PRICE_ONDEMAND)
		for i := 0; i < len(t.procs)-1; i++ {
			bids = append(bids, last)
		}
	} else {
		// Exponentential increase
		// for i := 0; i < len(t.procs); i++ {
		for i := 0; i < 1; i++ {
			bid := last + BID_INCREMENT*Price(len(t.procs))
			//bid := last + BID_INCREMENT
			//bid := last
			bids = append(bids, bid)
		}
	}
	return mkBid(t, bids)
}

// Bid one up from last, the lowest winning bid
func policyLast(t *Tenant, last Price) *Bid {
	bids := make([]Price, 0)
	for i := 0; i < len(t.procs); i++ {
		bids = append(bids, last+BID_INCREMENT)
	}
	return mkBid(t, bids)
}

func policyFixed(t *Tenant, last Price) *Bid {
	bids := make([]Price, 0)
	for i := 0; i < len(t.procs); i++ {
		bids = append(bids, PRICE_SPOT)
	}
	return mkBid(t, bids)
}

// Bid for new nodes if we have queued procs.  last is avg succesful
// bid in the last round.
func (t *Tenant) bid(last Price) *Bid {
	t.ngrant = 0
	t.nbid = len(t.procs)
	if len(t.procs) > 0 {
		return t.policy(t, last)
	}
	return nil
}

// mgr grants a node
func (t *Tenant) grantNode(n *Node) {
	t.ngrant++
	t.nodes = append(t.nodes, n)
	t.nodes.check()
	t.nodes.machines()
}

// After bidding, we may have received new nodes; use them.
func (t *Tenant) scheduleNodes() int {
	if DEBUG {
		if t.nbid > 0 && t.ngrant < t.nbid {
			fmt.Printf("%v %p: asked %d and received %d\n", tick, t, t.nbid, t.ngrant)
		}
	}

	t.procs = t.nodes[len(t.nodes)-t.ngrant:].schedule(t.procs)

	t.ntick += Tick(len(t.nodes))
	if len(t.nodes) > t.maxnode {
		t.maxnode = len(t.nodes)
	}
	t.nwait += Tick(uint64(len(t.procs)))
	return len(t.procs)
}

// Yield idle nodes, except if tenant "reserved" the node
func (t *Tenant) yieldIdle() {
	for i := 0; i < len(t.nodes); i++ {
		n := t.nodes[i]
		n.load = Load(0.0)
		if len(n.procs) == 0 && n.price != PRICE_ONDEMAND {
			t.nodes = append(t.nodes[0:i], t.nodes[i+1:]...)
			i--
			t.sim.mgr.yield(n)
		}
	}
}

// Manager is taking awy node n
func (t *Tenant) evict(n *Node) (uint64, uint64) {
	if t.nodes.remove(n) == nil {
		fmt.Printf("%p: n not found %v\n", t, n)
		panic("evict")
	}
	ms := t.nodes.machines()
	e := uint64(0)
	m := uint64(0)
	if n1 := ms.nodeOnMachine(n.mid); n1 != nil {
		if DEBUG {
			fmt.Printf("%v: Migrate %v to %v\n", tick, n, n1)
		}
		m += uint64(len(n.procs))
		n1.procs = append(n1.procs, n.procs...)
	} else {
		if DEBUG {
			fmt.Printf("%v: Evict %v\n", tick, n)
		}
		w, c := n.procs.wasted()
		e += uint64(len(n.procs))
		t.nwasted += w
		t.sim.mgr.nwasted += w
		t.sunkCost += c
	}
	t.nevict += e
	t.nmigrate += m
	n.procs = make(Procs, 0)
	return e, m
}

func (t *Tenant) isPresent(n *Node) bool {
	return t.nodes.isPresent(n)
}

func (t *Tenant) stats() {
	n := float64(NTICK)
	fmt.Printf("%p: p %dP l %v P/T %.2f T/P maxN %d work %v util %.2f nwait %v ndelay %v #migr %dP #evict %dP (waste %v) charge %v sunk %v tick %v\n", t, t.nproc, float64(t.nproc)/n, float64(t.ntick)/float64(t.nproc), t.maxnode, t.nwork, float64(t.nwork)/float64(t.ntick), t.nwait, t.ndelay, t.nmigrate, t.nevict, t.nwasted, t.cost, t.sunkCost, t.cost/Price(t.nwork))
}

//
// Manager assigns nodes to tenants
//

type Mgr struct {
	sim      *Sim
	free     Nodes
	cur      Nodes
	index    int
	revenue  Price
	nwork    int
	nidle    uint64
	nevict   uint64
	nmigrate uint64
	nwasted  FTick
	last     Price // lowest bid accepted in last tick
	avgbid   Price // avg bid in last tick
	high     Price // highest bid in last tick
}

func mkMgr(sim *Sim) *Mgr {
	m := &Mgr{}
	m.sim = sim
	ns := make(Nodes, NNODE, NNODE)
	for i, _ := range ns {
		ns[i] = &Node{}
		ns[i].mid = Tmid(i / nodes_per_machine)
	}
	m.free = ns
	m.last = PRICE_SPOT
	return m
}

func (m *Mgr) String() string {
	s := fmt.Sprintf("{mgr nodes:")
	for _, n := range m.cur {
		s += fmt.Sprintf("{%v} ", n)
	}
	return s + "}"
}

func (m *Mgr) stats() {
	n := NTICK * NNODE
	fmt.Printf("Mgr: last %v revenue %v avg rev/tick %v util %.2f idle %v nmigrate %dP nevict %dP nwasted %v\n", m.last, m.revenue, Price(float64(m.revenue)/float64(m.nwork)), float64(m.nwork)/float64(n), m.nidle, m.nmigrate, m.nevict, m.nwasted)
}

func (m *Mgr) yield(n *Node) {
	if DEBUG {
		fmt.Printf("%v: yield %v\n", tick, n)
	}
	n.tenant = nil
	m.free = append(m.free, n)
	m.cur.remove(n)
}

func (m *Mgr) collectBids() Bids {
	bids := make([]*Bid, 0)
	for i, _ := range m.sim.tenants {
		if b := m.sim.tenants[i].bid(m.last); b != nil {
			// sort the bids in b
			sort.Slice(b.bids, func(i, j int) bool {
				return b.bids[i] > b.bids[j]
			})
			bids = append(bids, b)
		}
	}
	return bids
}

func (m *Mgr) checkAssignment(s string) {
	for _, n := range m.cur {
		if !n.tenant.isPresent(n) {
			fmt.Printf("node %v\n", n)
			fmt.Printf("m.cur %v\n", m.cur)
			fmt.Printf("m.tenant.nodes %v\n", n.tenant.nodes)
			panic(s)
		}
	}
}

func (m *Mgr) assignNodes() (Nodes, Price) {
	m.checkAssignment("before")
	bids := m.collectBids()
	// fmt.Printf("bids %v #free nodes %d\n", bids, len(m.free))
	new := make(Nodes, 0)
	m.avgbid = Price(0.0)
	m.high = Price(0.0)
	naccept := 0

	for {
		t, bid := bids.PopHighest(m.sim.rand)
		if t == nil {
			break
		}

		ms := t.nodes.machines()

		m.last = bid
		if bid > m.high {
			m.high = bid
		}
		m.avgbid += bid

		// fmt.Printf("assignNodes: %p bid highest %v\n", t, bid)
		if n := m.free.findFree(ms); n != nil {
			n.tenant = t
			n.price = bid
			//fmt.Printf("assignNodes: allocate %p to %p at %v\n", n, t, bid)
			t.grantNode(n)
			new = append(new, n)
		} else if n := m.cur.findVictim(t, bid); n != nil {
			if DEBUG {
				fmt.Printf("%v: assignNodes: reallocate %v to %p at %v\n", tick, n, t, bid)
			}
			ev, mi := n.tenant.evict(n)
			n.tenant = t
			n.price = bid
			m.nevict += ev
			m.nmigrate += mi
			n.tenant.grantNode(n)
		} else {
			// fmt.Printf("assignNodes: no nodes left\n")
			break
		}
		naccept++
	}
	price := m.last
	m.cur = append(m.cur, new...)
	// fmt.Printf("assignment %d nodes: %v\n", len(m.cur), m.cur)
	m.checkAssignment("after")

	// if idle nodes, lower price
	idle := uint64(NNODE - len(m.cur))
	m.nidle += idle
	if idle > 0 {
		m.last -= BID_INCREMENT
	}

	// avg bid for stats
	if naccept > 0 {
		m.avgbid = m.avgbid / Price(naccept)
	}

	return m.cur, price
}

//
// Run simulation
//

type Sim struct {
	time     uint64
	tenants  [NTENANT]Tenant
	rand     *rand.Rand
	mgr      *Mgr
	nproc    int  // total # procs started
	len      Tick // sum of all procs len
	nprocq   uint64
	avgprice Price // avg price per tick
}

func mkSim(p Tpolicy) *Sim {
	sim := &Sim{}
	sim.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	sim.mgr = mkMgr(sim)
	for i := 0; i < NTENANT; i++ {
		t := &sim.tenants[i]
		t.procs = make([]*Proc, 0)
		t.sim = sim
		t.policy = p
		if i == 0 {
			t.poisson = &distuv.Poisson{Lambda: 10 * AVG_ARRIVAL_RATE}
		} else {
			t.poisson = &distuv.Poisson{Lambda: AVG_ARRIVAL_RATE}
		}
	}
	return sim
}

// At each tick, a tenants generates load in the form of procs that
// need to be run.
func (sim *Sim) genLoad() {
	for i, _ := range sim.tenants {
		p, l := sim.tenants[i].genProcs()
		sim.nproc += p
		sim.len += l
	}
}

func (sim *Sim) scheduleNodes() int {
	pq := 0
	for i, _ := range sim.tenants {
		pq += sim.tenants[i].scheduleNodes()
	}
	return pq
}

func (sim *Sim) printTenants(nn, pq int) {
	fmt.Printf("Tick %d nodes %d procq %d new price %v avgbid %v high %v", tick, nn, pq, sim.mgr.last, sim.mgr.avgbid, sim.mgr.high)
	for i, _ := range sim.tenants {
		t := &sim.tenants[i]
		if len(t.procs) > 0 || len(t.nodes) > 0 {
			fmt.Printf("\n%p: %v", t, t)
		}
	}
	fmt.Printf("\n")
}

func (sim *Sim) runProcs(ns Nodes, p Price) {
	sim.avgprice += p
	for _, n := range ns {
		l, d := n.procs.run(p)
		n.load = l
		n.tenant.ndelay += d
		n.tenant.cost += p
		n.tenant.nwork++
		sim.mgr.nwork++
	}
	sim.mgr.revenue += p * Price(len(ns))
}

func (sim *Sim) tick() {
	sim.genLoad()
	ns, p := sim.mgr.assignNodes()
	pq := sim.scheduleNodes()
	sim.nprocq += uint64(pq)

	if DEBUG {
		sim.printTenants(len(ns), pq)
	}
	sim.runProcs(ns, p)
}

func funcName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func runSim(p Tpolicy) {
	fmt.Printf("=== Policy %s (n/m %d)\n", funcName(p), nodes_per_machine)

	sim := mkSim(p)
	for tick = 0; tick < NTICK; tick++ {
		sim.tick()
	}
	if DEBUG {
		for i, _ := range sim.tenants {
			sim.tenants[i].stats()
		}
	} else {
		sim.tenants[0].stats()
		sim.tenants[1].stats()
		sim.tenants[2].stats()
	}
	sim.mgr.stats()
	n := float64(NTICK)
	fmt.Printf("nproc %dP len %v avg proclen %.2fT avg procq %.2fP/T avg price %v/T\n", sim.nproc, sim.len, float64(sim.len)/float64(sim.nproc), float64(sim.nprocq)/n, sim.avgprice/Price(n))
}

func main() {
	// policies := []Tpolicy{policyFixed, policyLast, policyBigMore}
	policies := []Tpolicy{policyBigMore}
	npm := []int{1, 5, NNODE}
	for i := 0; i < NTRIAL; i++ {
		for _, p := range policies {
			for _, n := range npm {
				nodes_per_machine = n
				runSim(p)
			}
		}
	}
}