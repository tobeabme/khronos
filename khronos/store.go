package khronos

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/abronan/valkeyrie"
	"github.com/abronan/valkeyrie/store"
	etcd "github.com/abronan/valkeyrie/store/etcd/v3"
	log "github.com/sirupsen/logrus"
)

const MaxExecutions = 200

type Store struct {
	Client   store.Store
	keyspace string
	backend  string
}

func init() {
	etcd.Register()
}

func NewStore(backend string, machines []string, keyspace string) *Store {
	s, err := valkeyrie.NewStore(store.Backend(backend), machines, nil)
	if err != nil {
		log.Error(err)
	}

	log.WithFields(log.Fields{
		"backend":  backend,
		"machines": machines,
		"keyspace": keyspace,
	}).Debug("store: Backend config")

	_, err = s.List(keyspace, nil)
	if err != store.ErrKeyNotFound && err != nil {
		log.WithError(err).Fatal("store: Store backend not reachable")
	}

	return &Store{Client: s, keyspace: keyspace, backend: backend}
}

// Store a job
func (s *Store) SetJob(job *Job) error {
	jobKey := fmt.Sprintf("%s/jobs/%s", s.keyspace, job.Name)

	// Get if the requested job already exist
	ej, err := s.GetJob(job.Name)
	if err != nil && err != store.ErrKeyNotFound {
		return err
	}

	if ej != nil {
		// When the job runs, these status vars are updated
		// otherwise use the ones that are stored
		if ej.Metadata.LastError.After(job.Metadata.LastError) {
			job.Metadata.LastError = ej.Metadata.LastError
		}
		if ej.Metadata.LastSuccess.After(job.Metadata.LastSuccess) {
			job.Metadata.LastSuccess = ej.Metadata.LastSuccess
		}
		if ej.Metadata.SuccessCount > job.Metadata.SuccessCount {
			job.Metadata.SuccessCount = ej.Metadata.SuccessCount
		}
		if ej.Metadata.ErrorCount > job.Metadata.ErrorCount {
			job.Metadata.ErrorCount = ej.Metadata.ErrorCount
		}
	}

	jobJSON, _ := json.Marshal(job)

	log.WithFields(log.Fields{
		"job":  job.Name,
		"json": string(jobJSON),
	}).Debug("store: Setting job")

	if err := s.Client.Put(jobKey, jobJSON, nil); err != nil {
		return err
	}

	return nil
}

// GetJobs returns all jobs
func (s *Store) GetJobs() ([]*Job, error) {
	res, err := s.Client.List(s.keyspace+"/jobs/", nil)
	if err != nil {
		if err == store.ErrKeyNotFound {
			log.Debug("store: No jobs found")
			return []*Job{}, nil
		}
		return nil, err
	}

	jobs := make([]*Job, 0)
	for _, node := range res {
		var job Job
		err := json.Unmarshal([]byte(node.Value), &job)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

// Get a job
func (s *Store) GetJob(name string) (*Job, error) {
	res, err := s.Client.Get(s.keyspace+"/jobs/"+name, nil)
	if err != nil {
		return nil, err
	}

	var job Job
	if err = json.Unmarshal([]byte(res.Value), &job); err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"job": job.Name,
	}).Debug("store: Retrieved job from datastore")

	return &job, nil
}

func (s *Store) DeleteJob(name string) (*Job, error) {
	job, err := s.GetJob(name)
	if err != nil {
		return nil, err
	}

	if err := s.Client.Delete(s.keyspace + "/jobs/" + name); err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Store) WatchJobsTree() (<-chan []*store.KVPair, error) {
	stopCh := make(<-chan struct{})
	dir := s.keyspace + "/jobs"

	isEx, _ := s.Client.Exists(dir, nil)
	if !isEx {
		j := &Job{
			Name:        "watch",
			Schedule:    "@yearly",
			JobType:     "rpc",
			Disabled:    true,
			Concurrency: "forbid",
			Application: "system",
		}
		if err := s.SetJob(j); err != nil {
			log.Error(err)
		}

	}

	events, err := s.Client.WatchTree(dir, stopCh, nil)
	if err != nil {
		log.Error(err)
	}

	return events, err
}

// Store a processor
func (s *Store) SetProcessor(p *Processor) error {
	addr := fmt.Sprintf("%s:%d", p.IP, p.Port)
	key := fmt.Sprintf("%s/processors/%s/%s", s.keyspace, p.Application, addr)

	// Get if the requested processor already exist
	// ej, err := s.GetProcessor(p.Application, addr)
	// if err != nil && err != store.ErrKeyNotFound {
	// 	return err
	// }
	// if ej != nil {
	// 	err = errors.New(fmt.Sprintf("the requested addr '%s' has already exist", addr))
	// 	return err
	// }

	pJSON, _ := json.Marshal(p)

	log.WithFields(log.Fields{
		"addr": addr,
		"json": string(pJSON),
	}).Debug("store: Setting processor")

	if err := s.Client.Put(key, pJSON, nil); err != nil {
		return err
	}

	return nil
}

