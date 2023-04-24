package sample

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/BSidesSF/ctf-2023/sefi/encoders"
	"github.com/BSidesSF/ctf-2023/sefi/rng"
	"github.com/BSidesSF/ctf-2023/sefi/types"
)

type Sample struct {
	Freqs     int
	TimeSlots int
	Samples   []byte
	StartTime int
}

func NewSample(freqs, timeSlots int) *Sample {
	return &Sample{
		Freqs:     freqs,
		TimeSlots: timeSlots,
		Samples:   make([]byte, freqs*timeSlots),
	}
}

func (s *Sample) FillWithNoise() {
	for f := 0; f < s.Freqs; f++ {
		for t := 0; t < s.TimeSlots; t++ {
			idx := t*s.Freqs + f
			s.Samples[idx] = byte(rng.ScaledNormalDistribution(0, 31))
		}
	}
}

func (s *Sample) EncodeBytes(freq int, data []byte, ticksPerBit int) error {
	e := encoders.NewEncoder8B10B()
	rawData := e.EncodeBytes(data)
	log.Printf("encoded bytes: %q", rawData)
	ticks := len(rawData) * 8 * ticksPerBit
	if ticks > s.TimeSlots {
		return fmt.Errorf("Need %d time slots, have %d", ticks, s.TimeSlots)
	}
	for i, v := range rawData {
		// msb first
		for b := 0; b < 8; b++ {
			if v&(1<<(7-b)) == 0 {
				continue
			}
			for t := 0; t < ticksPerBit; t++ {
				tickIdx := t + b*ticksPerBit + i*ticksPerBit*8
				sampleIdx := tickIdx*s.Freqs + freq
				if sampleIdx >= len(s.Samples) {
					return fmt.Errorf("We've failed: %d > %d", sampleIdx, len(s.Samples))
				}
				s.Samples[sampleIdx] += 224
				sides := []byte{224, 224, 208, 200, 192, 160, 128, 96, 64, 64, 32, 32, 32, 16}
				for i, v := range sides {
					if i < freq {
						sideIdx := sampleIdx - i - 1
						s.Samples[sideIdx] += v
					}
					if (freq + i) < s.Freqs {
						sideIdx := sampleIdx + i + 1
						s.Samples[sideIdx] += v
					}
				}
			}
		}
	}
	return nil
}

// Split into new samples, each timePer long (final one may be truncated)
func (s *Sample) Split(timePer int) []*Sample {
	var rv []*Sample
	for i := 0; i < s.TimeSlots; i += timePer {
		startPos := i * s.Freqs
		endPos := (i + timePer) * s.Freqs
		thisTime := timePer
		if endPos > len(s.Samples) {
			endPos = len(s.Samples)
			thisTime = (endPos - startPos) / s.Freqs
		}
		s := &Sample{
			Freqs:     s.Freqs,
			TimeSlots: thisTime,
			Samples:   s.Samples[startPos:endPos],
			StartTime: i,
		}
		rv = append(rv, s)
	}
	return rv
}

func (s *Sample) Encode(w io.Writer) error {
	outType := types.Sample{
		Freqs:     s.Freqs,
		TimeSlots: s.TimeSlots,
		Samples:   s.Samples[:],
		StartTime: s.StartTime,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(&outType)
}

func (s *Sample) EncodeToString() (string, error) {
	buf := &strings.Builder{}
	if err := s.Encode(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *Sample) Prepare() *types.Sample {
	return &types.Sample{
		Freqs:     s.Freqs,
		TimeSlots: s.TimeSlots,
		Samples:   s.Samples[:],
		StartTime: s.StartTime,
	}
}
