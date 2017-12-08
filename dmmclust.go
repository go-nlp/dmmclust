package dmmclust

import (
	"math/rand"
	"sync"

	"github.com/pkg/errors"
	randomkit "gorgonia.org/randomkit"
)

// ScoringFn is any function that can take a document and return the probabilities of it existing in those clusters
type ScoringFn func(doc Document, docs []Document, clusters []Cluster, conf Config) []float64

// Sampler is anything that can generate a index based on the given probability
type Sampler interface {
	Sample([]float64) int
}

// Config is a struct that configures the running of the algorithm.
type Config struct {
	// Maximum number of clusters expected
	K int

	// Vocabulary is the size of the vocabulary
	Vocabulary int

	// Number of iterations
	Iter int

	// Probability that controls whether a student will join an empty table
	Alpha float64

	// Betweenness affinity
	Beta float64

	// Score is a scoring function that will be used
	Score ScoringFn

	// Sampler is the sampler function
	Sampler Sampler
}

// valid checks that the config for errors
func (c *Config) valid() error {
	if c.Score == nil {
		return errors.New("Expected Score to not be nil")
	}
	if c.Sampler == nil {
		return errors.New("Expected Sampler to not be nil")
	}
	return nil
}

// TokenSet is a vector of word IDs for a document.
// Depending on algorithm, it may be uniquified using a set function that does not preserve order
type TokenSet []int

func (ts TokenSet) TokenSet() TokenSet { return ts }
func (ts TokenSet) Len() int           { return len(ts) }

// Document is anything that can return a TokenSet
type Document interface {
	// TokenSet returns a list of word IDs. It's preferable to return an ordered list, rather than a uniquified set.
	TokenSet() TokenSet

	// Len returns the number of words in the document
	Len() int
}

// Cluster is a representation of a cluster. It doesn't actually store any of the data, only the metadata of the cluster
type Cluster struct {
	id          int
	docs, words int
	dist        distro
}

func (c *Cluster) addDoc(doc Document) {
	c.docs++
	c.words += doc.Len()

	if c.dist == nil {
		c.dist = make(distro)
	}

	for _, tok := range doc.TokenSet() {
		c.dist[tok]++
	}
}

func (c *Cluster) removeDoc(doc Document) {
	c.docs--
	c.words -= doc.Len()
	for _, tok := range doc.TokenSet() {
		c.dist[tok]--
	}
}

// ID returns the ID of the cluster. This is not set or used until the results are returned
func (c *Cluster) ID() int { return c.id }

// Docs returns the number of documents in the cluster
func (c *Cluster) Docs() int { return c.docs }

// Freq returns the frequency of the ith word
func (c *Cluster) Freq(i int) float64 { return c.dist[i] }

// Wordcount returns the number of words in the cluster
func (c *Cluster) Wordcount() int { return c.words }

// Words returns the word IDs in the cluster
func (c *Cluster) Words() []int {
	retVal := make([]int, 0, len(c.dist))
	for k := range c.dist {
		retVal = append(retVal, k)
	}
	return retVal
}

// FindClusters is the main function to find clusters.
func FindClusters(docs []Document, conf Config) ([]Cluster, error) {
	if err := conf.valid(); err != nil {
		return nil, err
	}
	state := make([]Cluster, conf.K)
	probs := make([]float64, conf.K)
	for i := range probs {
		probs[i] = 1.0 / float64(conf.K)
	}

	// initialize the clusters
	dz := make([]int, len(docs))
	for i, doc := range docs {
		// randomly chuck docs into a cluster
		z := conf.Sampler.Sample(probs)

		dz[i] = z
		clust := &state[z]
		clust.addDoc(doc)
	}

	clusterCount := conf.K
	for i := 0; i < conf.Iter; i++ {
		var transfers int
		for j, doc := range docs {
			// remove from old cluster
			old := dz[j]
			clust := &state[old]
			clust.removeDoc(doc)

			// draw sample from distro to find new cluster
			p := conf.Score(doc, docs, state, conf)
			z2 := conf.Sampler.Sample(p)

			// transfer doc to new clusetr
			if z2 != old {
				transfers++
			}

			dz[j] = z2
			newClust := &state[z2]
			newClust.addDoc(doc)
		}

		// TODO: count new clusters
		var clusterCount2 int
		for i := range state {
			if state[i].docs > 0 {
				clusterCount2++
			}
		}

		if transfers == 0 && clusterCount2 == clusterCount && i > 25 {
			break // convergence achieved. Time to GTFO
		}
		clusterCount = clusterCount2
	}
	// return the clusters. As an additional niceness, we'll relabel the cluster IDs
	retVal := make([]Cluster, len(dz))
	reindex := make([]int, conf.K)
	for i := range reindex {
		reindex[i] = -1
	}
	var maxID int
	for i, clusterID := range dz {
		retVal[i] = state[clusterID]
		var cid int
		if cid = reindex[clusterID]; cid < 0 {
			cid = maxID
			maxID++
			reindex[clusterID] = cid
		}

		retVal[i].id = cid
	}
	return retVal, nil
}

