package main

import "fmt"
import "os"
import "unicode/utf8"
import "os/signal"
import "log"
import "flag"
import "time"
import "math/rand"
import "runtime"

import "./kbdscoring"
import "./kbdlayout"
import "./gen"

var scoringFuncs = map[string]kbdscoring.ScoringFunction{
	"monogram": &kbdscoring.MonogramScoringFunc{},
	"bigram":   &kbdscoring.BigramScoringFunc{},
}

var defaultMapping = kbdlayout.NewMapping("abcdefghijklmnopqrstuvwxyz.,;/")

var layouts = map[string]kbdlayout.KeyboardLayout{
	"qwerty":  kbdlayout.NewLayout(kbdlayout.Qwerty, defaultMapping),
	"abcde":   kbdlayout.NewLayout(kbdlayout.Abcde, defaultMapping),
	"dvorak":  kbdlayout.NewLayout(kbdlayout.Dvorak, defaultMapping),
	"colemak": kbdlayout.NewLayout(kbdlayout.Colemak, defaultMapping),
	"asset":   kbdlayout.NewLayout(kbdlayout.Asset, defaultMapping),
	"workman": kbdlayout.NewLayout(kbdlayout.Workman, defaultMapping),
	"nail":    kbdlayout.NewLayout(kbdlayout.Nail, defaultMapping),
	"layman":  kbdlayout.NewLayout(kbdlayout.Layman, defaultMapping),
}

func main() {

	var genCharactersParam = flag.String("characters", "abcdefghijklmnopqrstuvwxyz.,/;", "30 characters to use in the generator")
	var scoringFuncParam = flag.String("scoring-func", "monogram", "which function to use")
	var layoutParam = flag.String("layout", "", "all/qwerty/dvorak/colemak/asset/workman/nail/layman or custom (define with 30 characters)")

	flag.Parse()

	sf, ok := scoringFuncs[*scoringFuncParam]
	if !ok {
		fmt.Printf("could not find scoring func '%s'", *scoringFuncParam)
		return
	}
	if *layoutParam != "" {
		// scoring can happen on default mapping
		// TODO: will not work on all cases
		sf.Init(defaultMapping)

		if *layoutParam == "all" {
			// calculate scores for all layouts
			scoreAll(sf)
			return
		}

		// only calculate score for given layout
		layout, ok := layouts[*layoutParam]
		if !ok {
			if utf8.RuneCountInString(*layoutParam) == 30 {
				layout = kbdlayout.NewLayout(*layoutParam, defaultMapping)
			} else {
				log.Fatalf("could not find layout '%s' %d\n", *layoutParam, len(*layoutParam))
			}
		}

		scoreOne(sf, layout)
		return
	}

	// no layout defined

	if utf8.RuneCountInString(*genCharactersParam) != 30 {
		log.Fatalf("the generator needs exactly 30 characters, got %d", utf8.RuneCountInString(*genCharactersParam))
	}

	mapping := kbdlayout.NewMapping(*genCharactersParam)
	sf.Init(mapping)

	// start generating layouts
	generateLayouts(sf, mapping)
}

func scoreAll(sf kbdscoring.ScoringFunction) {
	for name, layout := range layouts {
		fmt.Printf("%16.12f - %s\n", sf.NormalizeScore(sf.CalculateScore(&layout)), name)
	}
}

func scoreOne(sf kbdscoring.ScoringFunction, layout kbdlayout.KeyboardLayout) {
	score := sf.NormalizeScore(sf.CalculateScore(&layout))
	fmt.Println("----")
	defaultMapping.PrintLayout(&layout)
	fmt.Println("----")
	fmt.Printf("%16.12f\n", score)
}

func generateLayouts(sf kbdscoring.ScoringFunction, mapping *kbdlayout.KeyboardMapping) {

	// we'll start an unending process, so lets hook up to a interrupt and kill signals
	//
	// buffer size of at least 1 is necessary, so we don't miss the signal in case we're not
	// listening it while it fires
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	// this is simple signaling channel
	done := make(chan struct{})
	// when this method exits, we're closing the done-channel to let everyone know
	// that we're exiting
	defer close(done)

	// goroutines will send best of each generation via this channel
	// make the buffer size 10 so no goroutine will need to pend writing
	generationBest := make(chan *gen.LayoutEntry, 10)

	// setup the maximum number of threads
	runtime.GOMAXPROCS(8)
	// setup random
	rand.Seed(time.Now().UnixNano())

	// create 10 goroutines to evolve
	for i := 0; i < 10; i++ {
		go gen.EvolvePopulation(sf, generationBest, done)
	}

	// keep count of generations
	var generation uint64

	// keep the best
	var bestOfTheBest *gen.LayoutEntry = nil

	// loop forever
	for {
		select {
		case sig := <-quit:
			// the quit channel signaled
			fmt.Printf("got signal: %s\n", sig.String())
			// just return from the function
			// this will trigger the close for the done channel
			return
		case next := <-generationBest:
			// some goroutine got one generation evolved
			generation++
			if generation%100 == 0 {
				fmt.Printf(".")
			}
			if bestOfTheBest == nil || bestOfTheBest.Score < next.Score {
				// got a new best
				bestOfTheBest = next
				fmt.Printf("\nnew best: %16.12f at generation %d\n", sf.NormalizeScore(next.Score), generation)
				mapping.PrintLayout(&next.Layout)
			}
		}
	}
}
