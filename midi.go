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
	Start    float64 // in seconds
	End      float64 // in seconds
	Channel  uint8
}

type TempoChange struct {
	Tick int64
	BPM  float64
}

func ParseMidi(filename string) ([]Note, error) {
	fmt.Printf("\n Loading MIDI file: %s\n", filename)

	res, err := smf.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read MIDI file: %w", err)
	}

	// Get resolution (ticks per quarter note)
	// smf.MetricTicks is a uint16
	resolution, ok := res.TimeFormat.(smf.MetricTicks)
	if !ok {
		return nil, fmt.Errorf("unsupported time format: %v", res.TimeFormat)
	}
	ticksPerQuarter := float64(resolution)

	fmt.Printf("   ├─ Tracks: %d\n", len(res.Tracks))
	fmt.Printf("   ├─ Resolution: %d ticks/quarter\n", resolution)

	// 1. Collect all tempo changes
	var tempoChanges []TempoChange
	// Default tempo is 120 BPM
	tempoChanges = append(tempoChanges, TempoChange{Tick: 0, BPM: 120.0})

	for _, track := range res.Tracks {
		var absTick int64 = 0
		for _, event := range track {
			absTick += int64(event.Delta)
			msg := event.Message
			var bpm float64
			if msg.GetMetaTempo(&bpm) {
				tempoChanges = append(tempoChanges, TempoChange{Tick: absTick, BPM: bpm})
			}
		}
	}

	// Sort tempo changes by tick
	sort.Slice(tempoChanges, func(i, j int) bool {
		return tempoChanges[i].Tick < tempoChanges[j].Tick
	})

	if len(tempoChanges) > 1 {
		fmt.Printf("   ├─ Tempo changes: %d (%.1f - %.1f BPM)\n", len(tempoChanges), tempoChanges[0].BPM, tempoChanges[len(tempoChanges)-1].BPM)
	} else {
		fmt.Printf("   ├─ Tempo: %.1f BPM\n", tempoChanges[0].BPM)
	}

	// Helper to convert ticks to seconds
	tickToSeconds := func(tick int64) float64 {
		var seconds float64 = 0
		var lastTick int64 = 0
		var currentBPM float64 = 120.0

		for _, tc := range tempoChanges {
			if tick < tc.Tick {
				break
			}
			// Add time for the duration between lastTick and tc.Tick
			delta := tc.Tick - lastTick
			seconds += (float64(delta) / ticksPerQuarter) * (60.0 / currentBPM)

			lastTick = tc.Tick
			currentBPM = tc.BPM
		}

		// Add remaining time
		if tick > lastTick {
			delta := tick - lastTick
			seconds += (float64(delta) / ticksPerQuarter) * (60.0 / currentBPM)
		}
		return seconds
	}

	var notes []Note

	for _, track := range res.Tracks {
		var absTick int64 = 0
		trackActiveNotes := make(map[uint8]map[uint8]*Note)

		for _, event := range track {
			absTick += int64(event.Delta)
			msg := event.Message

			var channel, key, velocity uint8
			switch {
			case msg.GetNoteOn(&channel, &key, &velocity):
				// Channel 9 is percussion (0-indexed), which produces non-pitched sounds.
				// We skip it to avoid "zapping" noises from playing drum keys as frequencies.
				if channel == 9 {
					continue
				}
				if velocity > 0 {
					if _, ok := trackActiveNotes[channel]; !ok {
						trackActiveNotes[channel] = make(map[uint8]*Note)
					}

					if existing, ok := trackActiveNotes[channel][key]; ok {
						existing.End = tickToSeconds(absTick)
						notes = append(notes, *existing)
					}

					newNote := &Note{
						Key:      key,
						Velocity: velocity,
						Start:    tickToSeconds(absTick),
						Channel:  channel,
					}
					trackActiveNotes[channel][key] = newNote
				} else {
					if chMap, ok := trackActiveNotes[channel]; ok {
						if note, ok := chMap[key]; ok {
							note.End = tickToSeconds(absTick)
							notes = append(notes, *note)
							delete(chMap, key)
						}
					}
				}
			case msg.GetNoteOff(&channel, &key, &velocity):
				if chMap, ok := trackActiveNotes[channel]; ok {
					if note, ok := chMap[key]; ok {
						note.End = tickToSeconds(absTick)
						notes = append(notes, *note)
						delete(chMap, key)
					}
				}
			}
		}
	}

	if len(notes) > 0 {
		duration := notes[len(notes)-1].End
		for _, n := range notes {
			if n.End > duration {
				duration = n.End
			}
		}
		fmt.Printf("   └─ Notes: %d (duration: %.1fs)\n", len(notes), duration)
	} else {
		fmt.Printf("   └─ Notes: 0\n")
	}
	return notes, nil
}

