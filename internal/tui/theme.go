package tui

import (
	"fmt"
	"math"
	"math/rand"
)

type Mood int

const (
	MoodVivid Mood = iota
	MoodPastel
	MoodNeon
	MoodMono
	MoodRetro
	MoodOcean
	MoodSunset
	moodCount
)

func (m Mood) String() string {
	switch m {
	case MoodVivid:
		return "VIVID"
	case MoodPastel:
		return "PASTEL"
	case MoodNeon:
		return "NEON"
	case MoodMono:
		return "MONO"
	case MoodRetro:
		return "RETRO"
	case MoodOcean:
		return "OCEAN"
	case MoodSunset:
		return "SUNSET"
	}
	return "VIVID"
}

type Palette struct {
	Primary   string
	Secondary string
	Accent    string
}

func RandomMood() Mood {
	return Mood(rand.Intn(int(moodCount)))
}

func (m Mood) GeneratePalette() Palette {
	switch m {
	case MoodVivid:
		return Palette{
			Primary:   hsl(rand.Float64()*360, 0.7+rand.Float64()*0.3, 0.5+rand.Float64()*0.2),
			Secondary: hsl(rand.Float64()*360, 0.6+rand.Float64()*0.3, 0.6+rand.Float64()*0.2),
			Accent:    hsl(rand.Float64()*360, 0.8+rand.Float64()*0.2, 0.4+rand.Float64()*0.2),
		}
	case MoodPastel:
		return Palette{
			Primary:   hsl(rand.Float64()*360, 0.3+rand.Float64()*0.2, 0.75+rand.Float64()*0.15),
			Secondary: hsl(rand.Float64()*360, 0.25+rand.Float64()*0.2, 0.8+rand.Float64()*0.1),
			Accent:    hsl(rand.Float64()*360, 0.35+rand.Float64()*0.2, 0.7+rand.Float64()*0.15),
		}
	case MoodNeon:
		return Palette{
			Primary:   hsl(rand.Float64()*360, 0.9+rand.Float64()*0.1, 0.5+rand.Float64()*0.15),
			Secondary: hsl(rand.Float64()*360, 0.85+rand.Float64()*0.15, 0.55+rand.Float64()*0.15),
			Accent:    hsl(rand.Float64()*360, 1.0, 0.45+rand.Float64()*0.15),
		}
	case MoodMono:
		h := rand.Float64() * 360
		return Palette{
			Primary:   hsl(h, 0.4+rand.Float64()*0.3, 0.5+rand.Float64()*0.2),
			Secondary: hsl(h, 0.3+rand.Float64()*0.2, 0.65+rand.Float64()*0.15),
			Accent:    hsl(h, 0.5+rand.Float64()*0.3, 0.4+rand.Float64()*0.15),
		}
	case MoodRetro:
		return Palette{
			Primary:   hsl(rand.Float64()*40+10, 0.6+rand.Float64()*0.2, 0.45+rand.Float64()*0.15),
			Secondary: hsl(rand.Float64()*30+20, 0.5+rand.Float64()*0.2, 0.55+rand.Float64()*0.15),
			Accent:    hsl(rand.Float64()*20+30, 0.7+rand.Float64()*0.2, 0.4+rand.Float64()*0.1),
		}
	case MoodOcean:
		return Palette{
			Primary:   hsl(180+rand.Float64()*60, 0.6+rand.Float64()*0.2, 0.45+rand.Float64()*0.15),
			Secondary: hsl(190+rand.Float64()*50, 0.5+rand.Float64()*0.2, 0.55+rand.Float64()*0.15),
			Accent:    hsl(170+rand.Float64()*40, 0.7+rand.Float64()*0.2, 0.4+rand.Float64()*0.15),
		}
	case MoodSunset:
		p := rand.Float64() * 60
		return Palette{
			Primary:   hsl(p, 0.7+rand.Float64()*0.2, 0.5+rand.Float64()*0.15),
			Secondary: hsl(p+30+rand.Float64()*30, 0.6+rand.Float64()*0.2, 0.55+rand.Float64()*0.15),
			Accent:    hsl(p+60+rand.Float64()*30, 0.8+rand.Float64()*0.15, 0.4+rand.Float64()*0.15),
		}
	}
	return Palette{Primary: "#FFFFFF", Secondary: "#AAAAAA", Accent: "#555555"}
}

func hsl(h, s, l float64) string {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s = math.Max(0, math.Min(1, s))
	l = math.Max(0, math.Min(1, l))

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	toHex := func(v float64) string {
		val := int(math.Round((v + m) * 255))
		if val > 255 {
			val = 255
		}
		if val < 0 {
			val = 0
		}
		return fmt.Sprintf("%02X", val)
	}

	return "#" + toHex(r) + toHex(g) + toHex(b)
}
