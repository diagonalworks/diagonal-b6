import { Line } from '@/components/system/Line';
import { useLineContext } from '@/lib/context/line';
import { AtomProto } from '@/types/generated/ui';
import { isUndefined } from 'lodash';
import { ChipAdapter } from './ChipAdapter';
import { ConditionalAdapter } from './ConditionalAdapter';
import { LabelledIconAdapter } from './LabelledIconAdapter';

export const AtomAdapter = ({ atom }: { atom: AtomProto }) => {
    const line = useLineContext();

    if (atom.value) {
        let value = atom.value;
        const toBold = atom.value.match(/_((?:[a-zA-Z]|\d)*)_/);
        if (toBold?.[0]) {
            value = value.replaceAll(toBold[0], `<b>${toBold[1]}</b>`);
        }
        return (
            <Line.Value className="text-sm">
                <div
                    dangerouslySetInnerHTML={{
                        __html: value,
                    }}
                />
            </Line.Value>
        );
    }

    if (atom.labelledIcon) {
        return <LabelledIconAdapter labelledIcon={atom.labelledIcon} />;
    }

    if (atom.chip) {
        const chipIndex = atom.chip.index;
        if (isUndefined(chipIndex)) {
            console.warn(`Chip index is undefined`, { line, atom });
            return null;
        }
        const lineChip = line.state.chips[chipIndex];
        if (lineChip) {
            return (
                <ChipAdapter
                    chip={lineChip}
                    onChange={(value: number) =>
                        line.setChipValue(lineChip.atom.index ?? 0, value)
                    }
                />
            );
        }
    }

    if (atom.conditional) {
        return <ConditionalAdapter atom={atom.conditional} />;
    }

    return (
        <>
            {atom.labelledIcon && (
                <LabelledIconAdapter labelledIcon={atom.labelledIcon} />
            )}
            {atom.value && <Line.Value>{atom.value}</Line.Value>}
            {/* @TODO: render other primitive atom types */}
            {atom.download && <span>{atom.download}</span>}
        </>
    );
};
