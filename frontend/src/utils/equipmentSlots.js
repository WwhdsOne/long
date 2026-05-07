export const EQUIPMENT_SLOTS = [
    {value: 'weapon', label: '武器'},
    {value: 'helmet', label: '头盔'},
    {value: 'chest', label: '胸甲'},
    {value: 'gloves', label: '手套'},
    {value: 'legs', label: '腿甲'},
    {value: 'accessory', label: '饰品'},
]

export function normalizeEquipmentSlot(slot) {
    const normalized = {
        weapon: 'weapon',
        武器: 'weapon',
        helmet: 'helmet',
        头盔: 'helmet',
        chest: 'chest',
        armor: 'chest',
        胸甲: 'chest',
        护甲: 'chest',
        gloves: 'gloves',
        手套: 'gloves',
        legs: 'legs',
        腿甲: 'legs',
        accessory: 'accessory',
        饰品: 'accessory',
    }

    return normalized[slot] || slot || ''
}

export function normalizeLoadout(loadout) {
    const normalized = Object.fromEntries(EQUIPMENT_SLOTS.map((slot) => [slot.value, null]))
    if (!loadout || typeof loadout !== 'object') {
        return normalized
    }

    for (const slot of EQUIPMENT_SLOTS) {
        normalized[slot.value] = loadout[slot.value] ?? null
    }
    if (!normalized.chest && loadout.armor) {
        normalized.chest = {
            ...loadout.armor,
            slot: normalizeEquipmentSlot(loadout.armor.slot),
        }
    }

    return normalized
}
