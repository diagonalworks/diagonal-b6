import { Chip } from '@/types/stack';
import { useMemo } from 'react';
import { SelectAdapter } from './SelectAdapter';

export const ChipAdapter = ({
    chip,
    onChange,
}: {
    chip: Chip;
    onChange: (v: number) => void;
}) => {
    const options = useMemo(
        () =>
            chip.atom.labels?.map((label, i) => ({
                value: i.toString(),
                label,
            })) ?? [],
        [chip.atom.labels]
    );

    return (
        <SelectAdapter
            options={options}
            value={chip.value.toString()}
            onValueChange={(v) => onChange(parseInt(v))}
        />
    );
};