/* Scoring Functions */

// Algorithm3 is the implementation of Equation 3 in the original paper.
// This assumes that a word can only occur once in a string. If the requirement is that words can appear multiple times,
// use Algorithm4.
func Algorithm3(doc Document, docs []Document, clusters []Cluster, conf Config) []float64 {
	docCount := float64(len(docs))
	k := float64(conf.K)
	vocab := float64(conf.Vocabulary)
	retVal := make([]float64, len(clusters))
	var wg sync.WaitGroup
	ts := doc.TokenSet()
	for i := range clusters {
		clust := clusters[i]
		wg.Add(1)
		go func(clust Cluster, i int, wg *sync.WaitGroup) {
			p := float64(clust.Docs()) + conf.Alpha/(docCount-1.0+k*conf.Alpha)
			num := algo3Numerator(clust, ts, conf.Beta)
			denom := algoDenominator(clust, ts, conf.Beta, vocab)
			retVal[i] = p * num / denom
			wg.Done()
		}(clust, i, &wg)
	}
	wg.Wait()

	norm := sum(retVal)
	if norm <= 0 {
		norm = 1
	}
	for i := range retVal {
		retVal[i] = retVal[i] / norm
	}
	return retVal
}

// Algorithm4 is the implementation of Equation 4 in the original paper.
// It allows for multiple words to be used in a document.
func Algorithm4(doc Document, docs []Document, clusters []Cluster, conf Config) []float64 {
	docCount := float64(len(docs))
	k := float64(conf.K)
	vocab := float64(conf.Vocabulary)
	retVal := make([]float64, len(clusters))
	var wg sync.WaitGroup
	ts := doc.TokenSet()
	for i := range clusters {
		clust := clusters[i]
		wg.Add(1)
		go func(clust Cluster, i int, wg *sync.WaitGroup) {

			p := float64(clust.Docs()) + conf.Alpha/(docCount-1.0+k*conf.Alpha)
			num := algo4Numerator(clust, ts, conf.Beta)
			denom := algoDenominator(clust, ts, conf.Beta, vocab)
			retVal[i] = p * num / denom
			wg.Done()
		}(clust, i, &wg)
	}
	wg.Wait()

	norm := sum(retVal)
	if norm <= 0 {
		norm = 1
	}
	for i := range retVal {
		retVal[i] = retVal[i] / norm
	}
	ddd++
	return retVal
}

func algoDenominator(clust Cluster, ts TokenSet, beta float64, vocab float64) float64 {
	retVal := 1.0
	wc := float64(clust.Wordcount())
	for i := 0; i < len(ts); i++ {
		retVal *= wc + vocab*beta + float64(i)
	}
	return retVal
}

func algo3Numerator(clust Cluster, ts TokenSet, beta float64) float64 {
	retVal := 1.0
	for _, tok := range ts {
		retVal *= (clust.Freq(tok) + beta)
	}
	return retVal
}

func algo4Numerator(clust Cluster, ts TokenSet, beta float64) float64 {
	d := make(kvs, 0, len(ts))
	for _, tok := range ts {
		d = d.add(tok)
		d.incr(tok)
	}

	retVal := 1.0
	for _, tok := range ts {
		prod := 1.0
		freq := d.val(tok)
		clustFreq := clust.Freq(tok)
		for j := 0.0; j < freq; j++ {
			prod *= clustFreq + beta + j
		}
		retVal *= prod
	}
	return retVal
}

/* Sampling Functions */

// Gibbs is the standard sampling function, as per the paper.
type Gibbs struct {
	randomkit.BinomialGenerator
}

func NewGibbs(rand *rand.Rand) *Gibbs {
	return &Gibbs{
		BinomialGenerator: randomkit.BinomialGenerator{
			Rand: rand,
		},
	}
}

// Gibbs returns the index sampled
func (s *Gibbs) Sample(p []float64) int {
	ret := s.BinomialGenerator.Multinomial(1, p, len(p))
	for i, v := range ret {
		if v != 0 {
			return i
		}
	}
	panic("Unreachable") // practically this part is unreachable
}
