# DMMClust #

package `dmmclust` is a package that provides functions for clustering small texts as described by [Yin and Wang (2014)](dbgroup.cs.tsinghua.edu.cn/wangjy/papers/KDD14-GSDMM.pdf) in *A Dirichlet Multinomial Mixture Model based Approach for Short Text Clustering*.

The clustering algorithm is remarkably elegant and simple, leading to a very minimal implementation. This package also exposes some types to allow for extensibility.

# Installing # 

`go get -u github.com/go-nlp/dmmclust`.

This package also provides a Gopkg.toml file for `dep` users.

This package uses SemVer 2.0 for versioning, and all releases are tagged.

# How To Use #

```
func main(){
	docs := getDocs()
	corp := getCorpus(docs)
	conf := dmmclust.Config{
		K:          10,                   // maximum 10 clusters expected
		Vocabulary: len(corp),            // simple example: the vocab is the same as the corpus size
		Iter:       100,                  // iterate 100 times
		Alpha:      0.0001,               // smaller probability of joining an empty group
		Beta:       0.1,                  // higher probability of joining groups like me
		Score:      dmmclust.Algorithm3,  // use Algorithm3 to score
		Sample:     dmmclust.Gibbs, // use Gibbs to sample
	}

	var clustered []dmmclust.Cluster // len(clustered) == len(docs)
	var err error
	if clustered, err = dmmclust.FindClusters(docs, conf); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Clusters:")
	for i, clust := range clustered {
		fmt.Printf("\t%d: %q\n", clust.ID(), data[i])
	}
}
```

# Plays Well #

This package plays well with other NLP packages (i.e. lingo). Much of the dependency on `lingo` has been removed for better interfacing. 

# Extensibility #

This package defines a Scoring Function as `type ScoringFn func(doc Document, docs []Document, clusters []Cluster, conf Config) []float64`. This allows for custom scoring functions to be used.

There are two scoring algorithms provided: `Algorithm3` and `Algorithm4`. I've been successful at using other scoring algorithms as well.

The sampling function is also customizable. The default is to use `Gibbs`. I've not had much success at other sampling algorithms.

# Licence #

This package is MIT licenced.