// Get a processor
func (s *Store) GetProcessor(app string, addr string) (*Processor, error) {
	key := fmt.Sprintf("%s/processors/%s/%s", s.keyspace, app, addr)
	res, err := s.Client.Get(key, nil)
	if err != nil {
		return nil, err
	}

	var p Processor
	if err = json.Unmarshal([]byte(res.Value), &p); err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"addr": addr,
	}).Debug("store: Retrieved processor from datastore")

	return &p, nil
}

func (s *Store) DeleteProcessor(app string, addr string) (*Processor, error) {
	key := fmt.Sprintf("%s/processors/%s/%s", s.keyspace, app, addr)
	p, err := s.GetProcessor(app, addr)
	if err != nil {
		return nil, err
	}

	if err := s.Client.Delete(key); err != nil {
		return nil, err
	}

	return p, nil
}

// GetProcessors returns all processor
func (s *Store) GetProcessors() ([]*Processor, error) {
	res, err := s.Client.List(s.keyspace+"/processors/", nil)
	if err != nil {
		if err == store.ErrKeyNotFound {
			log.Debug("store: No processors found")
			return []*Processor{}, nil
		}
		return nil, err
	}

	proces := make([]*Processor, 0)
	for _, node := range res {
		var p Processor
		err := json.Unmarshal([]byte(node.Value), &p)
		if err != nil {
			return nil, err
		}
		proces = append(proces, &p)
	}
	return proces, nil
}

// GetProcessors returns all processor of one app
func (s *Store) GetProcessorsByApp(app string) ([]*Processor, error) {
	res, err := s.Client.List(s.keyspace+"/processors/"+app, nil)
	if err != nil {
		if err == store.ErrKeyNotFound {
			log.Debug("store: No processors found in app: ", app)
			return []*Processor{}, nil
		}
		return nil, err
	}

	proces := make([]*Processor, 0)
	for _, node := range res {
		var p Processor
		err := json.Unmarshal([]byte(node.Value), &p)
		if err != nil {
			return nil, err
		}
		proces = append(proces, &p)
	}
	return proces, nil
}

func (s *Store) GetProcessorAddrs(app string) (map[string]string, error) {
	var addrKVPair map[string]string
	addrs, err := s.GetProcessorsByApp(app)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"app": app,
		}).Error("store.getWorkerRPCAddr called store.GetProcessorsByApp fail")
		return nil, err
	}
	for _, p := range addrs {
		addr := fmt.Sprintf("tcp@%s:%d", p.IP, p.Port)
		addrKVPair[p.NodeName] = addr
	}
	return addrKVPair, err

}

//You can use watches to watch modifications on a key. First you need to check if the key exists.
//If this is not the case, we need to create it using the Put function.
func (s *Store) WatchProcessorTree() (<-chan []*store.KVPair, error) {
	stopCh := make(<-chan struct{})
	dir := s.keyspace + "/processors"

	isEx, _ := s.Client.Exists(dir, nil)
	if !isEx {
		p := &Processor{
			Application: "system",
			NodeName:    "khronos01",
			IP:          "127.0.0.1",
			Port:        10005,
			Status:      true,
		}
		if err := s.SetProcessor(p); err != nil {
			log.Error(err)
		}

	}

	events, err := s.Client.WatchTree(dir, stopCh, nil)
	if err != nil {
		log.Error(err)
	}

	return events, err
}

func (s *Store) GetExecutionsAll() ([]*Execution, error) {
	prefix := fmt.Sprintf("%s/executions", s.keyspace)
	res, err := s.Client.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	var executions []*Execution

	for _, node := range res {
		var execution Execution
		err := json.Unmarshal([]byte(node.Value), &execution)
		if err != nil {
			return nil, err
		}
		executions = append(executions, &execution)
	}
	return executions, nil
}

func (s *Store) GetExecutions(jobName string) ([]*Execution, error) {
	prefix := fmt.Sprintf("%s/executions/%s", s.keyspace, jobName)
	res, err := s.Client.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	var executions []*Execution

	for _, node := range res {
		if store.Backend(s.backend) != store.ZK {
			path := store.SplitKey(node.Key)
			dir := path[len(path)-2]
			if dir != jobName {
				continue
			}
		}
		var execution Execution
		err := json.Unmarshal([]byte(node.Value), &execution)
		if err != nil {
			return nil, err
		}
		executions = append(executions, &execution)
	}
	return executions, nil
}

