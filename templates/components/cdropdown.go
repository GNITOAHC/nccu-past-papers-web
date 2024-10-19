package components

type Dropdown struct {
	TitleId string // Id of the Dropdown title
	Title   string // Title of the dropdown

	// Every option will have same attribute which is `optionid`
	// Use document.querySelectorAll("[optionid='theme-toggle']") to get all options
	OptionId string

	Menu map[string]string // Key: Option display text, Value: Option value

	ButtonClass string // TailwindCSS classes for the button
	PanelClass  string // TailwindCSS classes for the panel
	ListClass   string // TailwindCSS classes for the button list

	IsLink bool // If true, the value will be applied to the anchor tag href
}
