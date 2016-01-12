package kbdscoring

import "os"
import "log"
import "unicode"
import "unicode/utf8"
import "strconv"
import "bufio"
import "../kbdlayout"

type BigramScoringFunc struct {
	bigrams     [][]uint64 // bigrams[mapping.Rune2ID['e']][mapping.Rune2ID['s']] = 5234
	qwertyScore uint64     // will be used for normalizing
}

// will be mirrored to right side
var bigramBaseWeightsLeft = [15]uint64{
	25, 45, 60, 50, 25,
	50, 60, 80, 80, 50,
	20, 35, 45, 55, 20,
}

const oo = 0

// use oo when want to use base layer
// will be mirrored for right side
var bigramKeyWeightsLeft = [15][30]uint64{
	// top row
	[30]uint64{
		//
		20, 50, 65, 55, 30, oo, oo, oo, oo, oo,
		20, 60, 85, 90, 70, oo, oo, oo, oo, oo,
		10, 15, 35, 55, 20, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//  xx
		30, 25, 70, 55, 30, oo, oo, oo, oo, oo,
		35, 30, 85, 90, 70, oo, oo, oo, oo, oo,
		10, 10, 20, 55, 20, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//      xx
		oo, 55, 30, 55, 30, oo, oo, oo, oo, oo,
		oo, oo, 65, 90, 70, oo, oo, oo, oo, oo,
		oo, 25, 20, 50, oo, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//          xx
		oo, oo, oo, 30, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 65, 15, oo, oo, oo, oo, oo,
		oo, oo, 40, 45, 15, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//              xx
		oo, oo, oo, 35, 30, oo, oo, oo, oo, oo,
		oo, oo, oo, 65, 15, oo, oo, oo, oo, oo,
		oo, oo, 35, 35, 15, oo, oo, oo, oo, oo,
	},
	// middle row
	[30]uint64{
		//
		15, 55, 65, 55, 30, oo, oo, oo, oo, oo,
		35, 65, 85, 90, 65, oo, oo, oo, oo, oo,
		15, 35, 50, 65, 25, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//  xx
		20, 25, 65, 55, 30, oo, oo, oo, oo, oo,
		oo, 45, 85, 90, 65, oo, oo, oo, oo, oo,
		15, 15, 50, 65, 25, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//      xx
		oo, oo, 35, 55, 30, oo, oo, oo, oo, oo,
		oo, oo, 55, 90, 65, oo, oo, oo, oo, oo,
		oo, oo, 25, 65, 25, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//          xx
		oo, oo, oo, 30, 15, oo, oo, oo, oo, oo,
		oo, 65, 85, 55, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 40, 15, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//              xx
		oo, oo, oo, 30, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 65, 30, oo, oo, oo, oo, oo,
		oo, oo, oo, 40, 15, oo, oo, oo, oo, oo,
	},
	// bottom row
	[30]uint64{
		//
		15, 40, 60, 45, 20, oo, oo, oo, oo, oo,
		20, 60, 85, 85, 70, oo, oo, oo, oo, oo,
		20, 35, 50, 60, 25, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//  xx
		20, 20, 60, 45, 20, oo, oo, oo, oo, oo,
		30, 30, 85, 85, 70, oo, oo, oo, oo, oo,
		25, 25, 50, 60, 25, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//      xx
		oo, 40, 35, 45, 20, oo, oo, oo, oo, oo,
		oo, oo, 65, 85, 70, oo, oo, oo, oo, oo,
		oo, 40, 30, 60, 25, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//          xx
		oo, oo, oo, 45, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 65, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 30, 15, oo, oo, oo, oo, oo,
	},
	[30]uint64{
		//              xx
		oo, oo, oo, 40, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 65, 15, oo, oo, oo, oo, oo,
		oo, oo, oo, 50, 30, oo, oo, oo, oo, oo,
	},
}

var bigramKeyWeights = [30][30]uint64{}

func prepareWeights() {
	// first fill out the left part with base
	// all values of oo will be replaced from the base
	for i := 0; i < 15; i++ {
		for j := 0; j < 30; j++ {
			if bigramKeyWeightsLeft[i][j] == oo {
				col := j % 10
				row := j / 10
				if col < 5 {
					bigramKeyWeightsLeft[i][j] = bigramBaseWeightsLeft[col+row*5]
				} else {
					bigramKeyWeightsLeft[i][j] = bigramBaseWeightsLeft[9-col+row*5]
				}
			}
		}
	}

	// then fill the bigramKeyWeights with left and mirrored left
	for row := 0; row < 3; row++ {
		for pos := 0; pos < 5; pos++ {
			bigramKeyWeights[row*10+pos] = bigramKeyWeightsLeft[row*5+pos]
			for i := 0; i < 3; i++ {
				for j := 0; j < 10; j++ {
					bigramKeyWeights[row*10+9-pos][i*10+9-j] = bigramKeyWeights[row*10+pos][i*10+j]
				}
			}
		}
	}
}

func (s *BigramScoringFunc) Init(mapping *kbdlayout.KeyboardMapping) {
	file, err := os.Open("bigrams.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)

	s.bigrams = make([][]uint64, len(mapping.ID2Rune))
	for i := 0; i < len(mapping.ID2Rune); i++ {
		s.bigrams[i] = make([]uint64, len(mapping.ID2Rune))
	}

	for scanner.Scan() {
		line := scanner.Text()

		// read unicode letter
		letter1, size := utf8.DecodeRuneInString(line)
		// remove it from the line
		line = line[size:]

		letter2, size := utf8.DecodeRuneInString(line)
		// remove it and space from line
		line = line[size+1:]

		// make letters lowercase
		letter1 = unicode.ToLower(letter1)
		letter2 = unicode.ToLower(letter2)

		characterId1, ok := mapping.Rune2ID[letter1]
		if !ok {
			// there is no need for this letter, as there is no mapping for it
			continue
		}
		characterId2, ok := mapping.Rune2ID[letter2]
		if !ok {
			// there is no need for this letter, as there is no mapping for it
			continue
		}

		// parse count from the line
		count, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		s.bigrams[characterId1][characterId2] = count
	}

	prepareWeights()

	qwerty := kbdlayout.NewLayout(kbdlayout.Qwerty, mapping)
	s.qwertyScore = s.CalculateScore(&qwerty)
}

func (s *BigramScoringFunc) CalculateScore(layout *kbdlayout.KeyboardLayout) uint64 {

	var score uint64

	for i := 0; i < 30; i++ {
		charId1 := layout[i]
		for j := 0; j < 30; j++ {
			charId2 := layout[j]

			bigramFrequency := s.bigrams[charId1][charId2]
			bigramWeight := bigramKeyWeights[i][j]

			score += bigramFrequency * bigramWeight
		}
	}

	return score
}
func (s *BigramScoringFunc) NormalizeScore(score uint64) float64 {
	return float64(score) / float64(s.qwertyScore)
}
