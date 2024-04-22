import { Line } from '@/components/system/Line';
import { useLineContext } from '@/lib/context/line';
import { AtomProto } from '@/types/generated/ui';
import { ChipAdapter } from './ChipAdapter';
import { ConditionalAdapter } from './ConditionalAdapter';
import { LabelledIconAdapter } from './LabelledIconAdapter';

export const AtomAdapter = ({ atom }: { atom: AtomProto }) => {
    const line = useLineContext();

    if (atom.value) {
        return <Line.Value className="text-sm">{atom.value}</Line.Value>;
    }

    if (atom.labelledIcon) {
        return <LabelledIconAdapter labelledIcon={atom.labelledIcon} />;
    }

    if (atom.chip) {
        const lineChip = line.state.chips[atom.chip.index];
        if (lineChip) {
            return (
                <ChipAdapter
                    chip={lineChip}
                    onChange={(value: number) =>
                        line.setChipValue(lineChip.atom.index, value)
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
        </>
    );
};
