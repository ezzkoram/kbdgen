package gen

import "sort"
import "math/rand"

import "../kbdlayout"
import "../kbdscoring"

type LayoutEntry struct {
	Layout kbdlayout.KeyboardLayout
	Score  uint64
}

func EvolvePopulation(sf kbdscoring.ScoringFunction, generationBest chan<- *LayoutEntry, done <-chan struct{}) {

	// random population of 1000 layouts
	population := createRandomPopulation(1000)

	var currentBest uint64

	const numberOfParents = 35
	numberToRandomize := 100

	for {
		select {
		case <-done:
			// main thread signaled that we need to stop.
			return
		default:
			// no signal yet, do another generation

			// 1) sort the population so the best scores are on top
			sortPopulation(population, sf)

			// and send the best of the population to the main thread
			generationBest <- &LayoutEntry{
				Score:  population[0].Score,
				Layout: population[0].Layout,
			}

			// HACK: this will increase the randomness of the population
			// when the population doesn't evolve
			// TODO: figure out a better way to get past the local maximums
			if population[0].Score > currentBest {
				currentBest = population[0].Score
				numberToRandomize = 100
			} else {
				numberToRandomize += 10
				if numberToRandomize > 990 {
					currentBest = 0
					numberToRandomize = 100
				}
			}

			// 2) mix top n to have new generation
			num := numberOfParents
			for num < len(population)-numberToRandomize {
				for i := 0; i < numberOfParents && num < len(population)-numberToRandomize; i++ {
					for j := 0; j < numberOfParents && num < len(population)-numberToRandomize; j++ {
						if i == j {
							continue
						}
						mix(&population[num].Layout, &population[i].Layout, &population[j].Layout)
						num++
					}
				}
			}

			// 3) mutate all
			for i := 0; i < len(population)-numberToRandomize; i++ {
				mutate(&population[i].Layout)
			}

			// 4) randomize the rest
			for i := len(population) - numberToRandomize; i < len(population); i++ {
				randomizeLayout(&population[i].Layout)
			}

		}
	}
}

func createRandomPopulation(size uint64) []LayoutEntry {
	population := make([]LayoutEntry, size)
	for i := uint64(0); i < size; i++ {
		population[i] = LayoutEntry{}
		randomizeLayout(&population[i].Layout)
	}
	return population
}

func randomizeLayout(layout *kbdlayout.KeyboardLayout) {

	freeCharIds := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
	for i := 0; i < 30; i++ {
		idx := rand.Intn(len(freeCharIds))
		charId := freeCharIds[idx]
		freeCharIds[idx] = freeCharIds[len(freeCharIds)-1]
		freeCharIds = freeCharIds[:len(freeCharIds)-1]

		layout[i] = charId
	}
}

type ByScore []LayoutEntry

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score > a[j].Score }

func sortPopulation(population []LayoutEntry, sf kbdscoring.ScoringFunction) {
	for i := 0; i < len(population); i++ {
		population[i].Score = sf.CalculateScore(&population[i].Layout)
	}
	sort.Sort(ByScore(population))
}

// Mix two parents to get a child
func mix(child, parent1, parent2 *kbdlayout.KeyboardLayout) {
	// TODO: there should be easier way to mix two layouts
	freePositions := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
	charactersUsed := []bool{false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false}

	// mix ratio tells how much to take from parent1 vs parent2
	// 20% to 80%
	mixRatio := 0.2 + rand.Float64()*0.6

	// setup array for the parents for easier indexing
	parents := [2]*kbdlayout.KeyboardLayout{parent1, parent2}

	// flag to indicate if there is no usable parts left on the parent
	blocked := [2]bool{false, false}

	for i := 0; i < 30; i++ {

		// default to parent2
		p := 1
		if rand.Float64() < mixRatio {
			// use parent1
			p = 0
		}

		if !blocked[p] {

			// try to find a free location on our new layout
			// in the same location on parent we have a character, if this
			// character is still available (not yet put anywhere on the new layout)
			// then we have our pick.
			found := false
			for j := 0; j < len(freePositions); j++ {
				freePos := freePositions[j]
				charId := parents[p][freePos]
				if !charactersUsed[charId] {
					// found a character that we can take from parent
					child[freePos] = charId
					charactersUsed[charId] = true
					freePositions[j] = freePositions[len(freePositions)-1]
					freePositions = freePositions[:len(freePositions)-1]
					found = true
					break
				}
			}
			if found {
				// found a suitable free location and character
				continue
			}

			blocked[p] = true
		}

		// could not find suitable free location with free character on p
		// in this case we just pick a random free location with a free
		freePosId := rand.Intn(len(freePositions))
		freePos := freePositions[freePosId]
		freePositions[freePosId] = freePositions[len(freePositions)-1]
		freePositions = freePositions[:len(freePositions)-1]
		charId := uint8(0)
		for j := uint8(0); j < 30; j++ {
			if !charactersUsed[j] {
				charactersUsed[j] = true
				charId = j
				break
			}
		}
		child[freePos] = charId
	}
}

// Mutate the layout a bit with random
func mutate(layout *kbdlayout.KeyboardLayout) {
	// TODO: some better magic numbers needed here
	numMutations := rand.Intn(7) * rand.Intn(7)
	for i := 0; i < numMutations; i++ {
		p1 := rand.Intn(30)
		p2 := rand.Intn(30)
		r := layout[p1]
		layout[p1] = layout[p2]
		layout[p2] = r
	}
}