func GraphMidi(filename string) {
	fmt.Println("\n Graphing MIDI to Desmos...")
	graph("T=0")
	graph("x=T")

	notes, err := ParseMidi(filename)
	if err != nil {
		panic(err)
	}

	if len(notes) == 0 {
		fmt.Println(" No notes found in MIDI file")
		return
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Start < notes[j].Start
	})

	numChunks := (len(notes) + chunk - 1) / chunk
	fmt.Printf("   ├─ Chunk size: %d notes\n", chunk)
	fmt.Printf("   ├─ Total chunks: %d\n", numChunks)

	var fVars, sVars, eVars, vVars, vis, data, tones []string
	for i := 0; i < len(notes); i += chunk {
		end := min(i+chunk, len(notes))

		var sb strings.Builder
		sb.WriteString("[")
		for j, n := range notes[i:end] {
			if j > 0 {
				sb.WriteString(",")
			}
			fmt.Fprintf(&sb, "(%f, %d), (%f, %d), (0/0, 0/0)", n.Start, n.Key, n.End, n.Key)
		}
		sb.WriteString("]")
		vis = append(vis, sb.String())

		// Audio Lists
		var freqSB, startSB, endSB, velSB strings.Builder
		freqSB.WriteString("[")
		startSB.WriteString("[")
		endSB.WriteString("[")
		velSB.WriteString("[")

		for j, n := range notes[i:end] {
			if j > 0 {
				freqSB.WriteString(",")
				startSB.WriteString(",")
				endSB.WriteString(",")
				velSB.WriteString(",")
			}
			fmt.Fprintf(&freqSB, "%.2f", MidiToHz(int(n.Key)))
			fmt.Fprintf(&startSB, "%.2f", n.Start)
			fmt.Fprintf(&endSB, "%.2f", n.End)

			// Adjust volume based on pitch to balance loudness
			// Lower notes get a boost, higher notes get attenuated
			vol := float64(n.Velocity) / 127.0
			scale := 1.0 - (float64(n.Key)-60.0)*0.01

			// Reduce global volume significantly to prevent clipping with many notes
			finalVol := vol * scale * 1
			if finalVol > 1.0 {
				finalVol = 1.0
			} else if finalVol < 0.0 {
				finalVol = 0.0
			}
			fmt.Fprintf(&velSB, "%.2f", finalVol)
		}
		freqSB.WriteString("]")
		startSB.WriteString("]")
		endSB.WriteString("]")
		velSB.WriteString("]")

		// Create unique variable names for this chunk
		id := fmt.Sprintf("%d", i/chunk)
		fVar := "F_{" + id + "}"
		sVar := "S_{" + id + "}"
		eVar := "E_{" + id + "}"
		vVar := "V_{" + id + "}"

		data = append(data, fVar+"="+freqSB.String())
		data = append(data, sVar+"="+startSB.String())
		data = append(data, eVar+"="+endSB.String())
		data = append(data, vVar+"="+velSB.String())

		fVars = append(fVars, fVar)
		sVars = append(sVars, sVar)
		eVars = append(eVars, eVar)
		vVars = append(vVars, vVar)
	}

	for i := range fVars {
		id := fmt.Sprintf("%d", i)
		dVar := fmt.Sprintf("D_{%s}", id)
		fVar := fVars[i]
		sVar := sVars[i]
		eVar := eVars[i]
		vVar := vVars[i]

		data = append(data, fmt.Sprintf("%s = (T - %s) * (%s - T)", dVar, sVar, eVar))
		tones = append(tones, fmt.Sprintf("\\operatorname{tone}(%s\\{%s >= 0\\}, %s\\{%s >= 0\\})", fVar, dVar, vVar, dVar))
	}

	fmt.Printf("   ├─ Adding %d tone expressions...\n", len(tones))
	for _, v := range tones {
		graph(v)
	}

	fmt.Printf("   ├─ Adding %d visualization expressions...\n", len(vis))
	for _, v := range vis {
		graph(v)
	}

	fmt.Printf("   ├─ Adding %d data expressions...\n", len(data))
	for _, v := range data {
		graph(v)
	}

	fmt.Println("   └─ Done!")
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

