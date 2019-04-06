package reflector

// ExtractTagsFromStruct returns list of tag values
func ExtractTagsFromStruct(tagName string, i interface{}) []string {
	r := New(i)
	m := r.ExtractTags(tagName, WithoutEmpty(), WithoutMinus())
	res := make([]string, 0, len(m))
	for _, tag := range m {
		res = append(res, tag)
	}

	return res
}

// StructToMapByTags shortcut to Reflector.ExtractValues
func StructToMapByTags(tagName string, i interface{}, skipNils bool) map[string]interface{} {
	r := New(i)

	return r.ExtractValues(tagName, skipNils, WithoutEmpty(), WithoutMinus())
}
