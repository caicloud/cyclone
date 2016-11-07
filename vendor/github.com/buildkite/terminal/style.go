package terminal

import "strings"

var emptyStyle = style{}

type style struct {
	fgColor     string
	bgColor     string
	otherColors string
	classes     string
}

// True if both styles are equal (or are the same object)
func (s *style) isEqual(o *style) bool {
	return s == o || (s.fgColor == o.fgColor && s.bgColor == o.bgColor && s.otherColors == o.otherColors)
}

// CSS classes that make up the style
func (s *style) asClasses() string {
	if s.classes != "" || s.isEmpty() {
		return s.classes
	}

	var styles []string
	if s.fgColor != "" {
		styles = append(styles, s.fgColor)
	}
	if s.bgColor != "" {
		styles = append(styles, s.bgColor)
	}
	if s.otherColors != "" {
		styles = append(styles, s.otherColors)
	}
	s.classes = strings.TrimSpace(strings.Join(styles, " "))
	return s.classes
}

// True if style is empty
func (s *style) isEmpty() bool {
	return s.fgColor == "" && s.bgColor == "" && s.otherColors == ""
}

// Remove a particular 'other' colour from our style's colour list
func (s *style) removeOther(r string) {
	s.otherColors = strings.Replace(s.otherColors, r+" ", "", -1)
}

func (s *style) addOther(r string) {
	r = r + " "
	if strings.Index(s.otherColors, r) == -1 {
		s.otherColors = s.otherColors + r
	}
}

// Add colours to an existing style, potentially returning
// a new style.
func (s *style) color(colors []string) *style {
	if len(colors) == 1 && (colors[0] == "0" || colors[0] == "") {
		// Shortcut for full style reset
		return &emptyStyle
	}

	s = &style{fgColor: s.fgColor, bgColor: s.bgColor, otherColors: s.otherColors}

	if len(colors) >= 2 {
		if colors[0] == "38" && colors[1] == "5" {
			// Extended set foreground x-term color
			s.fgColor = "term-fgx" + colors[2]
			return s
		}

		// Extended set background x-term color
		if colors[0] == "48" && colors[1] == "5" {
			s.bgColor = "term-bgx" + colors[2]
			return s
		}
	}

	for _, cc := range colors {
		// If multiple colors are defined, i.e. \e[30;42m\e then loop through each
		// one, and assign it to s.fgColor or s.bgColor
		switch cc {
		case "0":
			// Reset all styles - don't use &emptyStyle here as we could end up adding colours
			// in this same action.
			s = &style{}
		case "21", "22":
			s.removeOther("term-fg1")
			s.removeOther("term-fg2")
			// Turn off italic
		case "23":
			s.removeOther("term-fg3")
			// Turn off underline
		case "24":
			s.removeOther("term-fg4")
			// Turn off crossed-out
		case "29":
			s.removeOther("term-fg9")
		case "39":
			s.fgColor = ""
		case "49":
			s.bgColor = ""
			// 30–37, then it's a foreground color
		case "30", "31", "32", "33", "34", "35", "36", "37":
			s.fgColor = "term-fg" + cc
			// 40–47, then it's a background color.
		case "40", "41", "42", "43", "44", "45", "46", "47":
			s.bgColor = "term-bg" + cc
			// 90-97 is like the regular fg color, but high intensity
		case "90", "91", "92", "93", "94", "95", "96", "97":
			s.fgColor = "term-fgi" + cc
			// 100-107 is like the regular bg color, but high intensity
		case "100", "101", "102", "103", "104", "105", "106", "107":
			s.fgColor = "term-bgi" + cc
			// 1-9 random other styles
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			s.addOther("term-fg" + cc)
		}
	}
	return s
}
