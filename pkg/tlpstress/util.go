package tlpstress

// Creates a set of common labels that should be applied to any object in association with
// the Stress instance. The name parameter is the Stress instance name.
func LabelsForStress(name string) map[string]string {
	return map[string]string{
		"app": "stress",
		"stress": name,
	}
}