func (s *Store) GetLastExecutionGroup(jobName string) ([]*Execution, error) {
	res, err := s.Client.List(fmt.Sprintf("%s/executions/%s", s.keyspace, jobName), nil)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return []*Execution{}, nil
	}

	var lastEx Execution
	var ex Execution
	// res does not guarantee any order,
	// so compare them by `StartedAt` time and get the last one
	for _, node := range res {
		err := json.Unmarshal([]byte(node.Value), &ex)
		if err != nil {
			return nil, err
		}
		if ex.StartedAt.After(lastEx.StartedAt) {
			lastEx = ex
		}
	}
	return s.GetExecutionGroup(&lastEx)
}

func (s *Store) GetExecutionGroup(execution *Execution) ([]*Execution, error) {
	res, err := s.Client.List(fmt.Sprintf("%s/executions/%s", s.keyspace, execution.JobName), nil)
	if err != nil {
		return nil, err
	}

	var executions []*Execution
	for _, node := range res {
		var ex Execution
		err := json.Unmarshal([]byte(node.Value), &ex)
		if err != nil {
			return nil, err
		}

		if ex.Group == execution.Group {
			executions = append(executions, &ex)
		}
	}
	return executions, nil
}

// Returns executions for a job grouped and with an ordered index
// to facilitate access.
func (s *Store) GetGroupedExecutions(jobName string) (map[int64][]*Execution, []int64, error) {
	execs, err := s.GetExecutions(jobName)
	if err != nil {
		return nil, nil, err
	}
	groups := make(map[int64][]*Execution)
	for _, exec := range execs {
		groups[exec.Group] = append(groups[exec.Group], exec)
	}

	// Build a separate data structure to show in order
	var byGroup int64arr
	for key := range groups {
		byGroup = append(byGroup, key)
	}
	sort.Sort(sort.Reverse(byGroup))

	return groups, byGroup, nil
}

// Save a new execution and returns the key of the new saved item or an error.
func (s *Store) SetExecution(execution *Execution) (string, error) {
	exJson, _ := json.Marshal(execution)
	key := execution.Key()

	log.WithFields(log.Fields{
		"job":       execution.JobName,
		"execution": key,
	}).Debug("store: Setting key")

	err := s.Client.Put(fmt.Sprintf("%s/executions/%s/%s", s.keyspace, execution.JobName, key), exJson, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"job":       execution.JobName,
			"execution": key,
			"error":     err,
		}).Debug("store: Failed to set key")
		return "", err
	}

	execs, err := s.GetExecutions(execution.JobName)
	if err != nil {
		log.Errorf("store: No executions found for job %s", execution.JobName)
	}

	// Delete all execution results over the limit, starting from olders
	if len(execs) > MaxExecutions {
		//sort the array of all execution groups by StartedAt time
		sort.Sort(ExecList(execs))
		for i := 0; i < len(execs)-MaxExecutions+100; i++ {
			log.WithFields(log.Fields{
				"job":       execs[i].JobName,
				"execution": execs[i].Key(),
			}).Debug("store: to detele key")
			err := s.Client.Delete(fmt.Sprintf("%s/executions/%s/%s", s.keyspace, execs[i].JobName, execs[i].Key()))
			if err != nil {
				log.Errorf("store: Trying to delete overflowed execution %s", execs[i].Key())
			}
		}
	}

	return key, nil
}

// Removes all executions of a job
func (s *Store) DeleteExecutions(jobName string) error {
	return s.Client.DeleteTree(fmt.Sprintf("%s/executions/%s", s.keyspace, jobName))
}

func (s *Store) DeleteExecutionsByNodeName(nodeName string) error {

	exs, err := s.GetExecutionsAll()
	if err == nil {
		for _, ex := range exs {
			key := ex.Key()

			log.WithFields(log.Fields{
				"nodeName":  nodeName,
				"key":       key,
				"execution": ex,
			}).Debug("store.DeleteExecutionsByNodeName: to detele executions of which node has downed. ")

			if ex.NodeName == nodeName {
				if ex.Success == false {

					err := s.Client.Delete(fmt.Sprintf("%s/executions/%s/%s", s.keyspace, ex.JobName, key))
					if err != nil {
						log.WithFields(log.Fields{
							"nodeName":  nodeName,
							"key":       key,
							"execution": ex,
						}).Error("store.DeleteExecutionsByNodeName: to detele executions of which node has downed. ")
					}
				}
			}
		}
	}

	return nil
}
