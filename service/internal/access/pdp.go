package access

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/opentdf/platform/protocol/go/policy"
	"github.com/opentdf/platform/service/logger"
)

type Pdp struct {
	logger *logger.Logger
}

func NewPdp(l *logger.Logger) *Pdp {
	return &Pdp{
		logger: l,
	}
}

// DetermineAccess will take data Attribute Values, entities mapped entityId to Attribute Value FQNs, and data AttributeDefinitions,
// compare every data Attribute against every entity's set of Attribute Values, generating a rolled-up decision
// result for each entity, as well as a detailed breakdown of every data comparison.
func (pdp *Pdp) DetermineAccess(
	ctx context.Context,
	dataAttributes []*policy.Value,
	entityAttributeSets map[string][]string,
	attributeDefinitions []*policy.Attribute,
) (map[string]*Decision, error) {
	pdp.logger.DebugContext(ctx, "DetermineAccess")
	// Group all the Data Attribute Values by their Definitions (that is, "<namespace>/attr/<attrname>").
	dataAttrValsByDefinition, err := GroupValuesByDefinition(dataAttributes)
	if err != nil {
		pdp.logger.ErrorContext(ctx, "error grouping data attributes by definition",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to group data attributes by definition: %w", err)
	}

	// Precompute grouped entity attribute FQNs by definition for all entities
	entityAttrGroups := make(map[string]map[string][]string)
	for entityID, attrFqns := range entityAttributeSets {
		grouped, err := GroupValueFqnsByDefinition(attrFqns)
		if err != nil {
			pdp.logger.ErrorContext(ctx, "error grouping entity attribute values for entity",
				slog.String("entityID", entityID),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to group entity attribute values for entity %s: %w", entityID, err)
		}
		entityAttrGroups[entityID] = grouped
	}

	// Unlike with Values, there should only be *one* Attribute Definition per FQN (e.g "https://namespace.org/attr/MyAttr")
	fqnToDefinitionMap, err := GetFqnToDefinitionMap(ctx, attributeDefinitions, pdp.logger)
	if err != nil {
		pdp.logger.ErrorContext(ctx, "error grouping attribute definitions by FQN",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to group attribute definitions by FQN: %w", err)
	}

	decisions, err := pdp.evaluateDataAttributesForEntities(ctx, dataAttrValsByDefinition, fqnToDefinitionMap, entityAttrGroups)
	if err != nil {
		return nil, err
	}

	return decisions, nil
}

// evaluateDataAttributesForEntities processes each data attribute definition and updates entity decisions accordingly.
func (pdp *Pdp) evaluateDataAttributesForEntities(
	ctx context.Context,
	dataAttrValsByDefinition map[string][]*policy.Value,
	fqnToDefinitionMap map[string]*policy.Attribute,
	entityAttrGroups map[string]map[string][]string,
) (map[string]*Decision, error) {
	decisions := make(map[string]*Decision)
	for definitionFqn, distinctValues := range dataAttrValsByDefinition {
		pdp.logger.DebugContext(ctx, "Evaluating data attribute fqn", "fqn:", definitionFqn)
		attrDefinition, ok := fqnToDefinitionMap[definitionFqn]
		if !ok {
			return nil, fmt.Errorf("expected an Attribute Definition under the FQN %s", definitionFqn)
		}

		var (
			entityRuleDecision map[string]DataRuleResult
			err                error
		)
		switch attrDefinition.GetRule() {
		case policy.AttributeRuleTypeEnum_ATTRIBUTE_RULE_TYPE_ENUM_ALL_OF:
			pdp.logger.DebugContext(ctx, "Evaluating under allOf", "name", definitionFqn)
			entityRuleDecision, err = pdp.allOfRule(ctx, distinctValues, entityAttrGroups)
		case policy.AttributeRuleTypeEnum_ATTRIBUTE_RULE_TYPE_ENUM_ANY_OF:
			pdp.logger.DebugContext(ctx, "Evaluating under anyOf", "name", definitionFqn)
			entityRuleDecision, err = pdp.anyOfRule(ctx, distinctValues, entityAttrGroups)
		case policy.AttributeRuleTypeEnum_ATTRIBUTE_RULE_TYPE_ENUM_HIERARCHY:
			pdp.logger.DebugContext(ctx, "Evaluating under hierarchy", "name", definitionFqn)
			entityRuleDecision, err = pdp.hierarchyRule(ctx, distinctValues, entityAttrGroups, attrDefinition.GetValues())
		case policy.AttributeRuleTypeEnum_ATTRIBUTE_RULE_TYPE_ENUM_UNSPECIFIED:
			return nil, fmt.Errorf("unset AttributeDefinition rule: %s", attrDefinition.GetRule())
		default:
			return nil, fmt.Errorf("unrecognized AttributeDefinition rule: %s", attrDefinition.GetRule())
		}
		if err != nil {
			return nil, fmt.Errorf("error evaluating rule: %s", err.Error())
		}

		for entityID, ruleResult := range entityRuleDecision {
			pdp.updateEntityDecision(decisions, entityID, ruleResult, attrDefinition)
		}
	}
	return decisions, nil
}

// updateEntityDecision updates or initializes the Decision for an entity.
func (pdp *Pdp) updateEntityDecision(
	decisions map[string]*Decision,
	entityID string,
	ruleResult DataRuleResult,
	attrDefinition *policy.Attribute,
) {
	entityDecision := decisions[entityID]
	ruleResult.RuleDefinition = attrDefinition
	if entityDecision == nil {
		decisions[entityID] = &Decision{
			Access:  ruleResult.Passed,
			Results: []DataRuleResult{ruleResult},
		}
	} else {
		entityDecision.Access = entityDecision.Access && ruleResult.Passed
		entityDecision.Results = append(entityDecision.Results, ruleResult)
	}
}

// AllOf the Data Attribute Values should be present in AllOf the Entity's entityAttributeValue sets
// Accepts
// - a set of data Attribute Values with the same FQN
// - a map of entity Attribute Values keyed by entity ID
// Returns a map of DataRuleResults keyed by Subject
func (pdp *Pdp) allOfRule(ctx context.Context, dataAttrValuesOfOneDefinition []*policy.Value, entityAttrGroups map[string]map[string][]string) (map[string]DataRuleResult, error) {
	ruleResultsByEntity := make(map[string]DataRuleResult)

	def, err := GetDefinitionFqnFromValue(dataAttrValuesOfOneDefinition[0])
	if err != nil {
		return nil, fmt.Errorf("error getting definition FQN from data attribute value: %s", err.Error())
	}
	pdp.logger.DebugContext(ctx, "Evaluating allOf decision", "attribute definition FQN", def)
	pdp.logger.TraceContext(ctx, "Attribute values for ", "attribute definition FQN", def, "values", dataAttrValuesOfOneDefinition)

	for entityID, groupedEntityAttrValsByDefinition := range entityAttrGroups {
		ruleResultsByEntity[entityID] = pdp.evaluateAllOfForEntity(ctx, dataAttrValuesOfOneDefinition, groupedEntityAttrValsByDefinition, entityID)
	}

	return ruleResultsByEntity, nil
}

// evaluateAllOfForEntity evaluates the allOf rule for a single entity.
func (pdp *Pdp) evaluateAllOfForEntity(
	ctx context.Context,
	dataAttrValuesOfOneDefinition []*policy.Value,
	groupedEntityAttrValsByDefinition map[string][]string,
	entityID string,
) DataRuleResult {
	var valueFailures []ValueFailure
	entityPassed := false

	for dvIndex, dataAttrVal := range dataAttrValuesOfOneDefinition {
		attrDefFqn, err := GetDefinitionFqnFromValue(dataAttrVal)
		if err != nil {
			pdp.logger.ErrorContext(ctx, "error getting definition FQN from data attribute value",
				slog.String("error", err.Error()),
				slog.String("entityID", entityID))
			continue
		}
		pdp.logger.DebugContext(ctx, "Evaluating allOf decision", "data attr fqn", attrDefFqn, "value", dataAttrVal.GetValue())
		// Build case-insensitive set for entity attribute FQNs
		fqnSet := make(map[string]struct{})
		for _, fqn := range groupedEntityAttrValsByDefinition[attrDefFqn] {
			fqnSet[strings.ToLower(fqn)] = struct{}{}
		}
		found := getIsValueFoundInFqnValuesSet(dataAttrValuesOfOneDefinition[dvIndex], fqnSet, pdp.logger)

		if !found {
			pdp.logger.WarnContext(ctx, "AllOf not satisfied",
				slog.String("dataAttrFqn", attrDefFqn),
				slog.String("value", dataAttrVal.GetValue()),
				slog.String("entityID", entityID))
			denialMsg := fmt.Sprintf("AllOf not satisfied for data attr %s with value %s and entity %s", attrDefFqn, dataAttrVal.GetValue(), entityID)
			valueFailures = append(valueFailures, ValueFailure{
				DataAttribute: dataAttrValuesOfOneDefinition[dvIndex],
				Message:       denialMsg,
			})
		}
	}

	if len(valueFailures) == 0 {
		entityPassed = true
	}
	return DataRuleResult{
		Passed:        entityPassed,
		ValueFailures: valueFailures,
	}
}

// AnyOf the Data Attribute Values can be present in AnyOf the Entity's Attribute Value FQN sets
// Accepts
// - a set of data Attribute Values with the same FQN
// - a map of entity Attribute Values keyed by entity ID
// Returns a map of DataRuleResults keyed by Subject entity ID
func (pdp *Pdp) anyOfRule(ctx context.Context, dataAttrValuesOfOneDefinition []*policy.Value, entityAttrGroups map[string]map[string][]string) (map[string]DataRuleResult, error) {
	ruleResultsByEntity := make(map[string]DataRuleResult)

	attrDefFqn, err := GetDefinitionFqnFromValue(dataAttrValuesOfOneDefinition[0])
	if err != nil {
		return nil, fmt.Errorf("error getting definition FQN from data attribute value: %s", err.Error())
	}
	pdp.logger.DebugContext(ctx, "Evaluating anyOf decision", "attribute definition FQN", attrDefFqn)
	pdp.logger.TraceContext(ctx, "Attribute values for ", "attribute definition FQN", attrDefFqn, "values", dataAttrValuesOfOneDefinition)

	for entityID, entityAttrGroup := range entityAttrGroups {
		ruleResultsByEntity[entityID] = pdp.evaluateAnyOfForEntity(ctx, dataAttrValuesOfOneDefinition, entityAttrGroup, entityID, attrDefFqn)
	}

	return ruleResultsByEntity, nil
}

// evaluateAnyOfForEntity evaluates the anyOf rule for a single entity.
func (pdp *Pdp) evaluateAnyOfForEntity(
	ctx context.Context,
	dataAttrValuesOfOneDefinition []*policy.Value,
	entityAttrGroup map[string][]string,
	entityID string,
	attrDefFqn string,
) DataRuleResult {
	var valueFailures []ValueFailure
	entityPassed := false

	// Build case-insensitive set for entity attribute FQNs
	fqnSet := make(map[string]struct{})
	for _, fqn := range entityAttrGroup[attrDefFqn] {
		fqnSet[strings.ToLower(fqn)] = struct{}{}
	}

	for dvIndex, dataAttrVal := range dataAttrValuesOfOneDefinition {
		pdp.logger.DebugContext(ctx, "Evaluating anyOf decision", "attribute definition FQN", attrDefFqn, "value", dataAttrVal.GetValue())
		found := getIsValueFoundInFqnValuesSet(dataAttrVal, fqnSet, pdp.logger)

		if !found {
			pdp.logger.WarnContext(ctx, "anyOf not satisfied",
				slog.String("dataAttrFqn", attrDefFqn),
				slog.String("value", dataAttrVal.GetValue()),
				slog.String("entityID", entityID))
			denialMsg := fmt.Sprintf("anyOf not satisfied for data attr %s with value %s and entity %s - anyOf is permissive, so this doesn't mean overall failure", attrDefFqn, dataAttrVal.GetValue(), entityID)
			valueFailures = append(valueFailures, ValueFailure{
				DataAttribute: dataAttrValuesOfOneDefinition[dvIndex],
				Message:       denialMsg,
			})
		}
	}

	if len(valueFailures) < len(dataAttrValuesOfOneDefinition) {
		pdp.logger.DebugContext(ctx, "anyOf satisfied", "attribute definition FQN", attrDefFqn, "entityId", entityID)
		entityPassed = true
	}
	return DataRuleResult{
		Passed:        entityPassed,
		ValueFailures: valueFailures,
	}
}

// Hierarchy rule compares the HIGHEST (that is, numerically lowest index) data Attribute Value for a given Attribute Value FQN
// with the LOWEST (that is, numerically highest index) entity value for a given Attribute Value FQN.
//
// If multiple data values (that is, Attribute Values) for a given hierarchy AttributeDefinition are present for the same FQN, the highest will be chosen and
// the others ignored.
//
// If multiple entity Attribute Values for a hierarchy AttributeDefinition are present for the same FQN, the lowest will be chosen,
// and the others ignored.
func buildFqnOrderMap(order []*policy.Value) map[string]int {
	m := make(map[string]int, len(order))
	for idx, v := range order {
		fqn := v.GetFqn()
		if fqn == "" {
			defFqn, err := GetDefinitionFqnFromValue(v)
			if err == nil && defFqn != "" && v.GetValue() != "" {
				fqn = fmt.Sprintf("%s/value/%s", defFqn, v.GetValue())
			}
		}
		// Only assign an index if FQN is valid (non-empty)
		if fqn != "" {
			m[fqn] = idx
		}
	}
	return m
}

func (pdp *Pdp) hierarchyRule(ctx context.Context, dataAttrValuesOfOneDefinition []*policy.Value, entityAttrGroups map[string]map[string][]string, order []*policy.Value) (map[string]DataRuleResult, error) {
	ruleResultsByEntity := make(map[string]DataRuleResult)
	fqnOrderMap := buildFqnOrderMap(order)

	highestDataAttrVal, err := pdp.getHighestRankedInstanceFromDataAttributes(ctx, order, fqnOrderMap, dataAttrValuesOfOneDefinition, pdp.logger)
	if err != nil {
		return nil, fmt.Errorf("error getting highest ranked instance from data attributes: %s", err.Error())
	}
	if highestDataAttrVal == nil {
		pdp.logger.WarnContext(ctx, "No data attribute value found that matches attribute definition allowed values! All entity access will be rejected!")
	} else {
		pdp.logger.DebugContext(ctx, "Highest ranked hierarchy value on data attributes found", "value", highestDataAttrVal.GetValue())
	}

	for entityID, entityAttrGroup := range entityAttrGroups {
		ruleResultsByEntity[entityID] = pdp.evaluateHierarchyForEntity(ctx, highestDataAttrVal, fqnOrderMap, entityAttrGroup, entityID)
	}

	return ruleResultsByEntity, nil
}

// evaluateHierarchyForEntity evaluates the hierarchy rule for a single entity.
func (pdp *Pdp) evaluateHierarchyForEntity(
	ctx context.Context,
	highestDataAttrVal *policy.Value,
	fqnOrderMap map[string]int,
	entityAttrGroup map[string][]string,
	entityID string,
) DataRuleResult {
	entityPassed := false
	valueFailures := []ValueFailure{}

	if highestDataAttrVal != nil {
		attrDefFqn, err := GetDefinitionFqnFromValue(highestDataAttrVal)
		if err != nil {
			pdp.logger.ErrorContext(ctx, "error getting definition FQN from data attribute value",
				slog.String("error", err.Error()),
				slog.String("entityID", entityID))
			return DataRuleResult{
				Passed:        false,
				ValueFailures: valueFailures,
			}
		}
		pdp.logger.DebugContext(ctx, "Evaluating hierarchy decision", "attribute definition fqn", attrDefFqn, "value", highestDataAttrVal.GetValue())
		pdp.logger.TraceContext(ctx, "Value obj", "value", highestDataAttrVal.GetValue(), "obj", highestDataAttrVal)

		passed, err := entityRankGreaterThanOrEqualToDataRank(fqnOrderMap, highestDataAttrVal, entityAttrGroup[attrDefFqn], pdp.logger)
		if err != nil {
			pdp.logger.ErrorContext(ctx, "error comparing entity rank to data rank",
				slog.String("error", err.Error()),
				slog.String("entityID", entityID))
			return DataRuleResult{
				Passed:        false,
				ValueFailures: valueFailures,
			}
		}
		entityPassed = passed

		if !entityPassed {
			pdp.logger.WarnContext(ctx, "Hierarchy not satisfied",
				slog.String("entityID", entityID),
				slog.String("dataValue", highestDataAttrVal.GetValue()))
			denialMsg := fmt.Sprintf("Hierarchy - Entity: %s hierarchy values rank below data hierarchy value of %s", entityID, highestDataAttrVal.GetValue())
			valueFailures = append(valueFailures, ValueFailure{
				DataAttribute: highestDataAttrVal,
				Message:       denialMsg,
			})
		}
	} else {
		pdp.logger.WarnContext(ctx, "Hierarchy - No data values found exist in attribute definition, no hierarchy comparison possible",
			slog.String("entityID", entityID))
		denialMsg := fmt.Sprintf("Hierarchy - No data values found exist in attribute definition, no hierarchy comparison possible, entity %s is denied", entityID)
		valueFailures = append(valueFailures, ValueFailure{
			DataAttribute: nil,
			Message:       denialMsg,
		})
	}
	return DataRuleResult{
		Passed:        entityPassed,
		ValueFailures: valueFailures,
	}
}

// It is possible that a data policy may have more than one Hierarchy value for the same data attribute definition
// name, e.g.:
// - "https://namespace.org/attr/MyHierarchyAttr/value/Value1"
// - "https://namespace.org/attr/MyHierarchyAttr/value/Value2"
// Since by definition hierarchy comparisons have to be one-data-value-to-many-entity-values, this won't work.
// So, in a scenario where there are multiple data values to choose from, grab the "highest" ranked value
// present in the set of data Attribute Values, and use that as the point of comparison, ignoring the "lower-ranked" data values.
// If we find a data value that does not exist in the attribute definition's list of valid values, we will skip it
// If NONE of the data values exist in the attribute definitions list of valid values, return a nil instance
func (pdp *Pdp) getHighestRankedInstanceFromDataAttributes(ctx context.Context, order []*policy.Value, fqnOrderMap map[string]int, dataAttributeGroup []*policy.Value, logger *logger.Logger) (*policy.Value, error) {
	// For hierarchy, convention is 0 == most privileged, 1 == less privileged, etc
	// So initialize with the LEAST privileged rank in the defined order
	highestDVIndex := len(order) - 1
	var highestRankedInstance *policy.Value
	for _, dataAttr := range dataAttributeGroup {
		foundRank, err := getOrderOfValueWithMap(fqnOrderMap, order, dataAttr, logger)
		if err != nil {
			return nil, fmt.Errorf("error getting order of value: %s", err.Error())
		}
		if foundRank == -1 {
			msg := fmt.Sprintf("Data value %s is not in order and is not a valid value for this attribute - ignoring this invalid value and continuing to look for a valid one...", dataAttr.GetValue())
			pdp.logger.WarnContext(ctx, "invalid data value for attribute",
				slog.String("value", dataAttr.GetValue()),
				slog.String("error", msg))
			continue
		}
		pdp.logger.DebugContext(ctx, "Found data", "rank", foundRank, "value", dataAttr.GetValue(), "maxRank", highestDVIndex)
		if foundRank <= highestDVIndex {
			pdp.logger.DebugContext(ctx, "Updating rank!")
			highestDVIndex = foundRank
			gotAttr := dataAttr
			highestRankedInstance = gotAttr
		}
	}
	return highestRankedInstance, nil
}

// Check for a match of a singular Attribute Value in a set of Attribute Value FQNs
func getIsValueFoundInFqnValuesSet(v *policy.Value, fqnSet map[string]struct{}, l *logger.Logger) bool {
	valFqn := v.GetFqn()
	if valFqn == "" {
		l.Error("Unexpected empty FQN for value",
			slog.Any("value", v))
		return false
	}
	_, found := fqnSet[strings.ToLower(valFqn)]
	return found
}

// Given set of ordered/ranked values, a data singular Attribute Value, and a set of entity Attribute Values,
// determine if the entity Attribute Values include a ranked value that equals or exceeds
// the rank of the data Attribute Value.
// For hierarchy, convention is 0 == most privileged, 1 == less privileged, etc
func entityRankGreaterThanOrEqualToDataRank(
	fqnOrderMap map[string]int, dataAttribute *policy.Value,
	entityAttrValueFqnsGroup []string,
	log *logger.Logger,
) (bool, error) {
	result := false
	dvIndex, err := getOrderOfValueWithMap(fqnOrderMap, nil, dataAttribute, log)
	if err != nil {
		return false, err
	}
	for _, entityAttributeFqn := range entityAttrValueFqnsGroup {
		dataAttrDefFqn, err := GetDefinitionFqnFromValue(dataAttribute)
		if err != nil {
			return false, fmt.Errorf("error getting definition FQN from data attribute value: %s", err.Error())
		}
		entityAttrDefFqn, err := GetDefinitionFqnFromValueFqn(entityAttributeFqn)
		if err != nil {
			return false, fmt.Errorf("error getting definition FQN from entity attribute value: %s", err.Error())
		}
		if dataAttrDefFqn == entityAttrDefFqn {
			evIndex := -1
			if idx, ok := fqnOrderMap[entityAttributeFqn]; ok {
				evIndex = idx
			} else {
				evIndex = len(fqnOrderMap) + 1
			}
			if evIndex > dvIndex || dvIndex == -1 {
				result = false
				return result, nil
			} else if evIndex <= dvIndex {
				result = true
			}
		}
	}
	return result, nil
}

// Given a set of ordered/ranked values and a singular Attribute Value, return the
// rank #/index of the singular Attribute Value. If the value is not found, return -1.
// For hierarchy, convention is 0 == most privileged, 1 == less privileged, etc.
func getOrderOfValueWithMap(fqnOrderMap map[string]int, order []*policy.Value, v *policy.Value, log *logger.Logger) (int, error) {
	valFqn := v.GetFqn()
	if valFqn == "" {
		log.Debug("Unexpected empty FQN in value",
			slog.Any("value", v))
		return -1, nil
	}
	if idx, ok := fqnOrderMap[valFqn]; ok {
		return idx, nil
	}
	return -1, nil
}

// Given a set of ordered/ranked values and a singular Attribute Value, return the
// rank #/index of the singular Attribute Value. If the value is not found, return -1.
// For hierarchy, convention is 0 == most privileged, 1 == less privileged, etc
/* getOrderOfValueByFqn is now obsolete, replaced by getOrderOfValueWithMap */

// A Decision represents the overall access decision for a specific entity,
// - that is, the aggregate result of comparing entity Attribute Values to every data Attribute Value.
type Decision struct {
	// The important bit - does this entity Have Access or not, for this set of data attribute values
	// This will be TRUE if, for *every* DataRuleResult in Results, EntityRuleResult.Passed == TRUE
	// Otherwise, it will be false
	Access bool `json:"access" example:"false"`
	// Results will contain at most 1 DataRuleResult for each data Attribute Value.
	// e.g. if we compare an entity's Attribute Values against 5 data Attribute Values,
	// then there will be 5 rule results, each indicating whether this entity "passed" validation
	// for that data Attribute Value or not.
	//
	// If an entity was skipped for a particular rule evaluation because of a GroupBy clause
	// on the AttributeDefinition for a given data Attribute Value, however, then there may be
	// FEWER DataRuleResults then there are DataRules
	//
	// e.g. there are 5 data Attribute Values, and two entities each with a set of Attribute Values,
	// the definition for one of those data Attribute Values has a GroupBy clause that excludes the second entity
	//-> the first entity will have 5 DataRuleResults with Passed = true
	//-> the second entity will have 4 DataRuleResults Passed = true
	//-> both will have Access == true.
	Results []DataRuleResult `json:"entity_rule_result"`
}

// DataRuleResult represents the rule-level (or AttributeDefinition-level) decision for a specific entity -
// the result of comparing entity Attribute Values to a single data AttributeDefinition/rule (with potentially many values)
//
// There may be multiple "instances" (that is, Attribute Values) of a single AttributeDefinition on both data and entities,
// each with a different value.
type DataRuleResult struct {
	// Indicates whether, for this specific data AttributeDefinition, an entity satisfied
	// the rule conditions (allof/anyof/hierarchy)
	Passed bool `json:"passed" example:"false"`
	// Contains the AttributeDefinition of the data attribute rule this result represents
	RuleDefinition *policy.Attribute `json:"rule_definition"`
	// May contain 0 or more ValueFailure types, depending on the RuleDefinition and which (if any)
	// data Attribute Values/values the entity failed against
	//
	// For an AllOf rule, there should be no value failures if Passed=TRUE
	// For an AnyOf rule, there should be fewer entity value failures than
	// there are data attribute values in total if Passed=TRUE
	// For a Hierarchy rule, there should be either no value failures if Passed=TRUE,
	// or exactly one value failure if Passed=FALSE
	ValueFailures []ValueFailure `json:"value_failures"`
}

// ValueFailure indicates, for a given entity and data Attribute Value, which data values
// (aka specific data Attribute Value) the entity "failed" on.
//
// There may be multiple "instances" (that is, Attribute Values) of a single AttributeDefinition on both data and entities,
// each with a different value.
//
// A ValueFailure does not necessarily mean the requirements for an AttributeDefinition were not or will not be met,
// it is purely informational - there will be one value failure, per entity, per rule, per value the entity lacks -
// it is up to the rule itself (anyof/allof/hierarchy) to translate this into an overall failure or not.
type ValueFailure struct {
	// The data attribute w/value that "caused" the denial
	DataAttribute *policy.Value `json:"data_attribute"`
	// Optional denial message
	Message string `json:"message" example:"Criteria NOT satisfied for entity: {entity_id} - lacked attribute value: {attribute}"`
}

// GroupDefinitionsByFqn takes a slice of Attribute Definitions and returns them as a map:
// FQN -> Attribute Definition
func GetFqnToDefinitionMap(ctx context.Context, attributeDefinitions []*policy.Attribute, log *logger.Logger) (map[string]*policy.Attribute, error) {
	grouped := make(map[string]*policy.Attribute)
	for _, def := range attributeDefinitions {
		a, err := GetDefinitionFqnFromDefinition(def)
		if err != nil {
			return nil, err
		}
		if v, ok := grouped[a]; ok {
			// TODO: is this really an error case, or is logging a warning okay?
			log.WarnContext(ctx, "duplicate Attribute Definition FQN found when building FQN map",
				slog.String("fqn", a))
			log.TraceContext(ctx, "duplicate attribute definitions found",
				slog.Any("attr1", v),
				slog.Any("attr2", def))
		}
		grouped[a] = def
	}
	return grouped, nil
}

// Groups Attribute Values by their parent Attribute Definition FQN
func GroupValuesByDefinition(values []*policy.Value) (map[string][]*policy.Value, error) {
	groupings := make(map[string][]*policy.Value)
	for _, v := range values {
		// If the parent Definition & its FQN are not nil, rely on them
		if v.GetAttribute() != nil {
			defFqn := v.GetAttribute().GetFqn()
			if defFqn != "" {
				groupings[defFqn] = append(groupings[defFqn], v)
				continue
			}
		}
		// Otherwise derive the grouping relation from the FQNs
		defFqn, err := GetDefinitionFqnFromValueFqn(v.GetFqn())
		if err != nil {
			return nil, err
		}
		groupings[defFqn] = append(groupings[defFqn], v)
	}
	return groupings, nil
}

func GroupValueFqnsByDefinition(valueFqns []string) (map[string][]string, error) {
	groupings := make(map[string][]string)
	for _, v := range valueFqns {
		defFqn, err := GetDefinitionFqnFromValueFqn(v)
		if err != nil {
			return nil, err
		}
		groupings[defFqn] = append(groupings[defFqn], v)
	}
	return groupings, nil
}

func GetDefinitionFqnFromValue(v *policy.Value) (string, error) {
	if v.GetAttribute() != nil {
		return GetDefinitionFqnFromDefinition(v.GetAttribute())
	}
	return GetDefinitionFqnFromValueFqn(v.GetFqn())
}

// Splits off the Value from the FQN to get the parent Definition FQN:
//
//	Input: https://<namespace>/attr/<attr name>/value/<value>
//	Output: https://<namespace>/attr/<attr name>
func GetDefinitionFqnFromValueFqn(valueFqn string) (string, error) {
	if valueFqn == "" {
		return "", fmt.Errorf("unexpected empty value FQN in GetDefinitionFqnFromValueFqn")
	}
	idx := strings.LastIndex(valueFqn, "/value/")
	if idx == -1 {
		return "", fmt.Errorf("value FQN (%s) is of unknown format with no '/value/' segment", valueFqn)
	}
	defFqn := valueFqn[:idx]
	if defFqn == "" {
		return "", fmt.Errorf("value FQN (%s) is of unknown format with no known parent Definition", valueFqn)
	}
	return defFqn, nil
}

func GetDefinitionFqnFromDefinition(def *policy.Attribute) (string, error) {
	// see if its FQN is already supplied
	fqn := def.GetFqn()
	if fqn != "" {
		return fqn, nil
	}
	// otherwise build it from the namespace and name
	ns := def.GetNamespace()
	if ns == nil {
		return "", fmt.Errorf("attribute definition has unexpectedly nil namespace")
	}
	nsName := ns.GetName()
	if nsName == "" {
		return "", fmt.Errorf("attribute definition's Namespace has unexpectedly empty name")
	}
	nsFqn := ns.GetFqn()
	attr := def.GetName()
	if attr == "" {
		return "", fmt.Errorf("attribute definition has unexpectedly empty name")
	}
	// Namespace FQN contains 'https://' scheme prefix, but Namespace Name does not
	if nsFqn != "" {
		return fmt.Sprintf("%s/attr/%s", nsFqn, attr), nil
	}
	return fmt.Sprintf("https://%s/attr/%s", nsName, attr), nil
}
