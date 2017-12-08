package dmmclust_test

import (
	"fmt"
	"math/rand"
	"strings"

	. "github.com/go-nlp/dmmclust"
	"github.com/xtgo/set"
)

// data is some sample tweets from @chewxy.
// They have been "preprocessed" to aid the simple tokenizer.
// one hashtag has been changed from #gopherdata to #gopher
var data = []string{
	// coffee related tweet
	"A Java prefix that I don't hate .",
	"Colleagues must have thought I was crazy on Friday, but @MeccaCoffee ' s latest Xade Burqa is nothing short of orgasm inducing .",

	// JavaScript hate
	"Let me take this time while I wait for your JavaScript to download to tell you to stop using so much JavaScript on your web page .",

	// On Error Resume Next
	"all future programming languages I implement will have On Error Resume Next . Even if it's a functional , expressions-only language . Because I can. ",
	"On Error Resume Next",
	"When I was younger , I used VB. My crutch was On Error Resume Next . I find it weird reimplementing it for a probabilistic parser .",

	// Gophers/Golang
	"Questions for #gopher and #golang people out there : how do you debug a slow compile ? ",
	"In case you missed it , 10000 words on generics in #golang :",
	"Data Science in Go https://speakerdeck.com/chewxy/data-science-in-go … Slides by @chewxy #gopher #golang",
	"Big heap , many pointers . GC killing me . Help ? Tips? #golang . Most pointers unavoidable .",
}

func makeCorpus(a []string) map[string]int {
	retVal := make(map[string]int)
	var id int
	for _, s := range a {
		for _, f := range strings.Fields(s) {
			if _, ok := retVal[f]; !ok {
				retVal[f] = id
				id++
			}
		}
	}
	return retVal
}

func makeDocuments(a []string, c map[string]int, allowRepeat bool) []Document {
	retVal := make([]Document, 0, len(a))
	for _, s := range a {
		var ts []int
		for _, f := range strings.Fields(s) {
			id := c[f]
			ts = append(ts, id)
		}
		if !allowRepeat {
			ts = set.Ints(ts) // this uniquifies the sentence
		}
		retVal = append(retVal, TokenSet(ts))
	}
	return retVal
}

func Example() {
	corp := makeCorpus(data)
	docs := makeDocuments(data, corp, false)
	r := rand.New(rand.NewSource(1337))
	conf := Config{
		K:          10,          // maximum 10 clusters expected
		Vocabulary: len(corp),   // simple example: the vocab is the same as the corpus size
		Iter:       1000,        // iterate 100 times
		Alpha:      0.0001,      // smaller probability of joining an empty group
		Beta:       0.1,         // higher probability of joining groups like me
		Score:      Algorithm3,  // use Algorithm3 to score
		Sampler:    NewGibbs(r), // use Gibbs to sample
	}
	var clustered []Cluster
	var err error
	if clustered, err = FindClusters(docs, conf); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Clusters (Algorithm3):")
	for i, clust := range clustered {
		fmt.Printf("\t%d: %q\n", clust.ID(), data[i])
	}

	// Using Algorithm4, where repeat words are allowed
	docs = makeDocuments(data, corp, true)
	conf.Score = Algorithm4
	if clustered, err = FindClusters(docs, conf); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\nClusters (Algorithm4):")
	for i, clust := range clustered {
		fmt.Printf("\t%d: %q\n", clust.ID(), data[i])
	}

	// Output:
	// Clusters (Algorithm3):
	//	0: "A Java prefix that I don't hate ."
	//	0: "Colleagues must have thought I was crazy on Friday, but @MeccaCoffee ' s latest Xade Burqa is nothing short of orgasm inducing ."
	//	1: "Let me take this time while I wait for your JavaScript to download to tell you to stop using so much JavaScript on your web page ."
	//	2: "all future programming languages I implement will have On Error Resume Next . Even if it's a functional , expressions-only language . Because I can. "
	//	2: "On Error Resume Next"
	//	2: "When I was younger , I used VB. My crutch was On Error Resume Next . I find it weird reimplementing it for a probabilistic parser ."
	//	3: "Questions for #gopher and #golang people out there : how do you debug a slow compile ? "
	//	3: "In case you missed it , 10000 words on generics in #golang :"
	//	3: "Data Science in Go https://speakerdeck.com/chewxy/data-science-in-go … Slides by @chewxy #gopher #golang"
	//	1: "Big heap , many pointers . GC killing me . Help ? Tips? #golang . Most pointers unavoidable ."
	//
	// Clusters (Algorithm4):
	//	0: "A Java prefix that I don't hate ."
	//	0: "Colleagues must have thought I was crazy on Friday, but @MeccaCoffee ' s latest Xade Burqa is nothing short of orgasm inducing ."
	//	1: "Let me take this time while I wait for your JavaScript to download to tell you to stop using so much JavaScript on your web page ."
	//	2: "all future programming languages I implement will have On Error Resume Next . Even if it's a functional , expressions-only language . Because I can. "
	//	2: "On Error Resume Next"
	//	2: "When I was younger , I used VB. My crutch was On Error Resume Next . I find it weird reimplementing it for a probabilistic parser ."
	//	3: "Questions for #gopher and #golang people out there : how do you debug a slow compile ? "
	//	3: "In case you missed it , 10000 words on generics in #golang :"
	//	3: "Data Science in Go https://speakerdeck.com/chewxy/data-science-in-go … Slides by @chewxy #gopher #golang"
	//	2: "Big heap , many pointers . GC killing me . Help ? Tips? #golang . Most pointers unavoidable ."
}
