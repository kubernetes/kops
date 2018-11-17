package uritemplates

func Expand(path string, values Values) (string, error) {
	template, err := Parse(path)
	if err != nil {
		return "", err
	}
	return template.Expand(values)
}
