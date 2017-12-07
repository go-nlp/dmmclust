package dmmclust

import (
	"github.com/pkg/errors"
	randomkit "gorgonia.org/randomkit"
)

// ScoringFn is any function that can take a document and return the probabilities of it existing in those clusters
type ScoringFn func(doc Document, docs []Document, clusters []Cluster, conf Config) []float64

// SamplingFn is a sampling function that take a slice of floats and returns an index
type SamplingFn func([]float64) int

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

	Score  ScoringFn
	Sample SamplingFn
}

// valid checks that the config for errors
func (c *Config) valid() error {
	if c.Score == nil {
		return errors.New("Expected Score to not be nil")
	}
	if c.Sample == nil {
		return errors.New("Expected Sample to not be nil")
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
	TokenSet() TokenSet
	Len() int
}

// distro represents the sparse distribution of words. The key is the corpus ID, while the value is the frequency
type distro map[int]float64

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
		z := conf.Sample(probs)

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
			z2 := conf.Sample(p)

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
	retVal := make([]float64, len(clusters))
	for i := 0; i < len(clusters); i++ {
		clust := clusters[i]
		p := float64(clust.Docs()) + conf.Alpha/(docCount-1.0+k*conf.Alpha)

		prod := 1.0
		for j, tok := range doc.TokenSet() {
			prod *= (clust.Freq(tok) + conf.Beta) / (float64(clust.Wordcount()) + float64(conf.Vocabulary)*conf.Beta + float64(j) - 1)
		}
		score := p * prod
		retVal[i] = score
	}

	norm := sum(retVal)
	if norm <= 0 {
		norm = 1
	}
	for i := range retVal {
		retVal[i] = retVal[i] / norm
	}
	return retVal
}

func Algorithm4(doc TokenSet, clusters []Cluster, conf Config) []float64 { return nil }

/* Sampling Functions */

func Gibbs(p []float64) int {
	ret := randomkit.Multinomial(1, p, len(p))
	for i, v := range ret {
		if v != 0 {
			return i
		}
	}
	panic("Unreachable") // practically this part is unreachable
}

/* UTILITIES */
func sum(a []float64) (retVal float64) {
	for _, v := range a {
		retVal += v
	}
	return
}
