package tlpstress

// Creates a set of common labels that should be applied to any object in association with
// the TLPStress instance. The name parameter is the TLPStress instance name.
func LabelsForTLPStress(name string) map[string]string {
	return map[string]string{
		"app": "tlpstress",
		"tlpstress": name,
	}
}
