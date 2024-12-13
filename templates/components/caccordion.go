package components

type Accordion struct {
	// Path to JSON file, should be formatted as follow:
	// {
	//   "itemList":[
	//     {"id": 0, "summary": "Summary 1", "detail": "Detail 1"}
	//   ]
	// }
	ItemPath string
}
