# DMMClust [![GoDoc](https://godoc.org/github.com/go-nlp/dmmclust?status.svg)](https://godoc.org/github.com/go-nlp/dmmclust) [![Build Status](https://travis-ci.org/go-nlp/dmmclust.svg?branch=master)](https://travis-ci.org/go-nlp/dmmclust) [![Coverage Status](https://coveralls.io/repos/github/go-nlp/dmmclust/badge.svg?branch=master)](https://coveralls.io/github/go-nlp/dmmclust?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/go-nlp/dmmclust)](https://goreportcard.com/report/github.com/go-nlp/dmmclust) #

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

## Hyperparameters ##

* `K` represents the maximum number of clusters expected. The final number of clusters can never exceed `K`.
* `Alpha` represents the probability of joining an empty group. If `Alpha` is `0.0` then once a group is empty, it'll stay empty for the rest of the 
* `Beta` represents the probability of joining groups that are similar. If `Beta` is `0.0`, then a document will never join a group if there are no common words between the groups and the documents. In some cases this is preferable (highly preprocessed inputs for example).

# Playing Well With Other Packages #

This package was originally built to play well with [lingo](https://github.com/chewxy/lingo). It's why it works on slices of integers. That's the only preprocessing necessary - converting a sentence into a slice of ints.

The `Document` interface is defined as:

```
type Document interface {
	TokenSet() TokenSet
	Len() int
}
```

`TokenSet` is simply a `[]int`, where each ith element represents the word ID of a corpus. The order is not important in the provided algorithms (Algorithm3 and Algorithm4), but may be important in some other scoring function.

# Extensibility #

This package defines a Scoring Function as `type ScoringFn func(doc Document, docs []Document, clusters []Cluster, conf Config) []float64`. This allows for custom scoring functions to be used.

There are two scoring algorithms provided: `Algorithm3` and `Algorithm4`. I've been successful at using other scoring algorithms as well.

The sampling function is also customizable. The default is to use `Gibbs`. I've not had much success at other sampling algorithms.

# Contributing #

To contribute to this package, simply file an issue, discuss and then send a pull request. Please ensure that tests are provided in any changes.

# Licence #

This package is MIT licenced.