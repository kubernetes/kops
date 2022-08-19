package termenv

import (
	"text/template"
)

// TemplateFuncs contains a few useful template helpers.
func TemplateFuncs(p Profile) template.FuncMap {
	return template.FuncMap{
		"Color": func(values ...interface{}) string {
			s := String(values[len(values)-1].(string))
			switch len(values) {
			case 2:
				s = s.Foreground(p.Color(values[0].(string)))
			case 3:
				s = s.
					Foreground(p.Color(values[0].(string))).
					Background(p.Color(values[1].(string)))
			}

			return s.String()
		},
		"Foreground": func(values ...interface{}) string {
			s := String(values[len(values)-1].(string))
			if len(values) == 2 {
				s = s.Foreground(p.Color(values[0].(string)))
			}

			return s.String()
		},
		"Background": func(values ...interface{}) string {
			s := String(values[len(values)-1].(string))
			if len(values) == 2 {
				s = s.Background(p.Color(values[0].(string)))
			}

			return s.String()
		},
		"Bold":      styleFunc(Style.Bold),
		"Faint":     styleFunc(Style.Faint),
		"Italic":    styleFunc(Style.Italic),
		"Underline": styleFunc(Style.Underline),
		"Overline":  styleFunc(Style.Overline),
		"Blink":     styleFunc(Style.Blink),
		"Reverse":   styleFunc(Style.Reverse),
		"CrossOut":  styleFunc(Style.CrossOut),
	}
}

func styleFunc(f func(Style) Style) func(...interface{}) string {
	return func(values ...interface{}) string {
		s := String(values[0].(string))
		return f(s).String()
	}
}
