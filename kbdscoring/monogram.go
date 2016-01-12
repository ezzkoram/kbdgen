package kbdscoring

import "os"
import "log"
import "unicode"
import "unicode/utf8"
import "strconv"
import "bufio"
import "../kbdlayout"

type MonogramScoringFunc struct {
	monograms   []uint64 // monograms[mapping.Rune2ID['e']] = 5234
	qwertyScore uint64   // will be used for normalizing
}

// Weights for each key location in layout
// The preferred locations have a higher weight
var monogramKeyWeights = [30]uint64{
	2, 4, 7, 3, 2, 2, 3, 7, 4, 2,
	7, 8, 9, 8, 5, 5, 8, 9, 8, 7,
	2, 3, 3, 7, 1, 1, 7, 3, 3, 2,
}

// Loads monograms from file and stores character counts with indices
func (s *MonogramScoringFunc) Init(mapping *kbdlayout.KeyboardMapping) {
	file, err := os.Open("monograms.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)

	s.monograms = make([]uint64, len(mapping.ID2Rune))
	for scanner.Scan() {
		line := scanner.Text()

		// read unicode letter
		letter, size := utf8.DecodeRuneInString(line)
		// remove it and space from line
		line = line[size+1:]

		// make letter lowercase
		letter = unicode.ToLower(letter)

		characterId, ok := mapping.Rune2ID[letter]
		if !ok {
			// there is no need for this letter, as there is no mapping for it
			continue
		}

		// parse count from the line
		count, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			// invalid format for the monograms file
			log.Fatal(err)
		}
		s.monograms[characterId] = count
	}

	// calculate the score for qwerty layout, so we can use it as a base.
	qwerty := kbdlayout.NewLayout(kbdlayout.Qwerty, mapping)
	s.qwertyScore = s.CalculateScore(&qwerty)
}

// Loop through the layout
// for each location, multiply the location weight and the character usage
func (s *MonogramScoringFunc) CalculateScore(layout *kbdlayout.KeyboardLayout) uint64 {

	var score uint64

	for i := 0; i < 30; i++ {
		charId := layout[i]
		charFrequency := s.monograms[charId]

		// weight for this position
		weight := monogramKeyWeights[i]

		score += charFrequency * weight
	}

	return score
}

// Normalize score so that qwerty is 1.0
func (s *MonogramScoringFunc) NormalizeScore(score uint64) float64 {
	return float64(score) / float64(s.qwertyScore)
}
