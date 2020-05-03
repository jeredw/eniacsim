package units

import (
	. "github.com/jeredw/eniacsim/lib"
)

func clearSettings() []BoolSwitchSetting {
	return []BoolSwitchSetting{
		{"C", true}, {"c", true},
		{"0", false},
	}
}

func ftOpSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"A-2", 0},
		{"A-1", 1},
		{"A0", 2},
		{"A+1", 3},
		{"A+2", 4},
		{"S+2", 5},
		{"S+1", 6},
		{"S0", 7},
		{"S-1", 8},
		{"S-2", 9},
	}
}

func ftArgSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"0", 0},
		{"NC", 1}, {"nc", 1},
		{"C", 2}, {"c", 2},
	}
}

func sfSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"0", 0},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
		{"5", 5},
		{"6", 6},
		{"7", 7},
		{"8", 8},
		{"9", 9},
		{"10", 10},
	}
}

func scSettings() []ByteSwitchSetting {
	return []ByteSwitchSetting{
		{"0", 0},
		{"SC", 1}, {"sc", 1},
	}
}

func accOpSettings() []ByteSwitchSetting {
	return []ByteSwitchSetting{
		{"α", 0}, {"a", 0}, {"alpha", 0},
		{"β", 1}, {"b", 1}, {"beta", 1},
		{"γ", 2}, {"g", 2}, {"gamma", 2},
		{"δ", 3}, {"d", 3}, {"delta", 3},
		{"ε", 4}, {"e", 4}, {"epsilon", 4},
		{"0", 5},
		{"A", 6},
		{"AS", 7},
		{"S", 8},
	}
}

func rpSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"1", 0},
		{"2", 1},
		{"3", 2},
		{"4", 3},
		{"5", 4},
		{"6", 5},
		{"7", 6},
		{"8", 7},
		{"9", 8},
	}
}

func adSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"A", 0},
		{"B", 1},
		{"C", 2},
	}
}

func argSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"α", 0}, {"a", 0}, {"alpha", 0},
		{"β", 1}, {"b", 1}, {"beta", 1},
		{"0", 2},
	}
}

func placeSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"D4", 0}, {"d4", 0},
		{"D7", 1}, {"d7", 1},
		{"D8", 2}, {"d8", 2},
		{"D9", 3}, {"d9", 3},
		{"D10", 4}, {"d10", 4},
		{"S4", 5}, {"s4", 5}, {"R4", 5}, {"r4", 5},
		{"S7", 6}, {"s7", 6}, {"R7", 6}, {"r7", 6},
		{"S8", 7}, {"s8", 7}, {"R8", 7}, {"r8", 7},
		{"S9", 8}, {"s9", 8}, {"R9", 8}, {"r9", 8},
		{"S10", 9}, {"s10", 9}, {"R10", 9}, {"r10", 9},
	}
}

func roSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"RO", 1}, {"ro", 1},
		{"NRO", 0}, {"nro", 0},
	}
}

func ilSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"I", 1}, {"i", 1},
		{"NI", 0}, {"ni", 0},
	}
}

func anSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"1", 0},
		{"2", 1},
		{"3", 2},
		{"4", 3},
		{"OFF", 4}, {"off", 4},
	}
}

func signSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"P", 0}, {"p", 0},
		{"M", 1}, {"m", 1},
		{"T", 2}, {"t", 2},
	}
}

func delSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"D", 1}, {"d", 1},
		{"O", 0}, {"o", 0},
	}
}

func consSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"0", 0},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
		{"5", 5},
		{"6", 6},
		{"7", 7},
		{"8", 8},
		{"9", 9},
		{"PM1", 10}, {"pm1", 10},
		{"PM2", 11}, {"pm2", 11},
	}
}

func subSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"S", 1}, {"s", 1},
		{"0", 0},
	}
}

func valSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"0", 0},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
		{"5", 5},
		{"6", 6},
		{"7", 7},
		{"8", 8},
		{"9", 9},
	}
}

func pmSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"P", 0}, {"p", 0},
		{"M", 1}, {"m", 1},
	}
}

func ninepSettings() []BoolSwitchSetting {
	return []BoolSwitchSetting{
		{"C", true}, {"c", true},
		{"Cpp", false}, {"cpp", false},
	}
}

func recvSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"α", 0}, {"a", 0}, {"alpha", 0},
		{"β", 1}, {"b", 1}, {"beta", 1},
		{"γ", 2}, {"g", 2}, {"gamma", 2},
		{"δ", 3}, {"d", 3}, {"delta", 3},
		{"ε", 4}, {"e", 4}, {"epsilon", 4},
		{"0", 5},
	}
}

func mclSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"C", 1}, {"c", 1},
		{"0", 0},
	}
}

func msfSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"0", 9}, {"O", 9},
		{"2", 8},
		{"3", 7},
		{"4", 6},
		{"5", 5},
		{"6", 4},
		{"7", 3},
		{"8", 2},
		{"9", 1},
		{"10", 0},
	}
}

func mplSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"2", 0},
		{"3", 1},
		{"4", 2},
		{"5", 3},
		{"6", 4},
		{"7", 5},
		{"8", 6},
		{"9", 7},
		{"10", 8},
	}
}

func prodSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"A", 0},
		{"S", 1},
		{"AS", 2},
		{"0", 3},
		{"AC", 4},
		{"SC", 5},
		{"ASC", 6},
	}
}

func printSettings() []BoolSwitchSetting {
	return []BoolSwitchSetting{
		{"P", true}, {"p", true},
		{"0", false},
	}
}

func couplingSettings() []BoolSwitchSetting {
	return []BoolSwitchSetting{
		{"C", true}, {"c", true},
		{"0", false},
	}
}

func constantSignSettings() []ByteSwitchSetting {
	return []ByteSwitchSetting{
		{"P", 0}, {"p", 0},
		{"M", 1}, {"m", 1},
	}
}

func constantDigitSettings() []ByteSwitchSetting {
	return []ByteSwitchSetting{
		{"0", 0},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
		{"5", 5},
		{"6", 6},
		{"7", 7},
		{"8", 8},
		{"9", 9},
	}
}

func mpDecadeSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"0", 0},
		{"1", 1},
		{"2", 2},
		{"3", 3},
		{"4", 4},
		{"5", 5},
		{"6", 6},
		{"7", 7},
		{"8", 8},
		{"9", 9},
	}
}

func mpClearSettings() []IntSwitchSetting {
	return []IntSwitchSetting{
		{"1", 0},
		{"2", 1},
		{"3", 2},
		{"4", 3},
		{"5", 4},
		{"6", 5},
	}
}
