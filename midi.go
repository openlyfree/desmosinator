package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"gitlab.com/gomidi/midi/v2/smf"
)

type Note struct {
	Key      uint8
	Velocity uint8
	Start    float64
	End      float64
	Channel  uint8
}

type TempoChange struct {
	Tick int64
	BPM  float64
}

func ParseMidi(filename string) ([]Note, error) {
	data, err := smf.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var notes []Note

	// Process each track independently
	for _, track := range data.Tracks {
		// Track active notes for this track: map[channel][key] -> list of (startTick, velocity)
		// Using a slice to handle multiple notes of same key overlapping
		type noteOn struct {
			tick     int64
			velocity uint8
		}
		activeNotes := make(map[uint8]map[uint8][]noteOn)

		var absTick int64
		for _, ev := range track {
			absTick += int64(ev.Delta)

			var channel, key, velocity uint8
			if ev.Message.GetNoteOn(&channel, &key, &velocity) {
				if velocity > 0 {
					// Note On - add to the list
					if activeNotes[channel] == nil {
						activeNotes[channel] = make(map[uint8][]noteOn)
					}
					activeNotes[channel][key] = append(activeNotes[channel][key], noteOn{tick: absTick, velocity: velocity})
				} else {
					// Note On with velocity 0 = Note Off
					if activeNotes[channel] != nil && len(activeNotes[channel][key]) > 0 {
						// Pop the first note (FIFO)
						on := activeNotes[channel][key][0]
						activeNotes[channel][key] = activeNotes[channel][key][1:]

						startTime := float64(data.TimeAt(on.tick)) / 1000000.0
						endTime := float64(data.TimeAt(absTick)) / 1000000.0

						// Skip zero-duration notes
						if endTime > startTime {
							notes = append(notes, Note{
								Key:      key,
								Velocity: on.velocity,
								Start:    startTime,
								End:      endTime,
								Channel:  channel,
							})
						}
					}
				}
			} else if ev.Message.GetNoteOff(&channel, &key, &velocity) {
				// Note Off
				if activeNotes[channel] != nil && len(activeNotes[channel][key]) > 0 {
					// Pop the first note (FIFO)
					on := activeNotes[channel][key][0]
					activeNotes[channel][key] = activeNotes[channel][key][1:]

					startTime := float64(data.TimeAt(on.tick)) / 1000000.0
					endTime := float64(data.TimeAt(absTick)) / 1000000.0

					// Skip zero-duration notes
					if endTime > startTime {
						notes = append(notes, Note{
							Key:      key,
							Velocity: on.velocity,
							Start:    startTime,
							End:      endTime,
							Channel:  channel,
						})
					}
				}
			}
		}
	}

	// Sort notes by start time
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Start < notes[j].Start
	})

	fmt.Printf("Parsed %d notes from MIDI file\n", len(notes))
	return notes, nil
}

func GraphMidi(filename string) {
	notes, err := ParseMidi(filename)
	if err != nil {
		fmt.Println("Error parsing MIDI:", err)
		return
	}

	// Filter out drum channel (channel 9, which is 10 in 1-indexed MIDI)
	var filteredNotes []Note
	for _, note := range notes {
		if note.Channel != 9 {
			filteredNotes = append(filteredNotes, note)
		}
	}
	notes = filteredNotes
	fmt.Printf("After filtering drums: %d notes\n", len(notes))

	// Find the total duration
	var maxEnd float64
	for _, note := range notes {
		if note.End > maxEnd {
			maxEnd = note.End
		}
	}

	// Create time slider
	graph("t=0")
	page.MustEval(`(id, min, max) => {
		Calc.setExpression({ id: id, latex: "t=0", sliderBounds: { min: min, max: max } });
	}`, fmt.Sprint(id), "0", fmt.Sprintf("%.2f", maxEnd))

	// Create individual tone expressions for each note
	// Volume is based on velocity (0-127 -> 0-1)
	// Round times to 3 decimal places to avoid floating point issues
	for _, note := range notes {
		freq := MidiToHz(int(note.Key))
		volume := float64(note.Velocity) / 127.0
		start := math.Round(note.Start*1000) / 1000
		end := math.Round(note.End*1000) / 1000
		// Tone is active only when t is within the note's time range
		toneExpr := fmt.Sprintf("\\operatorname{tone}(%.2f, %.3f\\{%.3f<t<%.3f\\})", freq, volume, start, end)
		graph(toneExpr)
	}

	// Group notes into chunks for Desmos visualization
	var chunks [][]Note
	for i := 0; i < len(notes); i += chunk {
		end := i + chunk
		if end > len(notes) {
			end = len(notes)
		}
		chunks = append(chunks, notes[i:end])
	}

	for _, chunk := range chunks {
		var points []string
		for _, note := range chunk {
			freq := MidiToHz(int(note.Key))
			// Create point: (start_time, frequency)
			points = append(points, fmt.Sprintf("(%.3f, %.2f)", note.Start, freq))
		}
		latex := strings.Join(points, ",")
		graph(latex)
	}

	fmt.Printf("Graphed %d notes in %d chunks (playable with slider t)\n", len(notes), len(chunks))
	fmt.Println("Animate the t slider to play the music!")
}

func GraphMidiNoVis(filename string) {
	notes, err := ParseMidi(filename)
	if err != nil {
		fmt.Println("Error parsing MIDI:", err)
		return
	}

	// Filter out drum channel (channel 9)
	var filteredNotes []Note
	for _, note := range notes {
		if note.Channel != 9 {
			filteredNotes = append(filteredNotes, note)
		}
	}
	notes = filteredNotes
	fmt.Printf("After filtering drums: %d notes\n", len(notes))

	// Find the total duration
	var maxEnd float64
	for _, note := range notes {
		if note.End > maxEnd {
			maxEnd = note.End
		}
	}

	// Create time slider
	graph("t=0")
	page.MustEval(`(id, min, max) => {
		Calc.setExpression({ id: id, latex: "t=0", sliderBounds: { min: min, max: max } });
	}`, fmt.Sprint(id), "0", fmt.Sprintf("%.2f", maxEnd))

	// Create individual tone expressions for each note
	// Volume is based on velocity (0-127 -> 0-1)
	for _, note := range notes {
		freq := MidiToHz(int(note.Key))
		volume := float64(note.Velocity) / 127.0
		start := math.Round(note.Start*1000) / 1000
		end := math.Round(note.End*1000) / 1000
		// Tone is active only when t is within the note's time range
		toneExpr := fmt.Sprintf("\\operatorname{tone}(%.2f, %.3f\\{%.3f<t<%.3f\\})", freq, volume, start, end)
		graph(toneExpr)
	}

	fmt.Printf("Graphed %d notes (no visualization, playable with slider t)\n", len(notes))
	fmt.Println("Animate the t slider to play the music!")
}

func MidiToHz(m int) float64 {
	const (
		a4Midi = 69
		a4Freq = 440.0
	)

	frequency := a4Freq * math.Pow(2.0, (float64(m-a4Midi)/12.0))

	return frequency
}

func graph(latex string) {

	page.MustEval(`(id, latex) => {
		Calc.setExpression({ id: id, latex: latex , pointSize: 2});
	}`, func() string {
		id++
		return fmt.Sprint(id)
	}(), latex)
}
