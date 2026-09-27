package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"deskorder/internal/model"
)

const productOptionIDBytes = 6

func generateProductOptionID(prefix string) (string, error) {
	randomBytes := make([]byte, productOptionIDBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(randomBytes), nil
}

func normalizeProductOptions(options model.ProductOptionGroups) (model.ProductOptionGroups, error) {
	if len(options) == 0 {
		return nil, nil
	}

	normalized := make(model.ProductOptionGroups, 0, len(options))
	groupIDs := make(map[string]struct{}, len(options))
	for groupIndex, group := range options {
		groupName := strings.TrimSpace(group.Name)
		if groupName == "" {
			return nil, fmt.Errorf("第%d组选项名称不能为空", groupIndex+1)
		}

		groupID := strings.TrimSpace(group.ID)
		if groupID == "" {
			generatedID, err := generateProductOptionID("pog_")
			if err != nil {
				return nil, errors.New("生成选项组标识失败")
			}
			groupID = generatedID
		}
		if _, exists := groupIDs[groupID]; exists {
			return nil, fmt.Errorf("选项组 %s 标识重复", groupName)
		}
		groupIDs[groupID] = struct{}{}

		selectType := strings.ToLower(strings.TrimSpace(group.SelectType))
		if selectType == "" {
			selectType = model.ProductOptionSelectTypeSingle
		}
		if selectType != model.ProductOptionSelectTypeSingle && selectType != model.ProductOptionSelectTypeMulti {
			return nil, fmt.Errorf("选项组 %s 类型无效", groupName)
		}
		if len(group.Values) == 0 {
			return nil, fmt.Errorf("选项组 %s 至少需要一个选项值", groupName)
		}

		valueIDs := make(map[string]struct{}, len(group.Values))
		normalizedValues := make([]model.ProductOptionValue, 0, len(group.Values))
		for valueIndex, value := range group.Values {
			valueName := strings.TrimSpace(value.Name)
			if valueName == "" {
				return nil, fmt.Errorf("选项组 %s 的第%d个选项值名称不能为空", groupName, valueIndex+1)
			}
			if value.PriceDelta < 0 {
				return nil, fmt.Errorf("选项组 %s 的选项值 %s 加价不能小于0", groupName, valueName)
			}

			valueID := strings.TrimSpace(value.ID)
			if valueID == "" {
				generatedID, err := generateProductOptionID("pov_")
				if err != nil {
					return nil, errors.New("生成选项值标识失败")
				}
				valueID = generatedID
			}
			if _, exists := valueIDs[valueID]; exists {
				return nil, fmt.Errorf("选项组 %s 存在重复选项值标识", groupName)
			}
			valueIDs[valueID] = struct{}{}

			normalizedValues = append(normalizedValues, model.ProductOptionValue{
				ID:         valueID,
				Name:       valueName,
				PriceDelta: value.PriceDelta,
				Sort:       value.Sort,
			})
		}

		required := group.Required
		minSelect := group.MinSelect
		maxSelect := group.MaxSelect
		if minSelect < 0 {
			minSelect = 0
		}
		if selectType == model.ProductOptionSelectTypeSingle {
			if required || minSelect > 0 {
				required = true
				minSelect = 1
			} else {
				minSelect = 0
			}
			maxSelect = 1
		} else {
			if required && minSelect == 0 {
				minSelect = 1
			}
			if maxSelect <= 0 || maxSelect > len(normalizedValues) {
				maxSelect = len(normalizedValues)
			}
			if minSelect > maxSelect {
				return nil, fmt.Errorf("选项组 %s 的最少选择数量不能大于最多选择数量", groupName)
			}
			required = required || minSelect > 0
		}

		normalized = append(normalized, model.ProductOptionGroup{
			ID:         groupID,
			Name:       groupName,
			SelectType: selectType,
			Required:   required,
			MinSelect:  minSelect,
			MaxSelect:  maxSelect,
			Sort:       group.Sort,
			Values:     normalizedValues,
		})
	}

	return normalized, nil
}

func resolveOrderItemSelectedOptions(product *model.Product, selections model.ProductOptionSelections) (model.OrderItemOptionGroupSnapshots, string, int, error) {
	if product == nil {
		return nil, "", 0, errors.New("商品不存在")
	}
	if len(product.Options) == 0 {
		if len(selections) > 0 {
			return nil, "", 0, fmt.Errorf("商品 %s 当前不支持规格选择", product.Name)
		}
		return nil, "", 0, nil
	}

	selectionMap := make(map[string][]string, len(selections))
	for _, selection := range selections {
		groupID := strings.TrimSpace(selection.GroupID)
		if groupID == "" {
			return nil, "", 0, errors.New("存在未指定选项组的商品选择")
		}
		if _, exists := selectionMap[groupID]; exists {
			return nil, "", 0, errors.New("同一选项组不能重复提交")
		}
		valueIDs := make([]string, 0, len(selection.ValueIDs))
		seenValueIDs := make(map[string]struct{}, len(selection.ValueIDs))
		for _, valueID := range selection.ValueIDs {
			trimmedValueID := strings.TrimSpace(valueID)
			if trimmedValueID == "" {
				continue
			}
			if _, exists := seenValueIDs[trimmedValueID]; exists {
				continue
			}
			seenValueIDs[trimmedValueID] = struct{}{}
			valueIDs = append(valueIDs, trimmedValueID)
		}
		selectionMap[groupID] = valueIDs
	}

	selectedGroups := make(model.OrderItemOptionGroupSnapshots, 0, len(selections))
	summaryParts := make([]string, 0, len(selections))
	totalDelta := 0
	for _, group := range product.Options {
		selectedValueIDs, exists := selectionMap[group.ID]
		if !exists {
			if group.Required || group.MinSelect > 0 {
				return nil, "", 0, fmt.Errorf("请选择%s", group.Name)
			}
			continue
		}
		delete(selectionMap, group.ID)

		selectedCount := len(selectedValueIDs)
		if group.SelectType == model.ProductOptionSelectTypeSingle {
			if selectedCount != 1 {
				return nil, "", 0, fmt.Errorf("%s 需要且只能选择1项", group.Name)
			}
		} else {
			if selectedCount < group.MinSelect {
				return nil, "", 0, fmt.Errorf("%s 至少选择%d项", group.Name, group.MinSelect)
			}
			if group.MaxSelect > 0 && selectedCount > group.MaxSelect {
				return nil, "", 0, fmt.Errorf("%s 最多选择%d项", group.Name, group.MaxSelect)
			}
		}

		valueMap := make(map[string]model.ProductOptionValue, len(group.Values))
		for _, value := range group.Values {
			valueMap[value.ID] = value
		}

		selectedValues := make([]model.OrderItemOptionValueSnapshot, 0, selectedCount)
		selectedNames := make([]string, 0, selectedCount)
		for _, valueID := range selectedValueIDs {
			value, exists := valueMap[valueID]
			if !exists {
				return nil, "", 0, fmt.Errorf("%s 包含无效选项值", group.Name)
			}
			selectedValues = append(selectedValues, model.OrderItemOptionValueSnapshot{
				ID:         value.ID,
				Name:       value.Name,
				PriceDelta: value.PriceDelta,
			})
			selectedNames = append(selectedNames, value.Name)
			totalDelta += value.PriceDelta
		}

		selectedGroups = append(selectedGroups, model.OrderItemOptionGroupSnapshot{
			GroupID:    group.ID,
			GroupName:  group.Name,
			SelectType: group.SelectType,
			Values:     selectedValues,
		})
		summaryParts = append(summaryParts, group.Name+":"+strings.Join(selectedNames, "/"))
	}

	if len(selectionMap) > 0 {
		return nil, "", 0, errors.New("存在无效的商品选项组")
	}

	return selectedGroups, strings.Join(summaryParts, "；"), totalDelta, nil
}
