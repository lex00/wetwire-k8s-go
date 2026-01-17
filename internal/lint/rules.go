package lint

// AllRules returns all available lint rules.
func AllRules() []Rule {
	return []Rule{
		RuleWK8001(),
		RuleWK8002(),
		RuleWK8003(),
		RuleWK8004(),
		RuleWK8005(),
		RuleWK8006(),
		RuleWK8041(),
		RuleWK8042(),
		RuleWK8101(),
		RuleWK8102(),
		RuleWK8103(),
		RuleWK8104(),
		RuleWK8105(),
		RuleWK8201(),
		RuleWK8202(),
		RuleWK8203(),
		RuleWK8204(),
		RuleWK8205(),
		RuleWK8207(),
		RuleWK8208(),
		RuleWK8209(),
		RuleWK8301(),
		RuleWK8302(),
		RuleWK8303(),
		RuleWK8304(),
		RuleWK8401(),
	}
}
