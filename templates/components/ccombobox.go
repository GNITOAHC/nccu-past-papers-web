package components

type ComboBox struct {
	// Path to JSON file, should be a list, should be formatted as follow:
	// {
	//   "itemList":[
	//     {"id": 0, name": "Item 1", "disabled": false}
	//   ]
	// }
	ItemPath    string
	QueryName   string
	Placeholder string
}
