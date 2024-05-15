import { Tooltip } from '@/components/system/Tooltip';
import { useLineContext } from '@/lib/context/line';
import { ConditionalProto } from '@/types/generated/ui';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { isUndefined } from 'lodash';
import { AtomAdapter } from './AtomAdapter';

export const ConditionalAdapter = ({ atom }: { atom: ConditionalProto }) => {
    const line = useLineContext();

    const atomIndex = atom.conditions?.findIndex((condition) => {
        const check = condition.indices?.map((index, i) => {
            const chip = line.state.chips[index];
            if (!chip) return false;
            return condition.values && chip.value === condition.values[i];
        });
        if (!check) return false;
        return check.every(Boolean);
    });

    if (atomIndex === -1 || isUndefined(atomIndex))
        return (
            <Tooltip content="Value not found">
                <ExclamationTriangleIcon className="text-graphite-50" />
            </Tooltip>
        );

    return (
        <>
            {atom.atoms?.[atomIndex] && (
                <AtomAdapter atom={atom.atoms[atomIndex]} />
            )}
            ;
        </>
    );
};
