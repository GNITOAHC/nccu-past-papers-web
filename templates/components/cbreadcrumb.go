package components

type Breadcrumb struct {
	// Links & Names should have the same length
	// For instance, if Links = {"a", "b", "c"} and Names = {"A", "B", "C"}, then the breadcrumb should look like:
	// [A](a) > [B](b) > [C](c)
	Links []string // Breadcrumbs' href attributes
	Names []string // Breadcrumbs' display names
}
