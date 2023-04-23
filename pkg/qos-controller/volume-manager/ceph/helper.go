package ceph

import (
	"regexp"

	vm "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller/volume-manager"
)

const (
	qosPattern = `^[1-9][0-9]*(M|G|T)?$`
)

var qosReg *regexp.Regexp = regexp.MustCompile(qosPattern)

// getQoSRulesFromMeta extracts QoS rules map from rbd image metadata
func getQoSRulesFromMeta(meta map[string]string) RBDQoSRules {
	rules := make(RBDQoSRules)
	for k, v := range meta {
		if _, ok := RBDQoSKeyMap[k]; ok {
			rules[k] = v
		}
	}
	return rules
}

// rbdQoSRules converts QoS settings map to rbd QoS rules map
func rbdQoSRules(settings vm.QoSSettings) RBDQoSRules {
	rules := make(RBDQoSRules)
	for k, v := range settings {
		rules[QoSKeyMap[k]] = v
	}
	return rules
}

// calSet calculates the QoS rules to be updated or added
func calSet(cur, spec RBDQoSRules) RBDQoSRules {
	rules := make(RBDQoSRules)
	for k, v := range spec {
		if cur[k] == v {
			continue
		}
		rules[k] = v
	}
	return rules
}

// calRemove calculates the QoS rules to be removed
func calRemove(cur, spec RBDQoSRules) RBDQoSRules {
	rules := make(RBDQoSRules)
	for k, v := range cur {
		if _, ok := spec[k]; ok {
			continue
		}
		rules[k] = v
	}
	return rules
}

// isQoSValueValid checks if the QoS setting value is valid
func isQoSValueValid(v string) bool {
	return qosReg.MatchString(v)
}