func GraphMidiNoVis(filename string) {
	fmt.Println("\n🎹 Graphing MIDI to Desmos (no visualization)...")
	graph("T=0")
	graph("x=T")

	notes, err := ParseMidi(filename)
	if err != nil {
		panic(err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes found in MIDI file")
		return
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Start < notes[j].Start
	})
	chunkSize := 500
	numChunks := (len(notes) + chunkSize - 1) / chunkSize
	fmt.Printf("   ├─ Chunk size: %d notes\n", chunkSize)
	fmt.Printf("   ├─ Total chunks: %d\n", numChunks)

	var fVars, sVars, eVars, vVars []string

	for i := 0; i < len(notes); i += chunkSize {
		end := min(i+chunkSize, len(notes))

		// Audio Lists
		var freqSB, startSB, endSB, velSB strings.Builder
		freqSB.WriteString("[")
		startSB.WriteString("[")
		endSB.WriteString("[")
		velSB.WriteString("[")

		for j, n := range notes[i:end] {
			if j > 0 {
				freqSB.WriteString(",")
				startSB.WriteString(",")
				endSB.WriteString(",")
				velSB.WriteString(",")
			}
			fmt.Fprintf(&freqSB, "%.2f", MidiToHz(int(n.Key)))
			fmt.Fprintf(&startSB, "%.2f", n.Start)
			fmt.Fprintf(&endSB, "%.2f", n.End)

			// Adjust volume based on pitch to balance loudness
			// Lower notes get a boost, higher notes get attenuated
			vol := float64(n.Velocity) / 127.0
			scale := 1.0 - (float64(n.Key)-60.0)*0.01

			// Reduce global volume significantly to prevent clipping with many notes
			finalVol := vol * scale * 1
			if finalVol > 1.0 {
				finalVol = 1.0
			} else if finalVol < 0.0 {
				finalVol = 0.0
			}
			fmt.Fprintf(&velSB, "%.2f", finalVol)
		}
		freqSB.WriteString("]")
		startSB.WriteString("]")
		endSB.WriteString("]")
		velSB.WriteString("]")

		// Create unique variable names for this chunk
		id := fmt.Sprintf("%d", i/chunkSize)
		fVar := "F_{" + id + "}"
		sVar := "S_{" + id + "}"
		eVar := "E_{" + id + "}"
		vVar := "V_{" + id + "}"

		graph(fVar + "=" + freqSB.String())
		graph(sVar + "=" + startSB.String())
		graph(eVar + "=" + endSB.String())
		graph(vVar + "=" + velSB.String())

		fVars = append(fVars, fVar)
		sVars = append(sVars, sVar)
		eVars = append(eVars, eVar)
		vVars = append(vVars, vVar)
	}

	var tones []string
	for i := range fVars {
		// We use the index as the ID, matching how we created the vars
		id := fmt.Sprintf("%d", i)
		dVar := fmt.Sprintf("D_{%s}", id)
		fVar := fVars[i]
		sVar := sVars[i]
		eVar := eVars[i]
		vVar := vVars[i]

		graph(fmt.Sprintf("%s = (T - %s) * (%s - T)", dVar, sVar, eVar))
		tones = append(tones, fmt.Sprintf("tone(%s[%s >= 0], %s[%s >= 0])", fVar, dVar, vVar, dVar))
	}
	fmt.Printf("   ├─ Adding %d tone expressions...\n", len(tones))
	for _, v := range tones {
		graph(v)
	}

	fmt.Println("   └─ Done!")
}
